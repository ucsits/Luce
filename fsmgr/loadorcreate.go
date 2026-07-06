package fsmgr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ucsits/Luce/blockchain"
)

func LoadOrCreate(dir string) (*blockchain.Blockchain, error) {
	var chain blockchain.Blockchain

	metaPath := filepath.Join(dir, ".luce", "metadata")
	_, err := os.Stat(metaPath)
	if os.IsNotExist(err) {
		Genesis(&chain)
		if err := Dump(dir, chain); err != nil {
			return nil, fmt.Errorf("persisting genesis: %w", err)
		}
		return &chain, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checking metadata: %w", err)
	}
	if err := Load(dir, &chain); err != nil {
		return nil, fmt.Errorf("loading blockchain: %w", err)
	}
	return &chain, nil
}
