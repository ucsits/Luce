package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ucsits/Luce/blockchain"
	"github.com/ucsits/Luce/fsmgr"
)

func newIntegrationServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	var chain blockchain.Blockchain
	fsmgrGenesis(&chain)
	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()
	s := NewServer(cfg, &chain)
	ts := httptest.NewServer(s.echo)
	return s, ts
}

func do(t *testing.T, ts *httptest.Server, method, path string, body interface{}) (int, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, ts.URL+path, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, data
}

func TestIntegrationFullFlow(t *testing.T) {
	s, ts := newIntegrationServer(t)
	defer ts.Close()

	code, data := do(t, ts, http.MethodGet, "/api/v1/chain/height", nil)
	if code != 200 {
		t.Fatalf("height: expected 200, got %d: %s", code, data)
	}
	t.Logf("height after genesis: %s", data)

	code, data = do(t, ts, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 1, Data: "first"})
	if code != 201 {
		t.Fatalf("append 1: expected 201, got %d: %s", code, data)
	}
	t.Logf("append 1: %s", data)

	code, data = do(t, ts, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 2, Data: "second"})
	if code != 201 {
		t.Fatalf("append 2: expected 201, got %d: %s", code, data)
	}

	code, data = do(t, ts, http.MethodGet, "/api/v1/chain/height", nil)
	if code != 200 {
		t.Fatalf("height: expected 200, got %d: %s", code, data)
	}
	t.Logf("height after 2 appends: %s", data)

	code, data = do(t, ts, http.MethodGet, "/api/v1/chain/validate", nil)
	if code != 200 {
		t.Fatalf("validate: expected 200, got %d: %s", code, data)
	}
	t.Logf("validate: %s", data)

	code, data = do(t, ts, http.MethodGet, "/api/v1/blocks", nil)
	if code != 200 {
		t.Fatalf("list: expected 200, got %d: %s", code, data)
	}
	t.Logf("list: %s", data)

	if s.chain.Height() != 3 {
		t.Fatalf("expected chain height 3, got %d", s.chain.Height())
	}
}

func TestAppendBlock_PersistedChainSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	chain, err := fsmgr.LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("load or create: %v", err)
	}
	cfg := DefaultConfig()
	cfg.DataDir = dir
	s := NewServer(cfg, chain)
	ts := httptest.NewServer(s.echo)
	defer ts.Close()

	code, _ := do(t, ts, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 9, Data: "persist me"})
	if code != http.StatusCreated {
		t.Fatalf("append: expected 201, got %d", code)
	}

	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, ".luce"))
	if err != nil {
		t.Fatalf("read .luce: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 files (metadata + genesis + 1 block), got %d", len(entries))
	}

	loaded, err := fsmgr.LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if loaded.Height() != 2 {
		t.Fatalf("expected reloaded height 2, got %d", loaded.Height())
	}
	if !loaded.Validate() {
		t.Fatal("reloaded chain should be valid")
	}
}

func TestAppendBlock_RollsBackOnPersistFailure(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}

	var chain blockchain.Blockchain
	fsmgrGenesis(&chain)
	cfg := DefaultConfig()
	cfg.DataDir = blocker
	s := NewServer(cfg, &chain)
	ts := httptest.NewServer(s.echo)
	defer ts.Close()

	s.mu.RLock()
	h0 := s.chain.Height()
	s.mu.RUnlock()

	code, _ := do(t, ts, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: 1, Data: "should roll back"})
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on persist failure, got %d", code)
	}

	s.mu.RLock()
	h1 := s.chain.Height()
	valid := s.chain.Validate()
	s.mu.RUnlock()

	if h1 != h0 {
		t.Fatalf("expected height rolled back to %d, got %d", h0, h1)
	}
	if !valid {
		t.Fatal("chain should remain valid after rollback")
	}
}

func TestConcurrentAccessNoRace(t *testing.T) {
	s, ts := newIntegrationServer(t)
	defer ts.Close()

	const (
		writers    = 4
		perWriter  = 10
		readers    = 2
		readerHits = 50
	)
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(author uint64) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				if code, _ := do(t, ts, http.MethodPost, "/api/v1/blocks", AppendBlockRequest{Author: author, Data: "x"}); code != http.StatusCreated {
					t.Errorf("writer %d: expected 201, got %d", author, code)
					return
				}
			}
		}(uint64(w))
	}
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < readerHits; i++ {
				if code, _ := do(t, ts, http.MethodGet, "/api/v1/chain/height", nil); code != 200 {
					t.Errorf("reader: expected 200, got %d", code)
					return
				}
			}
		}()
	}
	wg.Wait()

	s.mu.RLock()
	height := s.chain.Height()
	valid := s.chain.Validate()
	s.mu.RUnlock()

	if want := uint64(1 + writers*perWriter); height != want {
		t.Fatalf("height = %d, want %d (blocks were lost)", height, want)
	}
	if !valid {
		t.Fatal("chain should be valid after concurrent appends")
	}
}
