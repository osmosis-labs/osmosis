package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
)

func main() {
	var (
		nodeConfigStr string

		dataDir string

		existingGenesisDir string

		chainId string

		votingPeriod time.Duration

		stateSyncRPCServersStr string

		persistentPeersStr string

		trustHeight int64

		trustHash string
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&existingGenesisDir, "genesis-dir", "", "pre-existing genesis location")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.StringVar(&nodeConfigStr, "node-config", "", "serialized node config")
	flag.DurationVar(&votingPeriod, "voting-period", 30000000000, "voting period")
	flag.StringVar(&stateSyncRPCServersStr, "rpc-servers", "", "state sync RPC servers")
	flag.StringVar(&persistentPeersStr, "peers", "", "state sync RPC servers")
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

	persistenrPeers := strings.Split(persistentPeersStr, ",")
	if len(persistenrPeers) == 0 {
		panic("persistent peers are required, separated by commas")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	_, err = initialization.InitSingleNode(chainId, dataDir, existingGenesisDir, &nodeConfig, votingPeriod, trustHeight, trustHash, stateSyncRPCServers, persistenrPeers)
	if err != nil {
		panic(err)
	}
}
