package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"testing"

	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	appparams "github.com/osmosis-labs/osmosis/v14/app/params"
	"github.com/osmosis-labs/osmosis/v14/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v14/tests/e2e/initialization"
	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"gotest.tools/v3/assert"
	"github.com/stretchr/testify/require"
)

func TestConcentratedLiquidity(t *testing.T) {
	fmt.Println("TestConcentratedLiquidity", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	chainA := f.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	var (
		denom0                    string = "uion"
		denom1                    string = "uosmo"
		tickSpacing               uint64 = 1
		precisionFactorAtPriceOne int64  = -1
		frozenUntil               int64  = time.Unix(86400, 0).Unix()
		swapFee                          = "0.01"
	)
	poolID := node.CreateConcentratedPool(initialization.ValidatorWalletName, denom0, denom1, tickSpacing, precisionFactorAtPriceOne, swapFee)

	concentratedPool, err := node.QueryConcentratedPool(poolID)
	assert.NilError(t, err)

	// assert contents of the pool are valid
	assert.Equal(t, concentratedPool.GetId(), poolID)
	assert.Equal(t, concentratedPool.GetToken0(), denom0)
	assert.Equal(t, concentratedPool.GetToken1(), denom1)
	assert.Equal(t, concentratedPool.GetTickSpacing(), tickSpacing)
	assert.Equal(t, concentratedPool.GetPrecisionFactorAtPriceOne(), sdk.NewInt(precisionFactorAtPriceOne))
	assert.Equal(t, concentratedPool.GetSwapFee(sdk.Context{}), sdk.MustNewDecFromStr(swapFee))

	minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(sdk.NewInt(precisionFactorAtPriceOne))

	fundTokens := []string{"1000000uosmo", "1000000uion", "1000000stake"}
	// get 3 addresses to create positions
	address1 := node.CreateWalletAndFund("addr1", fundTokens)
	address2 := node.CreateWalletAndFund("addr2", fundTokens)
	address3 := node.CreateWalletAndFund("addr3", fundTokens)

	// Create 2 positions for node1: overlap together, overlap with 2 node3 positions)
	node.CreateConcentratedPosition(address1, "[-1200]", "400", fmt.Sprintf("1000%s", denom0), fmt.Sprintf("1000%s", denom1), 0, 0, frozenUntil, poolID)
	node.CreateConcentratedPosition(address1, "[-400]", "400", fmt.Sprintf("1000%s", denom0), fmt.Sprintf("1000%s", denom1), 0, 0, frozenUntil, poolID)

	// Create 1 position for node2: does not overlap with anything, ends at maximum
	node.CreateConcentratedPosition(address2, "2200", fmt.Sprintf("%d", maxTick), fmt.Sprintf("1000%s", denom0), fmt.Sprintf("1000%s", denom1), 0, 0, frozenUntil, poolID)

	// Create 2 positions for node3: overlap together, overlap with 2 node1 positions, one position starts from minimum
	node.CreateConcentratedPosition(address3, "[-1600]", "[-200]", fmt.Sprintf("1000%s", denom0), fmt.Sprintf("1000%s", denom1), 0, 0, frozenUntil, poolID)
	node.CreateConcentratedPosition(address3, fmt.Sprintf("[%d]", minTick), "1400", fmt.Sprintf("1000%s", denom0), fmt.Sprintf("1000%s", denom1), 0, 0, frozenUntil, poolID)

	// get newly created positions
	positionsAddress1 := node.QueryConcentratedPositions(address1)
	positionsAddress2 := node.QueryConcentratedPositions(address2)
	positionsAddress3 := node.QueryConcentratedPositions(address3)

	// assert number of positions per address
	assert.Equal(t, len(positionsAddress1), 2)
	assert.Equal(t, len(positionsAddress2), 1)
	assert.Equal(t, len(positionsAddress3), 2)

	// Assert returned positions:
	validateCLPosition := func(position types.FullPositionByOwnerResult, poolId uint64, lowerTick, upperTick int64) {
		assert.Equal(t, position.PoolId, poolId)
		assert.Equal(t, position.LowerTick, int64(lowerTick))
		assert.Equal(t, position.UpperTick, int64(upperTick))
	}

	// assert positions for address1
	addr1position1 := positionsAddress1[0]
	addr1position2 := positionsAddress1[1]
	// first position first address
	validateCLPosition(addr1position1, poolID, -1200, 400)
	// second position second address
	validateCLPosition(addr1position2, poolID, -400, 400)

	// assert positions for address2
	addr2position1 := positionsAddress2[0]
	// first position second address
	validateCLPosition(addr2position1, poolID, 2200, maxTick)

	// assert positions for address3
	addr3position1 := positionsAddress3[0]
	addr3position2 := positionsAddress3[1]
	// first position third address
	validateCLPosition(addr3position1, poolID, -1600, -200)
	// second position third address
	validateCLPosition(addr3position2, poolID, minTick, 1400)
}

// TestGeometricTwapMigration tests that the geometric twap record
// migration runs succesfully. It does so by attempting to execute
// the swap on the pool created pre-upgrade. When a pool is created
// pre-upgrade, twap records are initialized for a pool. By runnning
// a swap post-upgrade, we confirm that the geometric twap was initialized
// correctly and does not cause a chain halt. This test was created
// in-response to a testnet incident when performing the geometric twap
// upgrade. Upon adding the migrations logic, the tests began to pass.
func TestGeometricTwapMigration(t *testing.T) {
	fmt.Println("TestGeometricTwapMigration", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)
	if f.skipUpgrade {
		t.Skip("Skipping upgrade tests")
	}

	const (
		// Configurations for tests/e2e/scripts/pool1A.json
		// This pool gets initialized pre-upgrade.
		oldPoolId       = 1
		minAmountOut    = "1"
		otherDenom      = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
		migrationWallet = "migration"
	)

	chainA := f.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	uosmoIn := fmt.Sprintf("1000000%s", "uosmo")

	swapWalletAddr := node.CreateWallet(migrationWallet)

	node.BankSend(uosmoIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)

	// Swap to create new twap records on the pool that was created pre-upgrade.
	node.SwapExactAmountIn(uosmoIn, minAmountOut, fmt.Sprintf("%d", oldPoolId), otherDenom, swapWalletAddr)
}

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
// Additionally, it attempst to create a pool with IBC denoms.
func TestIBCTokenTransferAndCreatePool(t *testing.T) {
	fmt.Println("TestIBCTokenTransferAndCreatePool", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	if f.skipIBC {
		t.Skip("Skipping IBC tests")
	}
	chainA := f.configurer.GetChainConfig(0)
	chainB := f.configurer.GetChainConfig(1)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.StakeToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.StakeToken)

	chainANode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)
	chainANode.CreateBalancerPool("ibcDenomPool.json", initialization.ValidatorWalletName)
}

// TestSuperfluidVoting tests that superfluid voting is functioning as expected.
// It does so by doing the following:
// - creating a pool
// - attempting to submit a proposal to enable superfluid voting in that pool
// - voting yes on the proposal from the validator wallet
// - voting no on the proposal from the delegator wallet
// - ensuring that delegator's wallet overwrites the validator's vote
func TestSuperfluidVoting(t *testing.T) {
	fmt.Println("TestSuperfluidVoting", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	chainA := f.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	poolId := chainANode.CreateBalancerPool("nativeDenomPool.json", chainA.NodeConfigs[0].PublicAddress)

	// enable superfluid assets
	chainA.EnableSuperfluidAsset(fmt.Sprintf("gamm/pool/%d", poolId))

	// setup wallets and send gamm tokens to these wallets (both chains)
	superfluildVotingWallet := chainANode.CreateWallet("TestSuperfluidVoting")
	chainANode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), chainA.NodeConfigs[0].PublicAddress, superfluildVotingWallet)
	chainANode.LockTokens(fmt.Sprintf("%v%s", sdk.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId)), "240s", superfluildVotingWallet)
	chainA.LatestLockNumber += 1
	chainANode.SuperfluidDelegate(chainA.LatestLockNumber, chainA.NodeConfigs[1].OperatorAddress, superfluildVotingWallet)

	// create a text prop, deposit and vote yes
	chainANode.SubmitTextProposal("superfluid vote overwrite test", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)), false)
	chainA.LatestProposalNumber += 1
	chainANode.DepositProposal(chainA.LatestProposalNumber, false)
	for _, node := range chainA.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}

	// set delegator vote to no
	chainANode.VoteNoProposal(superfluildVotingWallet, chainA.LatestProposalNumber)

	require.Eventually(
		t,
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := chainANode.QueryPropTally(chainA.LatestProposalNumber)
			if err != nil {
				return false
			}
			if abstainTotal.Int64()+noTotal.Int64()+noWithVetoTotal.Int64()+yesTotal.Int64() <= 0 {
				return false
			}
			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve prop tally",
	)
	noTotal, _, _, _, _ := chainANode.QueryPropTally(chainA.LatestProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	assert.NilError(t, err)

	require.Eventually(
		t,
		func() bool {
			intAccountBalance, err := chainANode.QueryIntermediaryAccount(fmt.Sprintf("gamm/pool/%d", poolId), chainA.NodeConfigs[1].OperatorAddress)
			assert.NilError(t, err)
			if err != nil {
				return false
			}
			if noTotalFinal != intAccountBalance {
				fmt.Printf("noTotalFinal %v does not match intAccountBalance %v", noTotalFinal, intAccountBalance)
				return false
			}
			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"superfluid delegation vote overwrite not working as expected",
	)
}

// Copy a file from A to B with io.Copy
func copyFile(a, b string) error {
	source, err := os.Open(a)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(b)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}
	return nil
}

func TestIBCTokenTransferRateLimiting(t *testing.T) {
	fmt.Println("TestIBCTokenTransferRateLimiting", time.Now())
	t.Parallel()
	fixture := InitIntegrationFixture(t)

	if fixture.skipIBC {
		t.Skip("Skipping IBC tests")
	}
	chainA := fixture.configurer.GetChainConfig(0)
	chainB := fixture.configurer.GetChainConfig(1)

	node, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	osmoSupply, err := node.QuerySupplyOf("uosmo")
	assert.NilError(t, err)

	f, err := osmoSupply.ToDec().Float64()
	assert.NilError(t, err)

	over := f * 0.02

	// Sending >1%
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, sdk.NewInt64Coin(initialization.OsmoDenom, int64(over)))

	// copy the contract from x/rate-limit/testdata/
	wd, err := os.Getwd()
	assert.NilError(t, err)
	// co up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	fmt.Println(wd, projectDir)
	err = copyFile(projectDir+"/x/ibc-rate-limit/bytecode/rate_limiter.wasm", wd+"/scripts/rate_limiter.wasm")
	assert.NilError(t, err)

	node.StoreWasmCode("rate_limiter.wasm", initialization.ValidatorWalletName)
	chainA.LatestCodeId = int(node.QueryLatestWasmCodeID())
	node.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeId),
		fmt.Sprintf(`{"gov_module": "%s", "ibc_module": "%s", "paths": [{"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"testQuota", "duration": 86400, "send_recv": [1, 1]}] } ] }`, node.PublicAddress, node.PublicAddress, initialization.OsmoToken.Denom),
		initialization.ValidatorWalletName)

	contracts, err := node.QueryContractsFromId(chainA.LatestCodeId)
	assert.NilError(t, err)
	assert.Equal(t, len(contracts), 1, "Wrong number of contracts for the rate limiter")

	proposal := paramsutils.ParamChangeProposalJSON{
		Title:       "Param Change",
		Description: "Changing the rate limit contract param",
		Changes: paramsutils.ParamChangesJSON{
			paramsutils.ParamChangeJSON{
				Subspace: ibcratelimittypes.ModuleName,
				Key:      "contract",
				Value:    []byte(fmt.Sprintf(`"%s"`, contracts[0])),
			},
		},
		Deposit: "625000000uosmo",
	}
	proposalJson, err := json.Marshal(proposal)
	assert.NilError(t, err)

	node.SubmitParamChangeProposal(string(proposalJson), initialization.ValidatorWalletName)
	chainA.LatestProposalNumber += 1

	for _, n := range chainA.NodeConfigs {
		n.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}

	// The value is returned as a string, so we have to unmarshal twice
	type Params struct {
		Key      string `json:"key"`
		Subspace string `json:"subspace"`
		Value    string `json:"value"`
	}

	require.Eventually(
		t,
		func() bool {
			var params Params
			node.QueryParams(ibcratelimittypes.ModuleName, "contract", &params)
			var val string
			err := json.Unmarshal([]byte(params.Value), &val)
			if err != nil {
				return false
			}
			return val == contracts[0]
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve params",
	)

	// Sending <1%. Should work
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, sdk.NewInt64Coin(initialization.OsmoDenom, 1))
	// Sending >1%. Should fail
	node.FailIBCTransfer(initialization.ValidatorWalletName, chainB.NodeConfigs[0].PublicAddress, fmt.Sprintf("%duosmo", int(over)))

	// Removing the rate limit so it doesn't affect other tests
	node.WasmExecute(contracts[0], `{"remove_path": {"channel_id": "channel-0", "denom": "uosmo"}}`, initialization.ValidatorWalletName)
}

func TestLargeWasmUpload(t *testing.T) {
	fmt.Println("TestLargeWasmUpload", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	chainA := f.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	assert.NilError(t, err)
	node.StoreWasmCode("large.wasm", initialization.ValidatorWalletName)
}

func TestIBCWasmHooks(t *testing.T) {
	fmt.Println("TestIBCWasmHooks", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	if f.skipIBC {
		t.Skip("Skipping IBC tests")
	}
	chainA := f.configurer.GetChainConfig(0)
	chainB := f.configurer.GetChainConfig(1)

	nodeA, err := chainA.GetDefaultNode()
	assert.NilError(t, err)
	nodeB, err := chainB.GetDefaultNode()
	assert.NilError(t, err)

	// copy the contract from x/rate-limit/testdata/
	wd, err := os.Getwd()
	assert.NilError(t, err)
	// co up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	err = copyFile(projectDir+"/tests/ibc-hooks/bytecode/counter.wasm", wd+"/scripts/counter.wasm")
	assert.NilError(t, err)

	nodeA.StoreWasmCode("counter.wasm", initialization.ValidatorWalletName)
	chainA.LatestCodeId = int(nodeA.QueryLatestWasmCodeID())
	nodeA.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeId),
		`{"count": 0}`,
		initialization.ValidatorWalletName)

	contracts, err := nodeA.QueryContractsFromId(chainA.LatestCodeId)
	assert.NilError(t, err)
	assert.Equal(t, len(contracts), 1, "Wrong number of contracts for the counter")
	contractAddr := contracts[0]

	transferAmount := int64(10)
	validatorAddr := nodeB.GetWallet(initialization.ValidatorWalletName)
	nodeB.SendIBCTransfer(validatorAddr, contractAddr, fmt.Sprintf("%duosmo", transferAmount),
		fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr))

	// check the balance of the contract
	require.Eventually(
		t,
		func() bool {
			balance, err := nodeA.QueryBalances(contractAddr)
			assert.NilError(t, err)
			if len(balance) == 0 {
				return false
			}
			return balance[0].Amount.Int64() == transferAmount
		},
		1*time.Minute,
		10*time.Millisecond,
	)

	// sender wasm addr
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender("channel-0", validatorAddr, "osmo")

	var response map[string]interface{}
	require.Eventually(
		t,
		func() bool {
			response, err = nodeA.QueryWasmSmart(contractAddr, fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderBech32))
			totalFunds := response["total_funds"].([]interface{})[0]
			amount := totalFunds.(map[string]interface{})["amount"].(string)
			denom := totalFunds.(map[string]interface{})["denom"].(string)
			// check if denom contains "uosmo"
			return err == nil && amount == strconv.FormatInt(transferAmount, 10) && strings.Contains(denom, "ibc")
		},
		15*time.Second,
		10*time.Millisecond,
	)
}

// TestAddToExistingLockPostUpgrade ensures addToExistingLock works for locks created preupgrade.
func TestAddToExistingLockPostUpgrade(t *testing.T) {
	fmt.Println("TestAddToExistingLockPostUpgrade", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	if f.skipUpgrade {
		t.Skip("Skipping AddToExistingLockPostUpgrade test")
	}
	chainA := f.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)
	// ensure we can add to existing locks and superfluid locks that existed pre upgrade on chainA
	// we use the hardcoded gamm/pool/1 and these specific wallet names to match what was created pre upgrade
	lockupWalletAddr, lockupWalletSuperfluidAddr := chainANode.GetWallet("lockup-wallet"), chainANode.GetWallet("lockup-wallet-superfluid")
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletAddr)
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletSuperfluidAddr)
}

// TestAddToExistingLock tests lockups to both regular and superfluid locks.
func TestAddToExistingLock(t *testing.T) {
	fmt.Println("TestAddToExistingLock", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	chainA := f.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)
	// ensure we can add to new locks and superfluid locks
	// create pool and enable superfluid assets
	poolId := chainANode.CreateBalancerPool("nativeDenomPool.json", chainA.NodeConfigs[0].PublicAddress)
	chainA.EnableSuperfluidAsset(fmt.Sprintf("gamm/pool/%d", poolId))

	// setup wallets and send gamm tokens to these wallets on chainA
	lockupWalletAddr, lockupWalletSuperfluidAddr := chainANode.CreateWallet("TestAddToExistingLock"), chainANode.CreateWallet("TestAddToExistingLockSuperfluid")
	chainANode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), chainA.NodeConfigs[0].PublicAddress, lockupWalletAddr)
	chainANode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), chainA.NodeConfigs[0].PublicAddress, lockupWalletSuperfluidAddr)

	// ensure we can add to new locks and superfluid locks on chainA
	chainA.LockAndAddToExistingLock(sdk.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId), lockupWalletAddr, lockupWalletSuperfluidAddr)
}

// TestArithmeticTWAP tests TWAP by creating a pool, performing a swap.
// These two operations should create TWAP records.
// Then, we wait until the epoch for the records to be pruned.
// The records are guranteed to be pruned at the next epoch
// because twap keep time = epoch time / 4 and we use a timer
// to wait for at least the twap keep time.
func TestArithmeticTWAP(t *testing.T) {
	fmt.Println("TestArithmeticTWAP", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	const (
		poolFile   = "nativeDenomThreeAssetPool.json"
		walletName = "arithmetic-twap-wallet"

		denomA = "stake"
		denomB = "uion"
		denomC = "uosmo"

		minAmountOut = "1"

		epochIdentifier = "day"
	)

	coinAIn, coinBIn, coinCIn := fmt.Sprintf("2000000%s", denomA), fmt.Sprintf("2000000%s", denomB), fmt.Sprintf("2000000%s", denomC)

	chainA := f.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	// Triggers the creation of TWAP records.
	poolId := chainANode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainANode.CreateWallet(walletName)

	timeBeforeSwap := chainANode.QueryLatestBlockTime()
	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainA.WaitForNumHeights(1)

	t.Log("querying for the first TWAP to now before swap")
	twapFromBeforeSwapToBeforeSwapOneAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	assert.NilError(t, err)
	twapFromBeforeSwapToBeforeSwapOneBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	assert.NilError(t, err)
	twapFromBeforeSwapToBeforeSwapOneCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	assert.NilError(t, err)

	chainANode.BankSend(coinAIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)
	chainANode.BankSend(coinBIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)
	chainANode.BankSend(coinCIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)

	t.Log("querying for the second TWAP to now before swap, must equal to first")
	twapFromBeforeSwapToBeforeSwapTwoAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap.Add(50*time.Millisecond))
	assert.NilError(t, err)
	twapFromBeforeSwapToBeforeSwapTwoBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap.Add(50*time.Millisecond))
	assert.NilError(t, err)
	twapFromBeforeSwapToBeforeSwapTwoCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap.Add(50*time.Millisecond))
	assert.NilError(t, err)

	// Since there were no swaps between the two queries, the TWAPs should be the same.
	osmoassert.DecApproxEq(t, twapFromBeforeSwapToBeforeSwapOneAB, twapFromBeforeSwapToBeforeSwapTwoAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(t, twapFromBeforeSwapToBeforeSwapOneBC, twapFromBeforeSwapToBeforeSwapTwoBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(t, twapFromBeforeSwapToBeforeSwapOneCA, twapFromBeforeSwapToBeforeSwapTwoCA, sdk.NewDecWithPrec(1, 3))

	t.Log("performing swaps")
	chainANode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)
	chainANode.SwapExactAmountIn(coinBIn, minAmountOut, fmt.Sprintf("%d", poolId), denomC, swapWalletAddr)
	chainANode.SwapExactAmountIn(coinCIn, minAmountOut, fmt.Sprintf("%d", poolId), denomA, swapWalletAddr)

	keepPeriodCountDown := time.NewTimer(initialization.TWAPPruningKeepPeriod)

	// Make sure that we are still producing blocks and move far enough for the swap TWAP record to be created
	// so that we can measure start time post-swap (timeAfterSwap).
	chainA.WaitForNumHeights(2)

	// Measure time after swap and wait for a few blocks to be produced.
	// This is needed to ensure that start time is before the block time
	// when we query for TWAP.
	timeAfterSwap := chainANode.QueryLatestBlockTime()
	chainA.WaitForNumHeights(2)

	// TWAP "from before to after swap" should be different from "from before to before swap"
	// because swap introduces a new record with a different spot price.
	t.Log("querying for the TWAP from before swap to now after swap")
	twapFromBeforeSwapToAfterSwapAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	assert.NilError(t, err)
	twapFromBeforeSwapToAfterSwapBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	assert.NilError(t, err)
	twapFromBeforeSwapToAfterSwapCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	assert.NilError(t, err)
	// We had a swap of 2000000stake for some amount of uion,
	// 2000000uion for some amount of uosmo, and
	// 2000000uosmo for some amount of stake
	// Because we traded the same amount of all three assets, we expect the asset with the greatest
	// initial value (B, or uion) to have a largest negative price impact,
	// to the benefit (positive price impact) of the other two assets (A&C, or stake and uosmo)
	require.True(t, twapFromBeforeSwapToAfterSwapAB.GT(twapFromBeforeSwapToBeforeSwapOneAB))
	require.True(t, twapFromBeforeSwapToAfterSwapBC.LT(twapFromBeforeSwapToBeforeSwapOneBC))
	require.True(t, twapFromBeforeSwapToAfterSwapCA.GT(twapFromBeforeSwapToBeforeSwapOneCA))

	t.Log("querying for the TWAP from after swap to now")
	twapFromAfterToNowAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeAfterSwap)
	assert.NilError(t, err)
	twapFromAfterToNowBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeAfterSwap)
	assert.NilError(t, err)
	twapFromAfterToNowCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeAfterSwap)
	assert.NilError(t, err)
	// Because twapFromAfterToNow has a higher time weight for the after swap period,
	// we expect the results to be flipped from the previous comparison to twapFromBeforeSwapToBeforeSwapOne
	require.True(t, twapFromBeforeSwapToAfterSwapAB.LT(twapFromAfterToNowAB))
	require.True(t, twapFromBeforeSwapToAfterSwapBC.GT(twapFromAfterToNowBC))
	require.True(t, twapFromBeforeSwapToAfterSwapCA.LT(twapFromAfterToNowCA))

	t.Log("querying for the TWAP from after swap to after swap + 10ms")
	twapAfterSwapBeforePruning10MsAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	twapAfterSwapBeforePruning10MsBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	twapAfterSwapBeforePruning10MsCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	// Again, because twapAfterSwapBeforePruning10Ms has a higher time weight for the after swap period between the two,
	// we expect no change in the inequality
	require.True(t, twapFromBeforeSwapToAfterSwapAB.LT(twapAfterSwapBeforePruning10MsAB))
	require.True(t, twapFromBeforeSwapToAfterSwapBC.GT(twapAfterSwapBeforePruning10MsBC))
	require.True(t, twapFromBeforeSwapToAfterSwapCA.LT(twapAfterSwapBeforePruning10MsCA))

	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(t, twapAfterSwapBeforePruning10MsAB, twapFromAfterToNowAB, sdk.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(t, twapAfterSwapBeforePruning10MsBC, twapFromAfterToNowBC, sdk.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(t, twapAfterSwapBeforePruning10MsCA, twapFromAfterToNowCA, sdk.NewDecWithPrec(2, 3))

	// Make sure that the pruning keep period has passed.
	t.Logf("waiting for pruning keep period of (%.f) seconds to pass", initialization.TWAPPruningKeepPeriod.Seconds())
	<-keepPeriodCountDown.C

	// Epoch end triggers the prunning of TWAP records.
	// Records before swap should be pruned.
	chainA.WaitForNumEpochs(1, epochIdentifier)

	// We should not be able to get TWAP before swap since it should have been pruned.
	t.Log("pruning is now complete, querying TWAP for period that should be pruned")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	require.ErrorContains(t, err, "too old")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	require.ErrorContains(t, err, "too old")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	require.ErrorContains(t, err, "too old")

	// TWAPs for the same time range should be the same when we query for them before and after pruning.
	t.Log("querying for TWAP for period before pruning took place but should not have been pruned")
	twapAfterPruning10msAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	twapAfterPruning10msBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	twapAfterPruning10msCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	assert.NilError(t, err)
	require.Equal(t, twapAfterSwapBeforePruning10MsAB, twapAfterPruning10msAB)
	require.Equal(t, twapAfterSwapBeforePruning10MsBC, twapAfterPruning10msBC)
	require.Equal(t, twapAfterSwapBeforePruning10MsCA, twapAfterPruning10msCA)

	// TWAP "from after to after swap" should equal to "from after swap to after pruning"
	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	timeAfterPruning := chainANode.QueryLatestBlockTime()
	t.Log("querying for TWAP from after swap to after pruning")
	twapToNowPostPruningAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterPruning)
	assert.NilError(t, err)
	twapToNowPostPruningBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterPruning)
	assert.NilError(t, err)
	twapToNowPostPruningCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterPruning)
	assert.NilError(t, err)
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(t, twapToNowPostPruningAB, twapAfterSwapBeforePruning10MsAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(t, twapToNowPostPruningBC, twapAfterSwapBeforePruning10MsBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(t, twapToNowPostPruningCA, twapAfterSwapBeforePruning10MsCA, sdk.NewDecWithPrec(1, 3))
}

func TestStateSync(t *testing.T) {
	fmt.Println("TestStateSync", time.Now())
	t.Parallel()
	f := InitIntegrationFixture(t)

	if f.skipStateSync {
		t.Skip()
	}

	chainA := f.configurer.GetChainConfig(0)
	runningNode, err := chainA.GetDefaultNode()
	assert.NilError(t, err)

	persistentPeers := chainA.GetPersistentPeers()

	stateSyncHostPort := fmt.Sprintf("%s:26657", runningNode.Name)
	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

	// get trust height and trust hash.
	trustHeight, err := runningNode.QueryCurrentHeight()
	assert.NilError(t, err)

	trustHash, err := runningNode.QueryHashFromBlock(trustHeight)
	assert.NilError(t, err)

	stateSynchingNodeConfig := &initialization.NodeConfig{
		Name:               "state-sync",
		Pruning:            "default",
		PruningKeepRecent:  "0",
		PruningInterval:    "0",
		SnapshotInterval:   1500,
		SnapshotKeepRecent: 2,
	}

	tempDir, err := os.MkdirTemp("", "osmosis-e2e-statesync-")
	assert.NilError(t, err)

	// configure genesis and config files for the state-synchin node.
	nodeInit, err := initialization.InitSingleNode(
		chainA.Id,
		tempDir,
		filepath.Join(runningNode.ConfigDir, "config", "genesis.json"),
		stateSynchingNodeConfig,
		time.Duration(chainA.VotingPeriod),
		// time.Duration(chainA.ExpeditedVotingPeriod),
		trustHeight,
		trustHash,
		stateSyncRPCServers,
		persistentPeers,
	)
	assert.NilError(t, err)

	stateSynchingNode := chainA.CreateNode(nodeInit)

	// ensure that the running node has snapshots at a height > trustHeight.
	hasSnapshotsAvailable := func(syncInfo coretypes.SyncInfo) bool {
		snapshotHeight := runningNode.SnapshotInterval
		if uint64(syncInfo.LatestBlockHeight) < snapshotHeight {
			t.Logf("snapshot height is not reached yet, current (%d), need (%d)", syncInfo.LatestBlockHeight, snapshotHeight)
			return false
		}

		snapshots, err := runningNode.QueryListSnapshots()
		assert.NilError(t, err)

		for _, snapshot := range snapshots {
			if snapshot.Height > uint64(trustHeight) {
				t.Log("found state sync snapshot after trust height")
				return true
			}
		}
		t.Log("state sync snashot after trust height is not found")
		return false
	}
	runningNode.WaitUntil(hasSnapshotsAvailable)

	// start the state synchin node.
	err = stateSynchingNode.Run()
	assert.NilError(t, err)

	// ensure that the state synching node cathes up to the running node.
	require.Eventually(
		t,
		func() bool {
			stateSyncNodeHeight, err := stateSynchingNode.QueryCurrentHeight()
			assert.NilError(t, err)
			runningNodeHeight, err := runningNode.QueryCurrentHeight()
			assert.NilError(t, err)
			return stateSyncNodeHeight == runningNodeHeight
		},
		3*time.Minute,
		500*time.Millisecond,
	)

	// stop the state synching node.
	err = chainA.RemoveNode(stateSynchingNode.Name)
	assert.NilError(t, err)
}
