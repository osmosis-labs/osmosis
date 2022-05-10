package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ory/dockertest/v3/docker"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

func (s *IntegrationTestSuite) connectIBCChains() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chains[0].ChainMeta.Id, s.chains[1].ChainMeta.Id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			s.chains[0].ChainMeta.Id,
			s.chains[1].ChainMeta.Id,
			"--port-a=transfer",
			"--port-b=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Containsf(
		errBuf.String(),
		"successfully opened init channel",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chains[0].ChainMeta.Id, s.chains[1].ChainMeta.Id)
}

func (s *IntegrationTestSuite) sendIBC(srcChain *chain.Chain, dstChain *chain.Chain, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChain.ChainMeta.Id, dstChain.ChainMeta.Id, recipient)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"tx",
			"raw",
			"ft-transfer",
			dstChain.ChainMeta.Id,
			srcChain.ChainMeta.Id,
			"transfer",  // source chain port ID
			"channel-0", // since only one connection/channel exists, assume 0
			token.Amount.String(),
			fmt.Sprintf("--denom=%s", token.Denom),
			fmt.Sprintf("--receiver=%s", recipient),
			"--timeout-height-offset=1000",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "Success"),
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	currentHeight := s.chainHeight(s.valResources[srcChain.ChainMeta.Id][0].Container.ID)
	// required to wait 3 blocks (~3 seconds) in order to prevent account sequence mismatches
	s.Require().Eventually(
		func() bool {
			return s.chainHeight(s.valResources[srcChain.ChainMeta.Id][0].Container.ID) > currentHeight+2
		},
		5*time.Minute,
		time.Second,
	)

	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) submitProposal(c *chain.Chain, upgradeHeight int) {
	upgradeHeightStr := strconv.Itoa(upgradeHeight)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("submitting upgrade proposal on %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
		User:         "root",
		Cmd: []string{
			"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", "v8", "--title=\"v8 upgrade\"", "--description=\"v8 upgrade proposal\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to submit proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	currentHeight := s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID)

	s.Require().Eventually(
		func() bool {
			return s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID) > currentHeight+2
		},
		5*time.Minute,
		time.Second,
	)

	s.T().Log("successfully submitted proposal")
}

func (s *IntegrationTestSuite) depositProposal(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("depositing to upgrade proposal from %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
		User:         "root",
		Cmd: []string{
			"osmosisd", "tx", "gov", "deposit", "1", "10000000stake", "--from=val", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to deposit to upgrade proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	currentHeight := s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID)

	s.Require().Eventually(
		func() bool {
			return s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID) > currentHeight+2
		},
		5*time.Minute,
		time.Second,
	)

	s.T().Log("successfully deposited to proposal")

}

func (s *IntegrationTestSuite) voteProposal(c *chain.Chain, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("voting for upgrade proposal for chain-id: %s", c.ChainMeta.Id)
	for i := range c.Validators {
		exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
			Context:      ctx,
			AttachStdout: true,
			AttachStderr: true,
			Container:    s.valResources[c.ChainMeta.Id][i].Container.ID,
			User:         "root",
			Cmd: []string{
				"osmosisd", "tx", "gov", "vote", "1", "yes", "--from=val", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
			},
		})
		s.Require().NoError(err)

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
			Context:      ctx,
			Detach:       false,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
		})

		s.Require().NoErrorf(
			err,
			"failed to vote for proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		s.Require().Truef(
			strings.Contains(outBuf.String(), "code: 0"),
			"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		s.T().Logf("successfully voted for proposal from %s container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) chainStatus(containerId string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"osmosisd", "status",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	errBufByte := errBuf.Bytes()
	return errBufByte

}

func (s *IntegrationTestSuite) chainHeight(containerId string) int {
	var block syncInfo
	out := s.chainStatus(containerId)
	json.Unmarshal(out, &block)
	currentHeight, err := strconv.Atoi(block.SyncInfo.LatestHeight)
	s.Require().NoError(err)
	return currentHeight
}

func (s *IntegrationTestSuite) queryBalances(containerId string, addr string) (sdk.Coins, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"osmosisd", "query", "bank", "balances", addr, "--output=json",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	outBufByte := outBuf.Bytes()
	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(outBufByte, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.GetBalances(), nil

}

func (s *IntegrationTestSuite) createPool(c *chain.Chain, poolFile string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
		User:         "root",
		Cmd: []string{
			"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test",
		},
	})
	s.Require().NoError(err)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to create pool; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	currentHeight := s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID)

	s.Require().Eventually(
		func() bool {
			return s.chainHeight(s.valResources[c.ChainMeta.Id][0].Container.ID) > currentHeight+2
		},
		5*time.Minute,
		time.Second,
	)

	s.T().Logf("successfully created pool from %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

}
