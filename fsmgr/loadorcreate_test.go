package fsmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreate_CreatesGenesisWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	chain, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if chain.Height() != 1 {
		t.Fatalf("expected genesis height 1, got %d", chain.Height())
	}
}

func TestLoadOrCreate_LoadsExistingFromDir(t *testing.T) {
	dir := t.TempDir()

	first, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if first.Height() != 1 {
		t.Fatalf("first run: expected genesis height 1, got %d", first.Height())
	}
	first.AppendBlock(1, "in dir")
	if err := Dump(dir, *first); err != nil {
		t.Fatalf("dump: %v", err)
	}

	second, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if second.Height() != 2 {
		t.Fatalf("second run: expected loaded height 2, got %d", second.Height())
	}
	if !second.Validate() {
		t.Fatal("loaded chain should be valid")
	}

	if _, err := os.Stat(filepath.Join(dir, ".luce", "metadata")); err != nil {
		t.Fatalf("expected metadata in dir: %v", err)
	}
}
