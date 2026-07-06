package fsmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ucsits/Luce/blockchain"
)

func TestLoadOrCreate_EmptyLuceDirGenesisesNotFatal(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".luce"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	chain, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("expected no error on empty .luce dir, got: %v", err)
	}
	if chain.Height() != 1 {
		t.Fatalf("expected genesis height 1, got %d", chain.Height())
	}
}

func TestLoadOrCreate_MissingMetadataGenesises(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".luce"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".luce", "orphan"), []byte("junk"), 0644); err != nil {
		t.Fatalf("write orphan: %v", err)
	}
	chain, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("expected genesis on missing metadata, got: %v", err)
	}
	if chain.Height() != 1 {
		t.Fatalf("expected genesis height 1, got %d", chain.Height())
	}
}

func TestLoadOrCreate_CorruptMetadataReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".luce"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".luce", "metadata"), []byte("garbage"), 0644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	chain, err := LoadOrCreate(dir)
	if err == nil {
		t.Fatalf("expected error on corrupt metadata, got chain height %d", chain.Height())
	}
}

func TestLoadOrCreate_MissingBlockFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".luce"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	headHash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	if err := os.WriteFile(filepath.Join(dir, ".luce", "metadata"), []byte("1ꭣ"+headHash), 0644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	chain, err := LoadOrCreate(dir)
	if err == nil {
		t.Fatalf("expected error when referenced block file is missing, got chain height %d", chain.Height())
	}
}

func TestLoadOrCreate_PersistsGenesisOnCreate(t *testing.T) {
	dir := t.TempDir()
	chain, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if chain.Height() != 1 {
		t.Fatalf("expected genesis height 1, got %d", chain.Height())
	}
	if _, err := os.Stat(filepath.Join(dir, ".luce", "metadata")); err != nil {
		t.Fatalf("expected metadata persisted on create: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".luce"))
	if err != nil {
		t.Fatalf("read .luce: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 files (metadata + genesis block) after create, got %d", len(entries))
	}
}

func TestDump_NoTempFilesLeftBehind(t *testing.T) {
	dir := t.TempDir()
	var chain blockchain.Blockchain
	chain.AppendBlock(1, "a")
	chain.AppendBlock(2, "b")
	if err := Dump(dir, chain); err != nil {
		t.Fatalf("dump: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".luce"))
	if err != nil {
		t.Fatalf("read .luce: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
}

func TestLoadOrCreate_RestartPreservesAppendedBlocks(t *testing.T) {
	dir := t.TempDir()
	first, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	first.AppendBlock(1, "appended")
	if err := Dump(dir, *first); err != nil {
		t.Fatalf("dump after append: %v", err)
	}
	second, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if second.Height() != 2 {
		t.Fatalf("expected reloaded height 2, got %d", second.Height())
	}
	if !second.Validate() {
		t.Fatal("reloaded chain should be valid")
	}
}
