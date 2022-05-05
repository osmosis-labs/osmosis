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
		valConfig chain.ValidatorConfig
		dataDir   string
		chainId   string
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")

	flag.Parse()
	file := fmt.Sprintf("%v/%v-configEncode", dataDir, chainId)
	for i := 0; i < 10; i++ {
		encJson, _ := os.ReadFile(file)
		err := json.Unmarshal(encJson, &valConfig)
		if err == nil {
			break
		}

		if i == 9 {
			panic(err)
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	createdChain, err := chain.Init(chainId, dataDir, valConfig.NumVal, valConfig.Pruning, valConfig.PruningKeepRecent, valConfig.PruningInterval, valConfig.SnapshotInterval, valConfig.SnapshotKeepRecent)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(createdChain)
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainId)
	if err = os.WriteFile(fileName, b, 0o777); err != nil {
		panic(err)
	}
}
