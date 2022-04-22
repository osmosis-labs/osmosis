package main

import (
	"fmt"
	"io/ioutil"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func main() {
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")
	if err != nil {
		panic(err)
	}

	chain, err := chain.Init(chain.ChainAID, tmpDir)
	if err != nil {
		panic(err)
	}
	fmt.Println(chain)
}
