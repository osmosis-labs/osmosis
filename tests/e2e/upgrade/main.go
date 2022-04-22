package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func main() {
	var dataDir string
	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.Parse()

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

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
