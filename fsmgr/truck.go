package fsmgr

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ucsits/Luce/blockchain"
)

func mkluce(basepath string) error {
	err := os.MkdirAll(filepath.Join(basepath, ".luce"), 0755)
	if err != nil {
		return fmt.Errorf("creating .luce directory: %w", err)
	}
	return nil
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("syncing %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("closing %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming %s: %w", path, err)
	}
	return nil
}

func Dump(basepath string, c blockchain.Blockchain) error {
	encodedMetadata, err := c.Encode()
	if err != nil {
		return fmt.Errorf("encoding chain metadata: %w", err)
	}
	if err := mkluce(basepath); err != nil {
		return err
	}
	for i := uint64(0); i < c.Height(); i++ {
		block := c.GetBlock(i)
		blockHash := fmt.Sprintf("%x", block.Hash())
		if err := writeFileAtomic(filepath.Join(basepath, ".luce", blockHash), block.Encode()); err != nil {
			return fmt.Errorf("writing block %d: %w", i, err)
		}
	}
	if err := writeFileAtomic(filepath.Join(basepath, ".luce", "metadata"), []byte(encodedMetadata)); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}
	return nil
}

func PersistBlock(dir string, block blockchain.Block) error {
	if err := mkluce(dir); err != nil {
		return err
	}
	blockHash := fmt.Sprintf("%x", block.Hash())
	if err := writeFileAtomic(filepath.Join(dir, ".luce", blockHash), block.Encode()); err != nil {
		return fmt.Errorf("writing block %d: %w", block.Height, err)
	}
	metadata := fmt.Sprintf("%dꭣ%x", block.Height+1, block.Hash())
	if err := writeFileAtomic(filepath.Join(dir, ".luce", "metadata"), []byte(metadata)); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}
	return nil
}

func Load(basepath string, c *blockchain.Blockchain) error {
	data, err := os.ReadFile(filepath.Join(basepath, ".luce", "metadata"))
	if err != nil {
		return fmt.Errorf("reading metadata: %w", err)
	}

	var height uint64
	var headHashStr string
	var headHash [32]byte
	_, err = fmt.Sscanf(string(data), "%dꭣ%s", &height, &headHashStr)

	if err != nil {
		return fmt.Errorf("parsing metadata: %w", err)
	}

	dec, err := hex.DecodeString(headHashStr)
	if err != nil || len(dec) != 32 {
		return fmt.Errorf("decoding head hash: %w", err)
	}
	copy(headHash[:], dec)

	for i := uint64(0); i < height; i++ {
		block, err := blockchain.NewBlockFromFile(filepath.Join(basepath, ".luce", headHashStr))
		if err != nil {
			return fmt.Errorf("loading block %d: %w", i, err)
		}
		c.PrependBlock(block)

		headHashStr = fmt.Sprintf("%x", block.PrevBlockHash)
	}
	return nil
}
