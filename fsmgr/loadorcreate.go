package fsmgr

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ucsits/Luce/blockchain"
)

func LoadOrCreate(dir string) *blockchain.Blockchain {
	var chain blockchain.Blockchain

	_, err := os.Stat(filepath.Join(dir, ".luce"))
	if err == nil {
		if err := Load(dir, &chain); err != nil {
			log.Fatalf("loading blockchain: %v", err)
		}
	} else {
		Genesis(&chain)
	}

	return &chain
}
