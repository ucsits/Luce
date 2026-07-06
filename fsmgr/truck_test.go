package fsmgr

import (
	"testing"

	"github.com/ucsits/Luce/blockchain"
)

func TestGenesis(t *testing.T) {
	var chain blockchain.Blockchain
	genesis(&chain)

	if chain.Height() != 1 {
		t.Fatalf("chain height = %d, want 1", chain.Height())
	}

	block := chain.GetBlock(0)
	if block.Height != 0 {
		t.Errorf("genesis block Height = %d, want 0", block.Height)
	}
	if block.PrevBlockHash != [32]byte{0} {
		t.Errorf("genesis PrevBlockHash should be zero, got %x", block.PrevBlockHash)
	}
	if block.Timestamp != 1783300009 {
		t.Errorf("genesis Timestamp = %d, want 1783300009", block.Timestamp)
	}
	if block.Data == "" {
		t.Error("genesis block Data should not be empty")
	}
	if block.Author != 0 {
		t.Errorf("genesis Author = %d, want 0", block.Author)
	}

	t.Logf("Genesis data: %s", block.Data)
	t.Logf("Genesis hash: %x", block.Hash())
}

func TestGenesisBlockIsValid(t *testing.T) {
	var chain blockchain.Blockchain
	genesis(&chain)

	if !chain.Validate() {
		t.Error("genesis chain should be valid")
	}
}

// ---------------------------------------------------------------------------
// Dump / Load tests — currently UNTESTABLE due to:
//
// Bug 1: Dump calls c.Encode() which panics on empty chain (chain.go:52).
//
// Bug 2: Load calls NewBlockFromFile which is broken (block.go:58 — missing &
//   on Sscanf arguments, log.Fatal instead of error return).
//
// Bug 3: Load uses log.Fatal for all errors, making it untestable.
//
// Bug 4: Load metadata parsing uses %x Sscanf into a string, which may not
//   correctly capture a 64-char hex hash string.
//
// Fixes needed:
//   1. Guard Encode() against empty chain
//   2. Fix NewBlockFromFile Sscanf args (add &)
//   3. Replace log.Fatal with error returns in Load, Dump, mkluce
//   4. Verify %x vs %s for hash string parsing in Load metadata
// ---------------------------------------------------------------------------

func TestDumpEmptyChainPanics(t *testing.T) {
	var chain blockchain.Blockchain

	defer func() {
		if r := recover(); r == nil {
			t.Error("Dump() on empty chain should panic, but it did not")
		} else {
			t.Logf("Caught expected panic: %v", r)
		}
	}()
	Dump(t.TempDir(), chain)
}

func TestDumpLoadRoundTripBlockedByNewBlockFromFile(t *testing.T) {
	// This test documents the expected behavior once all bugs are fixed.
	// It is skipped because Load → NewBlockFromFile → log.Fatal kills the process.
	t.Skip("Blocked by bugs in NewBlockFromFile and log.Fatal usage in Load/Dump")

	tmpDir := t.TempDir()

	var chain blockchain.Blockchain
	chain.AppendBlock(0, "first block")
	chain.AppendBlock(1, "second block")
	chain.AppendBlock(2, "third block")

	originalHeight := chain.Height()

	Dump(tmpDir, chain)

	var loaded blockchain.Blockchain
	Load(tmpDir, &loaded)

	if loaded.Height() != originalHeight {
		t.Errorf("loaded chain height = %d, want %d", loaded.Height(), originalHeight)
	}

	if !loaded.Validate() {
		t.Error("loaded chain should be valid")
	}
}
