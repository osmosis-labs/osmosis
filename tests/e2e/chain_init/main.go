package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func main() {
	var dataDir string
	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.Parse()

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	chain, err := chain.Init(chain.ChainAID, dataDir)
	if err != nil {
		panic(err)
	}
	fmt.Println(chain)
}
