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
)

func (s *IntegrationTestSuite) TestConcentratedLiquidity() {
	chainA := s.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	// assert contents of the pool are valid
	s.Require().Equal(concentratedPool.GetId(), poolID)
	s.Require().Equal(concentratedPool.GetToken0(), denom0)
	s.Require().Equal(concentratedPool.GetToken1(), denom1)
	s.Require().Equal(concentratedPool.GetTickSpacing(), tickSpacing)
	s.Require().Equal(concentratedPool.GetPrecisionFactorAtPriceOne(), sdk.NewInt(precisionFactorAtPriceOne))
	s.Require().Equal(concentratedPool.GetSwapFee(sdk.Context{}), sdk.MustNewDecFromStr(swapFee))

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
	s.Require().Equal(len(positionsAddress1), 2)
	s.Require().Equal(len(positionsAddress2), 1)
	s.Require().Equal(len(positionsAddress3), 2)

	// Assert returned positions:
	validateCLPosition := func(position types.FullPositionByOwnerResult, poolId uint64, lowerTick, upperTick int64) {
		s.Require().Equal(position.PoolId, poolId)
		s.Require().Equal(position.LowerTick, int64(lowerTick))
		s.Require().Equal(position.UpperTick, int64(upperTick))
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
func (s *IntegrationTestSuite) TestGeometricTwapMigration() {
	if s.skipUpgrade {
		s.T().Skip("Skipping upgrade tests")
	}

	const (
		// Configurations for tests/e2e/scripts/pool1A.json
		// This pool gets initialized pre-upgrade.
		oldPoolId       = 1
		minAmountOut    = "1"
		otherDenom      = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
		migrationWallet = "migration"
	)

	chainA := s.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	s.Require().NoError(err)

	uosmoIn := fmt.Sprintf("1000000%s", "uosmo")

	swapWalletAddr := node.CreateWallet(migrationWallet)

	node.BankSend(uosmoIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)

	// Swap to create new twap records on the pool that was created pre-upgrade.
	node.SwapExactAmountIn(uosmoIn, minAmountOut, fmt.Sprintf("%d", oldPoolId), otherDenom, swapWalletAddr)
}

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
// Additionally, it attempst to create a pool with IBC denoms.
func (s *IntegrationTestSuite) TestIBCTokenTransferAndCreatePool() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.StakeToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.StakeToken)

	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)
	chainANode.CreateBalancerPool("ibcDenomPool.json", initialization.ValidatorWalletName)
}

// TestSuperfluidVoting tests that superfluid voting is functioning as expected.
// It does so by doing the following:
// - creating a pool
// - attempting to submit a proposal to enable superfluid voting in that pool
// - voting yes on the proposal from the validator wallet
// - voting no on the proposal from the delegator wallet
// - ensuring that delegator's wallet overwrites the validator's vote
func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

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

	s.Eventually(
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
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := chainANode.QueryIntermediaryAccount(fmt.Sprintf("gamm/pool/%d", poolId), chainA.NodeConfigs[1].OperatorAddress)
			s.Require().NoError(err)
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

func (s *IntegrationTestSuite) TestIBCTokenTransferRateLimiting() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)

	node, err := chainA.GetDefaultNode()
	s.NoError(err)

	osmoSupply, err := node.QuerySupplyOf("uosmo")
	s.NoError(err)

	f, err := osmoSupply.ToDec().Float64()
	s.NoError(err)

	over := f * 0.02

	// Sending >1%
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, sdk.NewInt64Coin(initialization.OsmoDenom, int64(over)))

	// copy the contract from x/rate-limit/testdata/
	wd, err := os.Getwd()
	s.NoError(err)
	// co up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	fmt.Println(wd, projectDir)
	err = copyFile(projectDir+"/x/ibc-rate-limit/bytecode/rate_limiter.wasm", wd+"/scripts/rate_limiter.wasm")
	s.NoError(err)

	node.StoreWasmCode("rate_limiter.wasm", initialization.ValidatorWalletName)
	chainA.LatestCodeId = int(node.QueryLatestWasmCodeID())
	node.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeId),
		fmt.Sprintf(`{"gov_module": "%s", "ibc_module": "%s", "paths": [{"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"testQuota", "duration": 86400, "send_recv": [1, 1]}] } ] }`, node.PublicAddress, node.PublicAddress, initialization.OsmoToken.Denom),
		initialization.ValidatorWalletName)

	contracts, err := node.QueryContractsFromId(chainA.LatestCodeId)
	s.NoError(err)
	s.Require().Len(contracts, 1, "Wrong number of contracts for the rate limiter")

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
	s.NoError(err)

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

	s.Eventually(
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

func (s *IntegrationTestSuite) TestLargeWasmUpload() {
	chainA := s.configurer.GetChainConfig(0)
	node, err := chainA.GetDefaultNode()
	s.NoError(err)
	node.StoreWasmCode("bytecode/large.wasm", initialization.ValidatorWalletName)
}

func (s *IntegrationTestSuite) TestIBCWasmHooks() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)

	nodeA, err := chainA.GetDefaultNode()
	s.NoError(err)
	nodeB, err := chainB.GetDefaultNode()
	s.NoError(err)

	// copy the contract from x/rate-limit/testdata/
	wd, err := os.Getwd()
	s.NoError(err)
	// co up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	err = copyFile(projectDir+"/tests/ibc-hooks/bytecode/counter.wasm", wd+"/scripts/counter.wasm")
	s.NoError(err)

	nodeA.StoreWasmCode("counter.wasm", initialization.ValidatorWalletName)
	chainA.LatestCodeId = int(nodeA.QueryLatestWasmCodeID())
	nodeA.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeId),
		`{"count": 0}`,
		initialization.ValidatorWalletName)

	contracts, err := nodeA.QueryContractsFromId(chainA.LatestCodeId)
	s.NoError(err)
	s.Require().Len(contracts, 1, "Wrong number of contracts for the counter")
	contractAddr := contracts[0]

	transferAmount := int64(10)
	validatorAddr := nodeB.GetWallet(initialization.ValidatorWalletName)
	nodeB.SendIBCTransfer(validatorAddr, contractAddr, fmt.Sprintf("%duosmo", transferAmount),
		fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr))

	// check the balance of the contract
	s.Eventually(func() bool {
		balance, err := nodeA.QueryBalances(contractAddr)
		s.Require().NoError(err)
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
	s.Eventually(func() bool {
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
func (s *IntegrationTestSuite) TestAddToExistingLockPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("Skipping AddToExistingLockPostUpgrade test")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)
	// ensure we can add to existing locks and superfluid locks that existed pre upgrade on chainA
	// we use the hardcoded gamm/pool/1 and these specific wallet names to match what was created pre upgrade
	lockupWalletAddr, lockupWalletSuperfluidAddr := chainANode.GetWallet("lockup-wallet"), chainANode.GetWallet("lockup-wallet-superfluid")
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletAddr)
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletSuperfluidAddr)
}

// TestAddToExistingLock tests lockups to both regular and superfluid locks.
func (s *IntegrationTestSuite) TestAddToExistingLock() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)
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
func (s *IntegrationTestSuite) TestArithmeticTWAP() {
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

	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	// Triggers the creation of TWAP records.
	poolId := chainANode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainANode.CreateWallet(walletName)

	timeBeforeSwap := chainANode.QueryLatestBlockTime()
	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainA.WaitForNumHeights(1)

	s.T().Log("querying for the first TWAP to now before swap")
	twapFromBeforeSwapToBeforeSwapOneAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().NoError(err)

	chainANode.BankSend(coinAIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)
	chainANode.BankSend(coinBIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)
	chainANode.BankSend(coinCIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)

	s.T().Log("querying for the second TWAP to now before swap, must equal to first")
	twapFromBeforeSwapToBeforeSwapTwoAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)

	// Since there were no swaps between the two queries, the TWAPs should be the same.
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneAB, twapFromBeforeSwapToBeforeSwapTwoAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneBC, twapFromBeforeSwapToBeforeSwapTwoBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneCA, twapFromBeforeSwapToBeforeSwapTwoCA, sdk.NewDecWithPrec(1, 3))

	s.T().Log("performing swaps")
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
	s.T().Log("querying for the TWAP from before swap to now after swap")
	twapFromBeforeSwapToAfterSwapAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().NoError(err)
	// We had a swap of 2000000stake for some amount of uion,
	// 2000000uion for some amount of uosmo, and
	// 2000000uosmo for some amount of stake
	// Because we traded the same amount of all three assets, we expect the asset with the greatest
	// initial value (B, or uion) to have a largest negative price impact,
	// to the benefit (positive price impact) of the other two assets (A&C, or stake and uosmo)
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.GT(twapFromBeforeSwapToBeforeSwapOneAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.LT(twapFromBeforeSwapToBeforeSwapOneBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.GT(twapFromBeforeSwapToBeforeSwapOneCA))

	s.T().Log("querying for the TWAP from after swap to now")
	twapFromAfterToNowAB, err := chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowBC, err := chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowCA, err := chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeAfterSwap)
	s.Require().NoError(err)
	// Because twapFromAfterToNow has a higher time weight for the after swap period,
	// we expect the results to be flipped from the previous comparison to twapFromBeforeSwapToBeforeSwapOne
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.LT(twapFromAfterToNowAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.GT(twapFromAfterToNowBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.LT(twapFromAfterToNowCA))

	s.T().Log("querying for the TWAP from after swap to after swap + 10ms")
	twapAfterSwapBeforePruning10MsAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	// Again, because twapAfterSwapBeforePruning10Ms has a higher time weight for the after swap period between the two,
	// we expect no change in the inequality
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.LT(twapAfterSwapBeforePruning10MsAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.GT(twapAfterSwapBeforePruning10MsBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.LT(twapAfterSwapBeforePruning10MsCA))

	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsAB, twapFromAfterToNowAB, sdk.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsBC, twapFromAfterToNowBC, sdk.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsCA, twapFromAfterToNowCA, sdk.NewDecWithPrec(2, 3))

	// Make sure that the pruning keep period has passed.
	s.T().Logf("waiting for pruning keep period of (%.f) seconds to pass", initialization.TWAPPruningKeepPeriod.Seconds())
	<-keepPeriodCountDown.C

	// Epoch end triggers the prunning of TWAP records.
	// Records before swap should be pruned.
	chainA.WaitForNumEpochs(1, epochIdentifier)

	// We should not be able to get TWAP before swap since it should have been pruned.
	s.T().Log("pruning is now complete, querying TWAP for period that should be pruned")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainANode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")

	// TWAPs for the same time range should be the same when we query for them before and after pruning.
	s.T().Log("querying for TWAP for period before pruning took place but should not have been pruned")
	twapAfterPruning10msAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	s.Require().Equal(twapAfterSwapBeforePruning10MsAB, twapAfterPruning10msAB)
	s.Require().Equal(twapAfterSwapBeforePruning10MsBC, twapAfterPruning10msBC)
	s.Require().Equal(twapAfterSwapBeforePruning10MsCA, twapAfterPruning10msCA)

	// TWAP "from after to after swap" should equal to "from after swap to after pruning"
	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	timeAfterPruning := chainANode.QueryLatestBlockTime()
	s.T().Log("querying for TWAP from after swap to after pruning")
	twapToNowPostPruningAB, err := chainANode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningBC, err := chainANode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningCA, err := chainANode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningAB, twapAfterSwapBeforePruning10MsAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningBC, twapAfterSwapBeforePruning10MsBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningCA, twapAfterSwapBeforePruning10MsCA, sdk.NewDecWithPrec(1, 3))
}

func (s *IntegrationTestSuite) TestStateSync() {
	if s.skipStateSync {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0)
	runningNode, err := chainA.GetDefaultNode()
	s.Require().NoError(err)

	persistentPeers := chainA.GetPersistentPeers()

	stateSyncHostPort := fmt.Sprintf("%s:26657", runningNode.Name)
	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

	// get trust height and trust hash.
	trustHeight, err := runningNode.QueryCurrentHeight()
	s.Require().NoError(err)

	trustHash, err := runningNode.QueryHashFromBlock(trustHeight)
	s.Require().NoError(err)

	stateSynchingNodeConfig := &initialization.NodeConfig{
		Name:               "state-sync",
		Pruning:            "default",
		PruningKeepRecent:  "0",
		PruningInterval:    "0",
		SnapshotInterval:   1500,
		SnapshotKeepRecent: 2,
	}

	tempDir, err := os.MkdirTemp("", "osmosis-e2e-statesync-")
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	stateSynchingNode := chainA.CreateNode(nodeInit)

	// ensure that the running node has snapshots at a height > trustHeight.
	hasSnapshotsAvailable := func(syncInfo coretypes.SyncInfo) bool {
		snapshotHeight := runningNode.SnapshotInterval
		if uint64(syncInfo.LatestBlockHeight) < snapshotHeight {
			s.T().Logf("snapshot height is not reached yet, current (%d), need (%d)", syncInfo.LatestBlockHeight, snapshotHeight)
			return false
		}

		snapshots, err := runningNode.QueryListSnapshots()
		s.Require().NoError(err)

		for _, snapshot := range snapshots {
			if snapshot.Height > uint64(trustHeight) {
				s.T().Log("found state sync snapshot after trust height")
				return true
			}
		}
		s.T().Log("state sync snashot after trust height is not found")
		return false
	}
	runningNode.WaitUntil(hasSnapshotsAvailable)

	// start the state synchin node.
	err = stateSynchingNode.Run()
	s.Require().NoError(err)

	// ensure that the state synching node cathes up to the running node.
	s.Require().Eventually(func() bool {
		stateSyncNodeHeight, err := stateSynchingNode.QueryCurrentHeight()
		s.Require().NoError(err)
		runningNodeHeight, err := runningNode.QueryCurrentHeight()
		s.Require().NoError(err)
		return stateSyncNodeHeight == runningNodeHeight
	},
		3*time.Minute,
		500*time.Millisecond,
	)

	// stop the state synching node.
	err = chainA.RemoveNode(stateSynchingNode.Name)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestExpeditedProposals() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	chainANode.SubmitTextProposal("expedited text proposal", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinExpeditedDeposit)), true)
	chainA.LatestProposalNumber += 1
	chainANode.DepositProposal(chainA.LatestProposalNumber, true)
	totalTimeChan := make(chan time.Duration, 1)
	go chainANode.QueryPropStatusTimed(chainA.LatestProposalNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)
	for _, node := range chainA.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}
	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	var elapsed time.Duration
	timeoutPeriod := time.Duration(2 * time.Minute)
	select {
	case elapsed = <-totalTimeChan:
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	}

	// compare the time it took to reach pass status to expected expedited voting period
	expeditedVotingPeriodDuration := time.Duration(chainA.ExpeditedVotingPeriod * float32(time.Second))
	timeDelta := elapsed - expeditedVotingPeriodDuration
	// ensure delta is within two seconds of expected time
	s.Require().Less(timeDelta, 2*time.Second)
	s.T().Logf("expeditedVotingPeriodDuration within two seconds of expected time: %v", timeDelta)
	close(totalTimeChan)
}

// TestGeometricTWAP tests geometric twap.
// It does the following:  creates a pool, queries twap, performs a swap , and queries twap again.
// Twap is expected to change after the swap.
// The pool is created with 1_000_000 uosmo and 2_000_000 stake and equal weights.
// Assuming base asset is uosmo, the initial twap is 2
// Upon swapping 1_000_000 uosmo for stake, supply changes, making uosmo less expensive.
// As a result of the swap, twap changes to 0.5.
func (s *IntegrationTestSuite) TestGeometricTWAP() {
	const (
		// This pool contains 1_000_000 uosmo and 2_000_000 stake.
		// Equals weights.
		poolFile   = "geometricPool.json"
		walletName = "geometric-twap-wallet"

		denomA = "uosmo" // 1_000_000 uosmo
		denomB = "stake" // 2_000_000 stake

		minAmountOut = "1"

		epochIdentifier = "day"
	)

	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	// Triggers the creation of TWAP records.
	poolId := chainANode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainANode.CreateWallet(walletName)

	// We add 5 ms to avoid landing directly on block time in twap. If block time
	// is provided as start time, the latest spot price is used. Otherwise
	// interpolation is done.
	timeBeforeSwapPlus5ms := chainANode.QueryLatestBlockTime().Add(5 * time.Millisecond)
	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainA.WaitForNumHeights(1)

	s.T().Log("querying for the first geometric TWAP to now (before swap)")
	// Assume base = uosmo, quote = stake
	// At pool creation time, the twap should be:
	// quote assset supply / base asset supply = 2_000_000 / 1_000_000 = 2
	initialTwapBOverA, err := chainANode.QueryGeometricTwapToNow(poolId, denomA, denomB, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(2), initialTwapBOverA)

	// Assume base = stake, quote = uosmo
	// At pool creation time, the twap should be:
	// quote assset supply / base asset supply = 1_000_000 / 2_000_000 = 0.5
	initialTwapAOverB, err := chainANode.QueryGeometricTwapToNow(poolId, denomB, denomA, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDecWithPrec(5, 1), initialTwapAOverB)

	coinAIn := fmt.Sprintf("1000000%s", denomA)
	chainANode.BankSend(coinAIn, chainA.NodeConfigs[0].PublicAddress, swapWalletAddr)

	s.T().Logf("performing swap of %s for %s", coinAIn, denomB)

	// stake out = stake supply * (1 - (uosmo supply before / uosmo supply after)^(uosmo weight / stake weight))
	//           = 2_000_000 * (1 - (1_000_000 / 2_000_000)^1)
	//           = 2_000_000 * 0.5
	//           = 1_000_000
	chainANode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)

	// New supply post swap:
	// stake = 2_000_000 - 1_000_000 - 1_000_000
	// uosmo = 1_000_000 + 1_000_000 = 2_000_000

	timeAfterSwap := chainANode.QueryLatestBlockTime()
	chainA.WaitForNumHeights(1)
	timeAfterSwapPlus1Height := chainANode.QueryLatestBlockTime()

	s.T().Log("querying for the TWAP from after swap to now")
	afterSwapTwapBOverA, err := chainANode.QueryGeometricTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwapPlus1Height)
	s.Require().NoError(err)

	// We swap uosmo so uosmo's supply will increase and stake will decrease.
	// The the price after will be smaller than the previous one.
	s.Require().True(initialTwapBOverA.GT(afterSwapTwapBOverA))

	// Assume base = uosmo, quote = stake
	// At pool creation, we had:
	// quote assset supply / base asset supply = 2_000_000 / 1_000_000 = 2
	// Next, we swapped 1_000_000 uosmo for stake.
	// Now, we roughly have
	// uatom = 1_000_000
	// uosmo = 2_000_000
	// quote assset supply / base asset supply = 1_000_000 / 2_000_000 = 0.5
	osmoassert.DecApproxEq(s.T(), sdk.NewDecWithPrec(5, 1), afterSwapTwapBOverA, sdk.NewDecWithPrec(1, 2))
}
