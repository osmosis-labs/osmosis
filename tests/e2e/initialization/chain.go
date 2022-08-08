package initialization

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

// internalChain contains the same info as chain, but with the validator structs instead using the internal validator
// representation, with more derived data
type internalChain struct {
	chainMeta ChainMeta
	nodes     []*internalNode
}

func new(id, dataDir string) (*internalChain, error) {
	chainMeta := ChainMeta{
		Id:      id,
		DataDir: dataDir,
	}
	return &internalChain{
		chainMeta: chainMeta,
	}, nil
}

func (c *internalChain) export() *Chain {
	exportNodes := make([]*Node, 0, len(c.nodes))
	for _, v := range c.nodes {
		exportNodes = append(exportNodes, v.export())
	}

	return &Chain{
		ChainMeta: c.chainMeta,
		Nodes:     exportNodes,
	}
}
