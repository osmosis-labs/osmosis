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
		valConfig       []*chain.NodeConfig
		dataDir         string
		chainId         string
		validatorConfig string
		votingPeriod    time.Duration
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.StringVar(&validatorConfig, "config", "", "serialized config")
	flag.DurationVar(&votingPeriod, "voting-period", 30000000000, "voting period")

	flag.Parse()

	fmt.Printf("--data-dir=%s\n", dataDir)
	fmt.Printf("--chain-id=%s\n", chainId)
	fmt.Printf("--config=%s\n", validatorConfig)
	fmt.Printf("--voting-period=%d\n", votingPeriod)

	if len(validatorConfig) == 0 {
		panic("validator config is required")
	}

	err := json.Unmarshal([]byte(validatorConfig), &valConfig)
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

	b, err := json.Marshal(createdChain)
	if err != nil {
		panic(err)
	}
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainId)
	if err = os.WriteFile(fileName, b, 0o777); err != nil {
		panic(err)
	}
}
