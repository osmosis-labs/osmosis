package chain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/bytes"

	appparams "github.com/osmosis-labs/osmosis/v17/app/params"
	"github.com/osmosis-labs/osmosis/v17/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v17/tests/e2e/initialization"
	"github.com/osmosis-labs/osmosis/v17/tests/e2e/util"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v17/x/ibc-rate-limit/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/p2p"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	app "github.com/osmosis-labs/osmosis/v17/app"

	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
)

// The value is returned as a string, so we have to unmarshal twice
type params struct {
	Key      string `json:"key"`
	Subspace string `json:"subspace"`
	Value    string `json:"value"`
}

func (n *NodeConfig) CreateBalancerPool(poolFile, from string) uint64 {
	n.LogActionF("creating balancer pool from file %s", poolFile)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000uosmo"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	// TODO: create a helper function for parsing pool ID and prop ID from the response
	startIndex := strings.Index(resp.String(), `{"key":"pool_id","value":"`) + len(`{"key":"pool_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	codeIdStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	poolID, _ := strconv.ParseUint(codeIdStr, 10, 64)

	n.LogActionF("successfully created balancer pool %d", poolID)
	return poolID
}

func (n *NodeConfig) CreateStableswapPool(poolFile, from string) uint64 {
	n.LogActionF("creating stableswap pool from file %s", poolFile)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), "--pool-type=stableswap", fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000uosmo"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	startIndex := strings.Index(resp.String(), `{"key":"pool_id","value":"`) + len(`{"key":"pool_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	codeIdStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	poolID, _ := strconv.ParseUint(codeIdStr, 10, 64)

	n.LogActionF("successfully created stableswap pool with ID %d", poolID)
	return poolID
}

// CollectSpreadRewards collects spread rewards earned by concentrated position in range of [lowerTick; upperTick] in pool with id of poolId
func (n *NodeConfig) CollectSpreadRewards(from, positionIds string) {
	n.LogActionF("collecting spread rewards from concentrated position")
	cmd := []string{"osmosisd", "tx", "concentratedliquidity", "collect-spread-rewards", positionIds, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	n.LogActionF("successfully collected spread rewards for account %s", from)
}

// CreateConcentratedPool creates a concentrated pool.
// Returns pool id of newly created pool on success
func (n *NodeConfig) CreateConcentratedPool(from, denom1, denom2 string, tickSpacing uint64, spreadFactor string) (uint64, error) {
	n.LogActionF("creating concentrated pool")

	cmd := []string{"osmosisd", "tx", "concentratedliquidity", "create-pool", denom1, denom2, fmt.Sprintf("%d", tickSpacing), spreadFactor, fmt.Sprintf("--from=%s", from)}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	if err != nil {
		return 0, err
	}

	startIndex := strings.Index(resp.String(), `{"key":"pool_id","value":"`) + len(`{"key":"pool_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	codeIdStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	poolID, _ := strconv.ParseUint(codeIdStr, 10, 64)

	n.LogActionF("successfully created concentrated pool with ID %d", poolID)
	return poolID, nil
}

// CreateConcentratedPosition creates a concentrated position from [lowerTick; upperTick] in pool with id of poolId
// token{0,1} - liquidity to create position with
func (n *NodeConfig) CreateConcentratedPosition(from, lowerTick, upperTick string, tokens string, token0MinAmt, token1MinAmt int64, poolId uint64) uint64 {
	n.LogActionF("creating concentrated position")
	// gas = 50,000 because e2e  default to 40,000, we hardcoded extra 10k gas to initialize tick
	// fees = 1250 (because 50,000 * 0.0025 = 1250)
	cmd := []string{"osmosisd", "tx", "concentratedliquidity", "create-position", fmt.Sprint(poolId), lowerTick, upperTick, tokens, fmt.Sprintf("%d", token0MinAmt), fmt.Sprintf("%d", token1MinAmt), fmt.Sprintf("--from=%s", from), "--gas=500000", "--fees=1250uosmo", "-o json"}
	outJson, _, err := n.containerManager.ExecTxCmdWithSuccessString(n.t, n.chainId, n.Name, cmd, "code\":0")
	require.NoError(n.t, err)

	var txResponse map[string]interface{}
	err = json.Unmarshal(outJson.Bytes(), &txResponse)
	require.NoError(n.t, err)

	positionIDString, err := GetPositionID(txResponse)
	require.NoError(n.t, err)

	positionID, err := strconv.ParseUint(positionIDString, 10, 64)
	require.NoError(n.t, err)

	n.LogActionF("successfully created concentrated position from %s to %s", lowerTick, upperTick)

	return positionID
}

func (n *NodeConfig) StoreWasmCode(wasmFile, from string) int {
	n.LogActionF("storing wasm code from file %s", wasmFile)
	cmd := []string{"osmosisd", "tx", "wasm", "store", wasmFile, fmt.Sprintf("--from=%s", from), "--gas=auto", "--gas-prices=0.1uosmo", "--gas-adjustment=1.3"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	startIndex := strings.Index(resp.String(), `{"key":"code_id","value":"`) + len(`{"key":"code_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	codeIdStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	codeId, _ := strconv.Atoi(codeIdStr)
	n.LogActionF("successfully stored")
	return codeId
}

func (n *NodeConfig) WithdrawPosition(from, liquidityOut string, positionId uint64) {
	n.LogActionF("withdrawing liquidity from position")
	cmd := []string{"osmosisd", "tx", "concentratedliquidity", "withdraw-position", fmt.Sprint(positionId), liquidityOut, fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000uosmo"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully withdrew %s liquidity from position %d", liquidityOut, positionId)
}

func (n *NodeConfig) InstantiateWasmContract(codeId, initMsg, from string) {
	n.LogActionF("instantiating wasm contract %s with %s", codeId, initMsg)
	cmd := []string{"osmosisd", "tx", "wasm", "instantiate", codeId, initMsg, fmt.Sprintf("--from=%s", from), "--no-admin", "--label=contract"}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully initialized")
}

func (n *NodeConfig) WasmExecute(contract, execMsg, from string) {
	n.LogActionF("executing %s on wasm contract %s from %s", execMsg, contract, from)
	cmd := []string{"osmosisd", "tx", "wasm", "execute", contract, execMsg, fmt.Sprintf("--from=%s", from)}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully executed")
}

// QueryParams extracts the params for a given subspace and key. This is done generically via json to avoid having to
// specify the QueryParamResponse type (which may not exist for all params).
func (n *NodeConfig) QueryParams(subspace, key string) string {
	cmd := []string{"osmosisd", "query", "params", "subspace", subspace, key, "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)

	result := &params{}
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
	return result.Value
}

func (n *NodeConfig) QueryGovModuleAccount() string {
	cmd := []string{"osmosisd", "query", "auth", "module-accounts", "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	var result map[string][]interface{}
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
	for _, acc := range result["accounts"] {
		account, ok := acc.(map[string]interface{})
		require.True(n.t, ok)
		if account["name"] == "gov" {
			moduleAccount, ok := account["base_account"].(map[string]interface{})["address"].(string)
			require.True(n.t, ok)
			return moduleAccount
		}
	}
	require.True(n.t, false, "gov module account not found")
	return ""
}

func (n *NodeConfig) SubmitParamChangeProposal(proposalJson, from string) int {
	n.LogActionF("submitting param change proposal %s", proposalJson)
	// ToDo: Is there a better way to do this?
	wd, err := os.Getwd()
	require.NoError(n.t, err)
	currentTime := time.Now().Format("20060102-150405.000")
	localProposalFile := wd + fmt.Sprintf("/scripts/param_change_proposal_%s.json", currentTime)
	f, err := os.Create(localProposalFile)
	require.NoError(n.t, err)
	_, err = f.WriteString(proposalJson)
	require.NoError(n.t, err)
	err = f.Close()
	require.NoError(n.t, err)

	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "param-change", fmt.Sprintf("/osmosis/param_change_proposal_%s.json", currentTime), fmt.Sprintf("--from=%s", from)}

	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	err = os.Remove(localProposalFile)
	require.NoError(n.t, err)

	startIndex := strings.Index(resp.String(), `[{"key":"proposal_id","value":"`) + len(`[{"key":"proposal_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	proposalIDStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	proposalID, _ := strconv.Atoi(proposalIDStr)

	n.LogActionF("successfully submitted param change proposal")

	return proposalID
}

func (n *NodeConfig) SendIBCTransfer(dstChain *Config, from, recipient, memo string, token sdk.Coin) {
	n.LogActionF("IBC sending %s from %s to %s. memo: %s", token.Amount.String(), from, recipient, memo)

	cmd := []string{"hermes", "tx", "ft-transfer", "--dst-chain", dstChain.Id, "--src-chain", n.chainId, "--src-port", "transfer", "--src-channel", "channel-0", "--amount", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000", "--memo", memo}
	_, _, err := n.containerManager.ExecHermesCmd(n.t, cmd, "SUCCESS")
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted sent IBC transfer")
}

func (n *NodeConfig) FailIBCTransfer(from, recipient, amount string) {
	n.LogActionF("IBC sending %s from %s to %s", amount, from, recipient)

	cmd := []string{"osmosisd", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, amount, fmt.Sprintf("--from=%s", from)}

	_, _, err := n.containerManager.ExecTxCmdWithSuccessString(n.t, n.chainId, n.Name, cmd, "rate limit exceeded")
	require.NoError(n.t, err)

	n.LogActionF("Failed to send IBC transfer (as expected)")
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

func (n *NodeConfig) JoinPoolExactAmountIn(tokenIn string, poolId uint64, shareOutMinAmount string, from string) {
	n.LogActionF("join-swap-extern-amount-in (%s)  (%s) from (%s), pool id (%d)", tokenIn, shareOutMinAmount, from, poolId)
	cmd := []string{"osmosisd", "tx", "gamm", "join-swap-extern-amount-in", tokenIn, shareOutMinAmount, fmt.Sprintf("--pool-id=%d", poolId), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully joined pool")
}

func (n *NodeConfig) ExitPool(from, minAmountsOut string, poolId uint64, shareAmountIn string) {
	n.LogActionF("exiting gamm pool")
	cmd := []string{"osmosisd", "tx", "gamm", "exit-pool", fmt.Sprintf("--min-amounts-out=%s", minAmountsOut), fmt.Sprintf("--share-amount-in=%s", shareAmountIn), fmt.Sprintf("--pool-id=%d", poolId), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully exited pool %d, minAmountsOut %s, shareAmountIn %s", poolId, minAmountsOut, shareAmountIn)
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64, initialDeposit sdk.Coin) {
	n.LogActionF("submitting upgrade proposal %s for height %d", upgradeVersion, upgradeHeight)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%d", upgradeHeight), "--upgrade-info=\"\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted upgrade proposal")
}

func (n *NodeConfig) SubmitSuperfluidProposal(asset string, initialDeposit sdk.Coin) int {
	n.LogActionF("submitting superfluid proposal for asset %s", asset)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	// Extract the proposal ID from the response
	startIndex := strings.Index(resp.String(), `[{"key":"proposal_id","value":"`) + len(`[{"key":"proposal_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	proposalIDStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	proposalID, _ := strconv.Atoi(proposalIDStr)

	n.LogActionF("successfully submitted superfluid proposal for asset %s", asset)

	return proposalID
}

func (n *NodeConfig) SubmitCreateConcentratedPoolProposal(initialDeposit sdk.Coin) int {
	n.LogActionF("Creating concentrated liquidity pool")
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "create-concentratedliquidity-pool-proposal", "--pool-records=stake,uosmo,100,-6,0.001", "--title=\"create concentrated pool\"", "--description=\"create concentrated pool", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	// Extract the proposal ID from the response
	startIndex := strings.Index(resp.String(), `[{"key":"proposal_id","value":"`) + len(`[{"key":"proposal_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	proposalIDStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	proposalID, _ := strconv.Atoi(proposalIDStr)

	n.LogActionF("successfully created a create concentrated liquidity pool proposal")

	return proposalID
}

func (n *NodeConfig) SubmitTextProposal(text string, initialDeposit sdk.Coin, isExpedited bool) int {
	n.LogActionF("submitting text gov proposal")
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	if isExpedited {
		cmd = append(cmd, "--is-expedited=true")
	}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	// Extract the proposal ID from the response
	startIndex := strings.Index(resp.String(), `[{"key":"proposal_id","value":"`) + len(`[{"key":"proposal_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the proposal ID substring
	proposalIDStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the proposal ID from string to int
	proposalID, _ := strconv.Atoi(proposalIDStr)

	n.LogActionF("successfully submitted text gov proposal")

	return proposalID
}

func (n *NodeConfig) SubmitTickSpacingReductionProposal(poolTickSpacingRecords string, initialDeposit sdk.Coin, isExpedited bool) {
	n.LogActionF("submitting tick spacing reduction gov proposal")
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "tick-spacing-decrease-proposal", "--title=\"test tick spacing reduction proposal title\"", "--description=\"test tick spacing reduction proposal\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit), fmt.Sprintf("--pool-tick-spacing-records=%s", poolTickSpacingRecords)}
	if isExpedited {
		cmd = append(cmd, "--is-expedited=true")
	}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted tick spacing reduction gov proposal")
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

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) int {
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}

	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	// Extract the lock ID from the response
	startIndex := strings.Index(resp.String(), `[{"key":"period_lock_id","value":"`) + len(`[{"key":"period_lock_id","value":"`)
	endIndex := strings.Index(resp.String()[startIndex:], `"`)

	// Extract the lock ID substring
	lockIDStr := resp.String()[startIndex : startIndex+endIndex]

	// Convert the lock ID from string to int
	lockID, err := strconv.Atoi(lockIDStr)
	require.NoError(n.t, err)

	n.LogActionF("successfully created lock")

	return lockID
}

func (n *NodeConfig) AddToExistingLock(tokens sdk.Int, denom, duration, from string, lockID int) {
	n.LogActionF("noting previous lockup amount")
	path := fmt.Sprintf("/osmosis/lockup/v1beta1/locked_by_id/%d", lockID)
	bz, err := n.QueryGRPCGateway(path)
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

func (n *NodeConfig) FundCommunityPool(sendAddress string, funds string) {
	n.LogActionF("funding community pool from address %s with %s", sendAddress, funds)
	cmd := []string{"osmosisd", "tx", "distribution", "fund-community-pool", funds, fmt.Sprintf("--from=%s", sendAddress), "--gas=600000", "--fees=1500uosmo"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully funded community pool from address %s with %s", sendAddress, funds)
}

// This method also funds fee tokens from the `initialization.ValidatorWalletName` account.
// TODO: Abstract this to be a fee token provider account.
func (n *NodeConfig) CreateWallet(walletName string) string {
	n.LogActionF("creating wallet %s", walletName)
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("created wallet %s, wallet address - %s", walletName, walletAddr)
	n.BankSend(initialization.WalletFeeTokens.String(), initialization.ValidatorWalletName, walletAddr)
	n.LogActionF("Sent fee tokens from %s", initialization.ValidatorWalletName)
	return walletAddr
}

func (n *NodeConfig) CreateWalletAndFund(walletName string, tokensToFund []string) string {
	return n.CreateWalletAndFundFrom(walletName, initialization.ValidatorWalletName, tokensToFund)
}

func (n *NodeConfig) CreateWalletAndFundFrom(newWalletName string, fundingWalletName string, tokensToFund []string) string {
	n.LogActionF("Sending tokens to %s", newWalletName)

	walletAddr := n.CreateWallet(newWalletName)
	for _, tokenToFund := range tokensToFund {
		n.BankSend(tokenToFund, fundingWalletName, walletAddr)
	}

	n.LogActionF("Successfully sent tokens to %s", newWalletName)
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

type validatorInfo struct {
	Address     bytes.HexBytes
	PubKey      cryptotypes.PubKey
	VotingPower int64
}

// ResultStatus is node's info, same as Tendermint, except that we use our own
// PubKey.
type resultStatus struct {
	NodeInfo      p2p.DefaultNodeInfo
	SyncInfo      coretypes.SyncInfo
	ValidatorInfo validatorInfo
}

func (n *NodeConfig) Status() (resultStatus, error) {
	cmd := []string{"osmosisd", "status"}
	_, errBuf, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	if err != nil {
		return resultStatus{}, err
	}

	cfg := app.MakeEncodingConfig()
	legacyAmino := cfg.Amino
	var result resultStatus
	err = legacyAmino.UnmarshalJSON(errBuf.Bytes(), &result)
	fmt.Println("result", result)

	if err != nil {
		return resultStatus{}, err
	}
	return result, nil
}

func GetPositionID(responseJson map[string]interface{}) (string, error) {
	return ParseEvent(responseJson, "position_id")
}

func ParseEvent(responseJson map[string]interface{}, field string) (string, error) {
	logs, ok := responseJson["logs"].([]interface{})
	if !ok {
		return "", fmt.Errorf("logs field not found in response")
	}

	if len(logs) == 0 {
		return "", fmt.Errorf("empty logs field in response")
	}

	log, ok := logs[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format of logs field")
	}

	events, ok := log["events"].([]interface{})
	if !ok {
		return "", fmt.Errorf("events field not found in logs")
	}

	for _, event := range events {
		attributes, ok := event.(map[string]interface{})["attributes"].([]interface{})
		if !ok {
			return "", fmt.Errorf("attributes field not found in event")
		}

		for _, attr := range attributes {
			switch v := attr.(type) {
			case map[string]interface{}:
				if v["key"] == field {
					fieldID, ok := v["value"].(string)
					if !ok {
						return "", fmt.Errorf("invalid format of %s field", field)
					}
					return fieldID, nil
				}
			default:
				return "", fmt.Errorf("invalid type for attributes field")
			}
		}
	}

	return "", fmt.Errorf("%s field not found in response", field)
}

func (n *NodeConfig) SendIBC(dstChain *Config, recipient string, token sdk.Coin) {
	n.t.Logf("IBC sending %s from %s to %s (%s)", token, n.chainId, dstChain.Id, recipient)

	dstNode, err := dstChain.GetDefaultNode()
	require.NoError(n.t, err)

	// removes the fee token from balances for calculating the difference in other tokens
	// before and after the IBC send. Since we run tests in parallel now, some tests may
	// send uosmo between accounts while this test is running. Since we don't care about
	// non ibc denoms, its safe to filter uosmo out.
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
	require.NoError(n.t, err)
	fmt.Println("balancesDstPre with no fee removed: ", balancesDstPreWithTxFeeBalance)
	balancesDstPre := removeFeeTokenFromBalance(balancesDstPreWithTxFeeBalance)
	cmd := []string{"hermes", "tx", "ft-transfer", "--dst-chain", dstChain.Id, "--src-chain", n.chainId, "--src-port", "transfer", "--src-channel", "channel-0", "--amount", token.Amount.String(), fmt.Sprintf("--denom=%s", token.Denom), fmt.Sprintf("--receiver=%s", recipient), "--timeout-height-offset=1000"}
	_, _, err = n.containerManager.ExecHermesCmd(n.t, cmd, "SUCCESS")
	require.NoError(n.t, err)

	require.Eventually(
		n.t,
		func() bool {
			balancesDstPostWithTxFeeBalance, err := dstNode.QueryBalances(recipient)
			require.NoError(n.t, err)
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
		5*time.Minute,
		time.Second,
		"tx not received on destination chain",
	)

	n.t.Log("successfully sent IBC tokens")
}

func (n *NodeConfig) EnableSuperfluidAsset(srcChain *Config, denom string) {
	propNumber := n.SubmitSuperfluidProposal(denom, sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)))
	srcChain.LatestProposalNumber += 1
	n.DepositProposal(propNumber, false)

	var wg sync.WaitGroup

	for _, n := range srcChain.NodeConfigs {
		wg.Add(1)
		go func(node *NodeConfig) {
			defer wg.Done()
			node.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()
}

func (n *NodeConfig) LockAndAddToExistingLock(srcChain *Config, amount sdk.Int, denom, lockupWalletAddr, lockupWalletSuperfluidAddr string) {
	// lock tokens
	lockID := n.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletAddr)
	srcChain.LatestLockNumber += 1
	fmt.Println("lock number: ", lockID)
	// add to existing lock
	n.AddToExistingLock(amount, denom, "240s", lockupWalletAddr, lockID)

	// superfluid lock tokens
	lockID = n.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletSuperfluidAddr)
	srcChain.LatestLockNumber += 1
	fmt.Println("lock number: ", lockID)
	n.SuperfluidDelegate(lockID, srcChain.NodeConfigs[1].OperatorAddress, lockupWalletSuperfluidAddr)
	// add to existing lock
	n.AddToExistingLock(amount, denom, "240s", lockupWalletSuperfluidAddr, lockID)
}

// TODO remove chain from this as input
func (n *NodeConfig) SetupRateLimiting(paths, gov_addr string, chain *Config) (string, error) {
	srcNode, err := chain.GetNodeAtIndex(1)
	require.NoError(n.t, err)

	// copy the contract from x/rate-limit/testdata/
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// go up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	fmt.Println(wd, projectDir)
	_, err = util.CopyFile(projectDir+"/x/ibc-rate-limit/bytecode/rate_limiter.wasm", wd+"/scripts/rate_limiter.wasm")
	if err != nil {
		return "", err
	}

	codeId := srcNode.StoreWasmCode("rate_limiter.wasm", initialization.ValidatorWalletName)
	chain.LatestCodeId = int(srcNode.QueryLatestWasmCodeID())
	srcNode.InstantiateWasmContract(
		strconv.Itoa(codeId),
		fmt.Sprintf(`{"gov_module": "%s", "ibc_module": "%s", "paths": [%s] }`, gov_addr, initialization.ValidatorWalletName, paths),
		initialization.ValidatorWalletName)

	contracts, err := srcNode.QueryContractsFromId(codeId)
	if err != nil {
		return "", err
	}

	contract := contracts[len(contracts)-1]

	err = srcNode.ParamChangeProposal(
		ibcratelimittypes.ModuleName,
		string(ibcratelimittypes.KeyContractAddress),
		[]byte(fmt.Sprintf(`"%s"`, contract)),
		chain,
	)
	if err != nil {
		return "", err
	}
	require.Eventually(
		n.t,
		func() bool {
			val := srcNode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
			return strings.Contains(val, contract)
		},
		1*time.Minute,
		10*time.Millisecond,
	)
	fmt.Println("contract address set to", contract)
	return contract, nil
}

func (n *NodeConfig) ParamChangeProposal(subspace, key string, value []byte, chain *Config) error {
	proposal := paramsutils.ParamChangeProposalJSON{
		Title:       "Param Change",
		Description: fmt.Sprintf("Changing the %s param", key),
		Changes: paramsutils.ParamChangesJSON{
			paramsutils.ParamChangeJSON{
				Subspace: subspace,
				Key:      key,
				Value:    value,
			},
		},
		Deposit: "625000000uosmo",
	}
	proposalJson, err := json.Marshal(proposal)
	if err != nil {
		return err
	}

	node, err := chain.GetDefaultNode()
	if err != nil {
		return err
	}
	propNumber := node.SubmitParamChangeProposal(string(proposalJson), initialization.ValidatorWalletName)
	chain.LatestProposalNumber += 1

	var wg sync.WaitGroup

	for _, n := range chain.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	require.Eventually(n.t, func() bool {
		status, err := node.QueryPropStatus(propNumber)
		if err != nil {
			return false
		}
		return status == proposalStatusPassed
	}, time.Minute*30, time.Millisecond*500)
	return nil
}
