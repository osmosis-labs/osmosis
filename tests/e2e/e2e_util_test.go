package e2e

import (
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
	"github.com/stretchr/testify/require"

	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

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
	_, _, err := s.containerManager.ExecCmd(s.T(), "", 0, cmd, "successfully opened init channel")
	s.Require().NoError(err)
	s.T().Logf("connected %s and %s chains via IBC", chainA.meta.Id, chainB.meta.Id)
}

func (s *IntegrationTestSuite) sendIBC(srcChain *chainConfig, dstChain *chainConfig, recipient string, token sdk.Coin) {
	cmd := []string{"hermes", "tx", "raw", "ft-transfer", dstChain.meta.Id, srcChain.meta.Id, "transfer", "channel-0", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	_, _, err := s.containerManager.ExecCmd(s.T(), "", 0, cmd, "Success")
	s.Require().NoError(err)

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
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("submitting upgrade proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.meta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
	s.T().Log("successfully submitted upgrade proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) submitSuperfluidProposal(c *chainConfig, asset string) {
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("submitting superfluid proposal for asset %s on %s container: %s", asset, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.meta.Id)}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
	s.T().Log("successfully submitted superfluid proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) submitTextProposal(c *chainConfig, text string) {
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("submitting text proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.meta.Id)}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
	s.T().Log("successfully submitted text proposal")
	c.latestProposalNumber = c.latestProposalNumber + 1
}

func (s *IntegrationTestSuite) depositProposal(c *chainConfig) {
	propStr := strconv.Itoa(c.latestProposalNumber)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("depositing to proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", propStr, "500000000uosmo", "--from=val", fmt.Sprintf("--chain-id=%s", c.meta.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
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
		_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, i, cmd, "code: 0")
		s.Require().NoError(err)
		validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, i)
		require.True(s.T(), exists)
		s.T().Logf("successfully voted yes on proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}
}

func (s *IntegrationTestSuite) voteNoProposal(c *chainConfig, validatorIndex int, from string) {
	propStr := strconv.Itoa(c.latestProposalNumber)
	s.T().Logf("voting no on proposal for chain-id: %s", c.meta.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "no", fmt.Sprintf("--from=%s", from), fmt.Sprintf("--chain-id=%s", c.meta.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, validatorIndex, cmd, "code: 0")
	s.Require().NoError(err)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, validatorIndex)
	require.True(s.T(), exists)
	s.T().Logf("successfully voted no for proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (s *IntegrationTestSuite) chainStatus(c *chainConfig, i int) []byte {
	cmd := []string{"osmosisd", "status"}
	_, errBuf, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, i, cmd, "")
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
	outBuf, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, i, cmd, "")
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
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("successfully created pool from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (s *IntegrationTestSuite) lockTokens(c *chainConfig, validatorIndex int, tokens string, duration string, from string) {
	s.T().Logf("locking %s for %s on chain-id: %s", tokens, duration, c.meta.Id)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--chain-id=%s", c.meta.Id), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, validatorIndex, cmd, "code: 0")
	s.Require().NoError(err)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, validatorIndex)
	require.True(s.T(), exists)
	s.T().Logf("successfully created lock %v from %s container: %s", c.latestLockNumber, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	c.latestLockNumber = c.latestLockNumber + 1
}

func (s *IntegrationTestSuite) superfluidDelegate(c *chainConfig, valAddress string, from string) {
	lockStr := strconv.Itoa(c.latestLockNumber)
	s.T().Logf("superfluid delegating lock %s to %s on chain-id: %s", lockStr, valAddress, c.meta.Id)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--chain-id=%s", c.meta.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, 0, cmd, "code: 0")
	s.Require().NoError(err)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, 0)
	require.True(s.T(), exists)
	s.T().Logf("successfully superfluid delegated from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (s *IntegrationTestSuite) sendTx(c *chainConfig, validatorIndex int, amount string, sendAddress string, receiveAddress string) {
	s.T().Logf("sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, c.meta.Id)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, fmt.Sprintf("--chain-id=%s", c.meta.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, validatorIndex, cmd, "code: 0")
	s.Require().NoError(err)
	validatorResource, exists := s.containerManager.GetValidatorResource(c.meta.Id, validatorIndex)
	require.True(s.T(), exists)
	s.T().Logf("successfully sent tx from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (s *IntegrationTestSuite) extractValidatorOperatorAddresses(config *chainConfig) {
	for i, val := range config.validators {
		if _, ok := config.skipRunValidatorIndexes[i]; ok {
			s.T().Logf("skipping %s validator with index %d from running...", val.validator.Name, i)
			continue
		}
		cmd := []string{"osmosisd", "debug", "addr", val.validator.PublicKey}
		s.T().Logf("extracting validator operator addresses for chain-id: %s", config.meta.Id)
		_, errBuf, err := s.containerManager.ExecCmd(s.T(), config.meta.Id, i, cmd, "")
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
	outBuf, _, err := s.containerManager.ExecCmd(s.T(), c.meta.Id, index, cmd, "")
	s.Require().NoError(err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	return walletAddr
}
