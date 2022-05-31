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

func (s *IntegrationTestSuite) ExecTx(chainId string, validatorIndex int, command []string, success string) (bytes.Buffer, bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var containerId string
	if chainId == "" {
		containerId = s.hermesResource.Container.ID
	} else {
		containerId = s.valResources[chainId][validatorIndex].Container.ID
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
				Container:    containerId,
				User:         "root",
				Cmd:          command,
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

			if err != nil {
				return false
			}

			if success != "" {
				return strings.Contains(outBuf.String(), success) || strings.Contains(errBuf.String(), success)
			}

			return true
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	return outBuf, errBuf, nil
}

func (s *IntegrationTestSuite) ExecQueryRPC(path string) ([]byte, error) {
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return nil, err
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (s *IntegrationTestSuite) connectIBCChains(chainA *chainConfig, chainB *chainConfig) {
	s.T().Logf("connecting %s and %s chains via IBC", chainA.meta.Id, chainB.meta.Id)
	cmd := []string{"hermes", "create", "channel", chainA.meta.Id, chainB.meta.Id, "--port-a=transfer", "--port-b=transfer"}
	s.ExecTx("", 0, cmd, "successfully opened init channel")
	s.T().Logf("connected %s and %s chains via IBC", chainA.meta.Id, chainB.meta.Id)
}

func (s *IntegrationTestSuite) sendIBC(srcChain *chainConfig, dstChain *chainConfig, recipient string, token sdk.Coin) {
	cmd := []string{"hermes", "tx", "raw", "ft-transfer", dstChain.meta.Id, srcChain.meta.Id, "transfer", "channel-0", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	s.ExecTx("", 0, cmd, "Success")

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChain.meta.Id, dstChain.meta.Id, recipient)
	balancesBPre, err := s.queryBalances(dstChain, 0, recipient)
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			balancesBPost, err := s.queryBalances(dstChain, 0, recipient)
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

func (s *IntegrationTestSuite) submitUpgradeProposal(c *chainConfig) {
	upgradeHeightStr := strconv.Itoa(c.propHeight)
	s.T().Logf("submitting upgrade proposal on %s container: %s", s.valResources[c.meta.Id][0].Container.Name[1:], s.valResources[c.meta.Id][0].Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.meta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json"}
	s.ExecTx(c.meta.Id, 0, cmd, "code: 0")
	s.T().Log("successfully submitted upgrade proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) submitSuperfluidProposal(c *chainConfig, asset string) {
	s.T().Logf("submitting superfluid proposal for asset %s on %s container: %s", asset, s.valResources[c.meta.Id][0].Container.Name[1:], s.valResources[c.meta.Id][0].Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.meta.Id)}
	s.ExecTx(c.meta.Id, 0, cmd, "code: 0")
	s.T().Log("successfully submitted superfluid proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) submitTextProposal(c *chainConfig, text string) {
	s.T().Logf("submitting text proposal on %s container: %s", s.valResources[c.meta.Id][0].Container.Name[1:], s.valResources[c.meta.Id][0].Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.meta.Id)}
	s.ExecTx(c.meta.Id, 0, cmd, "code: 0")
	s.T().Log("successfully submitted text proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) depositProposal(c *chainConfig) {
	propStr := strconv.Itoa(c.latestProposalNumber)
	s.T().Logf("depositing to proposal from %s container: %s", s.valResources[c.meta.Id][0].Container.Name[1:], s.valResources[c.meta.Id][0].Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", propStr, "500000000uosmo", "--from=val", fmt.Sprintf("--chain-id=%s", c.meta.Id), "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(c.meta.Id, 0, cmd, "code: 0")
	s.T().Log("successfully deposited to proposal")
}

func (s *IntegrationTestSuite) voteProposal(c *chainConfig) {
	propStr := strconv.Itoa(c.latestProposalNumber)
	s.T().Logf("voting yes on proposal for chain-id: %s", c.meta.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "yes", "--from=val", fmt.Sprintf("--chain-id=%s", c.meta.Id), "-b=block", "--yes", "--keyring-backend=test"}
	for i := range c.validators {
		if _, ok := c.skipRunValidatorIndexes[i]; ok {
			continue
		}
		s.ExecTx(c.meta.Id, i, cmd, "code: 0")
		s.T().Logf("successfully voted yes on proposal from %s container: %s", s.valResources[c.meta.Id][i].Container.Name[1:], s.valResources[c.meta.Id][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) voteNoProposal(c *chainConfig, i int, from string) {
	propStr := strconv.Itoa(c.latestProposalNumber)
	s.T().Logf("voting no on proposal for chain-id: %s", c.meta.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "no", fmt.Sprintf("--from=%s", from), fmt.Sprintf("--chain-id=%s", c.meta.Id), "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(c.meta.Id, i, cmd, "code: 0")
	s.T().Logf("successfully voted no for proposal from %s container: %s", s.valResources[c.meta.Id][i].Container.Name[1:], s.valResources[c.meta.Id][i].Container.ID)
}

func (s *IntegrationTestSuite) chainStatus(c *chainConfig, i int) []byte {
	cmd := []string{"osmosisd", "status"}
	_, errBuf, err := s.ExecTx(c.meta.Id, i, cmd, "")
	s.Require().NoError(err)
	return errBuf.Bytes()

}

func (s *IntegrationTestSuite) getCurrentChainHeight(c *chainConfig, i int) int {
	var block syncInfo
	s.Require().Eventually(
		func() bool {
			out := s.chainStatus(c, i)
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

func (s *IntegrationTestSuite) queryBalances(c *chainConfig, i int, addr string) (sdk.Coins, error) {
	cmd := []string{"osmosisd", "query", "bank", "balances", addr, "--output=json"}
	outBuf, _, err := s.ExecTx(c.meta.Id, i, cmd, "")
	s.Require().NoError(err)

	var balancesResp banktypes.QueryAllBalancesResponse
	err = util.Cdc.UnmarshalJSON(outBuf.Bytes(), &balancesResp)
	s.Require().NoError(err)

	return balancesResp.GetBalances(), nil
}

func (s *IntegrationTestSuite) queryPropTally(endpoint, addr string) (sdk.Int, sdk.Int, sdk.Int, sdk.Int, error) {
	path := fmt.Sprintf(
		"%s/cosmos/gov/v1beta1/proposals/%s/tally",
		endpoint, addr,
	)
	bz, err := s.ExecQueryRPC(path)
	s.Require().NoError(err)

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

func (s *IntegrationTestSuite) createPool(c *chainConfig, poolFile string) {
	s.T().Logf("creating pool for chain-id: %s", c.meta.Id)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--chain-id=%s", c.meta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(c.meta.Id, 0, cmd, "code: 0")
	s.T().Logf("successfully created pool from %s container: %s", s.valResources[c.meta.Id][0].Container.Name[1:], s.valResources[c.meta.Id][0].Container.ID)
}

func (s *IntegrationTestSuite) lockTokens(config *chainConfig, i int, tokens string, duration string, from string) {
	s.T().Logf("locking %s for %s on chain-id: %s", tokens, duration, config.meta.Id)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--chain-id=%s", config.meta.Id), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(config.meta.Id, i, cmd, "code: 0")
	s.T().Logf("successfully created lock %v from %s container: %s", config.latestLockNumber, s.valResources[config.meta.Id][i].Container.Name[1:], s.valResources[config.meta.Id][i].Container.ID)
	config.latestLockNumber = config.latestLockNumber + 1

}

func (s *IntegrationTestSuite) superfluidDelegate(config *chainConfig, valAddress string, from string) {
	lockStr := strconv.Itoa(config.latestLockNumber)
	s.T().Logf("superfluid delegating lock %s to %s on chain-id: %s", lockStr, valAddress, config.meta.Id)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--chain-id=%s", config.meta.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(config.meta.Id, 0, cmd, "code: 0")
	s.T().Logf("successfully superfluid delegated from %s container: %s", s.valResources[config.meta.Id][0].Container.Name[1:], s.valResources[config.meta.Id][0].Container.ID)

}

func (s *IntegrationTestSuite) sendTx(c *chainConfig, i int, amount string, sendAddress string, receiveAddress string) {
	s.T().Logf("sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, c.meta.Id)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, fmt.Sprintf("--chain-id=%s", c.meta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	s.ExecTx(c.meta.Id, i, cmd, "code: 0")
	s.T().Logf("successfully sent tx from %s container: %s", s.valResources[c.meta.Id][i].Container.Name[1:], s.valResources[c.meta.Id][i].Container.ID)

}

func (s *IntegrationTestSuite) extractValidatorOperatorAddresses(config *chainConfig) {

	for i, val := range config.validators {
		if _, ok := config.skipRunValidatorIndexes[i]; ok {
			s.T().Logf("skipping %s validator with index %d from running...", val.validator.Name, i)
			continue
		}
		cmd := []string{"osmosisd", "debug", "addr", val.validator.PublicKey}
		s.T().Logf("extracting validator operator addresses for chain-id: %s", config.meta.Id)
		_, errBuf, err := s.ExecTx(config.meta.Id, i, cmd, "")
		s.Require().NoError(err)
		re := regexp.MustCompile("osmovaloper(.{39})")
		operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
		config.validators[i].operatorAddress = strings.TrimSuffix(operAddr, "\n")
	}
}

func (s *IntegrationTestSuite) queryIntermediaryAccount(c *chainConfig, endpoint string, denom string, valAddr string) (int, error) {
	intAccount := superfluidtypes.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	path := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		endpoint, valAddr, intAccount,
	)
	bz, err := s.ExecQueryRPC(path)
	s.Require().NoError(err)

	var stakingResp stakingtypes.QueryDelegationResponse
	err = util.Cdc.UnmarshalJSON(bz, &stakingResp)
	s.Require().NoError(err)

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	s.Require().NoError(err)
	return intAccountBalance, err

}

func (s *IntegrationTestSuite) createWallet(c *chainConfig, index int, walletName string) string {
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := s.ExecTx(c.meta.Id, index, cmd, "")
	s.Require().NoError(err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	return walletAddr
}
