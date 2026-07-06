package fsmgr

import (
	"github.com/ucsits/Luce/blockchain"
)

func Genesis(chain *blockchain.Blockchain) {
	block := blockchain.NewBlock(
		0, [32]byte{0}, 0,
		"The social and economic conditions of an organization must not be such as to confine political participation to a small fraction of the population.",
	)
	block.Timestamp = 1783300009
	chain.PrependBlock(block)
}
