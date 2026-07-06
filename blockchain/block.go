package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
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

	var height uint64
	var hashStr string
	var hash [32]byte
	var author uint64
	var timestamp uint64
	var prevBlockHashStr string
	var prevBlockHash [32]byte
	var blockData string
	_, err = fmt.Sscanf(string(data), "%dꭣ%xꭣ%dꭣ%dꭣ%xꭣ%v", height, hashStr, author, timestamp, prevBlockHashStr, blockData)

	dec, err := hex.DecodeString(hashStr)
	if err == nil && len(dec) == 32 {
		copy(hash[:], dec)
	} else {
		log.Fatal(err)
	}

	dec, err = hex.DecodeString(hashStr)
	if err == nil && len(dec) == 32 {
		copy(prevBlockHash[:], dec)
	} else {
		log.Fatal(err)
	}

	b := NewBlock(height, prevBlockHash, author, blockData)
	b.Timestamp = timestamp

	return b
}
