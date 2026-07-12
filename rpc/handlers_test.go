package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ucsits/Luce/blockchain"
	"github.com/ucsits/Luce/fsmgr"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	var chain blockchain.Blockchain
	fsmgrGenesis(&chain)
	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()
	return NewServer(cfg, &chain)
}

func fsmgrGenesis(chain *blockchain.Blockchain) {
	block := blockchain.NewBlock(
		0, [32]byte{0}, 0,
		"genesis",
	)
	block.SetTimestamp(1783300009)
	chain.PrependBlock(block)
}

func request(t *testing.T, s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.RemoteAddr = "127.0.0.1:12345"
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func TestListBlocks(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PaginatedBlocksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 block, got %d", len(resp.Data))
	}
	if resp.Pagination.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.Page != 1 {
		t.Fatalf("expected page 1, got %d", resp.Pagination.Page)
	}
	if resp.Pagination.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", resp.Pagination.Limit)
	}
}

func TestListBlocks_Pagination(t *testing.T) {
	s := newTestServer(t)
	// Append more blocks
	for i := 0; i < 5; i++ {
		if err := fsmgr.PersistBlock(s.config.DataDir, s.chain.AppendBlock(uint64(i+1), "data")); err != nil {
			t.Fatalf("persisting block: %v", err)
		}
	}

	// Page 1, limit 2
	rec := request(t, s, http.MethodGet, "/api/v1/blocks?page=1&limit=2", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PaginatedBlocksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocks on page 1, got %d", len(resp.Data))
	}
	if resp.Pagination.Total != 6 {
		t.Fatalf("expected total 6, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.TotalPages != 3 {
		t.Fatalf("expected 3 pages, got %d", resp.Pagination.TotalPages)
	}
	// Verify first block is genesis (height 0)
	if resp.Data[0].Height != 0 {
		t.Fatalf("expected first block height 0, got %d", resp.Data[0].Height)
	}

	// Page 2, limit 2
	rec = request(t, s, http.MethodGet, "/api/v1/blocks?page=2&limit=2", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocks on page 2, got %d", len(resp.Data))
	}
	// Verify we got blocks at heights 2 and 3
	if resp.Data[0].Height != 2 {
		t.Fatalf("expected first block height 2 on page 2, got %d", resp.Data[0].Height)
	}

	// Page out of range
	rec = request(t, s, http.MethodGet, "/api/v1/blocks?page=99&limit=10", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 blocks on out-of-range page, got %d", len(resp.Data))
	}
}

func TestListBlocks_Desc(t *testing.T) {
	s := newTestServer(t)
	// Append more blocks
	for i := 0; i < 5; i++ {
		if err := fsmgr.PersistBlock(s.config.DataDir, s.chain.AppendBlock(uint64(i+1), "data")); err != nil {
			t.Fatalf("persisting block: %v", err)
		}
	}

	// Page 1, limit 2, desc=true — newest first
	rec := request(t, s, http.MethodGet, "/api/v1/blocks?page=1&limit=2&desc=true", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PaginatedBlocksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocks on page 1, got %d", len(resp.Data))
	}
	if resp.Pagination.Total != 6 {
		t.Fatalf("expected total 6, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.TotalPages != 3 {
		t.Fatalf("expected 3 pages, got %d", resp.Pagination.TotalPages)
	}
	// Verify first block is the newest (height 5)
	if resp.Data[0].Height != 5 {
		t.Fatalf("expected first block height 5 (newest), got %d", resp.Data[0].Height)
	}
	// Verify second block is height 4
	if resp.Data[1].Height != 4 {
		t.Fatalf("expected second block height 4, got %d", resp.Data[1].Height)
	}

	// Page 2, limit 2, desc=true
	rec = request(t, s, http.MethodGet, "/api/v1/blocks?page=2&limit=2&desc=true", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocks on page 2, got %d", len(resp.Data))
	}
	// Page 2 descending: blocks at heights 3 and 2
	if resp.Data[0].Height != 3 {
		t.Fatalf("expected first block height 3 on page 2, got %d", resp.Data[0].Height)
	}
	if resp.Data[1].Height != 2 {
		t.Fatalf("expected second block height 2 on page 2, got %d", resp.Data[1].Height)
	}

	// Page 3, limit 2, desc=true — oldest blocks
	rec = request(t, s, http.MethodGet, "/api/v1/blocks?page=3&limit=2&desc=true", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocks on page 3, got %d", len(resp.Data))
	}
	// Page 3 descending: blocks at heights 1 and 0
	if resp.Data[0].Height != 1 {
		t.Fatalf("expected first block height 1 on page 3, got %d", resp.Data[0].Height)
	}
	if resp.Data[1].Height != 0 {
		t.Fatalf("expected second block height 0 (genesis) on page 3, got %d", resp.Data[1].Height)
	}

	// desc=false should match the default ascending behaviour
	rec = request(t, s, http.MethodGet, "/api/v1/blocks?page=1&limit=2&desc=false", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if resp.Data[0].Height != 0 {
		t.Fatalf("expected first block height 0 with desc=false, got %d", resp.Data[0].Height)
	}
}

func TestGetBlockByHash(t *testing.T) {
	s := newTestServer(t)
	// Fetch the known block and extract its hash
	var resp PaginatedBlocksResponse
	rec := request(t, s, http.MethodGet, "/api/v1/blocks?limit=1", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling blocks: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 block, got %d", len(resp.Data))
	}
	hash := resp.Data[0].Hash

	// Look up by hash
	rec = request(t, s, http.MethodGet, "/api/v1/blocks/hash/"+hash, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var block BlockResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &block); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if block.Hash != hash {
		t.Fatalf("expected hash %s, got %s", hash, block.Hash)
	}
}

func TestGetBlockByHash_NotFound(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/hash/0000000000000000000000000000000000000000000000000000000000000000", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBlockByHash_InvalidHash(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/hash/xyz", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBlockByHash_ShortHash(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/hash/aabb", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChainSummary(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/chain/summary", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var summary ChainSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if summary.Height != 1 {
		t.Fatalf("expected height 1, got %d", summary.Height)
	}
	if summary.Blocks != 1 {
		t.Fatalf("expected blocks 1, got %d", summary.Blocks)
	}
	if summary.BestBlockHash == "" {
		t.Fatal("expected non-empty best_block_hash")
	}
	if summary.LastBlock == nil {
		t.Fatal("expected last_block to be present")
	}
	if summary.LastBlock.Height != 0 {
		t.Fatalf("expected last block height 0, got %d", summary.LastBlock.Height)
	}
}

func TestGetBlock_Valid(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/0", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var b BlockResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if b.Height != 0 {
		t.Fatalf("expected height 0, got %d", b.Height)
	}
}

func TestGetBlock_NotFound(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/99", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBlock_InvalidHeight(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/blocks/abc", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAppendBlock_Valid(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 42, Data: "hello"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var b BlockResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if b.Height != 1 {
		t.Fatalf("expected height 1, got %d", b.Height)
	}
	if s.chain.Height() != 2 {
		t.Fatalf("expected chain height 2, got %d", s.chain.Height())
	}
}

func TestAppendBlock_EmptyData(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 42, Data: ""})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAppendBlock_InvalidJSON(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/blocks", bytes.NewBufferString("{not json"))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestValidateChain(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/chain/validate", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if !resp["valid"] {
		t.Fatal("expected valid chain")
	}
}

func TestGetHeight(t *testing.T) {
	s := newTestServer(t)
	rec := request(t, s, http.MethodGet, "/api/v1/chain/height", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]uint64
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if resp["height"] != 1 {
		t.Fatalf("expected height 1, got %d", resp["height"])
	}
}

func TestListBlocks_DocumentContentStripped(t *testing.T) {
	s := newTestServer(t)

	// Create a document-type block with large content
	docData := `{"type":"document","v":1,"docId":"test-doc-id","title":"Test Document","content":"this is a large base64 encoded content that should be stripped","author":"123456","mimeType":"application/pdf"}`
	if err := fsmgr.PersistBlock(s.config.DataDir, s.chain.AppendBlock(1, docData)); err != nil {
		panic(err)
	}

	// Create a task-type block (should NOT strip content)
	taskData := `{"type":"task","v":1,"taskId":"task-123","title":"Test Task","description":"Task description"}`
	if err := fsmgr.PersistBlock(s.config.DataDir, s.chain.AppendBlock(1, taskData)); err != nil {
		panic(err)
	}

	rec := request(t, s, http.MethodGet, "/api/v1/blocks?limit=10", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp PaginatedLightweightBlocksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(resp.Data))
	}

	// Block 1 (genesis) - plain text, no stripping needed
	if resp.Data[0].Height != 0 {
		t.Fatalf("expected height 0, got %d", resp.Data[0].Height)
	}

	// Block 1 (document) - content should be stripped
	docBlock := resp.Data[1]
	if docBlock.Height != 1 {
		t.Fatalf("expected height 1, got %d", docBlock.Height)
	}
	var docParsed map[string]interface{}
	if err := json.Unmarshal([]byte(docBlock.Data), &docParsed); err != nil {
		t.Fatalf("parsing document data: %v", err)
	}
	if _, hasContent := docParsed["content"]; hasContent {
		t.Fatal("expected document content to be stripped from lightweight response")
	}
	if docParsed["type"] != "document" {
		t.Fatalf("expected type document, got %v", docParsed["type"])
	}
	if docParsed["title"] != "Test Document" {
		t.Fatalf("expected title Test Document, got %v", docParsed["title"])
	}

	// Block 2 (task) - content should NOT be stripped
	taskBlock := resp.Data[2]
	if taskBlock.Height != 2 {
		t.Fatalf("expected height 2, got %d", taskBlock.Height)
	}
	var taskParsed map[string]interface{}
	if err := json.Unmarshal([]byte(taskBlock.Data), &taskParsed); err != nil {
		t.Fatalf("parsing task data: %v", err)
	}
	if desc, ok := taskParsed["description"]; !ok || desc != "Task description" {
		t.Fatalf("expected task description to be preserved, got %v", desc)
	}
}
