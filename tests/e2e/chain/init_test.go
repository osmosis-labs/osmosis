package chain_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func TestChainInit(t *testing.T) {
	const (
		id = chain.ChainAID
	)

	var (
		nodeConfigs = []*chain.NodeConfig{
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
		dataDir, err = ioutil.TempDir("", "osmosis-e2e-testnet-test")
	)

	chain, err := chain.Init(id, dataDir, nodeConfigs, time.Second*3)
	require.NoError(t, err)

	require.Equal(t, chain.ChainMeta.DataDir, dataDir)
	require.Equal(t, chain.ChainMeta.Id, id)

	require.Equal(t, len(nodeConfigs), len(chain.Nodes))

	actualNodes := chain.Nodes

	expectedConfigFiles := []string{
		"app.toml", "config.toml", "genesis.json", "node_key.json", "priv_validator_key.json",
	}

	for i, expectedConfig := range nodeConfigs {
		actualNode := actualNodes[i]

		require.Equal(t, fmt.Sprintf("%s-node-%s", id, expectedConfig.Name), actualNode.Name)
		require.Equal(t, expectedConfig.IsValidator, actualNode.IsValidator)

		expectedPath := fmt.Sprintf("%s/%s/%s-node-%s", dataDir, id, id, expectedConfig.Name)

		require.Equal(t, expectedPath, actualNode.ConfigDir)

		require.NotEmpty(t, actualNode.Mnemonic)
		require.NotEmpty(t, actualNode.PublicAddress)
		require.NotEmpty(t, actualNode.PeerId)

		for _, expectedFileName := range expectedConfigFiles {
			expectedFilePath := path.Join(expectedPath, "config", expectedFileName)
			_, err := os.Stat(expectedFilePath)
			require.NoError(t, err)
		}
		_, err := os.Stat(path.Join(expectedPath, "keyring-test"))
		require.NoError(t, err)
	}
}

func TestNodeInit(t *testing.T) {
	const (
		id = chain.ChainAID
	)

	var (
		nodeConfigs = []*chain.NodeConfig{
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
		dataDir, err = ioutil.TempDir("", "osmosis-e2e-testnet-test")
	)

	chain, err := chain.Init(id, dataDir, nodeConfigs, time.Second*3)
	require.NoError(t, err)

	require.Equal(t, chain.ChainMeta.DataDir, dataDir)
	require.Equal(t, chain.ChainMeta.Id, id)

	require.Equal(t, len(nodeConfigs), len(chain.Nodes))

	actualNodes := chain.Nodes

	expectedConfigFiles := []string{
		"app.toml", "config.toml", "genesis.json", "node_key.json", "priv_validator_key.json",
	}

	for i, expectedConfig := range nodeConfigs {
		actualNode := actualNodes[i]

		require.Equal(t, fmt.Sprintf("%s-node-%s", id, expectedConfig.Name), actualNode.Name)
		require.Equal(t, expectedConfig.IsValidator, actualNode.IsValidator)

		expectedPath := fmt.Sprintf("%s/%s/%s-node-%s", dataDir, id, id, expectedConfig.Name)

		require.Equal(t, expectedPath, actualNode.ConfigDir)

		require.NotEmpty(t, actualNode.Mnemonic)
		require.NotEmpty(t, actualNode.PublicAddress)
		require.NotEmpty(t, actualNode.PeerId)

		for _, expectedFileName := range expectedConfigFiles {
			expectedFilePath := path.Join(expectedPath, "config", expectedFileName)
			_, err := os.Stat(expectedFilePath)
			require.NoError(t, err)
		}
		_, err := os.Stat(path.Join(expectedPath, "keyring-test"))
		require.NoError(t, err)
	}
}
