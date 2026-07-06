package fsmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ucsits/Luce/blockchain"
)

func TestPersistBlock_WritesBlockAndMetadata(t *testing.T) {
	dir := t.TempDir()
	var chain blockchain.Blockchain
	Genesis(&chain)
	if err := PersistBlock(dir, chain.GetBlock(0)); err != nil {
		t.Fatalf("persist: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".luce"))
	if err != nil {
		t.Fatalf("read .luce: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 files (block + metadata), got %d", len(entries))
	}
	var loaded blockchain.Blockchain
	if err := Load(dir, &loaded); err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Height() != 1 {
		t.Fatalf("expected loaded height 1, got %d", loaded.Height())
	}
	if !loaded.Validate() {
		t.Fatal("loaded chain should be valid")
	}
}

func TestPersistBlock_AppendAddsOneFile(t *testing.T) {
	dir := t.TempDir()
	var chain blockchain.Blockchain
	Genesis(&chain)
	if err := PersistBlock(dir, chain.GetBlock(0)); err != nil {
		t.Fatalf("persist genesis: %v", err)
	}

	block1 := chain.AppendBlock(1, "first")
	if err := PersistBlock(dir, block1); err != nil {
		t.Fatalf("persist block1: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, ".luce")); len(entries) != 3 {
		t.Fatalf("expected 3 files after 2 blocks, got %d", len(entries))
	}

	block2 := chain.AppendBlock(2, "second")
	if err := PersistBlock(dir, block2); err != nil {
		t.Fatalf("persist block2: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, ".luce")); len(entries) != 4 {
		t.Fatalf("expected 4 files after 3 blocks, got %d", len(entries))
	}

	var loaded blockchain.Blockchain
	if err := Load(dir, &loaded); err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Height() != 3 {
		t.Fatalf("expected loaded height 3, got %d", loaded.Height())
	}
	if !loaded.Validate() {
		t.Fatal("loaded chain should be valid")
	}
}
