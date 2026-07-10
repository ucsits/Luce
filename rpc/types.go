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

type PaginatedBlocksResponse struct {
	Data       []BlockResponse `json:"data"`
	Pagination PaginationMeta  `json:"pagination"`
}

type PaginationMeta struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      uint64 `json:"total"`
	TotalPages int    `json:"total_pages"`
}

type ChainSummaryResponse struct {
	Height        uint64        `json:"height"`
	Blocks        uint64        `json:"blocks"`
	BestBlockHash string        `json:"best_block_hash"`
	LastBlock     *BlockResponse `json:"last_block,omitempty"`
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
