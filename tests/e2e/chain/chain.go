package chain

import (
	"fmt"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

type ChainMeta struct {
	DataDir string
	Id      string
}

type Chain struct {
	ChainMeta  ChainMeta
	Validators []*Validator
}

func new(id, dataDir string) (*Chain, error) {
	// return &ChainMeta{
	// 	Id:      id,
	// 	DataDir: dataDir,
	// }, nil
	chain := &ChainMeta{
		Id:      id,
		DataDir: dataDir,
	}
	return &Chain{
		ChainMeta: *chain,
	}, nil
}

func (c *ChainMeta) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.Id)
}

func (c *Chain) createAndInitValidators(count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) createAndInitValidatorsWithMnemonics(count int, mnemonics []string) error {
	for i := 0; i < count; i++ {
		// create node
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

		// create keys
		if err := node.createKeyFromMnemonic("val", mnemonics[i]); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) createValidator(index int) *Validator {
	return &Validator{
		ChainMeta: c.ChainMeta,
		Index:     index,
		Moniker:   fmt.Sprintf("%s-osmosis-%d", c.ChainMeta.Id, index),
	}
}
