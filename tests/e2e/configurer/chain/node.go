package chain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

type NodeConfig struct {
	initialization.Node

	OperatorAddress  string
	chainId          string
	rpcClient        *rpchttp.HTTP
	t                *testing.T
	containerManager *containers.Manager
}

// NewNodeConfig returens new initialized NodeConfig.
func NewNodeConfig(t *testing.T, initNode *initialization.Node, chainId string, containerManager *containers.Manager) *NodeConfig {
	return &NodeConfig{
		Node:             *initNode,
		chainId:          chainId,
		containerManager: containerManager,
		t:                t,
	}
}

// Run runs a node container for the given nodeIndex.
// The node configuration must be already added to the chain config prior to calling this
// method.
func (n *NodeConfig) Run() error {
	n.t.Logf("starting %s validator container: %s", n.chainId, n.Name)
	resource, err := n.containerManager.RunNodeResource(n.chainId, n.Name, n.ConfigDir)
	if err != nil {
		return err
	}

	hostPort := resource.GetHostPort("26657/tcp")
	rpcClient, err := rpchttp.New("tcp://"+hostPort, "/websocket")
	if err != nil {
		return err
	}

	require.Eventually(
		n.t,
		func() bool {
			if _, err := rpcClient.Health(context.Background()); err != nil {
				return false
			}

			n.t.Logf("started %s node container: %s", resource.Container.Name[1:], resource.Container.ID)
			return true
		},
		5*time.Minute,
		time.Second,
		"Osmosis node failed to produce blocks",
	)

	n.rpcClient = rpcClient

	if err := n.extractOperatorAddressIfValidator(); err != nil {
		return err
	}

	return nil
}

// WaitUntil waits until node reaches doneCondition. Return nil
// if reached, error otherwise.
func (n *NodeConfig) WaitUntil(doneCondition func(syncInfo coretypes.SyncInfo) bool) error {
	var latestBlockHeight int64
	for i := 0; i < waitUntilrepeatMax; i++ {
		status, err := n.rpcClient.Status(context.Background())
		if err != nil {
			return err
		}
		latestBlockHeight = status.SyncInfo.LatestBlockHeight
		// let the node produce a few blocks
		if !doneCondition(status.SyncInfo) {
			time.Sleep(waitUntilRepeatPauseTime)
			continue
		}
		return nil
	}
	return fmt.Errorf("node %s timed out waiting for condition, latest block height was %d", n.Name, latestBlockHeight)
}

func (n *NodeConfig) extractOperatorAddressIfValidator() error {
	if !n.IsValidator {
		n.t.Logf("node (%s) is not a validator, skipping", n.Name)
		return nil
	}

	cmd := []string{"osmosisd", "debug", "addr", n.PublicKey}
	n.t.Logf("extracting validator operator addresses for validator: %s", n.Name)
	_, errBuf, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	if err != nil {
		return err
	}
	re := regexp.MustCompile("osmovaloper(.{39})")
	operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
	n.OperatorAddress = strings.TrimSuffix(operAddr, "\n")
	return nil
}
