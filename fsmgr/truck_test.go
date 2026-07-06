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
// Dump / Load tests
// ---------------------------------------------------------------------------

func TestDumpEmptyChainReturnsError(t *testing.T) {
	var chain blockchain.Blockchain

	err := Dump(t.TempDir(), chain)
	if err == nil {
		t.Error("Dump() on empty chain should return an error, but it did not")
	}
}

func TestDumpLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	var chain blockchain.Blockchain
	chain.AppendBlock(0, "first block")
	chain.AppendBlock(1, "second block")
	chain.AppendBlock(2, "third block")

	originalHeight := chain.Height()

	if err := Dump(tmpDir, chain); err != nil {
		t.Fatalf("Dump returned error: %v", err)
	}

	var loaded blockchain.Blockchain
	if err := Load(tmpDir, &loaded); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Height() != originalHeight {
		t.Errorf("loaded chain height = %d, want %d", loaded.Height(), originalHeight)
	}

	if !loaded.Validate() {
		t.Error("loaded chain should be valid")
	}

	// Verify each block matches
	for i := uint64(0); i < originalHeight; i++ {
		original := chain.GetBlock(i)
		loaded := loaded.GetBlock(i)
		if original.Hash() != loaded.Hash() {
			t.Errorf("block %d hash mismatch: got %x, want %x", i, loaded.Hash(), original.Hash())
		}
		if original.Data != loaded.Data {
			t.Errorf("block %d data mismatch: got %q, want %q", i, loaded.Data, original.Data)
		}
	}
}
