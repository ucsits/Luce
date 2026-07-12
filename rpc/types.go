package rpc

import (
	"encoding/hex"
	"encoding/json"

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

// LightweightBlockResponse is a performance-optimized response for block listings.
// For document-type blocks, it omits the large "content" field to reduce payload size.
type LightweightBlockResponse struct {
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

// PaginatedLightweightBlocksResponse returns blocks without large document content.
type PaginatedLightweightBlocksResponse struct {
	Data       []LightweightBlockResponse `json:"data"`
	Pagination PaginationMeta              `json:"pagination"`
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

// NewLightweightBlockResponse creates a response that omits the content field
// from document-type blocks to reduce payload size for list endpoints.
func NewLightweightBlockResponse(b blockchain.Block) LightweightBlockResponse {
	hash := b.Hash()
	prev := b.PrevBlockHash

	data := b.Data
	if stripped := stripDocumentContent(data); stripped != "" {
		data = stripped
	}

	return LightweightBlockResponse{
		Height:        b.Height,
		Hash:          hex.EncodeToString(hash[:]),
		Author:        b.Author,
		Timestamp:     b.Timestamp,
		PrevBlockHash: hex.EncodeToString(prev[:]),
		Data:          data,
	}
}

// stripDocumentContent parses JSON data and removes the "content" field
// for document-type blocks. Returns empty string if not a document block
// or if stripping fails.
func stripDocumentContent(data string) string {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return ""
	}

	// Only strip content from document-type blocks
	blockType, ok := parsed["type"].(string)
	if !ok || blockType != "document" {
		return ""
	}

	// Remove the content field
	delete(parsed, "content")

	// Re-marshal to JSON
	stripped, err := json.Marshal(parsed)
	if err != nil {
		return ""
	}

	return string(stripped)
}
