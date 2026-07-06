package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHeight(t *testing.T) {
	c := Blockchain{}
	if h := c.Height(); h != 0 {
		t.Errorf("Height() = %d, want 0 for empty chain", h)
	}
}

func TestHeightAfterOperations(t *testing.T) {
	c := Blockchain{}
	b := NewBlock(0, [32]byte{0}, 0, "block")
	c.PrependBlock(b)

	if h := c.Height(); h != 1 {
		t.Errorf("Height() = %d, want 1 after prepend", h)
	}
}

func TestAppendBlockToEmptyChain(t *testing.T) {
	c := Blockchain{}
	b := c.AppendBlock(1, "first block after genesis")

	if b.Height != 0 {
		t.Errorf("appended block height = %d, want 0", b.Height)
	}
	if b.PrevBlockHash != [32]byte{0} {
		t.Errorf("first block PrevBlockHash should be zero, got %x", b.PrevBlockHash)
	}
	if c.Height() != 1 {
		t.Errorf("chain height = %d, want 1", c.Height())
	}
}

func TestAppendBlockToChainWithOneBlock(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "genesis")
	c.AppendBlock(1, "second block")

	if c.Height() != 2 {
		t.Errorf("chain height = %d, want 2", c.Height())
	}

	second := c.GetBlock(1)
	first := c.GetBlock(0)
	if second.PrevBlockHash != first.Hash() {
		t.Errorf("second block PrevBlockHash = %x, want %x (hash of genesis)", second.PrevBlockHash, first.Hash())
	}
}

func TestAppendBlockHashLinking(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "block 0")
	c.AppendBlock(1, "block 1")
	c.AppendBlock(2, "block 2")
	c.AppendBlock(3, "block 3")

	for i := uint64(1); i < c.Height(); i++ {
		curr := c.GetBlock(i)
		prev := c.GetBlock(i - 1)
		if curr.PrevBlockHash != prev.Hash() {
			t.Errorf("block %d PrevBlockHash = %x, want %x", i, curr.PrevBlockHash, prev.Hash())
		}
	}
}

func TestPrependBlock(t *testing.T) {
	c := Blockchain{}

	b1 := NewBlock(0, [32]byte{0}, 0, "first")
	b1.Timestamp = 100
	c.PrependBlock(b1)

	b0 := NewBlock(0, [32]byte{0}, 0, "new first")
	b0.Timestamp = 50
	c.PrependBlock(b0)

	if c.Height() != 2 {
		t.Fatalf("height = %d, want 2", c.Height())
	}

	first := c.GetBlock(0)
	if first.Data != "new first" {
		t.Errorf("first block data = %q, want %q", first.Data, "new first")
	}

	second := c.GetBlock(1)
	if second.Data != "first" {
		t.Errorf("second block data = %q, want %q", second.Data, "first")
	}
}

func TestPrependBlockToEmptyChain(t *testing.T) {
	c := Blockchain{}
	b := NewBlock(0, [32]byte{0}, 0, "sole")
	c.PrependBlock(b)

	if c.Height() != 1 {
		t.Errorf("height = %d, want 1", c.Height())
	}
	if c.GetBlock(0).Data != "sole" {
		t.Errorf("block data = %q, want %q", c.GetBlock(0).Data, "sole")
	}
}

func TestGetBlock(t *testing.T) {
	c := Blockchain{}
	b := NewBlock(0, [32]byte{0}, 0, "test")
	c.PrependBlock(b)

	got := c.GetBlock(0)
	if got.Data != "test" {
		t.Errorf("GetBlock(0).Data = %q, want %q", got.Data, "test")
	}
}

func TestGetBlockOutOfBounds(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "only block")

	// Accessing index 1 on a chain of height 1 should panic.
	// This test documents the current behavior — no bounds checking exists.
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetBlock(1) on chain of height 1 should panic, but it did not")
		}
	}()
	c.GetBlock(1)
}

func TestGetBlockEmptyChain(t *testing.T) {
	c := Blockchain{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("GetBlock(0) on empty chain should panic, but it did not")
		}
	}()
	c.GetBlock(0)
}

func TestValidateEmptyChain(t *testing.T) {
	c := Blockchain{}
	if !c.Validate() {
		t.Error("Validate() should return true for empty chain")
	}
}

func TestValidateSingleBlock(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "genesis")

	if !c.Validate() {
		t.Error("Validate() should return true for a single-block chain")
	}
}

func TestValidateValidChain(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "block 0")
	c.AppendBlock(1, "block 1")
	c.AppendBlock(2, "block 2")

	if !c.Validate() {
		t.Error("Validate() should return true for a valid chain")
	}
}

func TestValidateHeightMismatch(t *testing.T) {
	c := Blockchain{}
	b := NewBlock(5, [32]byte{0}, 0, "wrong height block")
	c.PrependBlock(b)

	if c.Validate() {
		t.Error("Validate() should return false when block.Height != index")
	}
}

func TestValidateBrokenHashLink(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "block 0")
	c.AppendBlock(1, "block 1")

	block2 := c.GetBlock(1)
	block2.PrevBlockHash = sha256.Sum256([]byte("wrong"))
	c.blocks[1] = &block2

	if c.Validate() {
		t.Error("Validate() should return false when PrevBlockHash link is broken")
	}
}

func TestValidateTamperedData(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "genesis data")
	c.AppendBlock(1, "block 1 data")

	c.blocks[0].Data = "tampered data"

	if c.Validate() {
		t.Error("Validate() should return false when block data was tampered with")
	}
}

// Encode() returns an error on an empty chain because there is no last block to reference.
func TestEncodeEmptyChain(t *testing.T) {
	c := Blockchain{}

	_, err := c.Encode()
	if err == nil {
		t.Error("Encode() on empty chain should return an error, but it did not")
	}
}

func TestEncodeSingleBlockChain(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "genesis")

	encoded, err := c.Encode()
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encode() returned empty output for single-block chain")
	}

	parts := bytes.SplitN(encoded, []byte("ꭣ"), 2)
	if len(parts) != 2 {
		t.Fatalf("Encode() output missing ꭣ delimiter: %q", string(encoded))
	}

	heightStr := string(parts[0])
	expectedHeight := c.Height()
	if heightStr != fmt.Sprintf("%d", expectedHeight) {
		t.Errorf("Encode() height = %q, want %d", heightStr, expectedHeight)
	}

	expectedHash := fmt.Sprintf("%x", c.GetBlock(expectedHeight-1).Hash())
	if string(parts[1]) != expectedHash {
		t.Errorf("Encode() hash = %q, want %q", string(parts[1]), expectedHash)
	}
}

func TestEncodeMultiBlockChain(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "block 0")
	c.AppendBlock(1, "block 1")
	c.AppendBlock(2, "block 2")

	encoded, err := c.Encode()
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encode() returned empty output for multi-block chain")
	}

	parts := bytes.SplitN(encoded, []byte("ꭣ"), 2)
	if len(parts) != 2 {
		t.Fatalf("Encode() output missing ꭣ delimiter: %q", string(encoded))
	}

	heightStr := string(parts[0])
	expectedHeight := c.Height()
	if heightStr != fmt.Sprintf("%d", expectedHeight) {
		t.Errorf("Encode() height = %q, want %d", heightStr, expectedHeight)
	}

	expectedHash := fmt.Sprintf("%x", c.GetBlock(expectedHeight-1).Hash())
	if string(parts[1]) != expectedHash {
		t.Errorf("Encode() hash = %q, want %q", string(parts[1]), expectedHash)
	}
}
