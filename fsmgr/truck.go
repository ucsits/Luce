package fsmgr

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ucsits/Luce/blockchain"
)

func mkluce(basepath string) {
	err := os.MkdirAll(filepath.Join(basepath, ".luce"), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func Dump(basepath string, c blockchain.Blockchain) {
	encodedMetadata := c.Encode()
	data := []byte(encodedMetadata)

	mkluce(basepath)
	err := os.WriteFile(filepath.Join(basepath, ".luce", "metadata"), data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	for i := uint64(0); i < c.Height(); i++ {
		block := c.GetBlock(i)
		data := block.Encode()
		blockHash := fmt.Sprintf("%x", block.Hash())

		err := os.WriteFile(filepath.Join(basepath, ".luce", blockHash), data, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Load(basepath string, c *blockchain.Blockchain) {
	data, err := os.ReadFile(filepath.Join(basepath, ".luce", "metadata"))
	if err != nil {
		log.Fatal(err)
	}

	var height uint64
	var headHashStr string
	var headHash [32]byte
	_, err = fmt.Sscanf(string(data), "%dꭣ%x", &height, &headHashStr)

	if err != nil {
		log.Fatal(err)
	}

	dec, err := hex.DecodeString(headHashStr)
	if err == nil && len(dec) == 32 {
		copy(headHash[:], dec)
	} else {
		log.Fatal(err)
	}

	for i := uint64(0); i < height; i++ {
		block := blockchain.NewBlockFromFile(filepath.Join(basepath, ".luce", headHashStr))
		c.PrependBlock(block)

		headHashStr = fmt.Sprintf("%x", block.PrevBlockHash)
	}
}
