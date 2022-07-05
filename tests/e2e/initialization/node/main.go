package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"time"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v10/tests/e2e/initialization"
=======
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
>>>>>>> 1da14b8b (e2e: refactor initialization with single node logic (#1963))
)

func main() {
	var (
		nodeConfigStr string

		dataDir string

		existingGenesisDir string

		chainId string

		votingPeriod time.Duration

		stateSyncRPCServersStr string

		trustHeight int64

		trustHash string
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&existingGenesisDir, "genesis-dir", "", "pre-existing genesis location")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.StringVar(&nodeConfigStr, "node-config", "", "serialized node config")
	flag.DurationVar(&votingPeriod, "voting-period", 30000000000, "voting period")
	flag.StringVar(&stateSyncRPCServersStr, "rpc-servers", "", "state sync RPC servers")
	flag.Int64Var(&trustHeight, "trust-height", 0, "trust Height")
	flag.StringVar(&trustHash, "trust-hash", "", "trust hash")

	flag.Parse()

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	var nodeConfig initialization.NodeConfig
	err := json.Unmarshal([]byte(nodeConfigStr), &nodeConfig)
	if err != nil {
		panic(err)
	}

	stateSyncRPCServers := strings.Split(stateSyncRPCServersStr, ",")
	if len(stateSyncRPCServers) == 0 {
		panic("rpc-servers is required, separated by commas")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	_, err = initialization.InitSingleNode(chainId, dataDir, existingGenesisDir, &nodeConfig, votingPeriod, trustHeight, trustHash, stateSyncRPCServers)
	if err != nil {
		panic(err)
	}
}
