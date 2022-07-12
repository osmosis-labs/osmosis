package chain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

func (c *Config) CreatePool(poolFile, from string) {
	c.t.Logf("creating pool for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--chain-id=%s", c.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	require.NoError(c.t, err)

	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)
	c.t.Logf("successfully created pool from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (c *Config) SubmitUpgradeProposal(upgradeVersion string) {
	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)

	upgradeHeightStr := strconv.Itoa(c.PropHeight)
	c.t.Logf("submitting upgrade proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	require.NoError(c.t, err)
	c.t.Log("successfully submitted upgrade proposal")
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (c *Config) SubmitSuperfluidProposal(asset string) {
	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)

	c.t.Logf("submitting superfluid proposal for asset %s on %s container: %s", asset, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.Id)}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	require.NoError(c.t, err)
	c.t.Log("successfully submitted superfluid proposal")
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (c *Config) SubmitTextProposal(text string) {
	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)

	c.t.Logf("submitting text proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.Id)}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	c.t.Log("successfully submitted text proposal")
	require.NoError(c.t, err)
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (c *Config) DepositProposal() {
	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)

	propStr := strconv.Itoa(c.LatestProposalNumber)
	c.t.Logf("depositing to proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", propStr, "500000000uosmo", "--from=val", fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	require.NoError(c.t, err)
	c.t.Log("successfully deposited to proposal")
}

func (c *Config) VoteYesProposal() {
	propStr := strconv.Itoa(c.LatestProposalNumber)
	c.t.Logf("voting yes on proposal for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "yes", "--from=val", fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	for i := range c.NodeConfigs {
		_, _, err := c.containerManager.ExecCmd(c.t, c.Id, i, cmd, "code: 0")
		require.NoError(c.t, err)

		validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, i)
		require.True(c.t, exists)
		c.t.Logf("successfully voted yes on proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}
}

func (c *Config) VoteNoProposal(validatorIdx int, from string) {
	propStr := strconv.Itoa(c.LatestProposalNumber)
	c.t.Logf("voting no on proposal for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "no", fmt.Sprintf("--from=%s", from), fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIdx, cmd, "code: 0")
	require.NoError(c.t, err)

	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, validatorIdx)
	require.True(c.t, exists)
	c.t.Logf("successfully voted no for proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (c *Config) LockTokens(validatorIdx int, tokens string, duration string, from string) {
	c.t.Logf("locking %s for %s on chain-id: %s", tokens, duration, c.Id)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--chain-id=%s", c.Id), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIdx, cmd, "code: 0")
	require.NoError(c.t, err)

	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, validatorIdx)
	require.True(c.t, exists)
	c.t.Logf("successfully created lock %v from %s container: %s", c.LatestLockNumber, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	c.LatestLockNumber = c.LatestLockNumber + 1
}

func (c *Config) SuperfluidDelegate(valAddress string, from string) {
	lockStr := strconv.Itoa(c.LatestLockNumber)
	c.t.Logf("superfluid delegating lock %s to %s on chain-id: %s", lockStr, valAddress, c.Id)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--chain-id=%s", c.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, 0, cmd, "code: 0")
	require.NoError(c.t, err)

	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)
	c.t.Logf("successfully superfluid delegated from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (c *Config) BankSend(validatorIndex int, amount string, sendAddress string, receiveAddress string) {
	c.t.Logf("sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, c.Id)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, fmt.Sprintf("--chain-id=%s", c.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIndex, cmd, "code: 0")
	require.NoError(c.t, err)

	validatorResource, exists := c.containerManager.GetValidatorResource(c.Id, 0)
	require.True(c.t, exists)
	c.t.Logf("successfully sent tx from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (c *Config) CreateWallet(validatorIndex int, walletName string) string {
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIndex, cmd, "")
	require.NoError(c.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	return walletAddr
}

func (c *Config) SendIBC(dstChain *Config, recipient string, token sdk.Coin) {
	c.t.Logf("sending %s from %s to %s (%s)", token, c.Id, dstChain.Id, recipient)
	balancesBPre, err := dstChain.QueryBalances(0, recipient)
	require.NoError(c.t, err)

	cmd := []string{"hermes", "tx", "raw", "ft-transfer", dstChain.Id, c.Id, "transfer", "channel-0", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	_, _, err = c.containerManager.ExecCmd(c.t, "", 0, cmd, "Success")
	require.NoError(c.t, err)

	require.Eventually(
		c.t,
		func() bool {
			balancesBPost, err := dstChain.QueryBalances(0, recipient)
			require.NoError(c.t, err)
			ibcCoin := balancesBPost.Sub(balancesBPre)
			if ibcCoin.Len() == 1 {
				tokenPre := balancesBPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesBPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := initialization.OsmoToken.Amount
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

	c.t.Log("successfully sent IBC tokens")
}

func (c *Config) ExtractValidatorOperatorAddress(nodeIndex int) error {
	node := c.NodeConfigs[nodeIndex]

	if !node.IsValidator {
		return fmt.Errorf("node %s at index %d is not a validator", node.Name, nodeIndex)
	}

	cmd := []string{"osmosisd", "debug", "addr", node.PublicKey}
	c.t.Logf("extracting validator operator addresses for chain-id: %s", c.Id)
	_, errBuf, err := c.containerManager.ExecCmd(c.t, c.Id, nodeIndex, cmd, "")
	require.NoError(c.t, err)
	re := regexp.MustCompile("osmovaloper(.{39})")
	operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
	c.NodeConfigs[nodeIndex].OperatorAddress = strings.TrimSuffix(operAddr, "\n")
	return nil
}
