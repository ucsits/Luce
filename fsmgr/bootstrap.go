package fsmgr

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ucsits/Luce/blockchain"
)

func Bootstrap() {
	var chain blockchain.Blockchain

	dir, _ := os.Getwd()
	_, err := os.Stat(".luce")
	if err == nil {
		if err := Load(dir, &chain); err != nil {
			log.Fatalf("loading blockchain: %v", err)
		}
	} else {
		Genesis(&chain)
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
