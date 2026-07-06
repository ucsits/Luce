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

func Dump(basepath string, c blockchain.Blockchain) error {
	encodedMetadata, err := c.Encode()
	if err != nil {
		return fmt.Errorf("encoding chain metadata: %w", err)
	}
	data := []byte(encodedMetadata)

	if err := mkluce(basepath); err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(basepath, ".luce", "metadata"), data, 0644)
	if err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}

	for i := uint64(0); i < c.Height(); i++ {
		block := c.GetBlock(i)
		data := block.Encode()
		blockHash := fmt.Sprintf("%x", block.Hash())

		err := os.WriteFile(filepath.Join(basepath, ".luce", blockHash), data, 0644)
		if err != nil {
			return fmt.Errorf("writing block %d: %w", i, err)
		}
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
