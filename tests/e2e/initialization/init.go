package initialization

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/osmosis-labs/osmosis/v27/tests/e2e/util"
)

func InitChain(id, dataDir string, nodeConfigs []*NodeConfig, votingPeriod, expeditedVotingPeriod time.Duration, forkHeight int) (*Chain, error) {
	chain := new(id, dataDir)

	for _, nodeConfig := range nodeConfigs {
		newNode, err := newNode(chain, nodeConfig)
		if err != nil {
			return nil, err
		}
		chain.nodes = append(chain.nodes, newNode)
	}

	if err := initGenesis(chain, votingPeriod, expeditedVotingPeriod, forkHeight); err != nil {
		return nil, err
	}

	var peers []string
	for _, peer := range chain.nodes {
		peerID := fmt.Sprintf("%s@%s:26656", peer.getNodeKey().ID(), peer.moniker)
		peer.peerId = peerID
		peers = append(peers, peerID)
	}

	for _, node := range chain.nodes {
		if node.isValidator {
			if err := node.initNodeConfigs(peers); err != nil {
				return nil, err
			}
		}
	}
	return chain.export(), nil
}

func InitSingleNode(chainId, dataDir string, existingGenesisDir string, nodeConfig *NodeConfig, votingPeriod time.Duration, trustHeight int64, trustHash string, stateSyncRPCServers []string, persistentPeers []string) (*Node, error) {
	if nodeConfig.IsValidator {
		return nil, errors.New("creating individual validator nodes after starting up chain is not currently supported")
	}

	chain := new(chainId, dataDir)
	newNode, err := newNode(chain, nodeConfig)
	if err != nil {
		return nil, err
	}

	_, err = util.CopyFile(
		existingGenesisDir,
		filepath.Join(newNode.configDir(), "config", "genesis.json"),
	)
	if err != nil {
		return nil, err
	}

	if err := newNode.initNodeConfigs(persistentPeers); err != nil {
		return nil, err
	}

	if err := newNode.initStateSyncConfig(trustHeight, trustHash, stateSyncRPCServers); err != nil {
		return nil, err
	}

	return newNode.export(), nil
}
