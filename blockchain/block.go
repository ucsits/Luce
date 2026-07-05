package blockchain

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Height uint64
	Author uint64 // discord id

	Timestamp     uint64
	PrevBlockHash [32]byte
	Data          []rune
}

func NewBlock(height uint64, prevBlockHash [32]byte, author uint64, data []rune) *Block {
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

func (b Block) Hash() [32]byte {
	return sha256.Sum256(b.Format())
}
