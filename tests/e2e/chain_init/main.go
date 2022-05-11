package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func main() {
	var (
		valConfig    []*chain.ValidatorConfig
		dataDir      string
		chainId      string
		config       string
		votingPeriod time.Duration
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.StringVar(&config, "config", "", "serialized config")
	flag.DurationVar(&votingPeriod, "voting-period", 30000000000, "votingPeriod")

	flag.Parse()

	err := json.Unmarshal([]byte(config), &valConfig)
	if err != nil {
		panic(err)
	}

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	createdChain, err := chain.Init(chainId, dataDir, valConfig, votingPeriod)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(createdChain)
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainId)
	if err = os.WriteFile(fileName, b, 0o777); err != nil {
		panic(err)
	}
}
