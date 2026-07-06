package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
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
}

func NewBlock(height uint64, prevBlockHash [32]byte, author uint64, data string) *Block {
	return &Block{
		Height:        height,
		Author:        author,
		Timestamp:     uint64(time.Now().Unix()),
		PrevBlockHash: prevBlockHash,
		Data:          data,
	}
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

func NewBlockFromFile(filename string) *Block {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(string(data), "ꭣ")
	if len(parts) < 6 {
		log.Fatal("invalid block file format")
	}

	height, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	hashBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(hashBytes) != 32 {
		log.Fatal("invalid hash in block file")
	}
	var hash [32]byte
	copy(hash[:], hashBytes)

	author, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	timestamp, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	prevHashBytes, err := hex.DecodeString(parts[4])
	if err != nil || len(prevHashBytes) != 32 {
		log.Fatal("invalid prev block hash in block file")
	}
	var prevBlockHash [32]byte
	copy(prevBlockHash[:], prevHashBytes)

	// Join everything after the 5th delimiter back — data may contain
	// spaces and even the ꭣ delimiter itself
	blockData := strings.Join(parts[5:], "ꭣ")

	b := NewBlock(height, prevBlockHash, author, blockData)
	b.Timestamp = timestamp

	return b
}
