package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerconfig "github.com/osmosis-labs/osmosis/v7/tests/e2e/docker"
)

type Network struct {
	t     *testing.T
	index int
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	votingPeriod int64
	// upgrade proposal height for chain.
	proposalHeight   int64
	validatorRPC     []*rpchttp.HTTP
	chain            chain.Chain
	dockerResources  *dockerconfig.Resources
	dockerImages     *dockerconfig.ImageConfig
	workingDirectory string
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

const (
	// estimated number of blocks it takes to deposit for a proposal
	propDepositBlocks int = 10
	// estimated number of blocks it takes to submit for a proposal
	propSubmitBlocks int = 10
	// number of blocks it takes to vote for a single validator to vote for a proposal
	propVoteBlocks int = 1
	// number of blocks used as a calculation buffer
	propBufferBlocks int = 5

	repeatTime = 5 * time.Second
	repeatMax  = 20
)

func New(t *testing.T, index int, numValidators int, dockerResources *dockerconfig.Resources, dockerImages *dockerconfig.ImageConfig, workingDirectory string) *Network {
	return &Network{
		t:                t,
		index:            index,
		votingPeriod:     int64(propDepositBlocks + numValidators*propVoteBlocks + propBufferBlocks),
		dockerResources:  dockerResources,
		dockerImages:     dockerImages,
		workingDirectory: workingDirectory,
	}
}

func (n *Network) GetChain() *chain.Chain {
	return &n.chain
}

func (n *Network) GetVotingPeriod() int64 {
	return n.votingPeriod
}

// GetCurrentHeightFromValidator returns current height by querying a validator with
// validatorIndex.
func (n *Network) GetCurrentHeightFromValidator(validatorIndex int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repeatTime)
	defer cancel()
	status, err := n.validatorRPC[validatorIndex].Status(ctx)
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// GetHashFromBlock gets block hash at a specific height. Otherwise, error.
func (n *Network) GetHashFromBlock(height int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repeatTime)
	defer cancel()
	block, err := n.validatorRPC[0].Block(ctx, &height)
	if err != nil {
		return "", err
	}
	return block.BlockID.Hash.String(), nil
}

func (n *Network) GetProposalHeight() int64 {
	return n.proposalHeight
}

// WaitUntil waits until validator with validatorIndex reaches doneCondition. Return nil
// if reached, error otherwise.
func (n *Network) WaitUntil(validatorIndex int, doneCondition func(syncInfo coretypes.SyncInfo) bool) error {
	var latestBlockHeight int64
	for i := 0; i < repeatMax; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), repeatTime)
		defer cancel()
		status, err := n.validatorRPC[validatorIndex].Status(ctx)
		if err != nil {
			return err
		}
		latestBlockHeight = status.SyncInfo.LatestBlockHeight
		// let the node produce a few blocks
		if !doneCondition(status.SyncInfo) {
			time.Sleep(repeatTime)
			continue
		}
		return nil
	}
	return fmt.Errorf("validator with index %d timed out waiting for condition, latest block height was %d", validatorIndex, latestBlockHeight)
}

func (n *Network) CalclulateAndSetProposalHeight(currentHeight int64) {
	n.proposalHeight = currentHeight + n.votingPeriod + int64(propSubmitBlocks) + int64(propBufferBlocks)
}

func (n *Network) RemoveValidatorContainer(validatorIndex int) error {
	var opts docker.RemoveContainerOptions
	chainId := n.chain.ChainMeta.Id
	opts.ID = n.dockerResources.Validators[chainId][validatorIndex].Container.ID
	opts.Force = true
	if err := n.dockerResources.Pool.Client.RemoveContainer(opts); err != nil {
		return err
	}

	n.t.Logf("removed container: %s", n.dockerResources.Validators[chainId][validatorIndex].Container.Name[1:])
	return nil
}

func (n *Network) RunValidators() ([]*dockertest.Resource, error) {
	chain := n.chain
	n.dockerResources.Validators[n.chain.ChainMeta.Id] = make([]*dockertest.Resource, len(chain.Validators))
	n.validatorRPC = make([]*rpchttp.HTTP, len(chain.Validators))

	for i := range chain.Validators {
		_, err := n.RunValidator(i)
		if err != nil {
			n.t.Errorf("container for validator with index %d failed to run", i)
			return nil, err
		}
	}

	// Ensure the nodes are making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		return syncInfo.CatchingUp || syncInfo.LatestBlockHeight > 3
	}

	for i := range chain.Validators {
		if err := n.WaitUntil(i, doneCondition); err != nil {
			n.t.Errorf("validator with index %d failed to start", i)
			return nil, err
		}
	}

	return n.dockerResources.Validators[n.chain.ChainMeta.Id], nil
}

func (n *Network) RunValidator(validatorIndex int) (*dockertest.Resource, error) {
	runOpts := n.getValidatorOptions(validatorIndex)

	resource, err := n.dockerResources.Pool.RunWithOptions(runOpts, noRestart)
	if err != nil {
		return nil, err
	}
	n.dockerResources.Validators[n.chain.ChainMeta.Id][validatorIndex] = resource

	hostPort := resource.GetHostPort("26657/tcp")

	// This is needed to ensure that the Tenderming RPC server has enough time
	// to start up. We cannot deterministically predict how long it is going to take
	// so the value of one second is anecdotally chosen. If this sleep did not exist,
	// the first query to the Tendermint RPC could return "connection reset by peer".
	time.Sleep(time.Second)

	rpcClient, err := rpchttp.New("tcp://"+hostPort, "/websocket")
	if err != nil {
		return nil, err
	}
	n.validatorRPC[validatorIndex] = rpcClient

	n.t.Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	return resource, nil
}

func (c *Network) getValidatorOptions(valIndex int) *dockertest.RunOptions {
	validator := c.chain.Validators[valIndex]
	return &dockertest.RunOptions{
		Name:      validator.Name,
		NetworkID: c.dockerResources.Network.Network.ID,
		Mounts: []string{
			fmt.Sprintf("%s/:/osmosis/.osmosisd", validator.ConfigDir),
			fmt.Sprintf("%s/scripts:/osmosis", c.workingDirectory),
		},
		Repository: c.dockerImages.OsmosisRepository,
		Tag:        c.dockerImages.OsmosisTag,
		Cmd: []string{
			"start",
		},
	}
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
