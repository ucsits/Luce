package fsmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreate_CreatesGenesisWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	chain := LoadOrCreate(dir)
	if chain.Height() != 1 {
		t.Fatalf("expected genesis height 1, got %d", chain.Height())
	}
}

// TestLoadOrCreate_LoadsExistingFromDir guards against a regression where the
// existence check stats CWD (".luce") instead of filepath.Join(dir, ".luce"):
// without the dir-aware stat, a non-CWD data dir would be silently re-genesis'd
// on restart, discarding the persisted chain.
func TestLoadOrCreate_LoadsExistingFromDir(t *testing.T) {
	dir := t.TempDir()

	first := LoadOrCreate(dir)
	if first.Height() != 1 {
		t.Fatalf("first run: expected genesis height 1, got %d", first.Height())
	}
	first.AppendBlock(1, "in dir")
	if err := Dump(dir, *first); err != nil {
		t.Fatalf("dump: %v", err)
	}

	second := LoadOrCreate(dir)
	if second.Height() != 2 {
		t.Fatalf("second run: expected loaded height 2, got %d (stat used wrong path?)", second.Height())
	}
	if !second.Validate() {
		t.Fatal("loaded chain should be valid")
	}

	if _, err := os.Stat(filepath.Join(dir, ".luce", "metadata")); err != nil {
		t.Fatalf("expected metadata in dir: %v", err)
	}
}
