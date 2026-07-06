package blockchain

import (
	"fmt"
)

type Blockchain struct {
	blocks []*Block
}

func (c *Blockchain) PrependBlock(b *Block) {
	c.blocks = append([]*Block{b}, c.blocks...)
}

func (c *Blockchain) AppendBlock(author uint64, data string) Block {
	prevBlockHash := [32]byte{0}
	height := c.Height()
	if height > 0 {
		prevBlockHash = c.GetBlock(height - 1).Hash()
	}

	b := NewBlock(height, prevBlockHash, author, data)
	c.blocks = append(c.blocks, b)
	return *b
}

func (c Blockchain) GetBlock(idx uint64) Block {
	return *c.blocks[idx]
}

func (c Blockchain) Height() uint64 {
	return uint64(len(c.blocks))
}

func (c Blockchain) Validate() bool {
	for i := uint64(0); i < c.Height(); i++ {
		block := c.GetBlock(i)

		if i != block.Height {
			return false
		} else if block.Height > 0 && block.PrevBlockHash != c.GetBlock(i-1).Hash() {
			return false
		}
	}

	return true
}

func (c Blockchain) Encode() []byte {
	height := c.Height()
	lastBlock := c.GetBlock(height - 1)
	data := fmt.Sprintf("%dꭣ%x", height, lastBlock.Hash())

	return []byte(data)
}
