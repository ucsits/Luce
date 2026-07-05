package blockchain

type Blockchain struct {
	blocks []*Block
}

func (c Blockchain) AppendBlock(author uint64, data []rune) Block {
	height := c.Height()
	prevBlock := c.GetBlock(height - 2)
	b := NewBlock(height, prevBlock.Hash(), author, data)
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
