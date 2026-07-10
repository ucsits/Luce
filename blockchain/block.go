package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Height uint64
	Author uint64 // discord id

	Timestamp     uint64
	PrevBlockHash [32]byte
	Data          string

	// storedHash is the block's hash at creation time, used for self-consistency
	// validation. Set by NewBlock and NewBlockFromFile.
	storedHash [32]byte
}

func NewBlock(height uint64, prevBlockHash [32]byte, author uint64, data string) *Block {
	b := &Block{
		Height:        height,
		Author:        author,
		Timestamp:     uint64(time.Now().Unix()),
		PrevBlockHash: prevBlockHash,
		Data:          data,
	}
	b.storedHash = b.Hash()
	return b
}

// SetTimestamp overrides the block's timestamp and recomputes its stored hash.
// Use this instead of directly assigning to Timestamp to keep the hash consistent.
func (b *Block) SetTimestamp(timestamp uint64) {
	b.Timestamp = timestamp
	b.storedHash = b.Hash()
}

func (b Block) Format() []byte {
	formattedBlock := fmt.Sprintf("%dꭣ%dꭣ%dꭣ%xꭣ%v", b.Height, b.Author, b.Timestamp, b.PrevBlockHash, b.Data)
	return []byte(formattedBlock)
}

func (b Block) Encode() []byte {
	encodedBlock := fmt.Sprintf("%dꭣ%xꭣ%dꭣ%dꭣ%xꭣ%v", b.Height, b.Hash(), b.Author, b.Timestamp, b.PrevBlockHash, b.Data)
	return []byte(encodedBlock)
}

func (b Block) Hash() [32]byte {
	return sha256.Sum256(b.Format())
}

func NewBlockFromFile(filename string) (*Block, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading block file: %w", err)
	}

	parts := strings.Split(string(data), "ꭣ")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid block file format: expected at least 6 parts, got %d", len(parts))
	}

	height, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing block height: %w", err)
	}

	hashBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(hashBytes) != 32 {
		return nil, fmt.Errorf("invalid hash in block file")
	}
	var hash [32]byte
	copy(hash[:], hashBytes)

	author, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing block author: %w", err)
	}

	timestamp, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing block timestamp: %w", err)
	}

	prevHashBytes, err := hex.DecodeString(parts[4])
	if err != nil || len(prevHashBytes) != 32 {
		return nil, fmt.Errorf("invalid prev block hash in block file")
	}
	var prevBlockHash [32]byte
	copy(prevBlockHash[:], prevHashBytes)

	// Join everything after the 5th delimiter back — data may contain
	// spaces and even the ꭣ delimiter itself
	blockData := strings.Join(parts[5:], "ꭣ")

	b := NewBlock(height, prevBlockHash, author, blockData)
	b.SetTimestamp(timestamp)

	// Verify the reconstructed block's hash matches the stored hash.
	// This detects file corruption or tampering.
	if b.Hash() != hash {
		return nil, fmt.Errorf(
			"block hash mismatch: stored %x, computed %x", hash, b.Hash(),
		)
	}

	return b, nil
}
