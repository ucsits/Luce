package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {
	prevHash := [32]byte{0}
	before := uint64(time.Now().Unix())

	b := NewBlock(0, prevHash, 12345, "hello world")

	if b.Height != 0 {
		t.Errorf("Height = %d, want 0", b.Height)
	}
	if b.Author != 12345 {
		t.Errorf("Author = %d, want 12345", b.Author)
	}
	if b.PrevBlockHash != prevHash {
		t.Errorf("PrevBlockHash mismatch")
	}
	if b.Data != "hello world" {
		t.Errorf("Data = %q, want %q", b.Data, "hello world")
	}
	if b.Timestamp < before {
		t.Errorf("Timestamp %d is before test start %d", b.Timestamp, before)
	}
}

func TestHash(t *testing.T) {
	b1 := NewBlock(0, [32]byte{0}, 0, "test data")
	b1.Timestamp = 1000

	b2 := NewBlock(0, [32]byte{0}, 0, "test data")
	b2.SetTimestamp(1000)

	if b1.Hash() != b2.Hash() {
		t.Error("Hash() is not deterministic — same fields produce different hashes")
	}
}

func TestHashChangesWithDifferentFields(t *testing.T) {
	base := NewBlock(0, [32]byte{0}, 0, "data")
	base.Timestamp = 1000
	baseHash := base.Hash()

	tests := []struct {
		name  string
		block Block
	}{
		{"Height", Block{Height: 1, Author: 0, Timestamp: 1000, PrevBlockHash: [32]byte{0}, Data: "data"}},
		{"Author", Block{Height: 0, Author: 1, Timestamp: 1000, PrevBlockHash: [32]byte{0}, Data: "data"}},
		{"Timestamp", Block{Height: 0, Author: 0, Timestamp: 2000, PrevBlockHash: [32]byte{0}, Data: "data"}},
		{"PrevBlockHash", Block{Height: 0, Author: 0, Timestamp: 1000, PrevBlockHash: [32]byte{1}, Data: "data"}},
		{"Data", Block{Height: 0, Author: 0, Timestamp: 1000, PrevBlockHash: [32]byte{0}, Data: "different"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.block.Hash() == baseHash {
				t.Errorf("Hash did not change when %s was modified", tt.name)
			}
		})
	}
}

func TestHashKnownAnswer(t *testing.T) {
	b := NewBlock(0, [32]byte{0}, 0, "known")
	b.SetTimestamp(0)

	expected := sha256.Sum256(b.Format())
	if b.Hash() != expected {
		t.Errorf("Hash() = %x, expected sha256.Sum256(Format()) = %x", b.Hash(), expected)
	}
}

func TestFormat(t *testing.T) {
	b := NewBlock(1, [32]byte{0}, 42, "hello")
	b.SetTimestamp(1000)
	formatted := b.Format()

	if !bytes.Contains(formatted, []byte("ꭣ")) {
		t.Error("Format() output does not contain the ꭣ delimiter")
	}
	str := string(formatted)
	if !strings.Contains(str, "1") || !strings.Contains(str, "42") || !strings.Contains(str, "hello") {
		t.Errorf("Format() = %q, missing expected fields", str)
	}
}

func TestEncode(t *testing.T) {
	b := NewBlock(0, [32]byte{0}, 7, "encoded data")
	b.SetTimestamp(500)
	encoded := b.Encode()
	str := string(encoded)

	if !strings.Contains(str, "ꭣ") {
		t.Error("Encode() output does not contain the ꭣ delimiter")
	}
	if !strings.Contains(str, "7") {
		t.Error("Encode() output missing author field")
	}
	if !strings.Contains(str, "encoded data") {
		t.Error("Encode() output missing data field")
	}
}

func TestFormatVsEncode(t *testing.T) {
	b := NewBlock(0, [32]byte{0}, 1, "test")
	b.SetTimestamp(100)

	formatted := string(b.Format())
	encoded := string(b.Encode())

	if formatted == encoded {
		t.Error("Format() and Encode() produce identical output; Encode should include the hash")
	}

	if !strings.Contains(encoded, fmt.Sprintf("%x", b.Hash())) {
		t.Error("Encode() output should contain the block's hash")
	}
}

func TestNewBlockFromFileRoundTrip(t *testing.T) {
	b := NewBlock(0, [32]byte{0}, 7, "multi word test data")
	b.SetTimestamp(500)
	encoded := b.Encode()

	tmpFile := filepath.Join(t.TempDir(), "block")
	if err := os.WriteFile(tmpFile, encoded, 0644); err != nil {
		t.Fatal(err)
	}

	decoded, err := NewBlockFromFile(tmpFile)
	if err != nil {
		t.Fatalf("NewBlockFromFile returned error: %v", err)
	}
	if decoded.Hash() != b.Hash() {
		t.Errorf("round-trip hash mismatch: got %x, want %x", decoded.Hash(), b.Hash())
	}
	if decoded.Data != b.Data {
		t.Errorf("round-trip data mismatch: got %q, want %q", decoded.Data, b.Data)
	}
}
