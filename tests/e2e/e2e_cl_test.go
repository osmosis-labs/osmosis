package e2e

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/osmomath"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
	clmath "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// Note: do not use chain B in this test as it has taker fee set.
// This TWAP test depends on specific values that might be affected
// by the taker fee.
func (s *IntegrationTestSuite) CreateConcentratedLiquidityPoolVoting_And_TWAP() {
	chainA, chainANode := s.getChainACfgs()

	poolId, err := chainA.SubmitCreateConcentratedPoolProposal(chainANode, true)
	s.NoError(err)
	fmt.Println("poolId", poolId)

	var (
		expectedDenom0       = "stake"
		expectedDenom1       = "note"
		expectedTickspacing  = uint64(100)
		expectedSpreadFactor = "0.001000000000000000"
	)

	var concentratedPool cltypes.ConcentratedPoolExtension
	s.Eventually(
		func() bool {
			concentratedPool = s.updatedConcentratedPool(chainANode, poolId)
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

	fundTokens := []string{"100000000stake", "100000000note"}

	// Get address to create positions
	address1 := chainANode.CreateWalletAndFund("address1", fundTokens, chainA)

	// We add 5 ms to avoid landing directly on block time in twap. If block time
	// is provided as start time, the latest spot price is used. Otherwise
	// interpolation is done.
	timeBeforePositionCreationBeforeSwap := chainANode.QueryLatestBlockTime().Add(5 * time.Millisecond)
	s.T().Log("geometric twap, start time ", timeBeforePositionCreationBeforeSwap.Unix())

	// Wait for the next height so that the requested twap
	// start time (timeBeforePositionCreationBeforeSwap) is not equal to the block time.
	chainA.WaitForNumHeights(1)

	// Check initial TWAP
	// We expect this to error since there is no spot price yet.
	s.T().Log("initial twap check")
	initialTwapBOverA, err := chainANode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeBeforePositionCreationBeforeSwap)
	s.Require().Error(err)
	s.Require().Equal(osmomath.Dec{}, initialTwapBOverA)

	// Create a position and check that TWAP now returns a value.
	s.T().Log("creating first position")
	chainANode.CreateConcentratedPosition(address1, "[-120000]", "40000", fmt.Sprintf("10000000%s,20000000%s", concentratedPool.GetToken0(), concentratedPool.GetToken1()), 0, 0, concentratedPool.GetId())
	timeAfterPositionCreationBeforeSwap := chainANode.QueryLatestBlockTime()
	chainA.WaitForNumHeights(2)
	firstPositionTwapBOverA, err := chainANode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterPositionCreationBeforeSwap)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.5"), firstPositionTwapBOverA)

	// Run a swap and check that the TWAP updates.
	s.T().Log("run swap")
	coinAIn := fmt.Sprintf("1000000%s", concentratedPool.GetToken0())
	chainANode.SwapExactAmountIn(coinAIn, "1", fmt.Sprintf("%d", concentratedPool.GetId()), concentratedPool.GetToken1(), address1)

	timeAfterSwap := chainANode.QueryLatestBlockTime()
	chainANode.WaitForNumHeights(1)
	timeAfterSwapPlus1Height := chainANode.QueryLatestBlockTime()
	chainANode.WaitForNumHeights(1)

	s.T().Log("querying for the TWAP after swap")
	afterSwapTwapBOverA, err := chainANode.QueryGeometricTwap(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwap, timeAfterSwapPlus1Height)
	s.Require().NoError(err)

	// We swap stake so note's supply will decrease and stake will increase.
	// The price after will be larger than the previous one.
	s.Require().True(afterSwapTwapBOverA.GT(firstPositionTwapBOverA))

	// Remove the position and check that TWAP returns an error.
	s.T().Log("removing first position (pool is drained)")
	positions := chainANode.QueryConcentratedPositions(address1)
	chainANode.WithdrawPosition(address1, positions[0].Position.Liquidity.String(), positions[0].Position.PositionId)
	chainANode.WaitForNumHeights(1)

	s.T().Log("querying for the TWAP from after pool drained")
	afterRemoveTwapBOverA, err := chainANode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwapPlus1Height)
	s.Require().Error(err)
	s.Require().Equal(osmomath.Dec{}, afterRemoveTwapBOverA)

	// Create a position and check that TWAP now returns a value.
	s.T().Log("creating position")
	chainANode.CreateConcentratedPosition(address1, "[-120000]", "40000", fmt.Sprintf("20000000%s,10000000%s", concentratedPool.GetToken0(), concentratedPool.GetToken1()), 0, 0, concentratedPool.GetId())
	chainANode.WaitForNumHeights(1)
	timeAfterSwapRemoveAndCreate := chainANode.QueryLatestBlockTime()
	chainANode.WaitForNumHeights(1)
	secondTwapBOverA, err := chainANode.QueryGeometricTwapToNow(concentratedPool.GetId(), concentratedPool.GetToken1(), concentratedPool.GetToken0(), timeAfterSwapRemoveAndCreate)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewDec(2), secondTwapBOverA)
}

// Note: this test depends on taker fee being set.
// As a result, we use chain B. Chain A has zero taker fee.
// TODO: Move this test and its components to its own file, Its way too big and needs to be split up significantly.
func (s *IntegrationTestSuite) ConcentratedLiquidity() {
	var (
		denom0                 = "uion"
		denom1                 = "note"
		tickSpacing     uint64 = 100
		spreadFactor           = "0.001" // 0.1%
		spreadFactorDec        = osmomath.MustNewDecFromStr("0.001")
		takerFee               = osmomath.MustNewDecFromStr("0.0015")
	)

	// Use chain B node since it has taker fee enabled.
	chainB, chainBNode := s.getChainBCfgs()
	var adminWalletAddr string

	enablePermissionlessCl := func() {
		// Get the permisionless pool creation parameter.
		isPermisionlessCreationEnabledStr := chainBNode.QueryParams(cltypes.ModuleName, string(cltypes.KeyIsPermisionlessPoolCreationEnabled))
		if !strings.EqualFold(isPermisionlessCreationEnabledStr, "true") {
			// Change the parameter to enable permisionless pool creation.
			err := chainBNode.ParamChangeProposal("concentratedliquidity", string(cltypes.KeyIsPermisionlessPoolCreationEnabled), []byte("true"), chainB, true)
			s.Require().NoError(err)
		}

		// Confirm that the parameter has been changed.
		isPermisionlessCreationEnabledStr = chainBNode.QueryParams(cltypes.ModuleName, string(cltypes.KeyIsPermisionlessPoolCreationEnabled))
		if !strings.EqualFold(isPermisionlessCreationEnabledStr, "true") {
			s.T().Fatal("concentrated liquidity pool creation is not enabled")
		}

		go func() {
			s.T().Run("test update pool tick spacing", func(t *testing.T) {
				s.TickSpacingUpdateProp()
			})
		}()
	}

	changeProtorevAdminAndMaxPoolPoints := func() {
		// Update the protorev admin address to a known wallet we can control
		adminWalletAddr = chainBNode.CreateWalletAndFund("admin", []string{"4000000note"}, chainB)
		err := chainBNode.ParamChangeProposal("protorev", string(protorevtypes.ParamStoreKeyAdminAccount), []byte(fmt.Sprintf(`"%s"`, adminWalletAddr)), chainB, true)
		s.Require().NoError(err)

		// Update the weight of CL pools so that this test case is not back run by protorev.
		chainBNode.SetMaxPoolPointsPerTx(7, adminWalletAddr)
	}
	defer func() {
		// Reset the maximum number of pool points
		chainBNode.SetMaxPoolPointsPerTx(int(protorevtypes.DefaultMaxPoolPointsPerTx), adminWalletAddr)
	}()

	enablePermissionlessCl()
	changeProtorevAdminAndMaxPoolPoints()

	// Create concentrated liquidity pool when permisionless pool creation is enabled.
	poolID := chainBNode.CreateConcentratedPool(initialization.ValidatorWalletName, denom0, denom1, tickSpacing, spreadFactor)
	concentratedPool := s.updatedConcentratedPool(chainBNode, poolID)

	// Sanity check that pool initialized with valid parameters (the ones that we haven't explicitly specified)
	s.Require().Equal(concentratedPool.GetCurrentTick(), int64(0))
	s.Require().Equal(concentratedPool.GetCurrentSqrtPrice(), osmomath.ZeroBigDec())
	s.Require().Equal(concentratedPool.GetLiquidity(), osmomath.ZeroDec())

	// Assert contents of the pool are valid (that we explicitly specified)
	s.Require().Equal(concentratedPool.GetId(), poolID)
	s.Require().Equal(concentratedPool.GetToken0(), denom0)
	s.Require().Equal(concentratedPool.GetToken1(), denom1)
	s.Require().Equal(concentratedPool.GetTickSpacing(), tickSpacing)
	s.Require().Equal(concentratedPool.GetSpreadFactor(sdk.Context{}), osmomath.MustNewDecFromStr(spreadFactor))

	fundTokens := []string{"100000000note", "100000000uion", "100000000stake"}

	// Get 3 addresses to create positions
	address1 := chainBNode.CreateWalletAndFund("addr1", fundTokens, chainB)
	address2 := chainBNode.CreateWalletAndFund("addr2", fundTokens, chainB)
	address3 := chainBNode.CreateWalletAndFund("addr3", fundTokens, chainB)
	addresses := []string{address1, address2, address3}

	// When claiming rewards, a small portion of dust is forfeited and is redistributed to everyone. We must track the total
	// liquidity across all positions (even if not active), in order to calculate how much to increase the reward growth global per share by.
	totalLiquidity := osmomath.ZeroDec()

	// not sure what this is
	createPosFormat := fmt.Sprintf("10000000%s,10000000%s", denom0, denom1)
	createPosition := func(address string, lower int, upper int) math.LegacyDec {
		_, liquidity := chainBNode.CreateConcentratedPosition(address, formatCLIInt(lower), formatCLIInt(upper), createPosFormat, 0, 0, poolID)
		return liquidity
	}

	// Create 2 positions for address1: overlap together, overlap with 2 address3 positions
	type clposition struct {
		lower int
		upper int
	}
	positions := [][]clposition{
		{{-120000, 40000}, {-40000, 120000}},
		{{220000, int(cltypes.MaxTick)}},
		{{-160000, -20000}, {int(cltypes.MinInitializedTick), 140000}},
	}
	createdPositions := [][]model.FullPositionBreakdown{{}, {}, {}}
	// Create all positions, with each address' positions created in sequence, but all addresses' created concurrently
	var clwg sync.WaitGroup
	var mu sync.Mutex

	for i := range positions {
		clwg.Add(1)
		go func(i int) { // Launch a goroutine
			addr, positionSet := addresses[i], positions[i]
			defer clwg.Done() // Decrement the counter when the goroutine completes
			userLiquidity := math.LegacyZeroDec()
			for _, j := range positionSet {
				liquidity := createPosition(addr, j.lower, j.upper)
				userLiquidity.AddMut(liquidity)
			}
			mu.Lock() // Lock totalLiquidity for concurrent write
			totalLiquidity.AddMut(userLiquidity)
			mu.Unlock() // Unlock after write
			createdPositions[i] = chainBNode.QueryConcentratedPositions(addr)
		}(i)
	}

	clwg.Wait() // Block until all goroutines complete

	concentratedPool = s.updatedConcentratedPool(chainBNode, poolID)

	for i, posSet := range positions {
		s.Require().Equal(len(createdPositions[i]), len(posSet))
		for j, pos := range posSet {
			s.validateCLPosition(createdPositions[i][j].Position, poolID, int64(pos.lower), int64(pos.upper))
		}
	}

	// compat with old code
	positionsAddress1 := createdPositions[0]
	positionsAddress2 := createdPositions[1]
	positionsAddress3 := createdPositions[2]

	// Collect SpreadRewards

	var (
		// spreadRewardGrowthGlobal is a variable for tracking global spread reward growth
		spreadRewardGrowthGlobal = osmomath.ZeroDec()
		outMinAmt                = "1"
	)

	// Swap 1
	// Not crossing initialized ticks => performed in one swap step
	// Swap affects 3 positions: both that address1 has and one of address3's positions
	// Asserts that spread rewards are correctly collected for non cross-tick swaps
	var (
		// Swap parameters
		noteInDec_Swap1 = osmomath.NewBigDec(3465198)
		noteIn_Swap1    = fmt.Sprintf("%snote", noteInDec_Swap1.Dec().String())
	)
	// Perform swap (not crossing initialized ticks)
	chainBNode.SwapExactAmountIn(noteIn_Swap1, outMinAmt, fmt.Sprintf("%d", poolID), denom0, initialization.ValidatorWalletName)
	// Calculate and track global spread reward growth for swap 1
	noteInDec_Swap1_SubTakerFee := noteInDec_Swap1.Dec().Mul(osmomath.OneDec().Sub(takerFee)).TruncateDec()
	noteInDec_Swap1_SubTakerFee_SubSpreadFactor := noteInDec_Swap1_SubTakerFee.Mul(osmomath.OneDec().Sub(spreadFactorDec))
	totalSpreadReward := noteInDec_Swap1_SubTakerFee.Sub(noteInDec_Swap1_SubTakerFee_SubSpreadFactor).TruncateDec()

	spreadRewardGrowthGlobal.AddMut(calculateSpreadRewardGrowthGlobal(totalSpreadReward, concentratedPool.GetLiquidity()))
	// Check swap properties
	expectedSqrtPriceDelta := osmomath.BigDecFromDec(noteInDec_Swap1_SubTakerFee_SubSpreadFactor).QuoTruncate(osmomath.BigDecFromDec(concentratedPool.GetLiquidity())) // Δ(sqrtPrice) = Δy / L
	concentratedPoolAfterSwap := s.updatedConcentratedPool(chainBNode, poolID)
	s.assertClSwap(concentratedPool, concentratedPoolAfterSwap, expectedSqrtPriceDelta)
	concentratedPool = concentratedPoolAfterSwap

	// Collect SpreadRewards: Swap 1

	// Track balances for address1 position1
	addr1BalancesBefore := s.addrBalance(chainBNode, address1)
	chainBNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter := s.addrBalance(chainBNode, address1)

	// Assert that the balance changed only for tokenIn (note)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, false, true)

	// Assert Balances: Swap 1

	// Calculate uncollected spread rewards for address1 position1
	spreadRewardsUncollectedAddress1Position1_Swap1 := calculateUncollectedSpreadRewards(
		positionsAddress1[0].Position.Liquidity,
		osmomath.ZeroDec(), // no growth below
		osmomath.ZeroDec(), // no growth above
		osmomath.ZeroDec(), // no spreadRewardGrowthInsideLast - it is the first swap
		spreadRewardGrowthGlobal,
	)

	// Note the global spread reward growth before dust redistribution
	spreadRewardGrowthGlobalBeforeDustRedistribution := spreadRewardGrowthGlobal.Clone()

	// Determine forfeited dust amount
	forfeitedDustAmt := spreadRewardsUncollectedAddress1Position1_Swap1.Sub(spreadRewardsUncollectedAddress1Position1_Swap1.TruncateDec())
	forfeitedDust := sdk.NewDecCoins(sdk.NewDecCoinFromDec("note", forfeitedDustAmt))
	forfeitedDustPerShare := forfeitedDust.QuoDecTruncate(totalLiquidity)

	// Add forfeited dust back to the global spread reward growth
	spreadRewardGrowthGlobal.AddMut(forfeitedDustPerShare.AmountOf("note"))

	// Assert
	s.Require().Equal(
		addr1BalancesBefore.AmountOf("note").Add(spreadRewardsUncollectedAddress1Position1_Swap1.TruncateInt()).String(),
		addr1BalancesAfter.AmountOf("note").String(),
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
	sqrtPriceBeforeSwap := concentratedPool.GetCurrentSqrtPrice()
	liquidityBeforeSwap := concentratedPool.GetLiquidity()
	nextInitTick := int64(40000) // address1 position1's upper tick

	// Calculate sqrtPrice after and at the next initialized tick (upperTick of address1 position1 - 40000)
	sqrtPriceAfterNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick + tickOffset)
	s.Require().NoError(err)
	sqrtPriceAtNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick)
	s.Require().NoError(err)
	sqrtPriceAfterNextInitializedTickBigDec := sqrtPriceAfterNextInitializedTick
	sqrtPriceAtNextInitializedTickBigDec := sqrtPriceAtNextInitializedTick

	// Calculate Δ(sqrtPrice):
	// deltaSqrtPriceAfterNextInitializedTick = ΔsqrtP(40300) - ΔsqrtP(40000)
	// deltaSqrtPriceAtNextInitializedTick = ΔsqrtP(40000) - ΔsqrtP(currentTick)
	deltaSqrtPriceAfterNextInitializedTick := sqrtPriceAfterNextInitializedTickBigDec.Sub(sqrtPriceAtNextInitializedTickBigDec).Dec()
	deltaSqrtPriceAtNextInitializedTick := sqrtPriceAtNextInitializedTickBigDec.Sub(sqrtPriceBeforeSwap).Dec()

	// Calculate the amount of melody required to:
	// * amountInToGetToTickAfterInitialized - move price from next initialized tick (40000) to destination tick (40000 + tickOffset)
	// * amountInToGetToNextInitTick - move price from current tick to next initialized tick
	// Formula is as follows:
	// Δy = L * Δ(sqrtPrice)
	amountInToGetToTickAfterInitialized := deltaSqrtPriceAfterNextInitializedTick.Mul(liquidityBeforeSwap.Sub(positionsAddress1[0].Position.Liquidity))
	amountInToGetToNextInitTick := deltaSqrtPriceAtNextInitializedTick.Mul(liquidityBeforeSwap)

	var (
		// Swap parameters

		// noteInDec_Swap2_NoSpreadReward is calculated such that swapping this amount (not considering spread reward) moves the price over the next initialized tick
		noteInDec_Swap2_NoSpreadReward = amountInToGetToNextInitTick.Add(amountInToGetToTickAfterInitialized)
		noteInDec_Swap2                = noteInDec_Swap2_NoSpreadReward.Quo(osmomath.OneDec().Sub(spreadFactorDec)).TruncateDec() // account for spread factor of 1%

		spreadRewardGrowthGlobal_Swap1 = spreadRewardGrowthGlobalBeforeDustRedistribution.Clone()
	)

	noteInDec_Swap2_AddTakerFee := noteInDec_Swap2.Quo(osmomath.OneDec().Sub(takerFee)).TruncateDec() // account for taker fee
	noteIn_Swap2 := fmt.Sprintf("%snote", noteInDec_Swap2_AddTakerFee.String())

	// Perform a swap
	chainBNode.SwapExactAmountIn(noteIn_Swap2, outMinAmt, fmt.Sprintf("%d", poolID), denom0, initialization.ValidatorWalletName)

	// Calculate the amount of liquidity of the position that was kicked out during swap (address1 position1)
	liquidityOfKickedOutPosition := positionsAddress1[0].Position.Liquidity

	// Update pool and track pool's liquidity
	concentratedPool = s.updatedConcentratedPool(chainBNode, poolID)

	liquidityAfterSwap := concentratedPool.GetLiquidity()

	// Assert that net liquidity of kicked out position was successfully removed from current pool's liquidity
	s.Require().Equal(liquidityBeforeSwap.Sub(liquidityOfKickedOutPosition), liquidityAfterSwap)

	// Collect spread rewards: Swap 2

	// Calculate spread reward charges per each step

	// Step1: amountIn is note tokens that are swapped + note tokens that are paid for spread reward
	// hasReachedTarget in SwapStep is true, hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountIn * spreadFactor / (1 - spreadFactor)
	spreadRewardCharge_Swap2_Step1 := amountInToGetToNextInitTick.Mul(spreadFactorDec).Quo(osmomath.OneDec().Sub(spreadFactorDec))

	// Step2: hasReachedTarget in SwapStep is false (nextTick is 120000), hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountRemaining - amountOne
	amountRemainingAfterStep1 := noteInDec_Swap2.Sub(amountInToGetToNextInitTick).Sub(spreadRewardCharge_Swap2_Step1)
	spreadRewardCharge_Swap2_Step2 := amountRemainingAfterStep1.Sub(amountInToGetToTickAfterInitialized)

	// per unit of virtual liquidity
	spreadRewardCharge_Swap2_Step1.QuoMut(liquidityBeforeSwap)
	spreadRewardCharge_Swap2_Step2.QuoMut(liquidityAfterSwap)

	// Update spreadRewardGrowthGlobal
	spreadRewardGrowthGlobal.AddMut(spreadRewardCharge_Swap2_Step1.Add(spreadRewardCharge_Swap2_Step2))

	// Assert Balances: Swap 2

	// Assert that address1 position1 earned spread rewards only from first swap step

	// Track balances for address1 position1
	addr1BalancesBefore = s.addrBalance(chainBNode, address1)
	chainBNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter = s.addrBalance(chainBNode, address1)

	// Assert that the balance changed only for tokenIn (note)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, false, true)

	// Calculate uncollected spread rewards for position, which liquidity will only be live part of the swap
	spreadRewardsUncollectedAddress1Position1_Swap2 := calculateUncollectedSpreadRewards(
		positionsAddress1[0].Position.Liquidity,
		osmomath.ZeroDec(),
		spreadRewardCharge_Swap2_Step2,
		spreadRewardGrowthGlobal_Swap1,
		spreadRewardGrowthGlobal,
	)

	// Assert
	s.Require().Equal(
		addr1BalancesBefore.AmountOf("note").Add(spreadRewardsUncollectedAddress1Position1_Swap2.TruncateInt()),
		addr1BalancesAfter.AmountOf("note"),
	)

	// Assert that address3 position2 earned rewards from first and second swaps

	// Track balance off address3 position2: check that position that has not been kicked out earned full rewards
	addr3BalancesBefore := s.addrBalance(chainBNode, address3)
	chainBNode.CollectSpreadRewards(address3, fmt.Sprint(positionsAddress3[1].Position.PositionId))
	addr3BalancesAfter := s.addrBalance(chainBNode, address3)

	// Calculate uncollected spread rewards for address3 position2 earned from Swap 1
	spreadRewardsUncollectedAddress3Position2_Swap1 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		osmomath.ZeroDec(),
		osmomath.ZeroDec(),
		osmomath.ZeroDec(),
		spreadRewardGrowthGlobal_Swap1,
	)

	// Calculate uncollected spread rewards for address3 position2 (was active throughout both swap steps): Swap2
	spreadRewardsUncollectedAddress3Position2_Swap2 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		osmomath.ZeroDec(),
		osmomath.ZeroDec(),
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap1, osmomath.ZeroDec(), osmomath.ZeroDec()),
		spreadRewardGrowthGlobal,
	)

	// Total spread rewards earned by address3 position2 from 2 swaps
	totalUncollectedSpreadRewardsAddress3Position2 := spreadRewardsUncollectedAddress3Position2_Swap1.Add(spreadRewardsUncollectedAddress3Position2_Swap2)

	// Assert
	s.Require().Equal(
		addr3BalancesBefore.AmountOf("note").Add(totalUncollectedSpreadRewardsAddress3Position2.TruncateInt()),
		addr3BalancesAfter.AmountOf("note"),
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
	sqrtPricebBelowNextInitializedTick, err := clmath.TickToSqrtPrice(nextInitTick - tickOffset)
	s.Require().NoError(err)
	sqrtPriceAtNextInitializedTick, err = clmath.TickToSqrtPrice(nextInitTick)
	s.Require().NoError(err)
	sqrtPriceAtNextInitializedTickBigDec = sqrtPriceAtNextInitializedTick

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
	amountInToGetToTickBelowInitialized := liquidityBeforeSwap.Add(positionsAddress1[0].Position.Liquidity).Mul(fractionBelowNextInitializedTick.Dec())
	amountInToGetToNextInitTick = liquidityBeforeSwap.Mul(fractionAtNextInitializedTick.Dec())

	// Collect spread rewards for address1 position1 to avoid overhead computations (swap2 already asserted spread rewards are aggregated correctly from multiple swaps)
	chainBNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))

	var (
		// Swap parameters
		uionInDec_Swap3_NoSpreadReward = amountInToGetToNextInitTick.Add(amountInToGetToTickBelowInitialized)                     // amount of uion to move price from current to desired (not considering spreadFactor)
		uionInDec_Swap3                = uionInDec_Swap3_NoSpreadReward.Quo(osmomath.OneDec().Sub(spreadFactorDec)).TruncateDec() // consider spreadFactor

		// Save variables from previous swaps
		spreadRewardGrowthGlobal_Swap2                = spreadRewardGrowthGlobal.Clone()
		spreadRewardGrowthInsideAddress1Position1Last = spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap2_Step2).Clone()
	)

	uionInDec_Swap3_AddTakerFee := uionInDec_Swap3.Quo(osmomath.OneDec().Sub(takerFee)).TruncateDec() // account for taker fee
	uionIn_Swap3 := fmt.Sprintf("%suion", uionInDec_Swap3_AddTakerFee.String())

	// Perform a swap
	chainBNode.SwapExactAmountIn(uionIn_Swap3, outMinAmt, fmt.Sprintf("%d", poolID), denom1, initialization.ValidatorWalletName)

	// Assert liquidity of kicked in position was successfully added to the pool
	concentratedPool = s.updatedConcentratedPool(chainBNode, poolID)

	liquidityAfterSwap = concentratedPool.GetLiquidity()
	s.Require().Equal(liquidityBeforeSwap.Add(positionsAddress1[0].Position.Liquidity), liquidityAfterSwap)

	// Track balance of address1
	addr1BalancesBefore = s.addrBalance(chainBNode, address1)
	chainBNode.CollectSpreadRewards(address1, fmt.Sprint(positionsAddress1[0].Position.PositionId))
	addr1BalancesAfter = s.addrBalance(chainBNode, address1)

	// Assert that the balance changed only for tokenIn (uion)
	s.assertBalancesInvariants(addr1BalancesBefore, addr1BalancesAfter, true, false)

	// Assert the amount of collected spread rewards:

	// Step1: amountIn is uion tokens that are swapped + uion tokens that are paid for spread reward
	// hasReachedTarget in SwapStep is true, hence, to find spread rewards, calculate:
	// spreadRewardCharge = amountIn * spreadFactor / (1 - spreadFactor)
	spreadRewardCharge_Swap3_Step1 := amountInToGetToNextInitTick.Mul(spreadFactorDec).Quo(osmomath.OneDec().Sub(spreadFactorDec))

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
		osmomath.ZeroDec(),
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
	addr3BalancesBefore = s.addrBalance(chainBNode, address3)
	chainBNode.CollectSpreadRewards(address3, fmt.Sprint(positionsAddress3[1].Position.PositionId))
	addr3BalancesAfter = s.addrBalance(chainBNode, address3)

	// Assert that the balance changed only for tokenIn (uion)
	s.assertBalancesInvariants(addr3BalancesBefore, addr3BalancesAfter, true, false)

	// Was active throughout the whole swap, collects spread rewards from 2 steps

	// Step 1
	spreadRewardsUncollectedAddress3Position2_Swap3_Step1 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		osmomath.ZeroDec(), // no growth below
		osmomath.ZeroDec(), // no growth above
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap2, osmomath.ZeroDec(), osmomath.ZeroDec()), // snapshot of spread reward growth at swap 2
		spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap3_Step2),                                              // step 1 hasn't earned spread rewards from step 2
	)

	// Step 2
	spreadRewardsUncollectedAddress3Position2_Swap3_Step2 := calculateUncollectedSpreadRewards(
		positionsAddress3[1].Position.Liquidity,
		osmomath.ZeroDec(), // no growth below
		osmomath.ZeroDec(), // no growth above
		calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal_Swap2, osmomath.ZeroDec(), osmomath.ZeroDec()), // snapshot of spread reward growth at swap 2
		spreadRewardGrowthGlobal.Sub(spreadRewardCharge_Swap3_Step1),                                              // step 2 hasn't earned spread rewards from step 1
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
	assertUserThree := func() { s.ensureZeroRewardSpreads(chainBNode, address3, positionsAddress3[0].Position.PositionId) }

	// Address2's only position: [220000; 342000]
	assertUserTwo := func() { s.ensureZeroRewardSpreads(chainBNode, address2, positionsAddress2[0].Position.PositionId) }

	runFuncsInParallelAndBlock([]func(){assertUserThree, assertUserTwo})
	// Withdraw Position

	defaultLiquidityRemoval := "1000"
	chainB.WaitForNumHeights(1)

	// Assert removing some liquidity
	// 1) remove default liquidity from the 0th position of every address
	// 2) Afterwards, Remove the entire 0th position for addr 1 and 2
	clwg = sync.WaitGroup{}
	for i := 0; i < len(createdPositions); i++ {
		clwg.Add(1)
		go func(i int) { // Launch a goroutine
			defer clwg.Done()
			addr := addresses[i]
			posSet := createdPositions[i]
			posLiquidityBefore := posSet[0].Position.Liquidity
			chainBNode.WithdrawPosition(addr, defaultLiquidityRemoval, posSet[0].Position.PositionId)
			// assert
			createdPositions[i] = chainBNode.QueryConcentratedPositions(addr)
			s.Require().Equal(posLiquidityBefore, createdPositions[i][0].Position.Liquidity.Add(osmomath.MustNewDecFromStr(defaultLiquidityRemoval)))

			// remove 0th position for addr 1 and 2
			if i >= 2 {
				return
			}
			posLiquidity := createdPositions[i][0].Position.Liquidity
			chainBNode.WithdrawPosition(addr, posLiquidity.String(), createdPositions[i][0].Position.PositionId)
			finalPosition := chainBNode.QueryConcentratedPositions(addr)
			s.Require().Equal(len(finalPosition), len(createdPositions[i])-1)
		}(i)
	}
	clwg.Wait()
}

// This must be spawned from CL update test suite since it depends on permissionless pool creation
func (s *IntegrationTestSuite) TickSpacingUpdateProp() {
	var (
		denom0              = "uion"
		denom1              = "note"
		tickSpacing  uint64 = 100
		spreadFactor        = "0.001" // 0.1%
	)

	chainB, chainBNode := s.getChainBCfgs()

	// Test tick spacing reduction proposal
	poolID := chainBNode.CreateConcentratedPool(initialization.ValidatorWalletName, denom0, denom1, tickSpacing, spreadFactor)
	concentratedPool := s.updatedConcentratedPool(chainBNode, poolID)
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
	propNumber := chainBNode.SubmitTickSpacingReductionProposal(fmt.Sprintf("%d,%d", poolID, newTickSpacing), false, true)

	// TODO: simplify just querying w/ timeout
	totalTimeChan := make(chan time.Duration, 1)
	go chainBNode.QueryPropStatusTimed(propNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)

	chain.AllValsVoteOnProposal(chainB, propNumber)

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
	concentratedPool = s.updatedConcentratedPool(chainBNode, poolID)
	s.Require().Equal(newTickSpacing, concentratedPool.GetTickSpacing())
}

func (s *IntegrationTestSuite) assertClSwap(clpoolStart types.ConcentratedPoolExtension, clPoolAfter types.ConcentratedPoolExtension, expectedSqrtPriceDelta osmomath.BigDec) {
	// Update pool and track liquidity and sqrt price
	liquidityBeforeSwap := clpoolStart.GetLiquidity()
	sqrtPriceBeforeSwap := clpoolStart.GetCurrentSqrtPrice()
	liquidityAfterSwap := clPoolAfter.GetLiquidity()
	sqrtPriceAfterSwap := clPoolAfter.GetCurrentSqrtPrice()

	// Assert swaps don't change pool's liquidity amount
	s.Require().Equal(liquidityAfterSwap.String(), liquidityBeforeSwap.String())

	// Assert current sqrt price
	expectedSqrtPrice := sqrtPriceBeforeSwap.Add(expectedSqrtPriceDelta)
	s.Require().Equal(expectedSqrtPrice.String(), sqrtPriceAfterSwap.String())
}

// calculateSpreadRewardGrowthGlobal calculates spread reward growth global per unit of virtual liquidity based on swap parameters:
// amountIn - amount being swapped
// spreadFactor - pool's spread factor
// poolLiquidity - current pool liquidity
func calculateSpreadRewardGrowthGlobal(spreadRewardChargeTotal, poolLiquidity osmomath.Dec) osmomath.Dec {
	// First we get total spread reward charge for the swap (ΔY * spreadFactor)

	// Calculating spread reward growth global (dividing by pool liquidity to find spread reward growth per unit of virtual liquidity)
	spreadRewardGrowthGlobal := spreadRewardChargeTotal.QuoTruncate(poolLiquidity)
	return spreadRewardGrowthGlobal
}

// calculateSpreadRewardGrowthInside calculates spread reward growth inside range per unit of virtual liquidity
// spreadRewardGrowthGlobal - global spread reward growth per unit of virtual liquidity
// spreadRewardGrowthBelow - spread reward growth below lower tick
// spreadRewardGrowthAbove - spread reward growth above upper tick
// Formula: spreadRewardGrowthGlobal - spreadRewardGrowthBelowLowerTick - spreadRewardGrowthAboveUpperTick
func calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal, spreadRewardGrowthBelow, spreadRewardGrowthAbove osmomath.Dec) osmomath.Dec {
	return spreadRewardGrowthGlobal.Sub(spreadRewardGrowthBelow).Sub(spreadRewardGrowthAbove)
}

// Assert balances that are not affected by swap:
// * same amount of `stake` in balancesBefore and balancesAfter
// * amount of `e2e-default-feetoken` dropped by 1000 (default amount for fee per tx)
// * depending on `assertNoteBalanceIsConstant` and `assertUionBalanceIsConstant` parameters, check that those balances have also not been changed
func (s *IntegrationTestSuite) assertBalancesInvariants(balancesBefore, balancesAfter sdk.Coins, assertNoteBalanceIsConstant, assertUionBalanceIsConstant bool) {
	s.Require().True(balancesAfter.AmountOf("stake").Equal(balancesBefore.AmountOf("stake")))
	s.Require().True(balancesAfter.AmountOf("e2e-default-feetoken").Equal(balancesBefore.AmountOf("e2e-default-feetoken").Sub(defaultFeePerTx)))
	if assertUionBalanceIsConstant {
		s.Require().True(balancesAfter.AmountOf("uion").Equal(balancesBefore.AmountOf("uion")))
	}
	if assertNoteBalanceIsConstant {
		s.Require().True(balancesAfter.AmountOf("note").Equal(balancesBefore.AmountOf("note")))
	}
}

func (s *IntegrationTestSuite) ensureZeroRewardSpreads(node *chain.NodeConfig, addr string, positionId uint64) {
	addr2BalancesBefore := s.addrBalance(node, addr)
	node.CollectSpreadRewards(addr, fmt.Sprint(positionId))
	addr2BalancesAfter := s.addrBalance(node, addr)

	// Assert the balances did not change for every token
	s.assertBalancesInvariants(addr2BalancesBefore, addr2BalancesAfter, true, true)
}

// Get current (updated) pool
func (s *IntegrationTestSuite) updatedConcentratedPool(node *chain.NodeConfig, poolId uint64) types.ConcentratedPoolExtension {
	concentratedPool, err := node.QueryConcentratedPool(poolId)
	s.Require().NoError(err)
	return concentratedPool
}

// Assert returned positions:
func (s *IntegrationTestSuite) validateCLPosition(position model.Position, poolId uint64, lowerTick, upperTick int64) {
	s.Require().Equal(position.PoolId, poolId)
	s.Require().Equal(position.LowerTick, lowerTick)
	s.Require().Equal(position.UpperTick, upperTick)
}
