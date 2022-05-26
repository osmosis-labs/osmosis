package chain

import (
	"fmt"
	"time"
)

func Init(id, dataDir string, nodeConfigs []*NodeConfig, votingPeriod time.Duration) (*Chain, error) {
	chain, err := new(id, dataDir)
	if err != nil {
		return nil, err
	}

	for _, nodeConfig := range nodeConfigs {
		newNode, err := newNode(chain, nodeConfig)
		if err != nil {
			return nil, err
		}
		chain.nodes = append(chain.nodes, newNode)
	}

	if err := initGenesis(chain, votingPeriod); err != nil {
		return nil, err
	}

	var peers []string
	for i, peer := range chain.nodes {
		peerID := fmt.Sprintf("%s@%s%d:26656", peer.getNodeKey().ID(), peer.getMoniker(), i)
		peer.setPeerId(peerID)
		peers = append(peers, peerID)
	}

	for _, node := range chain.nodes {
		if node.isValidator {
			if err := node.initValidatorConfigs(chain, peers); err != nil {
				return nil, err
			}
		}
	}
	return chain.export(), nil
}
