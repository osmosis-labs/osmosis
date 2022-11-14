package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	appparams "github.com/osmosis-labs/osmosis/v12/app/params"
	"github.com/osmosis-labs/osmosis/v12/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v12/tests/e2e/initialization"
)

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
	chainANode.CreatePool("ibcDenomPool.json", initialization.ValidatorWalletName)
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

	poolId := chainANode.CreatePool("nativeDenomPool.json", chainA.NodeConfigs[0].PublicAddress)

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

	// balance, err := node.QueryBalances(chainA.NodeConfigs[1].PublicAddress)
	// s.NoError(err)

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
	chainA.LatestCodeId += 1
	node.InstantiateWasmContract(
		strconv.Itoa(chainA.LatestCodeId),
		fmt.Sprintf(`{"gov_module": "%s", "ibc_module": "%s", "paths": [{"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"testQuota", "duration": 86400, "send_recv": [1, 1]}] } ] }`, node.PublicAddress, node.PublicAddress, initialization.OsmoToken.Denom),
		initialization.ValidatorWalletName)

	// Using code_id 1 because this is the only contract right now. This may need to change if more contracts are added
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
			return val != ""
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
	poolId := chainANode.CreatePool("nativeDenomPool.json", chainA.NodeConfigs[0].PublicAddress)
	chainA.EnableSuperfluidAsset(fmt.Sprintf("gamm/pool/%d", poolId))

	// setup wallets and send gamm tokens to these wallets on chainA
	lockupWalletAddr, lockupWalletSuperfluidAddr := chainANode.CreateWallet("TestAddToExistingLock"), chainANode.CreateWallet("TestAddToExistingLockSuperfluid")
	chainANode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), chainA.NodeConfigs[0].PublicAddress, lockupWalletAddr)
	chainANode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), chainA.NodeConfigs[0].PublicAddress, lockupWalletSuperfluidAddr)

	// ensure we can add to new locks and superfluid locks on chainA
	chainA.LockAndAddToExistingLock(sdk.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId), lockupWalletAddr, lockupWalletSuperfluidAddr)
}

// TestTWAP tests TWAP by creating a pool, performing a swap.
// These two operations should create TWAP records.
// Then, we wait until the epoch for the records to be pruned.
// The records are guranteed to be pruned at the next epoch
// because twap keep time = epoch time / 4 and we use a timer
// to wait for at least the twap keep time.
func (s *IntegrationTestSuite) TestTWAP() {
	const (
		poolFile   = "nativeDenomThreeAssetPool.json"
		walletName = "swap-exact-amount-in-wallet"

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
	poolId := chainANode.CreatePool(poolFile, initialization.ValidatorWalletName)
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
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsAB, twapFromAfterToNowAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsBC, twapFromAfterToNowBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsCA, twapFromAfterToNowCA, sdk.NewDecWithPrec(1, 3))

	if !s.skipUpgrade {
		// TODO: we should reduce the pruning time in the v11
		// genesis to make this test run faster
		// Currenty, we are testing the upgrade from v11 to v12,
		// the pruning time is set to whatever is in the upgrade
		// handler (two days). Therefore, we cannot reasonably
		// test twap pruning post-upgrade.
		s.T().Skip("skipping TWAP Pruning test. This can be re-enabled post v12")
	}

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
