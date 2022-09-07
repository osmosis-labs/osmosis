//go:build e2e
// +build e2e

package initialization_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/tests/e2e/initialization"
)

const forkHeight = 10

var expectedConfigFiles = []string{
	"app.toml", "config.toml", "genesis.json", "node_key.json", "priv_validator_key.json",
}

// TestChainInit tests that chain initialization correctly initializes a full chain
// and produces the desired output with genesis, chain and validator configs.
func TestChainInit(t *testing.T) {
	const id = initialization.ChainAID

	var (
		nodeConfigs = []*initialization.NodeConfig{
			{
				Name:               "0",
				Pruning:            "default",
				PruningKeepRecent:  "0",
				PruningInterval:    "0",
				SnapshotInterval:   1500,
				SnapshotKeepRecent: 2,
				IsValidator:        true,
			},
			{
				Name:               "1",
				Pruning:            "nothing",
				PruningKeepRecent:  "0",
				PruningInterval:    "0",
				SnapshotInterval:   100,
				SnapshotKeepRecent: 1,
				IsValidator:        false,
			},
		}
		dataDir, err = os.MkdirTemp("", "osmosis-e2e-testnet-test")
	)

	chain, err := initialization.InitChain(id, dataDir, nodeConfigs, time.Second*3, time.Second, forkHeight)
	require.NoError(t, err)

	require.Equal(t, chain.ChainMeta.DataDir, dataDir)
	require.Equal(t, chain.ChainMeta.Id, id)

	require.Equal(t, len(nodeConfigs), len(chain.Nodes))

	actualNodes := chain.Nodes

	for i, expectedConfig := range nodeConfigs {
		actualNode := actualNodes[i]

		validateNode(t, id, dataDir, expectedConfig, actualNode)
	}
}

// TestSingleNodeInit tests that node initialization correctly initializes a single node
// and produces the desired output with genesis, chain and validator config.
func TestSingleNodeInit(t *testing.T) {
	const (
		id = initialization.ChainAID
	)

	var (
		existingChainNodeConfigs = []*initialization.NodeConfig{
			{
				Name:               "0",
				Pruning:            "default",
				PruningKeepRecent:  "0",
				PruningInterval:    "0",
				SnapshotInterval:   1500,
				SnapshotKeepRecent: 2,
				IsValidator:        true,
			},
			{
				Name:               "1",
				Pruning:            "nothing",
				PruningKeepRecent:  "0",
				PruningInterval:    "0",
				SnapshotInterval:   100,
				SnapshotKeepRecent: 1,
				IsValidator:        true,
			},
		}
		expectedConfig = &initialization.NodeConfig{
			Name:               "2",
			Pruning:            "everything",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   100,
			SnapshotKeepRecent: 1,
			IsValidator:        false,
		}
		dataDir, err = os.MkdirTemp("", "osmosis-e2e-testnet-test")
	)

	// Setup
	existingChain, err := initialization.InitChain(id, dataDir, existingChainNodeConfigs, time.Second*3, time.Second, forkHeight)
	require.NoError(t, err)

	actualNode, err := initialization.InitSingleNode(existingChain.ChainMeta.Id, dataDir, filepath.Join(existingChain.Nodes[0].ConfigDir, "config", "genesis.json"), expectedConfig, time.Second*3, 3, "testHash", []string{"some server"}, []string{"some server"})
	require.NoError(t, err)

	validateNode(t, id, dataDir, expectedConfig, actualNode)
}

func validateNode(t *testing.T, chainId string, dataDir string, expectedConfig *initialization.NodeConfig, actualNode *initialization.Node) {
	require.Equal(t, fmt.Sprintf("%s-node-%s", chainId, expectedConfig.Name), actualNode.Name)
	require.Equal(t, expectedConfig.IsValidator, actualNode.IsValidator)

	expectedPath := fmt.Sprintf("%s/%s/%s-node-%s", dataDir, chainId, chainId, expectedConfig.Name)

	require.Equal(t, expectedPath, actualNode.ConfigDir)

	require.NotEmpty(t, actualNode.Mnemonic)
	require.NotEmpty(t, actualNode.PublicAddress)

	if expectedConfig.IsValidator {
		require.NotEmpty(t, actualNode.PeerId)
	}

	for _, expectedFileName := range expectedConfigFiles {
		expectedFilePath := path.Join(expectedPath, "config", expectedFileName)
		_, err := os.Stat(expectedFilePath)
		require.NoError(t, err)
	}
	_, err := os.Stat(path.Join(expectedPath, "keyring-test"))
	require.NoError(t, err)
}
