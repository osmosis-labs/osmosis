package chain

import (
	"fmt"
	"io/ioutil"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

type Chain struct {
	DataDir    string
	Id         string
	Validators []*Validator
}

func New(id string) (*Chain, error) {
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &Chain{
		Id:      id,
		DataDir: tmpDir,
	}, nil
}

func (c *Chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.Id)
}

func (c *Chain) CreateAndInitValidators(count int) error {
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
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-osmosis-%d", c.Id, index),
	}
}
