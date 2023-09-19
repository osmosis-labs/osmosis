package chain

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/tests/e2e/configurer/config"

	"github.com/osmosis-labs/osmosis/v19/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v19/tests/e2e/initialization"
)

type Config struct {
	initialization.ChainMeta

	ValidatorInitConfigs []*initialization.NodeConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod          float32
	ExpeditedVotingPeriod float32
	// upgrade proposal height for chain.
	UpgradePropHeight int64

	NodeConfigs     []*NodeConfig
	NodeTempConfigs []*NodeConfig

	LatestCodeId int

	t                *testing.T
	containerManager *containers.Manager
}

const (
	// defaultNodeIndex to use for querying and executing transactions.
	// It is used when we are indifferent about the node we are working with.
	defaultNodeIndex = 0
	// waitUntilRepeatPauseTime is the time to wait between each check of the node status.
	waitUntilRepeatPauseTime = 2 * time.Second
	// waitUntilrepeatMax is the maximum number of times to repeat the wait until condition.
	waitUntilrepeatMax = 60

	proposalStatusPassed = "PROPOSAL_STATUS_PASSED"
)

func New(t *testing.T, containerManager *containers.Manager, id string, initValidatorConfigs []*initialization.NodeConfig) *Config {
	t.Helper()
	numVal := float32(len(initValidatorConfigs))
	return &Config{
		ChainMeta: initialization.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs:  initValidatorConfigs,
		VotingPeriod:          config.PropDepositBlocks + numVal*config.PropVoteBlocks + config.PropBufferBlocks,
		ExpeditedVotingPeriod: config.PropDepositBlocks + numVal*config.PropVoteBlocks + config.PropBufferBlocks - 2,
		t:                     t,
		containerManager:      containerManager,
	}
}

// CreateNode returns new initialized NodeConfig.
func (c *Config) CreateNode(initNode *initialization.Node) *NodeConfig {
	nodeConfig := &NodeConfig{
		Node:             *initNode,
		chainId:          c.Id,
		containerManager: c.containerManager,
		t:                c.t,
	}
	c.NodeConfigs = append(c.NodeConfigs, nodeConfig)
	return nodeConfig
}

// CreateNodeTemp returns new initialized NodeConfig and appends it to a separate list of temporary nodes.
// This is used for nodes that are intended to only exist for a single test. Without this separation,
// parallel tests will try and use this node and fail.
func (c *Config) CreateNodeTemp(initNode *initialization.Node) *NodeConfig {
	nodeConfig := &NodeConfig{
		Node:             *initNode,
		chainId:          c.Id,
		containerManager: c.containerManager,
		t:                c.t,
	}
	c.NodeTempConfigs = append(c.NodeTempConfigs, nodeConfig)
	return nodeConfig
}

// RemoveNode removes node and stops it from running.
func (c *Config) RemoveNode(nodeName string) error {
	for i, node := range c.NodeConfigs {
		if node.Name == nodeName {
			c.NodeConfigs = append(c.NodeConfigs[:i], c.NodeConfigs[i+1:]...)
			return node.Stop()
		}
	}
	return fmt.Errorf("node %s not found", nodeName)
}

// RemoveTempNode removes a temporary node and stops it from running.
func (c *Config) RemoveTempNode(nodeName string) error {
	for i, node := range c.NodeTempConfigs {
		if node.Name == nodeName {
			c.NodeTempConfigs = append(c.NodeTempConfigs[:i], c.NodeTempConfigs[i+1:]...)
			return node.Stop()
		}
	}
	return fmt.Errorf("node %s not found", nodeName)
}

// WaitUntilEpoch waits for the chain to reach the specified epoch.
func (c *Config) WaitUntilEpoch(epoch int64, epochIdentifier string) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	node.WaitUntil(func(_ coretypes.SyncInfo) bool {
		newEpochNumber := node.QueryCurrentEpoch(epochIdentifier)
		c.t.Logf("current epoch number is (%d), waiting to reach (%d)", newEpochNumber, epoch)
		return newEpochNumber >= epoch
	})
}

// WaitForNumEpochs waits for the chain to to go through a given number of epochs.
func (c *Config) WaitForNumEpochs(epochsToWait int64, epochIdentifier string) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	oldEpochNumber := node.QueryCurrentEpoch(epochIdentifier)
	c.WaitUntilEpoch(oldEpochNumber+epochsToWait, epochIdentifier)
}

// WaitUntilHeight waits for all validators to reach the specified height at the minimum.
// returns error, if any.
func (c *Config) WaitUntilHeight(height int64) {
	// Ensure the nodes are making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		curHeight := syncInfo.LatestBlockHeight

		if curHeight < height {
			c.t.Logf("current block height is %d, waiting to reach: %d", curHeight, height)
			return false
		}

		return !syncInfo.CatchingUp
	}

	for _, node := range c.NodeConfigs {
		c.t.Logf("node container: %s, waiting to reach height %d", node.Name, height)
		node.WaitUntil(doneCondition)
	}
}

// WaitForNumHeights waits for all nodes to go through a given number of heights.
func (c *Config) WaitForNumHeights(heightsToWait int64) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	currentHeight, err := node.QueryCurrentHeight()
	require.NoError(c.t, err)
	c.WaitUntilHeight(currentHeight + heightsToWait)
}

func (c *Config) SendIBC(dstChain *Config, recipient string, token sdk.Coin) {
	c.t.Logf("IBC sending %s from %s to %s (%s)", token, c.Id, dstChain.Id, recipient)

	dstNode, err := dstChain.GetDefaultNode()
	require.NoError(c.t, err)

	// removes the fee token from balances for calculating the difference in other tokens
	// before and after the IBC send. Since we run tests in parallel now, some tests may
	// send uosmo between accounts while this test is running. Since we don't care about
	// non ibc denoms, its safe to filter uosmo out.
	// TODO: we can probably improve this by specifying the denom we expect to be received
	// and just look out for that. This wasn't required prior to parallel tests, but
	// would be useful now.
	removeFeeTokenFromBalance := func(balance sdk.Coins) sdk.Coins {
		filteredCoinDenoms := []string{}
		for _, coin := range balance {
			if !strings.HasPrefix(coin.Denom, "ibc/") {
				filteredCoinDenoms = append(filteredCoinDenoms, coin.Denom)
			}
		}
		feeRewardTokenBalance := balance.FilterDenoms(filteredCoinDenoms)
		return balance.Sub(feeRewardTokenBalance)
	}

	balancesDstPreWithTxFeeBalance, err := dstNode.QueryBalances(recipient)
	require.NoError(c.t, err)
	balancesDstPre := removeFeeTokenFromBalance(balancesDstPreWithTxFeeBalance)
	cmd := []string{"hermes", "tx", "ft-transfer", "--dst-chain", dstChain.Id, "--src-chain", c.Id, "--src-port", "transfer", "--src-channel", "channel-0", "--amount", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	_, _, err = c.containerManager.ExecHermesCmd(c.t, cmd, "SUCCESS")
	require.NoError(c.t, err)

	cmd = []string{"hermes", "clear", "packets", "--chain", dstChain.Id, "--port", "transfer", "--channel", "channel-0"}
	_, _, err = c.containerManager.ExecHermesCmd(c.t, cmd, "SUCCESS")
	require.NoError(c.t, err)

	cmd = []string{"hermes", "clear", "packets", "--chain", c.Id, "--port", "transfer", "--channel", "channel-0"}
	_, _, err = c.containerManager.ExecHermesCmd(c.t, cmd, "SUCCESS")
	require.NoError(c.t, err)

	require.Eventually(
		c.t,
		func() bool {
			balancesDstPostWithTxFeeBalance, err := dstNode.QueryBalances(recipient)
			require.NoError(c.t, err)
			balancesDstPost := removeFeeTokenFromBalance(balancesDstPostWithTxFeeBalance)

			ibcCoin := balancesDstPost.Sub(balancesDstPre)
			if ibcCoin.Len() == 1 {
				tokenPre := balancesDstPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesDstPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := token.Amount
				resPost := tokenPost.Sub(tokenPre)
				return resPost.Uint64() == resPre.Uint64()
			} else {
				return false
			}
		},
		1*time.Minute,
		10*time.Millisecond,
		"tx not received on destination chain",
	)

	c.t.Log("successfully sent IBC tokens")
}

func (c *Config) GetAllChainNodes() []*NodeConfig {
	nodeConfigs := make([]*NodeConfig, len(c.NodeConfigs))
	copy(nodeConfigs, c.NodeConfigs)
	return nodeConfigs
}

// GetDefaultNode returns the default node of the chain.
// The default node is the first one created. Returns error if no
// ndoes created.
func (c *Config) GetDefaultNode() (*NodeConfig, error) {
	return c.getNodeAtIndex(defaultNodeIndex)
}

// GetPersistentPeers returns persistent peers from every node
// associated with a chain.
func (c *Config) GetPersistentPeers() []string {
	peers := make([]string, len(c.NodeConfigs))
	for i, node := range c.NodeConfigs {
		peers[i] = node.PeerId
	}
	return peers
}

// Returns the nodeIndex'th node on the chain
func (c *Config) GetNodeAtIndex(nodeIndex int) (*NodeConfig, error) {
	return c.getNodeAtIndex(nodeIndex)
}

func (c *Config) getNodeAtIndex(nodeIndex int) (*NodeConfig, error) {
	if nodeIndex > len(c.NodeConfigs) {
		return nil, fmt.Errorf("node index (%d) is greater than the number of nodes available (%d)", nodeIndex, len(c.NodeConfigs))
	}
	return c.NodeConfigs[nodeIndex], nil
}

func (c *Config) SubmitCreateConcentratedPoolProposal(chainANode *NodeConfig) (uint64, error) {
	propNumber := chainANode.SubmitCreateConcentratedPoolProposal(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.InitialMinDeposit)))

	chainANode.DepositProposal(propNumber, false)

	var wg sync.WaitGroup

	for _, n := range c.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	require.Eventually(c.t, func() bool {
		status, err := chainANode.QueryPropStatus(propNumber)
		if err != nil {
			return false
		}
		return status == proposalStatusPassed
	}, time.Second*30, 10*time.Millisecond)
	poolId := chainANode.QueryNumPools()
	return poolId, nil
}
