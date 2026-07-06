package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ucsits/Luce/blockchain"
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
	block.Timestamp = 1783300009
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
	var blocks []BlockResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &blocks); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
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
