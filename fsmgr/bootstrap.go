package fsmgr

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ucsits/Luce/blockchain"
)

func genesis(chain *blockchain.Blockchain) {
	block := blockchain.NewBlock(
		0, [32]byte{0}, 0,
		"The social and economic conditions of an organization must not be such as to confine political participation to a small fraction of the population.",
	)
	block.Timestamp = 1783300009
	chain.PrependBlock(block)
}

func Bootstrap() {
	var chain blockchain.Blockchain

	dir, _ := os.Getwd()
	_, err := os.Stat(".luce")
	if err == nil {
		if err := Load(dir, &chain); err != nil {
			log.Fatalf("loading blockchain: %v", err)
		}
	} else {
		genesis(&chain)
	}

	log.Print("Luce up and running!")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-sig:
			if err := Dump(dir, chain); err != nil {
				log.Fatalf("dumping blockchain: %v", err)
			}
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
