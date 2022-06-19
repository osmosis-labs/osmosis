package configurer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
)

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func (bc *baseConfigurer) CreatePool(c *chain.Config, poolFile string) {
	bc.t.Logf("creating pool for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--chain-id=%s", c.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	require.NoError(bc.t, err)

	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)
	bc.t.Logf("successfully created pool from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (bc *baseConfigurer) SendIBC(srcChain *chain.Config, dstChain *chain.Config, recipient string, token sdk.Coin) {
	cmd := []string{"hermes", "tx", "raw", "ft-transfer", dstChain.Id, srcChain.Id, "transfer", "channel-0", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, "", 0, cmd, "Success")
	require.NoError(bc.t, err)

	bc.t.Logf("sending %s from %s to %s (%s)", token, srcChain.Id, dstChain.Id, recipient)
	balancesBPre, err := bc.QueryBalances(dstChain, 0, recipient)
	require.NoError(bc.t, err)

	require.Eventually(
		bc.t,
		func() bool {
			balancesBPost, err := bc.QueryBalances(dstChain, 0, recipient)
			require.NoError(bc.t, err)
			ibcCoin := balancesBPost.Sub(balancesBPre)
			if ibcCoin.Len() == 1 {
				tokenPre := balancesBPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesBPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := chaininit.OsmoToken.Amount
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

	bc.t.Log("successfully sent IBC tokens")
}

func (bc *baseConfigurer) SubmitUpgradeProposal(c *chain.Config) {
	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)

	upgradeHeightStr := strconv.Itoa(c.PropHeight)
	bc.t.Logf("submitting upgrade proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", UpgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", UpgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%s", upgradeHeightStr), "--upgrade-info=\"\"", fmt.Sprintf("--chain-id=%s", c.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	require.NoError(bc.t, err)
	bc.t.Log("successfully submitted upgrade proposal")
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (bc *baseConfigurer) SubmitSuperfluidProposal(c *chain.Config, asset string) {
	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)

	bc.t.Logf("submitting superfluid proposal for asset %s on %s container: %s", asset, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.Id)}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	require.NoError(bc.t, err)
	bc.t.Log("successfully submitted superfluid proposal")
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (bc *baseConfigurer) SubmitTextProposal(c *chain.Config, text string) {
	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)

	bc.t.Logf("submitting text proposal on %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", "-b=block", "--yes", "--keyring-backend=test", "--log_format=json", fmt.Sprintf("--chain-id=%s", c.Id)}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	bc.t.Log("successfully submitted text proposal")
	require.NoError(bc.t, err)
	c.LatestProposalNumber = c.LatestProposalNumber + 1
}

func (bc *baseConfigurer) DepositProposal(c *chain.Config) {
	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)

	propStr := strconv.Itoa(c.LatestProposalNumber)
	bc.t.Logf("depositing to proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", propStr, "500000000uosmo", "--from=val", fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	require.NoError(bc.t, err)
	bc.t.Log("successfully deposited to proposal")
}

func (bc *baseConfigurer) VoteYesProposal(c *chain.Config) {
	propStr := strconv.Itoa(c.LatestProposalNumber)
	bc.t.Logf("voting yes on proposal for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "yes", "--from=val", fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	for i := range c.ValidatorConfigs {
		_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, i, cmd, "code: 0")
		require.NoError(bc.t, err)

		validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, i)
		require.True(bc.t, exists)
		bc.t.Logf("successfully voted yes on proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}
}

func (bc *baseConfigurer) VoteNoProposal(c *chain.Config, validatorIdx int, from string) {
	propStr := strconv.Itoa(c.LatestProposalNumber)
	bc.t.Logf("voting no on proposal for chain-id: %s", c.Id)
	cmd := []string{"osmosisd", "tx", "gov", "vote", propStr, "no", fmt.Sprintf("--from=%s", from), fmt.Sprintf("--chain-id=%s", c.Id), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, validatorIdx, cmd, "code: 0")
	require.NoError(bc.t, err)

	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, validatorIdx)
	require.True(bc.t, exists)
	bc.t.Logf("successfully voted no for proposal from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (bc *baseConfigurer) LockTokens(c *chain.Config, validatorIdx int, tokens string, duration string, from string) {
	bc.t.Logf("locking %s for %s on chain-id: %s", tokens, duration, c.Id)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--chain-id=%s", c.Id), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, validatorIdx, cmd, "code: 0")
	require.NoError(bc.t, err)

	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, validatorIdx)
	require.True(bc.t, exists)
	bc.t.Logf("successfully created lock %v from %s container: %s", c.LatestLockNumber, validatorResource.Container.Name[1:], validatorResource.Container.ID)
	c.LatestLockNumber = c.LatestLockNumber + 1
}

func (bc *baseConfigurer) SuperfluidDelegate(c *chain.Config, valAddress string, from string) {
	lockStr := strconv.Itoa(c.LatestLockNumber)
	bc.t.Logf("superfluid delegating lock %s to %s on chain-id: %s", lockStr, valAddress, c.Id)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--chain-id=%s", c.Id), fmt.Sprintf("--from=%s", from), "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, 0, cmd, "code: 0")
	require.NoError(bc.t, err)

	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)
	bc.t.Logf("successfully superfluid delegated from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (bc *baseConfigurer) BankSend(c *chain.Config, i int, amount string, sendAddress string, receiveAddress string) {
	bc.t.Logf("sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, c.Id)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, fmt.Sprintf("--chain-id=%s", c.Id), "--from=val", "-b=block", "--yes", "--keyring-backend=test"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, i, cmd, "code: 0")
	require.NoError(bc.t, err)

	validatorResource, exists := bc.containerManager.GetValidatorResource(c.Id, 0)
	require.True(bc.t, exists)
	bc.t.Logf("successfully sent tx from %s container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
}

func (bc *baseConfigurer) CreateWallet(c *chain.Config, index int, walletName string) string {
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := bc.containerManager.ExecCmd(bc.t, c.Id, index, cmd, "")
	require.NoError(bc.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	return walletAddr
}
