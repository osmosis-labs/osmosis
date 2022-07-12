package chain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

type Config struct {
	initialization.ChainMeta

	ValidatorInitConfigs []*initialization.NodeConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod float32
	// upgrade proposal height for chain.
	PropHeight           int
	LatestProposalNumber int
	LatestLockNumber     int
	NodeConfigs          []*NodeConfig

	t                *testing.T
	containerManager *containers.Manager
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

const (
	// waitUntilRepeatPauseTime is the time to wait between each check of the node status.
	waitUntilRepeatPauseTime = 500 * time.Millisecond
	// waitUntilrepeatMax is the maximum number of times to repeat the wait until condition.
	waitUntilrepeatMax = 20
)

func New(t *testing.T, containerManager *containers.Manager, id string, initValidatorConfigs []*initialization.NodeConfig) *Config {
	return &Config{
		ChainMeta: initialization.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs: initValidatorConfigs,
		t:                    t,
		containerManager:     containerManager,
	}
}

// RunNode runs a node container for the given nodeIndex.
// The node configuration must be already added to the chain config prior to calling this
// method.
func (c *Config) RunNode(nodeIndex int) error {
	c.t.Logf("starting %s validator containers...", c.Id)

	resource, err := c.containerManager.RunValidatorResource(c.Id, c.NodeConfigs[nodeIndex].Name, c.NodeConfigs[nodeIndex].ConfigDir)
	if err != nil {
		return err
	}

	hostPort := resource.GetHostPort("26657/tcp")
	rpcClient, err := rpchttp.New("tcp://"+hostPort, "/websocket")
	if err != nil {
		return err
	}

	require.Eventually(
		c.t,
		func() bool {
			if _, err := rpcClient.Health(context.Background()); err != nil {
				return false
			}

			c.t.Logf("started %s node container: %s", resource.Container.Name[1:], resource.Container.ID)
			return true
		},
		5*time.Minute,
		time.Second,
		"Osmosis node failed to produce blocks",
	)

	c.NodeConfigs[nodeIndex].rpcClient = rpcClient

	if c.NodeConfigs[nodeIndex].IsValidator {
		return c.ExtractValidatorOperatorAddress(nodeIndex)
	}

	return nil
}

// WaitUntil waits until validator with validatorIndex reaches doneCondition. Return nil
// if reached, error otherwise.
func (c *Config) WaitUntil(nodeIndex int, doneCondition func(syncInfo coretypes.SyncInfo) bool) error {
	var latestBlockHeight int64
	for i := 0; i < waitUntilrepeatMax; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), waitUntilRepeatPauseTime)
		defer cancel()
		status, err := c.NodeConfigs[nodeIndex].rpcClient.Status(ctx)
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
	return fmt.Errorf("validator with index %d timed out waiting for condition, latest block height was %d", nodeIndex, latestBlockHeight)
}

// WaitUntilHeight waits for all validators to reach the specified height at the minimum.
// returns error, if any.
func (c *Config) WaitUntilHeight(height int64) error {
	// Ensure the nodes are making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		curHeight := syncInfo.LatestBlockHeight

		if curHeight < height {
			c.t.Logf("current block height is %d, waiting to reach: %d", curHeight, height)
			return false
		}

		return !syncInfo.CatchingUp
	}

	for nodeIndex := range c.NodeConfigs {
		nodeResource, exists := c.containerManager.GetValidatorResource(c.Id, nodeIndex)
		container := nodeResource.Container
		c.t.Logf("node container: %s, id: %s, waiting to reach height %d", container.Name[1:], container.ID, height)
		if !exists {
			return fmt.Errorf("validator on chain %s  with index %d does not exist", c.Id, nodeIndex)
		}
		if err := c.WaitUntil(nodeIndex, doneCondition); err != nil {
			c.t.Errorf("validator with index %d failed to start", nodeIndex)
			return err
		}
	}
	return nil
}
