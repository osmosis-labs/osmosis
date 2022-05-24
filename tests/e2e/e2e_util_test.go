package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ory/dockertest/v3/docker"

	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

func (s *IntegrationTestSuite) connectIBCChains(chainA *chain.Chain, chainB *chain.Chain) {
	s.T().Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)

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
			chainA.ChainMeta.Id,
			chainB.ChainMeta.Id,
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

	s.T().Logf("connected %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
}

func (s *IntegrationTestSuite) sendIBC(srcChain *chain.Chain, dstChain *chain.Chain, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChain.ChainMeta.Id, dstChain.ChainMeta.Id, recipient)
	balancesBPre, err := s.queryBalances(s.valResources[dstChain.ChainMeta.Id][0].Container.ID, recipient)
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
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

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})

			return strings.Contains(outBuf.String(), "Success")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Eventually(
		func() bool {
			balancesBPost, err := s.queryBalances(s.valResources[dstChain.ChainMeta.Id][0].Container.ID, recipient)
			s.Require().NoError(err)
			ibcCoin := balancesBPost.Sub(balancesBPre)
			if ibcCoin.Len() == 1 {
				tokenPre := balancesBPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesBPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := chain.OsmoToken.Amount
				resPost := tokenPost.Sub(tokenPre)
				return resPost.Uint64() == resPre.Uint64()
			} else {
				return false
			}
		},
		5*time.Minute,
		time.Second,
		"tx not received on destination chain",
	)

	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) submitUpgradeProposal(c *chain.Chain, upgradeHeight int) {
	upgradeHeightStr := strconv.Itoa(upgradeHeight)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("submitting upgrade proposal on %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json",
				},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully submitted upgrade proposal")
	c.PropNumber = c.PropNumber + 1
}

func (s *IntegrationTestSuite) submitSuperfluidProposal(c *chain.Chain, asset string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("submitting superfluid proposal for asset %s on %s container: %s", asset, s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id),
				},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully submitted superfluid proposal")
	c.PropNumber = c.PropNumber + 1
}

func (s *IntegrationTestSuite) submitTextProposal(c *chain.Chain, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("submitting text proposal on %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id),
				},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully submitted text proposal")
	c.PropNumber = c.PropNumber + 1
}

func (s *IntegrationTestSuite) depositProposal(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	propStr := strconv.Itoa(c.PropNumber)

	s.T().Logf("depositing to proposal from %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "deposit", propStr, "500000000uosmo", "--from=val", fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully deposited to proposal")
}

func (s *IntegrationTestSuite) voteProposal(chainConfig *chainConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	chain := chainConfig.chain
	propStr := strconv.Itoa(chain.PropNumber)

	s.T().Logf("voting for upgrade proposal for chain-id: %s", chain.ChainMeta.Id)
	for i := range chain.Validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		s.Require().Eventually(
			func() bool {
				exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
					Context:      ctx,
					AttachStdout: true,
					AttachStderr: true,
					Container:    s.valResources[chain.ChainMeta.Id][i].Container.ID,
					User:         "root",
					Cmd: []string{
						"osmosisd", "tx", "gov", "vote", propStr, "yes", "--from=val", fmt.Sprintf("--chain-id=%s", chain.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
					},
				})
				s.Require().NoError(err)

				err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
					Context:      ctx,
					Detach:       false,
					OutputStream: &outBuf,
					ErrorStream:  &errBuf,
				})
				return strings.Contains(outBuf.String(), "code: 0")
			},
			time.Minute,
			time.Second,
			"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		s.T().Logf("successfully voted for proposal from %s container: %s", s.valResources[chain.ChainMeta.Id][i].Container.Name[1:], s.valResources[chain.ChainMeta.Id][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) voteNoProposal(c *chain.Chain, i int, from string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	propStr := strconv.Itoa(c.PropNumber)
	s.T().Logf("voting no for proposal for chain-id: %s", c.ChainMeta.Id)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][i].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "gov", "vote", propStr, "no", fmt.Sprintf("--from=%s", from), fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("successfully voted no for proposal from %s container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)

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

func (s *IntegrationTestSuite) getCurrentChainHeight(containerId string) int {
	var block syncInfo
	s.Require().Eventually(
		func() bool {
			out := s.chainStatus(containerId)
			err := json.Unmarshal(out, &block)
			if err != nil {
				return false
			}
			return true
		},
		1*time.Minute,
		time.Second,
		"Osmosis node failed to retrieve height info",
	)
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

func (s *IntegrationTestSuite) queryPropTally(endpoint, addr string) (sdk.Int, sdk.Int, sdk.Int, sdk.Int, error) {
	path := fmt.Sprintf(
		"%s/cosmos/gov/v1beta1/proposals/%s/tally",
		endpoint, addr,
	)
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("exceeded retry limit of %d with %d", retriesLeft, http.StatusServiceUnavailable)
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), err
	}

	var balancesResp govtypes.QueryTallyResultResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	noTotal := balancesResp.Tally.No
	yesTotal := balancesResp.Tally.Yes
	noWithVetoTotal := balancesResp.Tally.NoWithVeto
	abstainTotal := balancesResp.Tally.Abstain

	return noTotal, yesTotal, noWithVetoTotal, abstainTotal, nil
}

func (s *IntegrationTestSuite) createPool(c *chain.Chain, poolFile string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.T().Logf("creating pool for chain-id: %s", c.ChainMeta.Id)
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
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
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("successfully created pool from %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

}

func (s *IntegrationTestSuite) lockTokens(c *chain.Chain, i int, tokens string, duration string, from string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.T().Logf("locking %s for %s on chain-id: %s", tokens, duration, c.ChainMeta.Id)
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][i].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			s.Require().NoError(err)
			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		time.Minute,
		time.Second,
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
	c.LockNumber = c.LockNumber + 1
	s.T().Logf("successfully created lock %v from %s container: %s", c.LockNumber, s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)

}

func (s *IntegrationTestSuite) superfluidDelegate(c *chain.Chain, tokens string, valAddress string, from string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.T().Logf("superfluid delegating %s to %s on chain-id: %s", tokens, valAddress, c.ChainMeta.Id)
	lockStr := strconv.Itoa(c.LockNumber)
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][0].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			s.Require().NoError(err)
			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		5*time.Minute,
		time.Second,
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("successfully superfluid delegated from %s container: %s", s.valResources[c.ChainMeta.Id][0].Container.Name[1:], s.valResources[c.ChainMeta.Id][0].Container.ID)

}

func (s *IntegrationTestSuite) sendTx(c *chain.Chain, i int, amount string, sendAddress string, receiveAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.T().Logf("sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, c.ChainMeta.Id)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.ChainMeta.Id][i].Container.ID,
				User:         "root",
				Cmd: []string{
					"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, fmt.Sprintf("--chain-id=%s", c.ChainMeta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test",
				},
			})
			s.Require().NoError(err)
			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			return strings.Contains(outBuf.String(), "code: 0")
		},
		5*time.Minute,
		time.Second,
		"tx returned non code 0; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("successfully sent tx from %s container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)

}

func (s *IntegrationTestSuite) extractOperAddress(c *chain.Chain) {
	// var oper operInfo
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.T().Logf("extracting validator operator addresses for chain-id: %s", c.ChainMeta.Id)
	for i, val := range c.Validators {

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		s.Require().Eventually(
			func() bool {
				exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
					Context:      ctx,
					AttachStdout: true,
					AttachStderr: true,
					Container:    s.valResources[c.ChainMeta.Id][i].Container.ID,
					User:         "root",
					Cmd: []string{
						"osmosisd", "debug", "addr", val.PublicKey,
					},
				})
				s.Require().NoError(err)

				err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
					Context:      ctx,
					Detach:       false,
					OutputStream: &outBuf,
					ErrorStream:  &errBuf,
				})
				if err != nil {
					return false
				}
				return true
			},
			time.Minute,
			time.Second,
		)
		re := regexp.MustCompile("osmovaloper(.{39})")
		operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
		val.OperAddress = strings.TrimSuffix(operAddr, "\n")

	}
}

func (s *IntegrationTestSuite) queryIntermediaryAccount(c *chain.Chain, endpoint string, denom string, valAddr string) (int, error) {
	intAccount := superfluidtypes.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	path := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		endpoint, valAddr, intAccount,
	)
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return 0, fmt.Errorf("exceeded retry limit of %d with %d", retriesLeft, http.StatusServiceUnavailable)
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var stakingResp stakingtypes.QueryDelegationResponse
	if err := util.Cdc.UnmarshalJSON(bz, &stakingResp); err != nil {
		return 0, err
	}

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	s.Require().NoError(err)
	return intAccountBalance, err

}
