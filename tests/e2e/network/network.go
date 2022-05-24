package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerconfig "github.com/osmosis-labs/osmosis/v7/tests/e2e/docker"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/network/portoffset"
)

type Network struct {
	t     *testing.T
	index int
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	votingPeriod int
	// upgrade proposal height for chain.
	proposalHeight   int
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
		votingPeriod:     propDepositBlocks + numValidators*propVoteBlocks + propBufferBlocks,
		dockerResources:  dockerResources,
		dockerImages:     dockerImages,
		workingDirectory: workingDirectory,
	}
}

func (n *Network) GetChain() *chain.Chain {
	return &n.chain
}

func (n *Network) GetVotingPeriod() int {
	return n.votingPeriod
}

// GetCurrentHeightFromValidator returns current height by querying a validator with
// validatorIndex.
func (n *Network) GetCurrentHeightFromValidator(validatorIndex int) (int, error) {
	var block syncInfo
	out, err := n.chainStatus(validatorIndex)
	if err != nil {
		return 0, err
	}
	if err = json.Unmarshal(out, &block); err != nil {
		return 0, err
	}
	currentHeight, err := strconv.Atoi(block.SyncInfo.LatestHeight)
	if err != nil {
		return 0, err
	}
	return currentHeight, nil
}

func (n *Network) GetProposalHeight() int {
	return n.proposalHeight
}

// WaitUntil waits until validator with validatorIndex reaches doneCondition. Return nil
// if reached, error otherwise.
func (n *Network) WaitUntil(validatorIndex int, doneCondition func(syncInfo coretypes.SyncInfo) bool) error {
	hostPort := n.dockerResources.Validators[n.chain.ChainMeta.Id][validatorIndex].GetHostPort("26657/tcp")
	rpcClient, err := rpchttp.New("tcp://"+hostPort, "/websocket")
	if err != nil {
		return err
	}
	var latestBlockHeight int64
	for i := 0; i < repeatMax; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), repeatTime)
		defer cancel()
		status, err := rpcClient.Status(ctx)
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

func (n *Network) CalclulateAndSetProposalHeight(currentHeight int) {
	n.proposalHeight = currentHeight + int(n.votingPeriod) + int(propSubmitBlocks) + int(propBufferBlocks)
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
	for i := range chain.Validators {
		// expose the first two validators for state sync. State-sync needs at least
		// 2 RPC servers to be enabled to work.
		_, err := n.RunValidator(i, i == 0 || i == 1)
		if err != nil {
			return nil, err
		}

	}

	// Ensure the node is making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		return syncInfo.CatchingUp || syncInfo.LatestBlockHeight > 3
	}

	if err := n.WaitUntil(0, doneCondition); err != nil {
		return nil, err
	}
	return n.dockerResources.Validators[n.chain.ChainMeta.Id], nil
}

func (n *Network) RunValidator(validatorIndex int, shouldExposePorts bool) (*dockertest.Resource, error) {
	runOpts := n.getValidatorOptions(validatorIndex)
	if shouldExposePorts {
		runOpts.PortBindings = n.getPortBindings()
		n.t.Logf("exposing ports for validator %s with port mapping: \n%v\n", n.chain.Validators[validatorIndex].Name, runOpts.PortBindings)
	}
	resource, err := n.dockerResources.Pool.RunWithOptions(runOpts, noRestart)
	if err != nil {
		return nil, err
	}
	n.dockerResources.Validators[n.chain.ChainMeta.Id][validatorIndex] = resource
	n.t.Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	return resource, nil
}

func (n *Network) chainStatus(validatorIndex int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	containerId := n.dockerResources.Validators[n.chain.ChainMeta.Id][validatorIndex].Container.ID

	exec, err := n.dockerResources.Pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"osmosisd", "status",
		},
	})
	if err != nil {
		return nil, err
	}

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = n.dockerResources.Pool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	if err != nil {
		n.t.Errorf("failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String())
		return nil, err
	}
	return errBuf.Bytes(), nil

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

func (c *Network) getPortBindings() map[docker.Port][]docker.PortBinding {
	portOffset := portoffset.GetNext()
	return map[docker.Port][]docker.PortBinding{
		"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
		"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
		"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
		"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
		"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
		"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
		"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
		"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
		"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
		"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
	}
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
