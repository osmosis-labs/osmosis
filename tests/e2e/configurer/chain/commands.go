package chain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	appparams "github.com/osmosis-labs/osmosis/v12/app/params"
	"github.com/osmosis-labs/osmosis/v12/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v12/tests/e2e/util"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) CreatePool(poolFile, from string) uint64 {
	n.LogActionF("creating pool from file %s", poolFile)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	path := "osmosis/gamm/v1beta1/num_pools"

	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)

	var numPools gammtypes.QueryNumPoolsResponse
	err = util.Cdc.UnmarshalJSON(bz, &numPools)
	require.NoError(n.t, err)
	poolID := numPools.NumPools
	n.LogActionF("successfully created pool %d", poolID)
	return poolID
}

// SwapExactAmountIn swaps tokenInCoin to get at least tokenOutMinAmountInt of the other token's pool out.
// swapRoutePoolIds is the comma separated list of pool ids to swap through.
// swapRouteDenoms is the comma separated list of denoms to swap through.
// To reproduce locally:
// docker container exec <container id> osmosisd tx gamm swap-exact-amount-in <tokeinInCoin> <tokenOutMinAmountInt> --swap-route-pool-ids <swapRoutePoolIds> --swap-route-denoms <swapRouteDenoms> --chain-id=<id>--from=<address> --keyring-backend=test -b=block --yes --log_format=json
func (n *NodeConfig) SwapExactAmountIn(tokenInCoin, tokenOutMinAmountInt string, swapRoutePoolIds string, swapRouteDenoms string, from string) {
	n.LogActionF("swapping %s to get a minimum of %s with pool id routes (%s) and denom routes (%s)", tokenInCoin, tokenOutMinAmountInt, swapRoutePoolIds, swapRouteDenoms)
	cmd := []string{"osmosisd", "tx", "gamm", "swap-exact-amount-in", tokenInCoin, tokenOutMinAmountInt, fmt.Sprintf("--swap-route-pool-ids=%s", swapRoutePoolIds), fmt.Sprintf("--swap-route-denoms=%s", swapRouteDenoms), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully swapped")
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64, initialDeposit sdk.Coin) {
	n.LogActionF("submitting upgrade proposal %s for height %d", upgradeVersion, upgradeHeight)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%d", upgradeHeight), "--upgrade-info=\"\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted upgrade proposal")
}

func (n *NodeConfig) SubmitSuperfluidProposal(asset string, initialDeposit sdk.Coin) {
	n.LogActionF("submitting superfluid proposal for asset %s", asset)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted superfluid proposal for asset %s", asset)
}

func (n *NodeConfig) SubmitTextProposal(text string, initialDeposit sdk.Coin, isExpedited bool) {
	n.LogActionF("submitting text gov proposal")
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	if isExpedited {
		cmd = append(cmd, "--is-expedited=true")
	}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted text gov proposal")
}

func (n *NodeConfig) DepositProposal(proposalNumber int, isExpedited bool) {
	n.LogActionF("depositing on proposal: %d", proposalNumber)
	deposit := sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.MinDepositValue)).String()
	if isExpedited {
		deposit = sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.MinExpeditedDepositValue)).String()
	}
	cmd := []string{"osmosisd", "tx", "gov", "deposit", fmt.Sprintf("%d", proposalNumber), deposit, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully deposited on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteYesProposal(from string, proposalNumber int) {
	n.LogActionF("voting yes on proposal: %d", proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "yes", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted yes on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteNoProposal(from string, proposalNumber int) {
	n.LogActionF("voting no on proposal: %d", proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "no", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted no on proposal: %d", proposalNumber)
}

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) {
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created lock")
}

func (n *NodeConfig) AddToExistingLock(tokens sdk.Int, denom, duration, from string) {
	n.LogActionF("retrieving existing lock ID")
	durationPath := fmt.Sprintf("/osmosis/lockup/v1beta1/account_locked_longer_duration/%s?duration=%s", from, duration)
	bz, err := n.QueryGRPCGateway(durationPath)
	require.NoError(n.t, err)
	var accountLockedDurationResp lockuptypes.AccountLockedDurationResponse
	err = util.Cdc.UnmarshalJSON(bz, &accountLockedDurationResp)
	require.NoError(n.t, err)
	var lockID string
	for _, periodLock := range accountLockedDurationResp.Locks {
		if periodLock.Coins.AmountOf(denom).GT(sdk.ZeroInt()) {
			lockID = fmt.Sprintf("%v", periodLock.ID)
			break
		}
	}
	n.LogActionF("noting previous lockup amount")
	path := fmt.Sprintf("/osmosis/lockup/v1beta1/locked_by_id/%s", lockID)
	bz, err = n.QueryGRPCGateway(path)
	require.NoError(n.t, err)
	var lockedResp lockuptypes.LockedResponse
	err = util.Cdc.UnmarshalJSON(bz, &lockedResp)
	require.NoError(n.t, err)
	previousLockupAmount := lockedResp.Lock.Coins.AmountOf(denom)
	n.LogActionF("previous lockup amount is %v", previousLockupAmount)
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", fmt.Sprintf("%s%s", tokens, denom), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("noting new lockup amount")
	bz, err = n.QueryGRPCGateway(path)
	require.NoError(n.t, err)
	err = util.Cdc.UnmarshalJSON(bz, &lockedResp)
	require.NoError(n.t, err)
	newLockupAmount := lockedResp.Lock.Coins.AmountOf(denom)
	n.LogActionF("new lockup amount is %v", newLockupAmount)
	lockupDelta := newLockupAmount.Sub(previousLockupAmount)
	require.True(n.t, lockupDelta.Equal(tokens))
	n.LogActionF("successfully added to existing lock")
}

func (n *NodeConfig) SuperfluidDelegate(lockNumber int, valAddress string, from string) {
	lockStr := strconv.Itoa(lockNumber)
	n.LogActionF("superfluid delegating lock %s to %s", lockStr, valAddress)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully superfluid delegated lock %s to %s", lockStr, valAddress)
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.LogActionF("bank sending %s from address %s to %s", amount, sendAddress, receiveAddress)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully sent bank sent %s from address %s to %s", amount, sendAddress, receiveAddress)
}

func (n *NodeConfig) CreateWallet(walletName string) string {
	n.LogActionF("creating wallet %s", walletName)
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("created wallet %s, waller address - %s", walletName, walletAddr)
	return walletAddr
}

func (n *NodeConfig) GetWallet(walletName string) string {
	n.LogActionF("retrieving wallet %s", walletName)
	cmd := []string{"osmosisd", "keys", "show", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("wallet %s found, waller address - %s", walletName, walletAddr)
	return walletAddr
}

func (n *NodeConfig) QueryPropStatusTimed(proposalNumber int, desiredStatus string, totalTime chan time.Duration) {
	start := time.Now()
	require.Eventually(
		n.t,
		func() bool {
			status, err := n.QueryPropStatus(proposalNumber)
			if err != nil {
				return false
			}

			return status == desiredStatus
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve prop tally",
	)
	elapsed := time.Since(start)
	totalTime <- elapsed
}
