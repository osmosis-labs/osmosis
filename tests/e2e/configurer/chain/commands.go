package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/bytes"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/util"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/p2p"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"

	app "github.com/osmosis-labs/osmosis/v27/app"

	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
)

// The value is returned as a string, so we have to unmarshal twice
type params struct {
	Key      string `json:"key"`
	Subspace string `json:"subspace"`
	Value    string `json:"value"`
}

func (n *NodeConfig) SetMaxPoolPointsPerTx(maxPoolPoints int, from string) {
	n.LogActionF("setting max pool points per tx for protorev at %d points", maxPoolPoints)
	cmd := []string{"symphonyd", "tx", "protorev", "set-max-pool-points-per-tx", fmt.Sprintf("%d", maxPoolPoints), fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000note"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
}

func (n *NodeConfig) CreateBalancerPool(poolFile, from string) uint64 {
	n.LogActionF("creating balancer pool from file %s", poolFile)
	cmd := []string{"symphonyd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/symphony/%s", poolFile), fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000note"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	poolID, err := extractPoolIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully created balancer pool %d", poolID)
	return poolID
}

func (n *NodeConfig) CreateStableswapPool(poolFile, from string) uint64 {
	n.LogActionF("creating stableswap pool from file %s", poolFile)
	cmd := []string{"symphonyd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/symphony/%s", poolFile), "--pool-type=stableswap", fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000note"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	poolID, err := extractPoolIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully created stableswap pool with ID %d", poolID)
	return poolID
}

// CollectSpreadRewards collects spread rewards earned by concentrated position in range of [lowerTick; upperTick] in pool with id of poolId
func (n *NodeConfig) CollectSpreadRewards(from, positionIds string) {
	n.LogActionF("collecting spread rewards from concentrated position")
	cmd := []string{"symphonyd", "tx", "concentratedliquidity", "collect-spread-rewards", positionIds, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	n.LogActionF("successfully collected spread rewards for account %s", from)
}

// CreateConcentratedPool creates a concentrated pool.
// Returns pool id of newly created pool on success
func (n *NodeConfig) CreateConcentratedPool(from, denom1, denom2 string, tickSpacing uint64, spreadFactor string) uint64 {
	n.LogActionF("creating concentrated pool")

	cmd := []string{"symphonyd", "tx", "concentratedliquidity", "create-pool", denom1, denom2, fmt.Sprintf("%d", tickSpacing), spreadFactor, fmt.Sprintf("--from=%s", from)}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	poolID, err := extractPoolIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully created concentrated pool with ID %d", poolID)
	return poolID
}

// CreateConcentratedPosition creates a concentrated position from [lowerTick; upperTick] in pool with id of poolId
// token{0,1} - liquidity to create position with
func (n *NodeConfig) CreateConcentratedPosition(from, lowerTick, upperTick string, tokens string, token0MinAmt, token1MinAmt int64, poolId uint64) (uint64, osmomath.Dec) {
	n.LogActionF("creating concentrated position")
	// gas = 50,000 because e2e  default to 40,000, we hardcoded extra 10k gas to initialize tick
	// fees = 1250 (because 50,000 * 0.0025 = 1250)
	cmd := []string{"symphonyd", "tx", "concentratedliquidity", "create-position", fmt.Sprint(poolId), lowerTick, upperTick, tokens, fmt.Sprintf("%d", token0MinAmt), fmt.Sprintf("%d", token1MinAmt), fmt.Sprintf("--from=%s", from), "--gas=500000", "--fees=1250note", "-o json"}
	resp, _, err := n.containerManager.ExecTxCmdWithSuccessStringJSON(n.t, n.chainId, n.Name, cmd, "\"code\":0,")
	require.NoError(n.t, err)
	response := formatNonJsonResponse(resp.String())

	// Extract the position_id from the response
	r := regexp.MustCompile(`"position_id","value":"(\d+)"`)
	matches := r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		return 0, osmomath.ZeroDec()
	}

	// Convert the position_id from string to int
	positionID, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, osmomath.ZeroDec()
	}

	// Extract the liquidity from the response
	r = regexp.MustCompile(`"liquidity","value":"(\d+\.\d+)"`)
	matches = r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		return 0, osmomath.ZeroDec()
	}

	// Convert the liquidity from string to Dec
	liquidityStr := matches[1]

	n.LogActionF("successfully created concentrated position from %s to %s", lowerTick, upperTick)

	return uint64(positionID), osmomath.MustNewDecFromStr(liquidityStr)
}

func (n *NodeConfig) StoreWasmCode(wasmFile, from string) int {
	n.LogActionF("storing wasm code from file %s", wasmFile)
	cmd := []string{"symphonyd", "tx", "wasm", "store", wasmFile, fmt.Sprintf("--from=%s", from), "--gas=auto", "--gas-prices=0.1note", "--gas-adjustment=1.3"}
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	codeId, err := extractCodeIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully stored")
	return codeId
}

func (n *NodeConfig) WithdrawPosition(from, liquidityOut string, positionId uint64) {
	n.LogActionF("withdrawing liquidity from position %d", positionId)
	cmd := []string{"symphonyd", "tx", "concentratedliquidity", "withdraw-position", fmt.Sprint(positionId), liquidityOut, fmt.Sprintf("--from=%s", from), "--gas=700000", "--fees=5000note"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully withdrew %s liquidity from position %d", liquidityOut, positionId)
}

func (n *NodeConfig) InstantiateWasmContract(codeId, initMsg, from string) {
	n.LogActionF("instantiating wasm contract %s with %s", codeId, initMsg)
	cmd := []string{"symphonyd", "tx", "wasm", "instantiate", codeId, initMsg, fmt.Sprintf("--from=%s", from), "--no-admin", "--label=contract"}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully initialized")
}

func (n *NodeConfig) WasmExecute(contract, execMsg, from string) {
	n.LogActionF("executing %s on wasm contract %s from %s", execMsg, contract, from)
	cmd := []string{"symphonyd", "tx", "wasm", "execute", contract, execMsg, fmt.Sprintf("--from=%s", from)}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully executed")
}

// QueryParams extracts the params for a given subspace and key. This is done generically via json to avoid having to
// specify the QueryParamResponse type (which may not exist for all params).
func (n *NodeConfig) QueryParams(subspace, key string) string {
	cmd := []string{"symphonyd", "query", "params", "subspace", subspace, key, "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
	require.NoError(n.t, err)

	result := &params{}
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
	return result.Value
}

func (n *NodeConfig) QueryGovModuleAccount() string {
	cmd := []string{"symphonyd", "query", "auth", "module-accounts", "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
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

func (n *NodeConfig) SubmitParamChangeProposal(proposalJson, from string, isLegacy bool) int {
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

	var cmd []string
	if isLegacy {
		cmd = []string{"symphonyd", "tx", "gov", "submit-legacy-proposal", "param-change", fmt.Sprintf("/symphony/param_change_proposal_%s.json", currentTime), fmt.Sprintf("--from=%s", from)}
	} else {
		cmd = []string{"symphonyd", "tx", "gov", "submit-proposal", "param-change", fmt.Sprintf("/symphony/param_change_proposal_%s.json", currentTime), fmt.Sprintf("--from=%s", from)}
	}

	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	os.Remove(localProposalFile)

	proposalID, err := extractProposalIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted param change proposal")

	return proposalID
}

func (n *NodeConfig) SubmitNewV1ProposalType(proposalJson, from string) int {
	n.LogActionF("submitting new v1 proposal type %s", proposalJson)
	// ToDo: Is there a better way to do this?
	wd, err := os.Getwd()
	require.NoError(n.t, err)
	currentTime := time.Now().Format("20060102-150405.000")
	localProposalFile := wd + fmt.Sprintf("/scripts/new_v1_prop_%s.json", currentTime)
	f, err := os.Create(localProposalFile)
	require.NoError(n.t, err)
	_, err = f.WriteString(proposalJson)
	require.NoError(n.t, err)
	err = f.Close()
	require.NoError(n.t, err)

	cmd := []string{"symphonyd", "tx", "gov", "submit-proposal", fmt.Sprintf("/symphony/new_v1_prop_%s.json", currentTime), fmt.Sprintf("--from=%s", from)}

	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	os.Remove(localProposalFile)

	proposalID, err := extractProposalIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted new v1 proposal type")

	return proposalID
}

func (n *NodeConfig) SendIBCTransfer(dstChain *Config, from, recipient, memo string, token sdk.Coin) {
	n.LogActionF("IBC sending %s from %s to %s. memo: %s", token.Amount.String(), from, recipient, memo)

	cmd := []string{"symphonyd", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, token.String(), fmt.Sprintf("--from=%s", from), "--memo", memo}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	cmd = []string{"hermes", "clear", "packets", "--chain", dstChain.Id, "--port", "transfer", "--channel", "channel-0"}
	_, _, err = n.containerManager.ExecHermesCmd(n.t, cmd, "SUCCESS")
	require.NoError(n.t, err)

	cmd = []string{"hermes", "clear", "packets", "--chain", n.chainId, "--port", "transfer", "--channel", "channel-0"}
	_, _, err = n.containerManager.ExecHermesCmd(n.t, cmd, "SUCCESS")
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted sent IBC transfer")
}

func (n *NodeConfig) FailIBCTransfer(from, recipient, amount string) {
	n.LogActionF("IBC sending %s from %s to %s", amount, from, recipient)

	cmd := []string{"symphonyd", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, amount, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmdWithSuccessString(n.t, n.chainId, n.Name, cmd, "rate limit exceeded")
	require.NoError(n.t, err)

	n.LogActionF("Failed to send IBC transfer (as expected)")
}

// SwapExactAmountIn swaps tokenInCoin to get at least tokenOutMinAmountInt of the other token's pool out.
// swapRoutePoolIds is the comma separated list of pool ids to swap through.
// swapRouteDenoms is the comma separated list of denoms to swap through.
// To reproduce locally:
// docker container exec <container id> symphonyd tx gamm swap-exact-amount-in <tokeinInCoin> <tokenOutMinAmountInt> --swap-route-pool-ids <swapRoutePoolIds> --swap-route-denoms <swapRouteDenoms> --chain-id=<id>--from=<address> --keyring-backend=test --yes --log_format=json
func (n *NodeConfig) SwapExactAmountIn(tokenInCoin, tokenOutMinAmountInt string, swapRoutePoolIds string, swapRouteDenoms string, from string) {
	n.LogActionF("swapping %s to get a minimum of %s with pool id routes (%s) and denom routes (%s)", tokenInCoin, tokenOutMinAmountInt, swapRoutePoolIds, swapRouteDenoms)
	cmd := []string{"symphonyd", "tx", "gamm", "swap-exact-amount-in", tokenInCoin, tokenOutMinAmountInt, fmt.Sprintf("--swap-route-pool-ids=%s", swapRoutePoolIds), fmt.Sprintf("--swap-route-denoms=%s", swapRouteDenoms), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully swapped")
}

func (n *NodeConfig) JoinPoolExactAmountIn(tokenIn string, poolId uint64, shareOutMinAmount string, from string) {
	n.LogActionF("join-swap-extern-amount-in (%s)  (%s) from (%s), pool id (%d)", tokenIn, shareOutMinAmount, from, poolId)
	cmd := []string{"symphonyd", "tx", "gamm", "join-swap-extern-amount-in", tokenIn, shareOutMinAmount, fmt.Sprintf("--pool-id=%d", poolId), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully joined pool")
}

func (n *NodeConfig) ExitPool(from, minAmountsOut string, poolId uint64, shareAmountIn string) {
	n.LogActionF("exiting gamm pool")
	cmd := []string{"symphonyd", "tx", "gamm", "exit-pool", fmt.Sprintf("--min-amounts-out=%s", minAmountsOut), fmt.Sprintf("--share-amount-in=%s", shareAmountIn), fmt.Sprintf("--pool-id=%d", poolId), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully exited pool %d, minAmountsOut %s, shareAmountIn %s", poolId, minAmountsOut, shareAmountIn)
}

func (n *NodeConfig) SubmitProposal(cmdArgs []string, isExpedited bool, propDescriptionForLogs string, isLegacy bool) int {
	n.LogActionF("submitting proposal: %s", propDescriptionForLogs)
	var cmd []string
	if isLegacy {
		cmd = append([]string{"symphonyd", "tx", "gov", "submit-legacy-proposal"}, cmdArgs...)
	} else {
		cmd = append([]string{"symphonyd", "tx", "gov", "submit-proposal"}, cmdArgs...)
	}

	depositAmt := sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.InitialMinDeposit))
	if isExpedited {
		cmd = append(cmd, "--is-expedited=true")
		depositAmt = sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.MinExpeditedDepositValue))
	}
	cmd = append(cmd, fmt.Sprintf("--deposit=%s", depositAmt))
	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	proposalID, err := extractProposalIdFromResponse(resp.String())
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted proposal: %s", propDescriptionForLogs)

	return proposalID
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64, initialDeposit sdk.Coin, isLegacy bool) int {
	cmd := []string{"software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%d", upgradeHeight), "--upgrade-info=\"\"", "--no-validate", "--from=val"}
	return n.SubmitProposal(cmd, false, fmt.Sprintf("upgrade proposal %s for height %d", upgradeVersion, upgradeHeight), true)
}

func (n *NodeConfig) SubmitSuperfluidProposal(asset string, isLegacy bool) int {
	cmd := []string{"set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), "--title=\"superfluid asset prop\"", fmt.Sprintf("--summary=\"%s superfluid asset\"", asset), "--from=val", "--gas=700000", "--fees=5000note"}

	// TODO: no expedited flag for some reason
	return n.SubmitProposal(cmd, false, fmt.Sprintf("superfluid proposal for asset %s", asset), isLegacy)
}

func (n *NodeConfig) SubmitCreateConcentratedPoolProposal(isExpedited, isLegacy bool) int {
	cmd := []string{"create-concentratedliquidity-pool-proposal", "--pool-records=stake,note,100,0.001", "--title=\"create concentrated pool\"", "--summary=\"create concentrated pool\"", "--from=val", "--gas=400000", "--fees=5000note"}
	return n.SubmitProposal(cmd, isExpedited, "create concentrated liquidity pool", isLegacy)
}

func (n *NodeConfig) SubmitTextProposal(text string, isExpedited, isLegacy bool) int {
	cmd := []string{"--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val"}
	return n.SubmitProposal(cmd, isExpedited, "text proposal", isLegacy)
}

func (n *NodeConfig) SubmitTickSpacingReductionProposal(poolTickSpacingRecords string, isExpedited, isLegacy bool) int {
	cmd := []string{"tick-spacing-decrease-proposal", "--title=\"test tick spacing reduction proposal title\"", "--summary=\"test tick spacing reduction proposal\"", "--from=val", fmt.Sprintf("--pool-tick-spacing-records=%s", poolTickSpacingRecords)}
	return n.SubmitProposal(cmd, isExpedited, "tick spacing reduction proposal", isLegacy)
}

func (n *NodeConfig) DepositProposal(proposalNumber int, isExpedited bool) {
	n.LogActionF("depositing on proposal: %d", proposalNumber)
	deposit := sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.MinDepositValue)).String()
	if isExpedited {
		deposit = sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.MinExpeditedDepositValue)).String()
	}
	cmd := []string{"symphonyd", "tx", "gov", "deposit", fmt.Sprintf("%d", proposalNumber), deposit, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully deposited on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteYesProposal(from string, proposalNumber int) {
	n.LogActionF("voting yes on proposal: %d", proposalNumber)
	cmd := []string{"symphonyd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "yes", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted yes on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteNoProposal(from string, proposalNumber int) {
	n.LogActionF("voting no on proposal: %d", proposalNumber)
	cmd := []string{"symphonyd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "no", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted no on proposal: %d", proposalNumber)
}

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) int {
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"symphonyd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}

	resp, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	response := formatNonJsonResponse(resp.String())

	// Extract the lock ID from the response
	r := regexp.MustCompile(`period_lock_id value: "(\d+)"`)
	matches := r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		require.Fail(n.t, "period_lock_id not found")
	}

	// Convert the lock ID from string to int
	lockID, err := strconv.Atoi(matches[1])
	require.NoError(n.t, err)

	n.LogActionF("successfully created lock")

	return lockID
}

func (n *NodeConfig) AddToExistingLock(tokens osmomath.Int, denom, duration, from string, lockID int) {
	n.LogActionF("noting previous lockup amount")
	path := fmt.Sprintf("/symphony/lockup/v1beta1/locked_by_id/%d", lockID)
	bz, err := n.QueryGRPCGateway(path)
	require.NoError(n.t, err)
	var lockedResp lockuptypes.LockedResponse
	err = util.Cdc.UnmarshalJSON(bz, &lockedResp)
	require.NoError(n.t, err)
	previousLockupAmount := lockedResp.Lock.Coins.AmountOf(denom)
	n.LogActionF("previous lockup amount is %v", previousLockupAmount)
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"symphonyd", "tx", "lockup", "lock-tokens", fmt.Sprintf("%s%s", tokens, denom), fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
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
	cmd := []string{"symphonyd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully superfluid delegated lock %s to %s", lockStr, valAddress)
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.LogActionF("bank sending %s from address %s to %s", amount, sendAddress, receiveAddress)
	cmd := []string{"symphonyd", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully sent bank sent %s from address %s to %s", amount, sendAddress, receiveAddress)
}

func (n *NodeConfig) FundCommunityPool(sendAddress string, funds string) {
	n.LogActionF("funding community pool from address %s with %s", sendAddress, funds)
	cmd := []string{"symphonyd", "tx", "distribution", "fund-community-pool", funds, fmt.Sprintf("--from=%s", sendAddress), "--gas=600000", "--fees=1500note"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully funded community pool from address %s with %s", sendAddress, funds)
}

// This method also funds fee tokens from the `initialization.ValidatorWalletName` account.
// TODO: Abstract this to be a fee token provider account.
func (n *NodeConfig) CreateWallet(walletName string, chain *Config) string {
	n.LogActionF("creating wallet %s", walletName)
	cmd := []string{"symphonyd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, errBuf, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
	require.NoError(n.t, err)
	re := regexp.MustCompile("symphony1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")

	mnemonic, err := pullMnemonicFromResponse(errBuf.String())
	require.NoError(n.t, err)

	chainNodes := chain.GetAllChainNodes()
	for _, node := range chainNodes {
		if node.Name == n.Name {
			continue
		}
		node.AddExistingWallet(walletName, mnemonic)
	}

	n.LogActionF("created wallet %s, wallet address - %s", walletName, walletAddr)
	n.BankSend(initialization.WalletFeeTokens.String(), initialization.ValidatorWalletName, walletAddr)
	n.LogActionF("Sent fee tokens from %s", initialization.ValidatorWalletName)
	return walletAddr
}

func (n *NodeConfig) AddExistingWallet(walletName, mnemonic string) {
	n.LogActionF("adding existing wallet %s", walletName)
	cmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' | symphonyd keys add %s --keyring-backend=test --recover", mnemonic, walletName)}
	_, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
	require.NoError(n.t, err)

	n.LogActionF("added existing wallet %s", walletName)
}

func (n *NodeConfig) CreateWalletAndFund(walletName string, tokensToFund []string, chain *Config) string {
	return n.CreateWalletAndFundFrom(walletName, initialization.ValidatorWalletName, tokensToFund, chain)
}

func (n *NodeConfig) CreateWalletAndFundFrom(newWalletName string, fundingWalletName string, tokensToFund []string, chain *Config) string {
	n.LogActionF("Sending tokens to %s", newWalletName)

	walletAddr := n.CreateWallet(newWalletName, chain)
	for _, tokenToFund := range tokensToFund {
		n.BankSend(tokenToFund, fundingWalletName, walletAddr)
	}

	n.LogActionF("Successfully sent tokens to %s", newWalletName)
	return walletAddr
}

func (n *NodeConfig) GetWallet(walletName string) string {
	n.LogActionF("retrieving wallet %s", walletName)
	cmd := []string{"symphonyd", "keys", "show", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
	require.NoError(n.t, err)
	re := regexp.MustCompile("symphony1(.{38})")
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
		"Symphony node failed to retrieve prop tally",
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
	cmd := []string{"symphonyd", "status"}
	_, errBuf, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "", false, false)
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

func GetLiquidity(responseJson map[string]interface{}) (string, error) {
	return ParseEvent(responseJson, "liquidity")
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

// TODO: Make test usage so that we can eliminate this!
var addrMutexMap = make(map[string]*sync.Mutex)

func IbcLockAddrs(addrs []string) func() {
	for _, addr := range addrs {
		if _, exists := addrMutexMap[addr]; !exists {
			addrMutexMap[addr] = &sync.Mutex{}
		}
	}
	for _, addr := range addrs {
		addrMutexMap[addr].Lock()
	}
	return func() {
		for _, addr := range addrs {
			addrMutexMap[addr].Unlock()
		}
	}
}

func (n *NodeConfig) SendIBC(srcChain, dstChain *Config, recipient string, token sdk.Coin) {
	// We add a mutex here since we don't want multiple IBC transfers to happen at the same time
	// Otherwise, when we check if the receiving end has the correct balance, we might get the balance
	// of a previous transfer.
	sender := n.GetWallet(initialization.ValidatorWalletName)

	// Create or get the mutex for the specific sender and recipient
	unlockFn := IbcLockAddrs([]string{recipient, sender})
	defer unlockFn()

	n.SendIBCNoMutex(srcChain, dstChain, recipient, token)
}

func (n *NodeConfig) SendIBCNoMutex(srcChain, dstChain *Config, recipient string, token sdk.Coin) {
	n.t.Logf("IBC sending %s from %s to %s (%s)", token, n.chainId, dstChain.Id, recipient)

	sender := n.GetWallet(initialization.ValidatorWalletName)
	dstNode, err := dstChain.GetDefaultNode()
	require.NoError(n.t, err)

	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", token.Denom))
	ibcDenom := denomTrace.IBCDenom()

	balancePre, err := dstNode.QueryBalance(recipient, ibcDenom)
	require.NoError(n.t, err)

	n.SendIBCTransfer(dstChain, sender, recipient, "", token)

	require.Eventually(
		n.t,
		func() bool {
			balancePost, err := dstNode.QueryBalance(recipient, ibcDenom)
			require.NoError(n.t, err)

			return balancePost.Amount.Equal(balancePre.Amount.Add(token.Amount))
		},
		3*time.Minute, // TODO: Lower this
		10*time.Millisecond,
		"tx not received on destination chain",
	)

	n.t.Log("successfully sent IBC tokens")
}

func (n *NodeConfig) EnableSuperfluidAsset(srcChain *Config, denom string, isLegacy bool) {
	propNumber := n.SubmitSuperfluidProposal(denom, isLegacy)
	n.DepositProposal(propNumber, false)

	AllValsVoteOnProposal(srcChain, propNumber)
}

func (n *NodeConfig) LockAndAddToExistingLock(srcChain *Config, amount osmomath.Int, denom, lockupWalletAddr, lockupWalletSuperfluidAddr string) {
	// lock tokens
	lockID := n.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletAddr)

	// add to existing lock
	n.AddToExistingLock(amount, denom, "240s", lockupWalletAddr, lockID)

	// superfluid lock tokens
	lockID = n.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletSuperfluidAddr)

	n.SuperfluidDelegate(lockID, srcChain.NodeConfigs[1].OperatorAddress, lockupWalletSuperfluidAddr)
	// add to existing lock
	n.AddToExistingLock(amount, denom, "240s", lockupWalletSuperfluidAddr, lockID)
}

// TODO remove chain from this as input
func (n *NodeConfig) SetupRateLimiting(paths, gov_addr string, chain *Config, isLegacy bool) (string, error) {
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
		isLegacy,
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

func (n *NodeConfig) ParamChangeProposal(subspace, key string, value []byte, chain *Config, isLegacy bool) error {
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
		Deposit: strconv.Itoa(int(config.InitialMinExpeditedDeposit)) + appparams.BaseCoinUnit,
	}
	proposalJson, err := json.Marshal(proposal)
	if err != nil {
		return err
	}

	propNumber := n.SubmitParamChangeProposal(string(proposalJson), initialization.ValidatorWalletName, isLegacy)

	AllValsVoteOnProposal(chain, propNumber)

	require.Eventually(n.t, func() bool {
		status, err := n.QueryPropStatus(propNumber)
		if err != nil {
			return false
		}
		return status == proposalStatusPassed
	}, time.Minute*2, 10*time.Millisecond)
	return nil
}

func AllValsVoteOnProposal(chain *Config, propNumber int) {
	var wg sync.WaitGroup

	for _, n := range chain.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()
}

func extractProposalIdFromResponse(response string) (int, error) {
	response = formatNonJsonResponse(response)

	// Extract the proposal ID from the response
	r := regexp.MustCompile(`proposal_id value: "(\d+)"`)
	matches := r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		return 0, errors.New("proposal_id not found")
	}

	// Convert the proposal ID from string to int
	proposalID, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return proposalID, nil
}

func extractPoolIdFromResponse(response string) (uint64, error) {
	response = formatNonJsonResponse(response)

	// Extract the pool ID from the response
	//fmt.Println(response)
	r := regexp.MustCompile(`pool_id value: "(\d+)"`)
	matches := r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		return 0, errors.New("pool_id not found")
	}

	// Convert the pool ID from string to uint64
	poolID, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return poolID, nil
}

func extractCodeIdFromResponse(response string) (int, error) {
	response = formatNonJsonResponse(response)

	r := regexp.MustCompile(`code_id value: "(\d+)"`)
	matches := r.FindStringSubmatch(response)

	// Check if we found a match
	if len(matches) < 2 {
		return 0, errors.New("code_id not found")
	}

	// Convert the code ID from string to int
	codeId, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return codeId, nil
}

func pullMnemonicFromResponse(response string) (string, error) {
	// Using regex to get mnemonic
	r := regexp.MustCompile(`(?m)(?i)^(\w+\s){23}\w+$`)
	mnemonicMatch := r.FindString(response)

	// Check if we found the mnemonic
	if mnemonicMatch == "" {
		return "", errors.New("mnemonic not found")
	}

	// Split the mnemonicMatch on spaces to get individual words
	mnemonicWords := strings.Split(mnemonicMatch, " ")

	// Check if we got 24 words
	if len(mnemonicWords) != 24 {
		return "", errors.New("mnemonic does not contain 24 words")
	}

	return mnemonicMatch, nil
}

func formatNonJsonResponse(resp string) string {
	resp = strings.ReplaceAll(resp, "\n", " ")
	resp = strings.ReplaceAll(resp, "-", " ")
	return strings.Join(strings.Fields(resp), " ")
}
