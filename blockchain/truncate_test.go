package blockchain

import "testing"

func TestTruncateLast_RemovesLastBlock(t *testing.T) {
	c := Blockchain{}
	c.AppendBlock(0, "genesis")
	c.AppendBlock(1, "second")
	c.AppendBlock(2, "third")

	if c.Height() != 3 {
		t.Fatalf("height = %d, want 3", c.Height())
	}

	removed := c.TruncateLast()
	if c.Height() != 2 {
		t.Fatalf("after truncate: height = %d, want 2", c.Height())
	}
	if removed.Data != "third" {
		t.Fatalf("removed data = %q, want %q", removed.Data, "third")
	}
	if !c.Validate() {
		t.Fatal("chain should remain valid after truncate")
	}

	c.TruncateLast()
	c.TruncateLast()
	if c.Height() != 0 {
		t.Fatalf("height = %d, want 0 after truncating all", c.Height())
	}

	_ = c.TruncateLast()
	if c.Height() != 0 {
		t.Fatalf("height = %d, want 0 after truncate on empty", c.Height())
	}
}
