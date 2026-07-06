package rpc

import (
	"encoding/hex"

	"github.com/ucsits/Luce/blockchain"
)

type BlockResponse struct {
	Height        uint64 `json:"height"`
	Hash          string `json:"hash"`
	Author        uint64 `json:"author"`
	Timestamp     uint64 `json:"timestamp"`
	PrevBlockHash string `json:"prev_block_hash"`
	Data          string `json:"data"`
}

type AppendBlockRequest struct {
	Author uint64 `json:"author"`
	Data   string `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewBlockResponse(b blockchain.Block) BlockResponse {
	hash := b.Hash()
	prev := b.PrevBlockHash
	return BlockResponse{
		Height:        b.Height,
		Hash:          hex.EncodeToString(hash[:]),
		Author:        b.Author,
		Timestamp:     b.Timestamp,
		PrevBlockHash: hex.EncodeToString(prev[:]),
		Data:          b.Data,
	}
}
