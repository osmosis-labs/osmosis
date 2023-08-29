package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/address"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/iancoleman/orderedmap"

	packetforwardingtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v19/x/ibc-rate-limit/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	appparams "github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v19/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v19/tests/e2e/initialization"
	clmath "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

var (
	// minDecTolerance minimum tolerance for sdk.Dec, given its precision of 18.
	minDecTolerance = sdk.MustNewDecFromStr("0.000000000000000001")
)

// TODO: Find more scalable way to do this
func (s *IntegrationTestSuite) TestAllE2E() {
	// Reset the default taker fee to 0.15%, so we can actually run tests with it activated
	s.T().Run("SetDefaultTakerFeeBothChains", func(t *testing.T) {
		s.T().Log("resetting the default taker fee to 0.15%")
		s.SetDefaultTakerFeeBothChains()
	})

	// Zero Dependent Tests
	s.T().Run("CreateConcentratedLiquidityPoolVoting_And_TWAP", func(t *testing.T) {
		t.Parallel()
		s.CreateConcentratedLiquidityPoolVoting_And_TWAP()
	})

	s.T().Run("ProtoRev", func(t *testing.T) {
		t.Parallel()
		s.ProtoRev()
	})

	s.T().Run("ConcentratedLiquidity", func(t *testing.T) {
		t.Parallel()
		s.ConcentratedLiquidity()
	})

	s.T().Run("SuperfluidVoting", func(t *testing.T) {
		t.Parallel()
		s.SuperfluidVoting()
	})

	s.T().Run("AddToExistingLock", func(t *testing.T) {
		t.Parallel()
		s.AddToExistingLock()
	})

	s.T().Run("ExpeditedProposals", func(t *testing.T) {
		t.Parallel()
		s.ExpeditedProposals()
	})

	s.T().Run("GeometricTWAP", func(t *testing.T) {
		t.Parallel()
		s.GeometricTWAP()
	})

	s.T().Run("LargeWasmUpload", func(t *testing.T) {
		t.Parallel()
		s.LargeWasmUpload()
	})

	// Test currently disabled
	// s.T().Run("ArithmeticTWAP", func(t *testing.T) {
	// 	t.Parallel()
	// 	s.ArithmeticTWAP()
	// })

	// State Sync Dependent Tests

	if s.skipStateSync {
		s.T().Skip()
	} else {
		s.T().Run("StateSync", func(t *testing.T) {
			t.Parallel()
			s.StateSync()
		})
	}

	// Upgrade Dependent Tests

	if s.skipUpgrade {
		s.T().Skip("Skipping StableSwapPostUpgrade test")
	} else {
		s.T().Run("StableSwapPostUpgrade", func(t *testing.T) {
			t.Parallel()
			s.StableSwapPostUpgrade()
		})
	}

	if s.skipUpgrade {
		s.T().Skip("Skipping GeometricTwapMigration test")
	} else {
		s.T().Run("GeometricTwapMigration", func(t *testing.T) {
			t.Parallel()
			s.GeometricTwapMigration()
		})
	}

	if s.skipUpgrade {
		s.T().Skip("Skipping AddToExistingLockPostUpgrade test")
	} else {
		s.T().Run("AddToExistingLockPostUpgrade", func(t *testing.T) {
			t.Parallel()
			s.AddToExistingLockPostUpgrade()
		})
	}

	// IBC Dependent Tests

	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	} else {
		s.T().Run("IBCTokenTransferRateLimiting", func(t *testing.T) {
			t.Parallel()
			s.IBCTokenTransferRateLimiting()
		})
	}

	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	} else {
		s.T().Run("IBCTokenTransferAndCreatePool", func(t *testing.T) {
			t.Parallel()
			s.IBCTokenTransferAndCreatePool()
		})
	}

	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	} else {
		s.T().Run("IBCWasmHooks", func(t *testing.T) {
			t.Parallel()
			s.IBCWasmHooks()
		})
	}

	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	} else {
		s.T().Run("PacketForwarding", func(t *testing.T) {
			t.Parallel()
			s.PacketForwarding()
		})
	}
}

// TestProtoRev is a test that ensures that the protorev module is working as expected. In particular, this tests and ensures that:
// 1. The protorev module is correctly configured on init
// 2. The protorev module can correctly back run a swap
// 3. the protorev module correctly tracks statistics
func (s *IntegrationTestSuite) ProtoRev() {
	const (
		poolFile1 = "protorevPool1.json"
		poolFile2 = "protorevPool2.json"
		poolFile3 = "protorevPool3.json"

		walletName = "swap-that-creates-an-arb"

		denomIn      = initialization.LuncIBCDenom
		denomOut     = initialization.UstIBCDenom
		amount       = "10000"
		minAmountOut = "1"

		epochIdentifier = "day"
	)

	// NOTE: Uses chainA since IBC denoms are hard coded.
	chainA, chainANode, err := s.getChainACfgs()
	s.Require().NoError(err)

	sender := chainANode.GetWallet(initialization.ValidatorWalletName)

	// --------------- Module init checks ---------------- //
	// The module should be enabled by default.
	enabled, err := chainANode.QueryProtoRevEnabled()
	s.T().Logf("checking that the protorev module is enabled: %t", enabled)
	s.Require().NoError(err)
	s.Require().True(enabled)

	// The module should have no new hot routes by default.
	hotRoutes, err := chainANode.QueryProtoRevTokenPairArbRoutes()
	s.T().Logf("checking that the protorev module has no new hot routes: %v", hotRoutes)
	s.Require().NoError(err)
	s.Require().Len(hotRoutes, 0)

	// The module should have no trades by default.
	_, err = chainANode.QueryProtoRevNumberOfTrades()
	s.T().Logf("checking that the protorev module has no trades on init: %s", err)
	s.Require().Error(err)

	// The module should have pool weights by default.
	info, err := chainANode.QueryProtoRevInfoByPoolType()
	s.T().Logf("checking that the protorev module has pool info on init: %v", info)
	s.Require().NoError(err)
	s.Require().NotNil(info)

	// The module should have max pool points per tx by default.
	maxPoolPointsPerTx, err := chainANode.QueryProtoRevMaxPoolPointsPerTx()
	s.T().Logf("checking that the protorev module has max pool points per tx on init: %d", maxPoolPointsPerTx)
	s.Require().NoError(err)

	// The module should have max pool points per block by default.
	maxPoolPointsPerBlock, err := chainANode.QueryProtoRevMaxPoolPointsPerBlock()
	s.T().Logf("checking that the protorev module has max pool points per block on init: %d", maxPoolPointsPerBlock)
	s.Require().NoError(err)

	// The module should have only osmosis as a supported base denom by default.
	supportedBaseDenoms, err := chainANode.QueryProtoRevBaseDenoms()
	s.T().Logf("checking that the protorev module has only osmosis as a supported base denom on init: %v", supportedBaseDenoms)
	s.Require().NoError(err)
	s.Require().Len(supportedBaseDenoms, 1)
	s.Require().Equal(supportedBaseDenoms[0].Denom, "uosmo")

	// --------------- Set up for a calculated backrun ---------------- //
	// Create all of the pools that will be used in the test.
	swapPoolId1 := chainANode.CreateBalancerPool(poolFile1, initialization.ValidatorWalletName)
	swapPoolId2 := chainANode.CreateBalancerPool(poolFile2, initialization.ValidatorWalletName)
	swapPoolId3 := chainANode.CreateBalancerPool(poolFile3, initialization.ValidatorWalletName)

	// Wait for the creation to be propogated to the other nodes + for the protorev module to
	// correctly update the highest liquidity pools.
	s.T().Logf("waiting for the protorev module to update the highest liquidity pools (wait %.f sec) after the week epoch duration", initialization.EpochDayDuration.Seconds())
	chainA.WaitForNumEpochs(1, epochIdentifier)

	// Create a wallet to use for the swap txs.
	swapWalletAddr := chainANode.CreateWallet(walletName, chainA)
	coinIn := fmt.Sprintf("%s%s", amount, denomIn)
	chainANode.BankSend(coinIn, sender, swapWalletAddr)

	// Check supplies before swap.
	supplyBefore, err := chainANode.QuerySupply()
	s.Require().NoError(err)
	s.Require().NotNil(supplyBefore)

	// Performing the swap that creates a cyclic arbitrage opportunity.
	s.T().Logf("performing a swap that creates a cyclic arbitrage opportunity")
	chainANode.SwapExactAmountIn(coinIn, minAmountOut, fmt.Sprintf("%d", swapPoolId2), denomOut, swapWalletAddr)

	// --------------- Module checks after a calculated backrun ---------------- //
	// Check that the supplies have not changed.
	s.T().Logf("checking that the supplies have not changed")
	supplyAfter, err := chainANode.QuerySupply()
	s.Require().NoError(err)
	s.Require().NotNil(supplyAfter)
	s.Require().Equal(supplyBefore, supplyAfter)

	// Check that the number of trades executed by the protorev module is 1.
	numTrades, err := chainANode.QueryProtoRevNumberOfTrades()
	s.T().Logf("checking that the protorev module has executed 1 trade")
	s.Require().NoError(err)
	s.Require().NotNil(numTrades)
	s.Require().Equal(uint64(1), numTrades.Uint64())

	// Check that the profits of the protorev module are not nil.
	profits, err := chainANode.QueryProtoRevProfits()
	s.T().Logf("checking that the protorev module has non-nil profits: %s", profits)
	s.Require().NoError(err)
	s.Require().NotNil(profits)
	s.Require().Len(profits, 1)

	// Check that the route statistics of the protorev module are not nil.
	routeStats, err := chainANode.QueryProtoRevAllRouteStatistics()
	s.T().Logf("checking that the protorev module has non-nil route statistics: %x", routeStats)
	s.Require().NoError(err)
	s.Require().NotNil(routeStats)
	s.Require().Len(routeStats, 1)
	s.Require().Equal(sdk.OneInt(), routeStats[0].NumberOfTrades)
	s.Require().Equal([]uint64{swapPoolId1, swapPoolId2, swapPoolId3}, routeStats[0].Route)
	s.Require().Equal(profits, routeStats[0].Profits)
}

func (s *IntegrationTestSuite) ConcentratedLiquidity() {
	var (
		denom0                 = "uion"
		denom1                 = "uosmo"
		tickSpacing     uint64 = 100
		spreadFactor           = "0.001" // 0.1%
		spreadFactorDec        = sdk.MustNewDecFromStr("0.001")
		takerFee               = sdk.MustNewDecFromStr("0.0015")
	)

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	// Get the permisionless pool creation parameter.
	isPermisionlessCreationEnabledStr := chainABNode.QueryParams(cltypes.ModuleName, string(cltypes.KeyIsPermisionlessPoolCreationEnabled))
	if !strings.EqualFold(isPermisionlessCreationEnabledStr, "false") {
		s.T().Fatal("concentrated liquidity pool creation is enabled when should not have been")
	}

	// Change the parameter to enable permisionless pool creation.
	err = chainABNode.ParamChangeProposal("concentratedliquidity", string(cltypes.KeyIsPermisionlessPoolCreationEnabled), []byte("true"), chainAB)
	s.Require().NoError(err)

	// Update the protorev admin address to a known wallet we can control
	adminWalletAddr := chainABNode.CreateWalletAndFund("admin", []string{"4000000uosmo"}, chainAB)
	err = chainABNode.ParamChangeProposal("protorev", string(protorevtypes.ParamStoreKeyAdminAccount), []byte(fmt.Sprintf(`"%s"`, adminWalletAddr)), chainAB)
	s.Require().NoError(err)

	// Update the weight of CL pools so that this test case is not back run by protorev.
	chainABNode.SetMaxPoolPointsPerTx(7, adminWalletAddr)

	// Confirm that the parameter has been changed.
	isPermisionlessCreationEnabledStr = chainABNode.QueryParams(cltypes.ModuleName, string(cltypes.KeyIsPermisionlessPoolCreationEnabled))
	if !strings.EqualFold(isPermisionlessCreationEnabledStr, "true") {
		s.T().Fatal("concentrated liquidity pool creation is not enabled")
	}

	// Create concentrated liquidity pool when permisionless pool creation is enabled.
	poolID := chainABNode.CreateConcentratedPool(initialization.ValidatorWalletName, denom0, denom1, tickSpacing, spreadFactor)

	concentratedPool := s.updatedConcentratedPool(chainABNode, poolID)

	// Sanity check that pool initialized with valid parameters (the ones that we haven't explicitly specified)
	s.Require().Equal(concentratedPool.GetCurrentTick(), int64(0))
	s.Require().Equal(concentratedPool.GetCurrentSqrtPrice(), osmomath.ZeroDec())
	s.Require().Equal(concentratedPool.GetLiquidity(), sdk.ZeroDec())

	// Assert contents of the pool are valid (that we explicitly specified)
	s.Require().Equal(concentratedPool.GetId(), poolID)
	s.Require().Equal(concentratedPool.GetToken0(), denom0)
	s.Require().Equal(concentratedPool.GetToken1(), denom1)
	s.Require().Equal(concentratedPool.GetTickSpacing(), tickSpacing)
	s.Require().Equal(concentratedPool.GetExponentAtPriceOne(), cltypes.ExponentAtPriceOne)
	s.Require().Equal(concentratedPool.GetSpreadFactor(sdk.Context{}), sdk.MustNewDecFromStr(spreadFactor))

	fundTokens := []string{"100000000uosmo", "100000000uion", "100000000stake"}

	// Get 3 addresses to create positions
	address1 := chainABNode.CreateWalletAndFund("addr1", fundTokens, chainAB)
	address2 := chainABNode.CreateWalletAndFund("addr2", fundTokens, chainAB)
	address3 := chainABNode.CreateWalletAndFund("addr3", fundTokens, chainAB)

	// When claiming rewards, a small portion of dust is forfeited and is redistributed to everyone. We must track the total
	// liquidity across all positions (even if not active), in order to calculate how much to increase the reward growth global per share by.
	totalLiquidity := sdk.ZeroDec()

	// Create 2 positions for address1: overlap together, overlap with 2 address3 positions
	_, liquidity := chainABNode.CreateConcentratedPosition(address1, "[-120000]", "40000", fmt.Sprintf("10000000%s,10000000%s", denom0, denom1), 0, 0, poolID)
	totalLiquidity = totalLiquidity.Add(liquidity)
	_, liquidity = chainABNode.CreateConcentratedPosition(address1, "[-40000]", "120000", fmt.Sprintf("10000000%s,10000000%s", denom0, denom1), 0, 0, poolID)
	totalLiquidity = totalLiquidity.Add(liquidity)

	// Create 1 position for address2: does not overlap with anything, ends at maximum
	_, liquidity = chainABNode.CreateConcentratedPosition(address2, "220000", fmt.Sprintf("%d", cltypes.MaxTick), fmt.Sprintf("10000000%s,10000000%s", denom0, denom1), 0, 0, poolID)
	totalLiquidity = totalLiquidity.Add(liquidity)

	// Create 2 positions for address3: overlap together, overlap with 2 address1 positions, one position starts from minimum
	_, liquidity = chainABNode.CreateConcentratedPosition(address3, "[-160000]", "[-20000]", fmt.Sprintf("10000000%s,10000000%s", denom0, denom1), 0, 0, poolID)
	totalLiquidity = totalLiquidity.Add(liquidity)
	_, liquidity = chainABNode.CreateConcentratedPosition(address3, fmt.Sprintf("[%d]", cltypes.MinInitializedTick), "140000", fmt.Sprintf("10000000%s,10000000%s", denom0, denom1), 0, 0, poolID)
	totalLiquidity = totalLiquidity.Add(liquidity)

	// Get newly created positions
	positionsAddress1 := chainABNode.QueryConcentratedPositions(address1)
	positionsAddress2 := chainABNode.QueryConcentratedPositions(address2)
	positionsAddress3 := chainABNode.QueryConcentratedPositions(address3)

	concentratedPool = s.updatedConcentratedPool(chainABNode, poolID)

	// Assert number of positions per address
	s.Require().Equal(len(positionsAddress1), 2)
	s.Require().Equal(len(positionsAddress2), 1)
	s.Require().Equal(len(positionsAddress3), 2)

	// Assert positions for address1
	addr1position1 := positionsAddress1[0].Position
	addr1position2 := positionsAddress1[1].Position
	// First position first address
	s.validateCLPosition(addr1position1, poolID, -120000, 40000)
	// Second position second address
	s.validateCLPosition(addr1position2, poolID, -40000, 120000)

	// Assert positions for address2
	addr2position1 := positionsAddress2[0].Position
	// First position second address
	s.validateCLPosition(addr2position1, poolID, 220000, cltypes.MaxTick)

	// Assert positions for address3
	addr3position1 := positionsAddress3[0].Position
	addr3position2 := positionsAddress3[1].Position
	// First position third address
	s.validateCLPosition(addr3position1, poolID, -160000, -20000)
	// Second position third address
	s.validateCLPosition(addr3position2, poolID, cltypes.MinInitializedTick, 140000)

	// Collect SpreadRewards

	var (
		// spreadRewardGrowthGlobal is a variable for tracking global spread reward growth
		spreadRewardGrowthGlobal = sdk.ZeroDec()
		outMinAmt                = "1"
	)

	// Swap 1
	// Not crossing initialized ticks => performed in one swap step
	// Swap affects 3 positions: both that address1 has and one of address3's positions
	// Asserts that spread rewards are correctly collected for non cross-tick swaps
	var (
		// Swap parameters
		uosmoInDec_Swap1 = osmomath.NewBigDec(3465198)
		uosmoIn_Swap1    = fmt.Sprintf("%suosmo", uosmoInDec_Swap1.SDKDec().String())
	)
	// Perform swap (not crossing initialized ticks)
	chainABNode.SwapExactAmountIn(uosmoIn_Swap1, outMinAmt, fmt.Sprintf("%d", poolID), denom0, initialization.ValidatorWalletName)
	// Calculate and track global spread reward growth for swap 1
	uosmoInDec_Swap1_SubTakerFee := uosmoInDec_Swap1.SDKDec().Mul(sdk.OneDec().Sub(takerFee)).TruncateDec()
	uosmoInDec_Swap1_SubTakerFee_SubSpreadFactor := uosmoInDec_Swap1_SubTakerFee.Mul(sdk.OneDec().Sub(spreadFactorDec))
	totalSpreadReward := uosmoInDec_Swap1_SubTakerFee.Sub(uosmoInDec_Swap1_SubTakerFee_SubSpreadFactor).TruncateDec()

	spreadRewardGrowthGlobal.AddMut(calculateSpreadRewardGrowthGlobal(totalSpreadReward, concentratedPool.GetLiquidity()))

	// Update pool and track liquidity and sqrt price
	liquidityBeforeSwap := concentratedPool.GetLiquidity()
	sqrtPriceBeforeSwap := concentratedPool.GetCurrentSqrtPrice()

	concentratedPool = s.updatedConcentratedPool(chainABNode, poolID)

	liquidityAfterSwap := concentratedPool.GetLiquidity()
	sqrtPriceAfterSwap := concentratedPool.GetCurrentSqrtPrice()

	// Assert swaps don't change pool's liquidity amount
	s.Require().Equal(liquidityAfterSwap.String(), liquidityBeforeSwap.String())

	// Assert current sqrt price
	expectedSqrtPriceDelta := osmomath.BigDecFromSDKDec(uosmoInDec_Swap1_SubTakerFee_SubSpreadFactor).QuoTruncate(osmomath.BigDecFromSDKDec(concentratedPool.GetLiquidity())) // Δ(sqrtPrice) = Δy / L
	expectedSqrtPrice := sqrtPriceBeforeSwap.Add(expectedSqrtPriceDelta)
	s.Require().Equal(expectedSqrtPrice.String(), sqrtPriceAfterSwap.String())

	// Collect SpreadRewards: Swap 1

	// Track balances for address1 position1
	addr1BalancesBefore := s.addrBalance(chainABNode, address1)
	chainABNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter := s.addrBalance(chainABNode, address1)

	// Assert that the balance changed only for tokenIn (uosmo)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, false, true)

	// Assert Balances: Swap 1

	// Calculate uncollected spread rewards for address1 position1
	spreadRewardsUncollectedAddress1Position1_Swap1 := calculateUncollectedSpreadRewards(
		positionsAddress1[0].Position.Liquidity,
		sdk.ZeroDec(), // no growth below
		sdk.ZeroDec(), // no growth above
		sdk.ZeroDec(), // no spreadRewardGrowthInsideLast - it is the first swap
		spreadRewardGrowthGlobal,
	)

	// Note the global spread reward growth before dust redistribution
	spreadRewardGrowthGlobalBeforeDustRedistribution := spreadRewardGrowthGlobal.Clone()

	// Determine forfeited dust amount
	forfeitedDustAmt := spreadRewardsUncollectedAddress1Position1_Swap1.Sub(spreadRewardsUncollectedAddress1Position1_Swap1.TruncateDec())
	forfeitedDust := sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", forfeitedDustAmt))
	forfeitedDustPerShare := forfeitedDust.QuoDecTruncate(totalLiquidity)

	// Add forfeited dust back to the global spread reward growth
	spreadRewardGrowthGlobal.AddMut(forfeitedDustPerShare.AmountOf("uosmo"))

	// Assert
	s.Require().Equal(
		addr1BalancesBefore.AmountOf("uosmo").Add(spreadRewardsUncollectedAddress1Position1_Swap1.TruncateInt()).String(),
		addr1BalancesAfter.AmountOf("uosmo").String(),
	)

	// Swap 2
	//
	// Cross-tick swap:
	// * Part of swap happens in range of liquidity of 3 positions: both of address1 and one for address3 (until tick 40000 - upper tick of address1 position1)
	// * Another part happens in range of liquidity of 2 positions: one from address1 and address3
	//
	// Asserts:
	// * Net liquidity is kicked out when crossing initialized tick
	// * Liquidity of position that was kicked out after first swap step does not earn rewards from second swap step
	// * Uncollected spread rewards from multiple swaps are correctly summed up and collected

	// tickOffset is a tick index after the next initialized tick to which this swap needs to move the current price
	tickOffset := int64(300)
	sqrtPriceBeforeSwap = concentratedPool.GetCurrentSqrtPrice()
	liquidityBeforeSwap = concentratedPool.GetLiquidity()
	nextInitTick := int64(40000) // address1 position1's upper tick

	// Calculate sqrtPrice after and at the next initialized tick (upperTick of address1 position1 - 40000)
	_, sqrtPriceAfterNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick + tickOffset)
	s.Require().NoError(err)
	_, sqrtPriceAtNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick)
	s.Require().NoError(err)
	sqrtPriceAfterNextInitializedTickBigDec := osmomath.BigDecFromSDKDec(sqrtPriceAfterNextInitializedTick)
	sqrtPriceAtNextInitializedTickBigDec := osmomath.BigDecFromSDKDec(sqrtPriceAtNextInitializedTick)

	// Calculate Δ(sqrtPrice):
	// deltaSqrtPriceAfterNextInitializedTick = ΔsqrtP(40300) - ΔsqrtP(40000)
	// deltaSqrtPriceAtNextInitializedTick = ΔsqrtP(40000) - ΔsqrtP(currentTick)
	deltaSqrtPriceAfterNextInitializedTick := sqrtPriceAfterNextInitializedTickBigDec.Sub(sqrtPriceAtNextInitializedTickBigDec).SDKDec()
	deltaSqrtPriceAtNextInitializedTick := sqrtPriceAtNextInitializedTickBigDec.Sub(sqrtPriceBeforeSwap).SDKDec()

	// Calculate the amount of osmo required to:
	// * amountInToGetToTickAfterInitialized - move price from next initialized tick (40000) to destination tick (40000 + tickOffset)
	// * amountInToGetToNextInitTick - move price from current tick to next initialized tick
	// Formula is as follows:
	// Δy = L * Δ(sqrtPrice)
	amountInToGetToTickAfterInitialized := deltaSqrtPriceAfterNextInitializedTick.Mul(liquidityBeforeSwap.Sub(positionsAddress1[0].Position.Liquidity))
	amountInToGetToNextInitTick := deltaSqrtPriceAtNextInitializedTick.Mul(liquidityBeforeSwap)

	var (
		// Swap parameters

		// uosmoInDec_Swap2_NoSpreadReward is calculated such that swapping this amount (not considering spread reward) moves the price over the next initialized tick
		uosmoInDec_Swap2_NoSpreadReward = amountInToGetToNextInitTick.Add(amountInToGetToTickAfterInitialized)
		uosmoInDec_Swap2                = uosmoInDec_Swap2_NoSpreadReward.Quo(sdk.OneDec().Sub(spreadFactorDec)).TruncateDec() // account for spread factor of 1%

		spreadRewardGrowthGlobal_Swap1 = spreadRewardGrowthGlobalBeforeDustRedistribution.Clone()
	)

	uosmoInDec_Swap2_AddTakerFee := uosmoInDec_Swap2.Quo(sdk.OneDec().Sub(takerFee)).TruncateDec() // account for taker fee
	uosmoIn_Swap2 := fmt.Sprintf("%suosmo", uosmoInDec_Swap2_AddTakerFee.String())

	// Perform a swap
	chainABNode.SwapExactAmountIn(uosmoIn_Swap2, outMinAmt, fmt.Sprintf("%d", poolID), denom0, initialization.ValidatorWalletName)

	// Calculate the amount of liquidity of the position that was kicked out during swap (address1 position1)
	liquidityOfKickedOutPosition := positionsAddress1[0].Position.Liquidity

	// Update pool and track pool's liquidity
	concentratedPool = s.updatedConcentratedPool(chainABNode, poolID)

	liquidityAfterSwap = concentratedPool.GetLiquidity()

	// Assert that net liquidity of kicked out position was successfully removed from current pool's liquidity
	s.Require().Equal(liquidityBeforeSwap.Sub(liquidityOfKickedOutPosition), liquidityAfterSwap)

	// Collect spread rewards: Swap 2

	// Calculate spread reward charges per each step

	// Step1: amountIn is uosmo tokens that are swapped + uosmo tokens that are paid for spread reward
	// hasReachedTarget in SwapStep is true, hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountIn * spreadFactor / (1 - spreadFactor)
	spreadRewardCharge_Swap2_Step1 := amountInToGetToNextInitTick.Mul(spreadFactorDec).Quo(sdk.OneDec().Sub(spreadFactorDec))

	// Step2: hasReachedTarget in SwapStep is false (nextTick is 120000), hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountRemaining - amountOne
	amountRemainingAfterStep1 := uosmoInDec_Swap2.Sub(amountInToGetToNextInitTick).Sub(spreadRewardCharge_Swap2_Step1)
	spreadRewardCharge_Swap2_Step2 := amountRemainingAfterStep1.Sub(amountInToGetToTickAfterInitialized)

	// per unit of virtual liquidity
	spreadRewardCharge_Swap2_Step1.QuoMut(liquidityBeforeSwap)
	spreadRewardCharge_Swap2_Step2.QuoMut(liquidityAfterSwap)

	// Update spreadRewardGrowthGlobal
	spreadRewardGrowthGlobal.AddMut(spreadRewardCharge_Swap2_Step1.Add(spreadRewardCharge_Swap2_Step2))

	// Assert Balances: Swap 2

	// Assert that address1 position1 earned spread rewards only from first swap step

	// Track balances for address1 position1
	addr1BalancesBefore = s.addrBalance(chainABNode, address1)
	chainABNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter = s.addrBalance(chainABNode, address1)

	// Assert that the balance changed only for tokenIn (uosmo)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, false, true)

	// Calculate uncollected spread rewards for position, which liquidity will only be live part of the swap
	spreadRewardsUncollectedAddress1Position1_Swap2 := calculateUncollectedSpreadRewards(
		positionsAddress1[0].Position.Liquidity,
		sdk.ZeroDec(),
		spreadRewardCharge_Swap2_Step2,
		spreadRewardGrowthGlobal_Swap1,
		spreadRewardGrowthGlobal,
	)

	// Assert
	s.Require().Equal(
		addr1BalancesBefore.AmountOf("uosmo").Add(spreadRewardsUncollectedAddress1Position1_Swap2.TruncateInt()),
		addr1BalancesAfter.AmountOf("uosmo"),
	)

	// Assert that address3 position2 earned rewards from first and second swaps

	// Track balance off address3 position2: check that position that has not been kicked out earned full rewards
	addr3BalancesBefore := s.addrBalance(chainABNode, address3)
	chainABNode.CollectSpreadRewards(address3, fmt.Sprint(positionsAddress3[1].Position.PositionId))
	addr3BalancesAfter := s.addrBalance(chainABNode, address3)

	// Calculate uncollected spread rewards for address3 position2 earned from Swap 1
	spreadRewardsUncollectedAddress3Position2_Swap1 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		spreadRewardGrowthGlobal_Swap1,
	)

	// Calculate uncollected spread rewards for address3 position2 (was active throughout both swap steps): Swap2
	spreadRewardsUncollectedAddress3Position2_Swap2 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap1, sdk.ZeroDec(), sdk.ZeroDec()),
		spreadRewardGrowthGlobal,
	)

	// Total spread rewards earned by address3 position2 from 2 swaps
	totalUncollectedSpreadRewardsAddress3Position2 := spreadRewardsUncollectedAddress3Position2_Swap1.Add(spreadRewardsUncollectedAddress3Position2_Swap2)

	// Assert
	s.Require().Equal(
		addr3BalancesBefore.AmountOf("uosmo").Add(totalUncollectedSpreadRewardsAddress3Position2.TruncateInt()),
		addr3BalancesAfter.AmountOf("uosmo"),
	)

	// Swap 3
	// Asserts:
	// * swapping amountZero for amountOne works correctly
	// * liquidity of positions that come in range are correctly kicked in

	// tickOffset is a tick index after the next initialized tick to which this swap needs to move the current price
	tickOffset = 300
	sqrtPriceBeforeSwap = concentratedPool.GetCurrentSqrtPrice()
	liquidityBeforeSwap = concentratedPool.GetLiquidity()
	nextInitTick = 40000

	// Calculate amount required to get to
	// 1) next initialized tick
	// 2) tick below next initialized (-300)
	// Using: CalcAmount0Delta = liquidity * ((sqrtPriceB - sqrtPriceA) / (sqrtPriceB * sqrtPriceA))

	// Calculate sqrtPrice after and at the next initialized tick (which is upperTick of address1 position1 - 40000)
	_, sqrtPricebBelowNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick - tickOffset)
	s.Require().NoError(err)
	_, sqrtPriceAtNextInitializedTick, err = clmath.TickToSqrtPrice(nextInitTick)
	s.Require().NoError(err)
	sqrtPriceAtNextInitializedTickBigDec = osmomath.BigDecFromSDKDec(sqrtPriceAtNextInitializedTick)

	// Calculate numerators
	numeratorBelowNextInitializedTick := sqrtPriceAtNextInitializedTick.Sub(sqrtPricebBelowNextInitializedTick)
	numeratorNextInitializedTick := sqrtPriceBeforeSwap.Sub(sqrtPriceAtNextInitializedTickBigDec)

	// Calculate denominators
	denominatorBelowNextInitializedTick := sqrtPriceAtNextInitializedTick.Mul(sqrtPricebBelowNextInitializedTick)
	denominatorNextInitializedTick := sqrtPriceBeforeSwap.Mul(sqrtPriceAtNextInitializedTickBigDec)

	// Calculate fractions
	fractionBelowNextInitializedTick := numeratorBelowNextInitializedTick.Quo(denominatorBelowNextInitializedTick)
	fractionAtNextInitializedTick := numeratorNextInitializedTick.Quo(denominatorNextInitializedTick)

	// Calculate amounts of uionIn needed
	amountInToGetToTickBelowInitialized := liquidityBeforeSwap.Add(positionsAddress1[0].Position.Liquidity).Mul(fractionBelowNextInitializedTick)
	amountInToGetToNextInitTick = liquidityBeforeSwap.Mul(fractionAtNextInitializedTick.SDKDec())

	// Collect spread rewards for address1 position1 to avoid overhead computations (swap2 already asserted spread rewards are aggregated correctly from multiple swaps)
	chainABNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))

	var (
		// Swap parameters
		uionInDec_Swap3_NoSpreadReward = amountInToGetToNextInitTick.Add(amountInToGetToTickBelowInitialized)                // amount of uion to move price from current to desired (not considering spreadFactor)
		uionInDec_Swap3                = uionInDec_Swap3_NoSpreadReward.Quo(sdk.OneDec().Sub(spreadFactorDec)).TruncateDec() // consider spreadFactor

		// Save variables from previous swaps
		spreadRewardGrowthGlobal_Swap2                = spreadRewardGrowthGlobal.Clone()
		spreadRewardGrowthInsideAddress1Position1Last = spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap2_Step2).Clone()
	)

	uionInDec_Swap3_AddTakerFee := uionInDec_Swap3.Quo(sdk.OneDec().Sub(takerFee)).TruncateDec() // account for taker fee
	uionIn_Swap3 := fmt.Sprintf("%suion", uionInDec_Swap3_AddTakerFee.String())

	// Perform a swap
	chainABNode.SwapExactAmountIn(uionIn_Swap3, outMinAmt, fmt.Sprintf("%d", poolID), denom1, initialization.ValidatorWalletName)

	// Assert liquidity of kicked in position was successfully added to the pool
	concentratedPool = s.updatedConcentratedPool(chainABNode, poolID)

	liquidityAfterSwap = concentratedPool.GetLiquidity()
	s.Require().Equal(liquidityBeforeSwap.Add(positionsAddress1[0].Position.Liquidity), liquidityAfterSwap)

	// Track balance of address1
	addr1BalancesBefore = s.addrBalance(chainABNode, address1)
	chainABNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter = s.addrBalance(chainABNode, address1)

	// Assert that the balance changed only for tokenIn (uion)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, true, false)

	// Assert the amount of collected spread rewards:

	// Step1: amountIn is uion tokens that are swapped + uion tokens that are paid for spread reward
	// hasReachedTarget in SwapStep is true, hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountIn * spreadFactor / (1 - spreadFactor)
	spreadRewardCharge_Swap3_Step1 := amountInToGetToNextInitTick.Mul(spreadFactorDec).Quo(sdk.OneDec().Sub(spreadFactorDec))

	// Step2: hasReachedTarget in SwapStep is false (next initialized tick is -20000), hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountRemaining - amountZero
	amountRemainingAfterStep1 = uionInDec_Swap3.Sub(amountInToGetToNextInitTick).Sub(spreadRewardCharge_Swap3_Step1)
	spreadRewardCharge_Swap3_Step2 := amountRemainingAfterStep1.Sub(amountInToGetToTickBelowInitialized)

	// Per unit of virtual liquidity
	spreadRewardCharge_Swap3_Step1.QuoMut(liquidityBeforeSwap)
	spreadRewardCharge_Swap3_Step2.QuoMut(liquidityAfterSwap)

	// Update spreadRewardGrowthGlobal
	spreadRewardGrowthGlobal.AddMut(spreadRewardCharge_Swap3_Step1.Add(spreadRewardCharge_Swap3_Step2))

	// Assert position that was active throughout second swap step (address1 position1) only earned spread rewards for this step:

	// Only collects spread rewards for second swap step
	spreadRewardsUncollectedAddress1Position1_Swap3 := calculateUncollectedSpreadRewards(
		positionsAddress1[0].Position.Liquidity,
		sdk.ZeroDec(),
		spreadRewardCharge_Swap2_Step2.Add(spreadRewardCharge_Swap3_Step1), // spread rewards acquired by swap2 step2 and swap3 step1 (steps happened above upper tick of this position)
		spreadRewardGrowthInsideAddress1Position1Last,                      // spreadRewardGrowthInside from first and second swaps
		spreadRewardGrowthGlobal,
	)

	// Assert
	s.Require().Equal(
		addr1BalancesBefore.AmountOf("uion").Add(spreadRewardsUncollectedAddress1Position1_Swap3.TruncateInt()),
		addr1BalancesAfter.AmountOf("uion"),
	)

	// Assert position that was active throughout the whole swap:

	// Track balance of address3
	addr3BalancesBefore = s.addrBalance(chainABNode, address3)
	chainABNode.CollectSpreadRewards(address3, fmt.Sprint(positionsAddress3[1].Position.PositionId))
	addr3BalancesAfter = s.addrBalance(chainABNode, address3)

	// Assert that the balance changed only for tokenIn (uion)
	s.assertBalancesInvariants(addr3BalancesBefore, addr3BalancesAfter, true, false)

	// Was active throughout the whole swap, collects spread rewards from 2 steps

	// Step 1
	spreadRewardsUncollectedAddress3Position2_Swap3_Step1 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		sdk.ZeroDec(), // no growth below
		sdk.ZeroDec(), // no growth above
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap2, sdk.ZeroDec(), sdk.ZeroDec()), // snapshot of spread reward growth at swap 2
		spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap3_Step2),                                    // step 1 hasn't earned spread rewards from step 2
	)

	// Step 2
	spreadRewardsUncollectedAddress3Position2_Swap3_Step2 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		sdk.ZeroDec(), // no growth below
		sdk.ZeroDec(), // no growth above
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap2, sdk.ZeroDec(), sdk.ZeroDec()), // snapshot of spread reward growth at swap 2
		spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap3_Step1),                                    // step 2 hasn't earned spread rewards from step 1
	)

	// Calculate total spread rewards acquired by address3 position2 from all swap steps
	totalUncollectedSpreadRewardsAddress3Position2 = spreadRewardsUncollectedAddress3Position2_Swap3_Step1.Add(spreadRewardsUncollectedAddress3Position2_Swap3_Step2)

	// Assert
	s.Require().Equal(
		addr3BalancesBefore.AmountOf("uion").Add(totalUncollectedSpreadRewardsAddress3Position2.TruncateInt()),
		addr3BalancesAfter.AmountOf("uion"),
	)

	// Collect SpreadRewards: Sanity Checks

	// Assert that positions, which were not included in swaps, were not affected

	// Address3 Position1: [-160000; -20000]
	addr3BalancesBefore = s.addrBalance(chainABNode, address3)
	chainABNode.CollectSpreadRewards(address3, fmt.Sprint(positionsAddress3[0].Position.PositionId))
	addr3BalancesAfter = s.addrBalance(chainABNode, address3)

	// Assert that balances did not change for any token
	s.assertBalancesInvariants(addr3BalancesBefore, addr3BalancesAfter, true, true)

	// Address2's only position: [220000; 342000]
	addr2BalancesBefore := s.addrBalance(chainABNode, address2)
	chainABNode.CollectSpreadRewards(address2, fmt.Sprint(positionsAddress2[0].Position.PositionId))
	addr2BalancesAfter := s.addrBalance(chainABNode, address2)

	// Assert the balances did not change for every token
	s.assertBalancesInvariants(addr2BalancesBefore, addr2BalancesAfter, true, true)

	// Withdraw Position

	// Withdraw Position parameters
	defaultLiquidityRemoval := "1000"

	chainAB.WaitForNumHeights(2)

	// Assert removing some liquidity
	// address1: check removing some amount of liquidity
	address1position1liquidityBefore := positionsAddress1[0].Position.Liquidity
	chainABNode.WithdrawPosition(address1, defaultLiquidityRemoval, positionsAddress1[0].Position.PositionId)
	// assert
	positionsAddress1 = chainABNode.QueryConcentratedPositions(address1)
	s.Require().Equal(address1position1liquidityBefore, positionsAddress1[0].Position.Liquidity.Add(sdk.MustNewDecFromStr(defaultLiquidityRemoval)))

	// address2: check removing some amount of liquidity
	address2position1liquidityBefore := positionsAddress2[0].Position.Liquidity
	chainABNode.WithdrawPosition(address2, defaultLiquidityRemoval, positionsAddress2[0].Position.PositionId)
	// assert
	positionsAddress2 = chainABNode.QueryConcentratedPositions(address2)
	s.Require().Equal(address2position1liquidityBefore, positionsAddress2[0].Position.Liquidity.Add(sdk.MustNewDecFromStr(defaultLiquidityRemoval)))

	// address3: check removing some amount of liquidity
	address3position1liquidityBefore := positionsAddress3[0].Position.Liquidity
	chainABNode.WithdrawPosition(address3, defaultLiquidityRemoval, positionsAddress3[0].Position.PositionId)
	// assert
	positionsAddress3 = chainABNode.QueryConcentratedPositions(address3)
	s.Require().Equal(address3position1liquidityBefore, positionsAddress3[0].Position.Liquidity.Add(sdk.MustNewDecFromStr(defaultLiquidityRemoval)))

	// Assert removing all liquidity
	// address2: no more positions left
	allLiquidityAddress2Position1 := positionsAddress2[0].Position.Liquidity
	chainABNode.WithdrawPosition(address2, allLiquidityAddress2Position1.String(), positionsAddress2[0].Position.PositionId)
	positionsAddress2 = chainABNode.QueryConcentratedPositions(address2)
	s.Require().Empty(positionsAddress2)

	// address1: one position left
	allLiquidityAddress1Position1 := positionsAddress1[0].Position.Liquidity
	chainABNode.WithdrawPosition(address1, allLiquidityAddress1Position1.String(), positionsAddress1[0].Position.PositionId)
	positionsAddress1 = chainABNode.QueryConcentratedPositions(address1)
	s.Require().Equal(len(positionsAddress1), 1)

	// Test tick spacing reduction proposal

	// Get the current tick spacing
	currentTickSpacing := concentratedPool.GetTickSpacing()

	// Get the index of the current tick spacing in relation to authorized tick spacings
	indexOfCurrentTickSpacing := uint64(0)
	for i, tickSpacing := range cltypes.AuthorizedTickSpacing {
		if tickSpacing == currentTickSpacing {
			indexOfCurrentTickSpacing = uint64(i)
			break
		}
	}

	// The new tick spacing will be the next lowest authorized tick spacing
	newTickSpacing := cltypes.AuthorizedTickSpacing[indexOfCurrentTickSpacing-1]

	// Run the tick spacing reduction proposal
	propNumber := chainABNode.SubmitTickSpacingReductionProposal(fmt.Sprintf("%d,%d", poolID, newTickSpacing), sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinExpeditedDeposit)), true)

	chainABNode.DepositProposal(propNumber, true)
	totalTimeChan := make(chan time.Duration, 1)
	go chainABNode.QueryPropStatusTimed(propNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)
	var wg sync.WaitGroup

	// TODO: create a helper function for all these go routine yes vote calls.
	for _, n := range chainAB.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *chain.NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	timeoutPeriod := 2 * time.Minute
	select {
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	case <-totalTimeChan:
		// The goroutine finished before the timeout period, continue execution.
	}

	// Check that the tick spacing was reduced to the expected new tick spacing
	concentratedPool = s.updatedConcentratedPool(chainABNode, poolID)
	s.Require().Equal(newTickSpacing, concentratedPool.GetTickSpacing())
}

func (s *IntegrationTestSuite) StableSwapPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("Skipping StableSwapPostUpgrade test")
	}

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	index := s.getChainIndex(chainAB)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	const (
		denomA = "stake"
		denomB = "uosmo"

		minAmountOut = "1"
	)

	coinAIn, coinBIn := fmt.Sprintf("20000%s", denomA), fmt.Sprintf("2%s", denomB)

	chainABNode.BankSend(initialization.WalletFeeTokens.String(), sender, config.StableswapWallet[index])
	chainABNode.BankSend(coinAIn, sender, config.StableswapWallet[index])
	chainABNode.BankSend(coinBIn, sender, config.StableswapWallet[index])

	s.T().Log("performing swaps")
	chainABNode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[index]), denomB, config.StableswapWallet[index])
	chainABNode.SwapExactAmountIn(coinBIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[index]), denomA, config.StableswapWallet[index])
}

// TestGeometricTwapMigration tests that the geometric twap record
// migration runs successfully. It does so by attempting to execute
// the swap on the pool created pre-upgrade. When a pool is created
// pre-upgrade, twap records are initialized for a pool. By runnning
// a swap post-upgrade, we confirm that the geometric twap was initialized
// correctly and does not cause a chain halt. This test was created
// in-response to a testnet incident when performing the geometric twap
// upgrade. Upon adding the migrations logic, the tests began to pass.
func (s *IntegrationTestSuite) GeometricTwapMigration() {
	if s.skipUpgrade {
		s.T().Skip("Skipping upgrade tests")
	}

	var (
		// Configurations for tests/e2e/scripts/pool1A.json
		// This pool gets initialized pre-upgrade.
		minAmountOut    = "1"
		otherDenom      = []string{"ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518", "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}
		migrationWallet = "migration"
	)

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	index := s.getChainIndex(chainAB)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	uosmoIn := fmt.Sprintf("1000000%s", "uosmo")

	swapWalletAddr := chainABNode.CreateWallet(migrationWallet, chainAB)

	chainABNode.BankSend(uosmoIn, sender, swapWalletAddr)

	// Swap to create new twap records on the pool that was created pre-upgrade.
	chainABNode.SwapExactAmountIn(uosmoIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradePoolId[index]), otherDenom[index], swapWalletAddr)
}

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
// Additionally, it attempst to create a pool with IBC denoms.
func (s *IntegrationTestSuite) IBCTokenTransferAndCreatePool() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode, err := s.getChainACfgs()
	s.Require().NoError(err)
	chainB, chainBNode, err := s.getChainBCfgs()
	s.Require().NoError(err)

	chainANode.SendIBC(chainA, chainB, chainBNode.PublicAddress, initialization.OsmoToken)
	chainBNode.SendIBC(chainB, chainA, chainANode.PublicAddress, initialization.OsmoToken)
	chainANode.SendIBC(chainA, chainB, chainBNode.PublicAddress, initialization.StakeToken)
	chainBNode.SendIBC(chainB, chainA, chainANode.PublicAddress, initialization.StakeToken)

	chainANode.CreateBalancerPool("ibcDenomPool.json", initialization.ValidatorWalletName)
}

// TestSuperfluidVoting tests that superfluid voting is functioning as expected.
// It does so by doing the following:
// - creating a pool
// - attempting to submit a proposal to enable superfluid voting in that pool
// - voting yes on the proposal from the validator wallet
// - voting no on the proposal from the delegator wallet
// - ensuring that delegator's wallet overwrites the validator's vote
func (s *IntegrationTestSuite) SuperfluidVoting() {
	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	poolId := chainABNode.CreateBalancerPool("nativeDenomPool.json", initialization.ValidatorWalletName)

	// enable superfluid assets
	chainABNode.EnableSuperfluidAsset(chainAB, fmt.Sprintf("gamm/pool/%d", poolId))

	// setup wallets and send gamm tokens to these wallets (both chains)
	superfluidVotingWallet := chainABNode.CreateWallet("TestSuperfluidVoting", chainAB)
	chainABNode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), initialization.ValidatorWalletName, superfluidVotingWallet)
	lockId := chainABNode.LockTokens(fmt.Sprintf("%v%s", sdk.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId)), "240s", superfluidVotingWallet)
	chainABNode.SuperfluidDelegate(lockId, chainABNode.OperatorAddress, superfluidVotingWallet)

	// create a text prop, deposit and vote yes
	propNumber := chainABNode.SubmitTextProposal("superfluid vote overwrite test", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)), false)
	chainABNode.DepositProposal(propNumber, false)

	var wg sync.WaitGroup

	for _, n := range chainAB.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *chain.NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	// set delegator vote to no
	chainABNode.VoteNoProposal(superfluidVotingWallet, propNumber)

	s.Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := chainABNode.QueryPropTally(propNumber)
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
	noTotal, _, _, _, _ := chainABNode.QueryPropTally(propNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := chainABNode.QueryIntermediaryAccount(fmt.Sprintf("gamm/pool/%d", poolId), chainABNode.OperatorAddress)
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

func (s *IntegrationTestSuite) CreateConcentratedLiquidityPoolVoting_And_TWAP() {
	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	poolId, err := chainAB.SubmitCreateConcentratedPoolProposal(chainABNode)
	s.NoError(err)
	fmt.Println("poolId", poolId)

	var (
		expectedDenom0       = "stake"
		expectedDenom1       = "uosmo"
		expectedTickspacing  = uint64(100)
		expectedSpreadFactor = "0.001000000000000000"
	)

	var concentratedPool cltypes.ConcentratedPoolExtension
	s.Eventually(
		func() bool {
			concentratedPool = s.updatedConcentratedPool(chainABNode, poolId)
			s.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())
			s.Require().Equal(expectedDenom0, concentratedPool.GetToken0())
			s.Require().Equal(expectedDenom1, concentratedPool.GetToken1())
			s.Require().Equal(expectedTickspacing, concentratedPool.GetTickSpacing())
			s.Require().Equal(expectedSpreadFactor, concentratedPool.GetSpreadFactor(sdk.Context{}).String())

			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"create concentrated liquidity pool was not successful.",
	)

	fundTokens := []string{"100000000stake", "100000000uosmo"}

	// Get address to create positions
	address1 := chainABNode.CreateWalletAndFund("address1", fundTokens, chainAB)

	// We add 5 ms to avoid landing directly on block time in twap. If block time
	// is provided as start time, the latest spot price is used. Otherwise
	// interpolation is done.
	timeBeforePositionCreationBeforeSwap := chainABNode.QueryLatestBlockTime().Add(5 * time.Millisecond)
	s.T().Log("geometric twap, start time ", timeBeforePositionCreationBeforeSwap.Unix())

	// Wait for the next height so that the requested twap
	// start time (timeBeforePositionCreationBeforeSwap) is not equal to the block time.
	chainAB.WaitForNumHeights(1)

	// Check initial TWAP
	// We expect this to error since there is no spot price yet.
	s.T().Log("initial twap check")
	initialTwapBOverA, err := chainABNode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeBeforePositionCreationBeforeSwap)
	s.Require().Error(err)
	s.Require().Equal(sdk.Dec{}, initialTwapBOverA)

	// Create a position and check that TWAP now returns a value.
	s.T().Log("creating first position")
	chainABNode.CreateConcentratedPosition(address1, "[-120000]", "40000", fmt.Sprintf("10000000%s,20000000%s", concentratedPool.GetToken0(), concentratedPool.GetToken1()), 0, 0, concentratedPool.GetId())
	timeAfterPositionCreationBeforeSwap := chainABNode.QueryLatestBlockTime()
	chainAB.WaitForNumHeights(2)
	firstPositionTwapBOverA, err := chainABNode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterPositionCreationBeforeSwap)
	s.Require().NoError(err)
	s.Require().Equal(sdk.MustNewDecFromStr("0.5"), firstPositionTwapBOverA)

	// Run a swap and check that the TWAP updates.
	s.T().Log("run swap")
	coinAIn := fmt.Sprintf("1000000%s", concentratedPool.GetToken0())
	chainABNode.SwapExactAmountIn(coinAIn, "1", fmt.Sprintf("%d", concentratedPool.GetId()), concentratedPool.GetToken1(), address1)

	timeAfterSwap := chainABNode.QueryLatestBlockTime()
	chainAB.WaitForNumHeights(1)
	timeAfterSwapPlus1Height := chainABNode.QueryLatestBlockTime()

	s.T().Log("querying for the TWAP after swap")
	afterSwapTwapBOverA, err := chainABNode.QueryGeometricTwap(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwap, timeAfterSwapPlus1Height)
	s.Require().NoError(err)

	// We swap stake so uosmo's supply will decrease and stake will increase.
	// The price after will be larger than the previous one.
	s.Require().True(afterSwapTwapBOverA.GT(firstPositionTwapBOverA))

	// Remove the position and check that TWAP returns an error.
	s.T().Log("removing first position (pool is drained)")
	positions := chainABNode.QueryConcentratedPositions(address1)
	chainABNode.WithdrawPosition(address1, positions[0].Position.Liquidity.String(), positions[0].Position.PositionId)
	chainAB.WaitForNumHeights(1)

	s.T().Log("querying for the TWAP from after pool drained")
	afterRemoveTwapBOverA, err := chainABNode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwapPlus1Height)
	s.Require().Error(err)
	s.Require().Equal(sdk.Dec{}, afterRemoveTwapBOverA)

	// Create a position and check that TWAP now returns a value.
	// Should be equal to 1 since the position contains equal amounts of both tokens.
	s.T().Log("creating position")
	chainABNode.CreateConcentratedPosition(address1, "[-120000]", "40000", fmt.Sprintf("10000000%s,10000000%s", concentratedPool.GetToken0(), concentratedPool.GetToken1()), 0, 0, concentratedPool.GetId())
	chainAB.WaitForNumHeights(1)
	timeAfterSwapRemoveAndCreatePlus1Height := chainABNode.QueryLatestBlockTime()
	secondTwapBOverA, err := chainABNode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwapRemoveAndCreatePlus1Height)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(1), secondTwapBOverA)
}

func (s *IntegrationTestSuite) IBCTokenTransferRateLimiting() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode, err := s.getChainACfgs()
	s.Require().NoError(err)
	chainB, chainBNode, err := s.getChainBCfgs()
	s.Require().NoError(err)

	receiver := chainBNode.GetWallet(initialization.ValidatorWalletName)

	// If the RL param is already set. Remember it to set it back at the end
	param := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
	fmt.Println("param", param)

	osmoSupply, err := chainANode.QuerySupplyOf("uosmo")
	s.Require().NoError(err)

	f, err := osmoSupply.ToDec().Float64()
	s.Require().NoError(err)

	over := f * 0.02

	paths := fmt.Sprintf(`{"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"testQuota", "duration": 86400, "send_recv": [1, 1]}] }`, initialization.OsmoToken.Denom)

	// Sending >1%
	fmt.Println("Sending >1%")
	chainANode.SendIBC(chainA, chainB, receiver, sdk.NewInt64Coin(initialization.OsmoDenom, int64(over)))

	contract, err := chainANode.SetupRateLimiting(paths, chainANode.PublicAddress, chainA)
	s.Require().NoError(err)

	s.Eventually(
		func() bool {
			val := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
			return strings.Contains(val, contract)
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve params",
	)

	// Sending <1%. Should work
	fmt.Println("Sending <1%. Should work")
	chainANode.SendIBC(chainA, chainB, receiver, sdk.NewInt64Coin(initialization.OsmoDenom, 1))
	// Sending >1%. Should fail
	fmt.Println("Sending >1%. Should fail")
	chainANode.FailIBCTransfer(initialization.ValidatorWalletName, receiver, fmt.Sprintf("%duosmo", int(over)))

	// Removing the rate limit so it doesn't affect other tests
	chainANode.WasmExecute(contract, `{"remove_path": {"channel_id": "channel-0", "denom": "uosmo"}}`, initialization.ValidatorWalletName)
	// reset the param to the original contract if it existed
	if param != "" {
		err = chainANode.ParamChangeProposal(
			ibcratelimittypes.ModuleName,
			string(ibcratelimittypes.KeyContractAddress),
			[]byte(param),
			chainA,
		)
		s.Require().NoError(err)
		s.Eventually(func() bool {
			val := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
			return strings.Contains(val, param)
		}, time.Second*30, 10*time.Millisecond)
	}
}

func (s *IntegrationTestSuite) LargeWasmUpload() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.Require().NoError(err)
	validatorAddr := chainANode.GetWallet(initialization.ValidatorWalletName)
	chainANode.StoreWasmCode("bytecode/large.wasm", validatorAddr)
}

func (s *IntegrationTestSuite) IBCWasmHooks() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode, err := s.getChainACfgs()
	s.Require().NoError(err)
	_, chainBNode, err := s.getChainBCfgs()
	s.Require().NoError(err)

	contractAddr := s.UploadAndInstantiateCounter(chainA)

	transferAmount := int64(10)
	validatorAddr := chainBNode.GetWallet(initialization.ValidatorWalletName)
	fmt.Println("Sending IBC transfer IBCWasmHooks")
	coin := sdk.NewCoin("uosmo", sdk.NewInt(transferAmount))
	chainBNode.SendIBCTransfer(chainA, validatorAddr, contractAddr,
		fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr), coin)

	// check the balance of the contract
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "uosmo"))
	ibcDenom := denomTrace.IBCDenom()
	s.CallCheckBalance(chainANode, contractAddr, ibcDenom, transferAmount)

	// sender wasm addr
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender("channel-0", validatorAddr, "osmo")

	var response map[string]interface{}
	s.Require().Eventually(func() bool {
		response, err = chainANode.QueryWasmSmartObject(contractAddr, fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderBech32))
		if err != nil {
			return false
		}

		totalFundsIface, ok := response["total_funds"].([]interface{})
		if !ok || len(totalFundsIface) == 0 {
			return false
		}

		totalFunds, ok := totalFundsIface[0].(map[string]interface{})
		if !ok {
			return false
		}

		amount, ok := totalFunds["amount"].(string)
		if !ok {
			return false
		}

		denom, ok := totalFunds["denom"].(string)
		if !ok {
			return false
		}

		// check if denom contains "uosmo"
		return amount == strconv.FormatInt(transferAmount, 10) && strings.Contains(denom, "ibc")
	},

		15*time.Second,
		10*time.Millisecond,
	)
}

// TestPacketForwarding sends a packet from chainA to chainB, and forwards it
// back to chainA with a custom memo to execute the counter contract on chain A
func (s *IntegrationTestSuite) PacketForwarding() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode, err := s.getChainACfgs()
	s.Require().NoError(err)
	chainB := s.configurer.GetChainConfig(1)

	// Instantiate the counter contract on chain A
	contractAddr := s.UploadAndInstantiateCounter(chainA)

	transferAmount := int64(10)
	validatorAddr := chainANode.GetWallet(initialization.ValidatorWalletName)
	// Specify that the counter contract should be called on chain A when the packet is received
	contractCallMemo := []byte(fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr))
	// Generate the forward metadata
	forwardMetadata := packetforwardingtypes.ForwardMetadata{
		Receiver: contractAddr,
		Port:     "transfer",
		Channel:  "channel-0",
		Next:     packetforwardingtypes.NewJSONObject(false, contractCallMemo, orderedmap.OrderedMap{}), // The packet sent to chainA will have this memo
	}
	memoData := packetforwardingtypes.PacketMetadata{Forward: &forwardMetadata}
	forwardMemo, err := json.Marshal(memoData)
	s.NoError(err)
	// Send the transfer from chainA to chainB. ChainB will parse the memo and forward the packet back to chainA
	coin := sdk.NewCoin("uosmo", sdk.NewInt(transferAmount))
	chainANode.SendIBCTransfer(chainB, validatorAddr, validatorAddr, string(forwardMemo), coin)

	// check the balance of the contract
	s.CallCheckBalance(chainANode, contractAddr, "uosmo", transferAmount)

	// Getting the sender as set by PFM
	senderStr := fmt.Sprintf("channel-0/%s", validatorAddr)
	senderHash32 := address.Hash(packetforwardingtypes.ModuleName, []byte(senderStr)) // typo intended
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := "osmo"
	pfmSender, err := sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
	s.Require().NoError(err)

	// sender wasm addr
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender("channel-0", pfmSender, "osmo")
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		response, err := chainANode.QueryWasmSmartObject(contractAddr, fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderBech32))
		if err != nil {
			return false
		}
		countValue, ok := response["count"].(float64)
		if !ok {
			return false
		}
		return countValue == 0
	},
		15*time.Second,
		10*time.Millisecond,
	)
}

// TestAddToExistingLockPostUpgrade ensures addToExistingLock works for locks created preupgrade.
func (s *IntegrationTestSuite) AddToExistingLockPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("Skipping AddToExistingLockPostUpgrade test")
	}

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)
	index := s.getChainIndex(chainAB)

	// ensure we can add to existing locks and superfluid locks that existed pre upgrade on chainA
	// we use the hardcoded gamm/pool/1 and these specific wallet names to match what was created pre upgrade
	preUpgradePoolShareDenom := fmt.Sprintf("gamm/pool/%d", config.PreUpgradePoolId[index])

	lockupWalletAddr, lockupWalletSuperfluidAddr := chainABNode.GetWallet("lockup-wallet"), chainABNode.GetWallet("lockup-wallet-superfluid")
	chainABNode.AddToExistingLock(sdk.NewInt(1000000000000000000), preUpgradePoolShareDenom, "240s", lockupWalletAddr, 1)
	chainABNode.AddToExistingLock(sdk.NewInt(1000000000000000000), preUpgradePoolShareDenom, "240s", lockupWalletSuperfluidAddr, 2)
}

// TestAddToExistingLock tests lockups to both regular and superfluid locks.
func (s *IntegrationTestSuite) AddToExistingLock() {
	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	funder := chainABNode.GetWallet(initialization.ValidatorWalletName)
	// ensure we can add to new locks and superfluid locks
	// create pool and enable superfluid assets
	poolId := chainABNode.CreateBalancerPool("nativeDenomPool.json", funder)
	chainABNode.EnableSuperfluidAsset(chainAB, fmt.Sprintf("gamm/pool/%d", poolId))

	// setup wallets and send gamm tokens to these wallets on chainA
	gammShares := fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId)
	fundTokens := []string{gammShares, initialization.WalletFeeTokens.String()}
	lockupWalletAddr := chainABNode.CreateWalletAndFundFrom("TestAddToExistingLock", funder, fundTokens, chainAB)
	lockupWalletSuperfluidAddr := chainABNode.CreateWalletAndFundFrom("TestAddToExistingLockSuperfluid", funder, fundTokens, chainAB)

	// ensure we can add to new locks and superfluid locks on chainA
	chainABNode.LockAndAddToExistingLock(chainAB, sdk.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId), lockupWalletAddr, lockupWalletSuperfluidAddr)
}

// TestArithmeticTWAP tests TWAP by creating a pool, performing a swap.
// These two operations should create TWAP records.
// Then, we wait until the epoch for the records to be pruned.
// The records are guaranteed to be pruned at the next epoch
// because twap keep time = epoch time / 4 and we use a timer
// to wait for at least the twap keep time.
func (s *IntegrationTestSuite) ArithmeticTWAP() {
	s.T().Skip("TODO: investigate further: https://github.com/osmosis-labs/osmosis/issues/4342")

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

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	// Triggers the creation of TWAP records.
	poolId := chainABNode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainABNode.CreateWalletAndFund(walletName, []string{initialization.WalletFeeTokens.String()}, chainAB)

	timeBeforeSwap := chainABNode.QueryLatestBlockTime()
	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainAB.WaitForNumHeights(2)

	s.T().Log("querying for the first TWAP to now before swap")
	twapFromBeforeSwapToBeforeSwapOneAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().NoError(err)

	chainABNode.BankSend(coinAIn, sender, swapWalletAddr)
	chainABNode.BankSend(coinBIn, sender, swapWalletAddr)
	chainABNode.BankSend(coinCIn, sender, swapWalletAddr)

	s.T().Log("querying for the second TWAP to now before swap, must equal to first")
	twapFromBeforeSwapToBeforeSwapTwoAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)

	// Since there were no swaps between the two queries, the TWAPs should be the same.
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneAB, twapFromBeforeSwapToBeforeSwapTwoAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneBC, twapFromBeforeSwapToBeforeSwapTwoBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneCA, twapFromBeforeSwapToBeforeSwapTwoCA, sdk.NewDecWithPrec(1, 3))

	s.T().Log("performing swaps")
	chainABNode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)
	chainABNode.SwapExactAmountIn(coinBIn, minAmountOut, fmt.Sprintf("%d", poolId), denomC, swapWalletAddr)
	chainABNode.SwapExactAmountIn(coinCIn, minAmountOut, fmt.Sprintf("%d", poolId), denomA, swapWalletAddr)

	keepPeriodCountDown := time.NewTimer(initialization.TWAPPruningKeepPeriod)

	// Make sure that we are still producing blocks and move far enough for the swap TWAP record to be created
	// so that we can measure start time post-swap (timeAfterSwap).
	chainAB.WaitForNumHeights(2)

	// Measure time after swap and wait for a few blocks to be produced.
	// This is needed to ensure that start time is before the block time
	// when we query for TWAP.
	timeAfterSwap := chainABNode.QueryLatestBlockTime()
	chainAB.WaitForNumHeights(2)

	// TWAP "from before to after swap" should be different from "from before to before swap"
	// because swap introduces a new record with a different spot price.
	s.T().Log("querying for the TWAP from before swap to now after swap")
	twapFromBeforeSwapToAfterSwapAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
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
	twapFromAfterToNowAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeAfterSwap)
	s.Require().NoError(err)
	// Because twapFromAfterToNow has a higher time weight for the after swap period,
	// we expect the results to be flipped from the previous comparison to twapFromBeforeSwapToBeforeSwapOne
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.LT(twapFromAfterToNowAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.GT(twapFromAfterToNowBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.LT(twapFromAfterToNowCA))

	s.T().Log("querying for the TWAP from after swap to after swap + 10ms")
	twapAfterSwapBeforePruning10MsAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
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
	chainAB.WaitForNumEpochs(1, epochIdentifier)

	// We should not be able to get TWAP before swap since it should have been pruned.
	s.T().Log("pruning is now complete, querying TWAP for period that should be pruned")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")

	// TWAPs for the same time range should be the same when we query for them before and after pruning.
	s.T().Log("querying for TWAP for period before pruning took place but should not have been pruned")
	twapAfterPruning10msAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	s.Require().Equal(twapAfterSwapBeforePruning10MsAB, twapAfterPruning10msAB)
	s.Require().Equal(twapAfterSwapBeforePruning10MsBC, twapAfterPruning10msBC)
	s.Require().Equal(twapAfterSwapBeforePruning10MsCA, twapAfterPruning10msCA)

	// TWAP "from after to after swap" should equal to "from after swap to after pruning"
	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	timeAfterPruning := chainABNode.QueryLatestBlockTime()
	s.T().Log("querying for TWAP from after swap to after pruning")
	twapToNowPostPruningAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningAB, twapAfterSwapBeforePruning10MsAB, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningBC, twapAfterSwapBeforePruning10MsBC, sdk.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningCA, twapAfterSwapBeforePruning10MsCA, sdk.NewDecWithPrec(1, 3))
}

func (s *IntegrationTestSuite) StateSync() {
	if s.skipStateSync {
		s.T().Skip()
	}

	// This test benefits from the use of chainA's default node, since it has
	// the shortest snapshot interval.
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.Require().NoError(err)

	persistentPeers := chainA.GetPersistentPeers()

	stateSyncHostPort := fmt.Sprintf("%s:26657", chainANode.Name)
	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

	// get trust height and trust hash.
	trustHeight, err := chainANode.QueryCurrentHeight()
	s.Require().NoError(err)

	trustHash, err := chainANode.QueryHashFromBlock(trustHeight)
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
		filepath.Join(chainANode.ConfigDir, "config", "genesis.json"),
		stateSynchingNodeConfig,
		time.Duration(chainA.VotingPeriod),
		// time.Duration(chainA.ExpeditedVotingPeriod),
		trustHeight,
		trustHash,
		stateSyncRPCServers,
		persistentPeers,
	)
	s.Require().NoError(err)

	// Call tempNode method here to not add the node to the list of nodes.
	// This messes with the nodes running in parallel if we add it to the regular list.
	stateSynchingNode := chainA.CreateNodeTemp(nodeInit)

	// ensure that the running node has snapshots at a height > trustHeight.
	hasSnapshotsAvailable := func(syncInfo coretypes.SyncInfo) bool {
		snapshotHeight := chainANode.SnapshotInterval
		if uint64(syncInfo.LatestBlockHeight) < snapshotHeight {
			s.T().Logf("snapshot height is not reached yet, current (%d), need (%d)", syncInfo.LatestBlockHeight, snapshotHeight)
			return false
		}

		snapshots, err := chainANode.QueryListSnapshots()
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
	chainANode.WaitUntil(hasSnapshotsAvailable)

	// start the state synchin node.
	err = stateSynchingNode.Run()
	s.Require().NoError(err)

	// ensure that the state synching node cathes up to the running node.
	s.Require().Eventually(func() bool {
		stateSyncNodeHeight, err := stateSynchingNode.QueryCurrentHeight()
		s.Require().NoError(err)
		runningNodeHeight, err := chainANode.QueryCurrentHeight()
		s.Require().NoError(err)
		return stateSyncNodeHeight == runningNodeHeight
	},
		1*time.Minute,
		10*time.Millisecond,
	)

	// stop the state synching node.
	err = chainA.RemoveTempNode(stateSynchingNode.Name)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) ExpeditedProposals() {
	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	propNumber := chainABNode.SubmitTextProposal("expedited text proposal", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinExpeditedDeposit)), true)

	chainABNode.DepositProposal(propNumber, true)
	totalTimeChan := make(chan time.Duration, 1)
	go chainABNode.QueryPropStatusTimed(propNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)

	var wg sync.WaitGroup

	for _, n := range chainAB.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *chain.NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	var elapsed time.Duration
	timeoutPeriod := 2 * time.Minute
	select {
	case elapsed = <-totalTimeChan:
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	}

	// compare the time it took to reach pass status to expected expedited voting period
	expeditedVotingPeriodDuration := time.Duration(chainAB.ExpeditedVotingPeriod * float32(time.Second))
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
func (s *IntegrationTestSuite) GeometricTWAP() {
	const (
		// This pool contains 1_000_000 uosmo and 2_000_000 stake.
		// Equals weights.
		poolFile   = "geometricPool.json"
		walletName = "geometric-twap-wallet"

		denomA = "uosmo" // 1_000_000 uosmo
		denomB = "stake" // 2_000_000 stake

		minAmountOut = "1"
	)

	chainAB, chainABNode, err := s.getChainCfgs()
	s.Require().NoError(err)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	// Triggers the creation of TWAP records.
	poolId := chainABNode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainABNode.CreateWalletAndFund(walletName, []string{initialization.WalletFeeTokens.String()}, chainAB)

	// We add 5 ms to avoid landing directly on block time in twap. If block time
	// is provided as start time, the latest spot price is used. Otherwise
	// interpolation is done.
	timeBeforeSwapPlus5ms := chainABNode.QueryLatestBlockTime().Add(5 * time.Millisecond)
	s.T().Log("geometric twap, start time ", timeBeforeSwapPlus5ms.Unix())

	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainAB.WaitForNumHeights(4)

	s.T().Log("querying for the first geometric TWAP to now (before swap)")
	// Assume base = uosmo, quote = stake
	// At pool creation time, the twap should be:
	// quote assset supply / base asset supply = 2_000_000 / 1_000_000 = 2
	curBlockTime := chainABNode.QueryLatestBlockTime().Unix()
	s.T().Log("geometric twap, end time ", curBlockTime)

	initialTwapBOverA, err := chainABNode.QueryGeometricTwapToNow(poolId, denomA, denomB, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(2), initialTwapBOverA)

	// Assume base = stake, quote = uosmo
	// At pool creation time, the twap should be:
	// quote assset supply / base asset supply = 1_000_000 / 2_000_000 = 0.5
	initialTwapAOverB, err := chainABNode.QueryGeometricTwapToNow(poolId, denomB, denomA, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDecWithPrec(5, 1), initialTwapAOverB)

	coinAIn := fmt.Sprintf("1000000%s", denomA)
	chainABNode.BankSend(coinAIn, sender, swapWalletAddr)

	s.T().Logf("performing swap of %s for %s", coinAIn, denomB)

	// stake out = stake supply * (1 - (uosmo supply before / uosmo supply after)^(uosmo weight / stake weight))
	//           = 2_000_000 * (1 - (1_000_000 / 2_000_000)^1)
	//           = 2_000_000 * 0.5
	//           = 1_000_000
	chainABNode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)

	// New supply post swap:
	// stake = 2_000_000 - 1_000_000 - 1_000_000
	// uosmo = 1_000_000 + 1_000_000 = 2_000_000

	timeAfterSwap := chainABNode.QueryLatestBlockTime()
	chainAB.WaitForNumHeights(4)
	timeAfterSwapPlus1Height := chainABNode.QueryLatestBlockTime()

	s.T().Log("querying for the TWAP from after swap to now")
	afterSwapTwapBOverA, err := chainABNode.QueryGeometricTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwapPlus1Height)
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

func (s *IntegrationTestSuite) SetDefaultTakerFeeBothChains() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Chain A

	go func() {
		defer wg.Done()
		chainA, chainANode, err := s.getChainACfgs()
		s.Require().NoError(err)
		s.SetDefaultTakerFee(chainA, chainANode)
	}()

	// Chain B

	go func() {
		defer wg.Done()
		chainB, chainBNode, err := s.getChainBCfgs()
		s.Require().NoError(err)
		s.SetDefaultTakerFee(chainB, chainBNode)
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}

func (s *IntegrationTestSuite) SetDefaultTakerFee(chain *chain.Config, chainNode *chain.NodeConfig) {
	// Change the parameter to set the default taker fee to a non zero value

	err := chainNode.ParamChangeProposal("poolmanager", string(poolmanagertypes.KeyDefaultTakerFee), json.RawMessage(`"0.001500000000000000"`), chain)
	s.Require().NoError(err)
}
