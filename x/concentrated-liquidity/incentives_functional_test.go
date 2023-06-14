package concentrated_liquidity_test

import (
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

type FunctionalIncentivesTestSuite struct {
	apptesting.KeeperTestHelper
}

type balancerLinkConfig struct {
	balancerOwner            sdk.AccAddress
	initialBalancerLiquidity sdk.Coins
	clPoolId                 uint64
	balancerPoolId           uint64
	tokenZero                sdk.Coin
	tokenOne                 sdk.Coin

	joinLPShares   sdk.Int
	bondedFraction sdk.Dec
}

type positionConfig struct {
	// These are configured at the start
	isFullRange bool
	owner       sdk.AccAddress
	poolId      uint64
	coins       sdk.Coins
	// These imply how many tick spacings away from the current tick the position is in
	// Negative means to the left of the current tick. Positive means to the right.
	lowerTickTickSpacingsAway int64
	upperTickTickSpacingsAway int64

	// These are set during configuration.
	positionId uint64
	liquidity  sdk.Dec

	// These ticks are the actual position ticks at which the position is created.
	lowerTick int64
	upperTick int64
}

type positionsConfig []*positionConfig

type incentiveRecordConfig struct {
	creator       sdk.AccAddress
	poolId        uint64
	emissionRate  sdk.Dec
	remainingCoin sdk.Coin
	startTime     time.Time
	uptime        time.Duration
}

type incentiveRecordsConfig []incentiveRecordConfig

type swapConfig struct {
	poolId  uint64
	swapper sdk.AccAddress
	tokenIn sdk.Coin

	// This is used to estimate the swap amount
	// From knowing the direction of the swap and
	/// boundTick that must be the next initialized tick
	// we can estimate the swap amount
	// The resulting swap amount is multiplied by this
	// amountInMultiplier to get the actual swap amount
	isZeroForOne       bool
	boundTick          int64
	amountInMultiplier sdk.Dec
}

type blockOperation interface {
	Apply(s *FunctionalIncentivesTestSuite)
}

var _ blockOperation = &balancerLinkConfig{}

type blockConfiguration struct {
	blockTime       time.Duration
	blockOperations []blockOperation

	propagateExecutionCtxNextBlock func()
}

func TestFunctionalIncentivesTestSuite(t *testing.T) {
	suite.Run(t, new(FunctionalIncentivesTestSuite))
}

func (s *FunctionalIncentivesTestSuite) TestIncentives_Functional_NoTickCrossing() {
	s.Setup()
	s.TestAccs = apptesting.CreateRandomAccounts(6)

	s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

	var (
		defaultPositionCoins = sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1))
	)

	// BLOCK 1:
	// 1. Create balancer link
	// 2. Create 3 positions: 1) full-range 2) in-range, 3) completely out of range
	// 3. Create incentive record
	//
	// BLOCK 2:
	// 4. Increase block time and perform a swap
	//
	// BLOCK 3:
	// 5. Increase block time and collect incentives for 1) 2) 3)
	//   * 1) claims correct amount
	//   * 2) claims correct amount
	//   * 3) claims nothing

	// BLOCK 1:

	// Create pools
	clPool := s.PrepareConcentratedPool()
	clPoolId := clPool.GetId()

	ownerOne := s.TestAccs[0]
	ownerTwo := s.TestAccs[1]
	ownerThree := s.TestAccs[2]
	balancerOwner := s.TestAccs[3]

	balancerConfig := balancerLinkConfig{
		clPoolId:                 clPoolId,
		balancerOwner:            balancerOwner,
		initialBalancerLiquidity: sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)),

		// Choosing joinLPShares to be an initial pool share supply implies that we
		// LP with 50% of total pool shares. By locking up half of that, we end up
		// with 25% of total pool shares locked.
		joinLPShares:   gammtypes.InitPoolSharesSupply,
		bondedFraction: sdk.NewDecWithPrec(5, 1),
	}

	positions := positionsConfig{
		{

			isFullRange: true,
			owner:       ownerOne,
			poolId:      clPoolId,
			coins:       defaultPositionCoins,
		},
		{
			owner:                     ownerTwo,
			poolId:                    clPoolId,
			coins:                     defaultPositionCoins,
			lowerTickTickSpacingsAway: -1,
			upperTickTickSpacingsAway: 1,
		},
		{
			owner:                     ownerThree,
			poolId:                    clPoolId,
			coins:                     defaultPositionCoins,
			lowerTickTickSpacingsAway: 10000,
			upperTickTickSpacingsAway: 10001,
		},
	}

	remainingCoin := sdk.NewCoin(ETH, sdk.NewInt(1000000000000000000))
	emissionRate := sdk.NewDec(1000000) // 1 per second
	incentiveRecordCreator := s.TestAccs[4]
	defaultIncentiveRecordConfig := incentiveRecordConfig{
		remainingCoin: remainingCoin,
		emissionRate:  emissionRate,
		poolId:        clPoolId,
		creator:       incentiveRecordCreator,
		startTime:     defaultBlockTime,
		uptime:        time.Nanosecond,
	}
	incentiveRecords := incentiveRecordsConfig{defaultIncentiveRecordConfig}

	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)

	// Perform a swap in the opposite direction of position 3 so that we never activate it.
	// This is to test that we don't claim rewards for inactive positions.
	// Since USDC is token 1 and ETH is token 0, we swap ETH for USDC to move to the left.
	swapper := s.TestAccs[5]
	swapConfig := swapConfig{
		isZeroForOne: true,
		swapper:      swapper,
		poolId:       clPoolId,

		amountInMultiplier: sdk.NewDecWithPrec(5, 1),

		// Note, the bound tick is set in-between blocks:
		// by propagateExecutionCtxNextBlock()
	}

	blocks := []blockConfiguration{
		{
			blockOperations: []blockOperation{
				// 1. Create balancer <> CL link
				&balancerConfig,

				// 2. Create positions
				&positions,

				// 3. Create incentive records
				&incentiveRecords,
			},
			blockTime: time.Second,
			propagateExecutionCtxNextBlock: func() {
				// For the swap in the next block, we use the second position's lower tick.
				swapConfig.boundTick = positions[1].lowerTick
			},
		},
		{
			blockOperations: []blockOperation{
				&swapConfig,
			},

			blockTime:                      time.Second,
			propagateExecutionCtxNextBlock: func() {},
		},
	}

	for _, block := range blocks {
		for _, blockOp := range block.blockOperations {
			blockOp.Apply(s)
		}

		// Block time is increased by the desired block time.
		s.CommitWithBlockTime(block.blockTime)

		// Propagate context from previous block execution to next
		block.propagateExecutionCtxNextBlock()
	}

	positionOne := positions[0]
	positionTwo := positions[1]
	positionThree := positions[2]

	// Compute how much liquidity does balancer full range position contribute to incentives.s
	balancerLiquidityContributionToFullRange := s.getBalancerLiquidityContribution(clPoolId, balancerConfig.tokenZero.Amount, balancerConfig.tokenOne.Amount)

	// refetch CL pool
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)
	currentTickLiquidity := clPool.GetLiquidity()
	totalLiquidity := currentTickLiquidity.Add(balancerLiquidityContributionToFullRange)

	// 2 blocks have passed, so we expect 2 * emission rate
	expectedTotalAmountCollected := emissionRate.MulInt64(2)

	balancerShare := expectedTotalAmountCollected.Mul(balancerLiquidityContributionToFullRange).Quo(totalLiquidity).TruncateInt()
	ownerOneShare := expectedTotalAmountCollected.Mul(positionOne.liquidity).Quo(totalLiquidity).TruncateInt()
	ownerTwoShare := expectedTotalAmountCollected.Mul(positionTwo.liquidity).Quo(totalLiquidity).TruncateInt()

	// Validate that gauge is not getting updated unless a mutative action occurs within
	// a block that transfers the rewards to the gauge.
	balancerGaugeCoins := s.getBalancerGaugeCoins(balancerConfig.balancerPoolId)
	s.Require().Equal(sdk.Coins(nil), balancerGaugeCoins)

	// Collect incentives for all positions
	// Attempt to claim from wrong address and fail.
	_, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, ownerOne, positionTwo.positionId)
	s.Require().Error(err)

	// 1) claims correct amount
	s.collectAndValidateIncentives(ownerOne, positionOne.positionId, sdk.NewCoins(sdk.NewCoin(ETH, ownerOneShare)))

	// Refetch the gauge and check that balancer gauge has been updated
	balancerGaugeCoins = s.getBalancerGaugeCoins(balancerConfig.balancerPoolId)
	s.Require().Equal(sdk.NewCoins(sdk.NewCoin(ETH, balancerShare)), balancerGaugeCoins)

	// 2) claims correct amount
	s.collectAndValidateIncentives(ownerTwo, positionTwo.positionId, sdk.NewCoins(sdk.NewCoin(ETH, ownerTwoShare)))

	// 3) claims nothing
	s.collectAndValidateIncentives(ownerThree, positionThree.positionId, sdk.Coins(nil))
}

func (s *FunctionalIncentivesTestSuite) TestIncentives_Functional_TickCrossing() {
	s.Setup()
	s.TestAccs = apptesting.CreateRandomAccounts(6)

	s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

	var (
		defaultPositionCoins = sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1))
	)

	// BLOCK 1:
	// 1. Create balancer link
	// 2. Create 3 positions: 1) full-range 2) in-range, 3) completely out of range
	// 3. Create incentive record
	//
	// BLOCK 2:
	// 4. Increase block time and perform a swap
	//
	// BLOCK 3:
	// 5. Increase block time and collect incentives for 1) 2) 3)
	//   * 1) claims correct amount
	//   * 2) claims correct amount
	//   * 3) claims nothing

	// BLOCK 1:

	// Create pools
	clPool := s.PrepareConcentratedPool()
	clPoolId := clPool.GetId()

	ownerOne := s.TestAccs[0]
	ownerTwo := s.TestAccs[1]
	ownerThree := s.TestAccs[2]
	balancerOwner := s.TestAccs[3]

	balancerConfig := balancerLinkConfig{
		clPoolId:                 clPoolId,
		balancerOwner:            balancerOwner,
		initialBalancerLiquidity: sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)),

		// Choosing joinLPShares to be an initial pool share supply implies that we
		// LP with 50% of total pool shares. By locking up half of that, we end up
		// with 25% of total pool shares locked.
		joinLPShares:   gammtypes.InitPoolSharesSupply,
		bondedFraction: sdk.NewDecWithPrec(5, 1),
	}

	positions := positionsConfig{
		{

			isFullRange: true,
			owner:       ownerOne,
			poolId:      clPoolId,
			coins:       defaultPositionCoins,
		},
		{
			owner:                     ownerTwo,
			poolId:                    clPoolId,
			coins:                     defaultPositionCoins,
			lowerTickTickSpacingsAway: -1,
			upperTickTickSpacingsAway: 1,
		},
		{
			owner:                     ownerThree,
			poolId:                    clPoolId,
			coins:                     defaultPositionCoins,
			lowerTickTickSpacingsAway: 10000,
			upperTickTickSpacingsAway: 10001,
		},
	}

	remainingCoin := sdk.NewCoin(ETH, sdk.NewInt(1000000000000000000))
	emissionRate := sdk.NewDec(1000000) // 1 per second
	incentiveRecordCreator := s.TestAccs[4]
	defaultIncentiveRecordConfig := incentiveRecordConfig{
		remainingCoin: remainingCoin,
		emissionRate:  emissionRate,
		poolId:        clPoolId,
		creator:       incentiveRecordCreator,
		startTime:     defaultBlockTime,
		uptime:        time.Nanosecond,
	}
	incentiveRecords := incentiveRecordsConfig{defaultIncentiveRecordConfig}

	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)

	// Perform a swap in the opposite direction of position 3 so that we never activate it.
	// This is to test that we don't claim rewards for inactive positions.
	// Since USDC is token 1 and ETH is token 0, we swap ETH for USDC to move to the left.
	swapper := s.TestAccs[5]
	swapConfig := swapConfig{
		isZeroForOne: true,
		swapper:      swapper,
		poolId:       clPoolId,

		// Expecting to cross the tick
		amountInMultiplier: sdk.OneDec(),

		// Note, the bound tick is set in-between blocks:
		// by propagateExecutionCtxNextBlock()
	}

	swapConfigTwo := swapConfig

	blocks := []blockConfiguration{
		{
			blockOperations: []blockOperation{
				// 1. Create balancer <> CL link
				&balancerConfig,

				// 2. Create positions
				&positions,

				// 3. Create incentive records
				&incentiveRecords,
			},
			blockTime: time.Second,
			propagateExecutionCtxNextBlock: func() {
				// For the swap in the next block, we use the second position's lower tick.
				swapConfig.boundTick = positions[1].lowerTick

				swapConfigTwo.boundTick = positions[1].lowerTick - int64(clPool.GetTickSpacing())
			},
		},
		{
			blockOperations: []blockOperation{
				&swapConfig,
				&swapConfigTwo,
			},

			blockTime:                      time.Second,
			propagateExecutionCtxNextBlock: func() {},
		},
	}

	for _, block := range blocks {
		for _, blockOp := range block.blockOperations {
			blockOp.Apply(s)
		}

		// Block time is increased by the desired block time.
		s.CommitWithBlockTime(block.blockTime)

		// Propagate context from previous block execution to next
		block.propagateExecutionCtxNextBlock()
	}

	positionOne := positions[0]
	positionTwo := positions[1]
	positionThree := positions[2]

	// Compute how much liquidity does balancer full range position contribute to incentives.s
	balancerLiquidityContributionToFullRange := s.getBalancerLiquidityContribution(clPoolId, balancerConfig.tokenZero.Amount, balancerConfig.tokenOne.Amount)

	// refetch CL pool
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)
	currentTickLiquidity := clPool.GetLiquidity()
	totalLiquidity := currentTickLiquidity.Add(balancerLiquidityContributionToFullRange)

	// 2 blocks have passed, so we expect 2 * emission rate
	expectedTotalAmountCollected := emissionRate.MulInt64(2)

	balancerShare := expectedTotalAmountCollected.Mul(balancerLiquidityContributionToFullRange).Quo(totalLiquidity).TruncateInt()
	ownerOneShare := expectedTotalAmountCollected.Mul(positionOne.liquidity).Quo(totalLiquidity).TruncateInt()
	ownerTwoShare := expectedTotalAmountCollected.Mul(positionTwo.liquidity).Quo(totalLiquidity).TruncateInt()

	// Validate that gauge is not getting updated unless a mutative action occurs within
	// a block that transfers the rewards to the gauge.
	balancerGaugeCoins := s.getBalancerGaugeCoins(balancerConfig.balancerPoolId)
	s.Require().Equal(sdk.Coins(nil), balancerGaugeCoins)

	// Collect incentives for all positions
	// Attempt to claim from wrong address and fail.
	_, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, ownerOne, positionTwo.positionId)
	s.Require().Error(err)

	// 1) claims correct amount
	s.collectAndValidateIncentives(ownerOne, positionOne.positionId, sdk.NewCoins(sdk.NewCoin(ETH, ownerOneShare)))

	// Refetch the gauge and check that balancer gauge has been updated
	balancerGaugeCoins = s.getBalancerGaugeCoins(balancerConfig.balancerPoolId)
	s.Require().Equal(sdk.NewCoins(sdk.NewCoin(ETH, balancerShare)), balancerGaugeCoins)

	// 2) claims correct amount
	s.collectAndValidateIncentives(ownerTwo, positionTwo.positionId, sdk.NewCoins(sdk.NewCoin(ETH, ownerTwoShare)))

	// 3) claims nothing
	s.collectAndValidateIncentives(ownerThree, positionThree.positionId, sdk.Coins(nil))
}

// CONTRACTs:
// * CL pool has been created and config.clPoolId has been set.
// * Balancer pool and position owner is set at config.balancerOwner
func (s FunctionalIncentivesTestSuite) balancerPoolIdLinkSetup(config *balancerLinkConfig) {
	s.Require().NotZero(config.clPoolId)
	s.Require().NotNil(config.balancerOwner)
	s.Require().NotEmpty(config.initialBalancerLiquidity)

	// Create balancer pool.
	config.balancerPoolId = s.PrepareBalancerPoolWithCoins(config.initialBalancerLiquidity...)

	// Create balancer link
	s.App.GAMMKeeper.SetMigrationRecords(s.Ctx, gammtypes.MigrationRecords{
		BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
			{
				BalancerPoolId: config.balancerPoolId,
				ClPoolId:       config.clPoolId,
			},
		},
	})

	// Join balancer pool so that shares are bonded.
	balancerShares := gammtypes.InitPoolSharesSupply
	tokenInMaxs := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1))
	s.FundAcc(config.balancerOwner, tokenInMaxs)
	balancerTokens, bondedShareAmount, err := s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, config.balancerOwner, config.balancerPoolId, balancerShares, tokenInMaxs)
	s.Require().NoError(err)

	// Lock shares
	// Choosing balancerShares to be an initial pool share supply implies that we
	// LP with 50% of total pool shares. By locking up half of that, we end up
	// with 25% of total pool shares locked.
	fraction := sdk.NewDecWithPrec(5, 1)
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)
	lockAmt := bondedShareAmount.ToDec().Mul(fraction).TruncateInt()
	lockCoins := sdk.NewCoins(sdk.NewCoin(gammtypes.GetPoolShareDenom(config.balancerPoolId), lockAmt))
	_, err = s.App.LockupKeeper.CreateLock(s.Ctx, config.balancerOwner, lockCoins, longestLockableDuration)
	s.Require().NoError(err)

	for i, balancerToken := range balancerTokens {
		// 25%, see reasoning above balancer position creation
		balancerTokens[i].Amount = balancerToken.Amount.QuoRaw(2)
	}

	config.tokenZero = balancerTokens[0]
	config.tokenOne = balancerTokens[1]
}

// getBalancerGaugeCoins returns the coins in the longest lockable duration
// balancer gauge for the given balancer pool id.
func (s FunctionalIncentivesTestSuite) getBalancerGaugeCoins(balancerPoolId uint64) sdk.Coins {
	// Get longest lockable duration
	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)

	// Get balancer gauge id
	balancerGaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, balancerPoolId, longestLockableDuration)
	s.Require().NoError(err)

	// Validate that gauge is not getting updated unless a mutative action occurs within
	// a block that transfers the rewards to the gauge.
	balancerGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, balancerGaugeId)
	s.Require().NoError(err)

	return balancerGauge.Coins
}

func (s FunctionalIncentivesTestSuite) getLiquidityFromAmounts(clPoolId uint64, token0Amount sdk.Int, token1Amount sdk.Int) sdk.Dec {
	// Refetech pool
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)
	liquidity := math.GetLiquidityFromAmounts(clPool.GetCurrentSqrtPrice(), types.MinSqrtPrice, types.MaxSqrtPrice, token0Amount, token1Amount)
	return liquidity
}

func (s FunctionalIncentivesTestSuite) getBalancerLiquidityContribution(clPoolId uint64, token0Amount sdk.Int, token1Amount sdk.Int) sdk.Dec {
	balancerLiquidityContributionToFullRange := s.getLiquidityFromAmounts(clPoolId, token0Amount, token1Amount)
	params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	balancerRatio := sdk.OneDec().Sub(params.BalancerSharesRewardDiscount)
	balancerLiquidityContributionToFullRange = balancerLiquidityContributionToFullRange.Mul(balancerRatio)

	return balancerLiquidityContributionToFullRange
}

func (s FunctionalIncentivesTestSuite) collectAndValidateIncentives(owner sdk.AccAddress, positionId uint64, expectedCollectedAmount sdk.Coins) {
	collected, forfeited, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, owner, positionId)
	s.Require().NoError(err)
	s.Require().Equal(expectedCollectedAmount.String(), collected.String())
	s.Require().Equal(sdk.DecCoins{}, forfeited)
}

func (s FunctionalIncentivesTestSuite) computeAmountInBetweenTicksZFO(clPoolId uint64, lowerTick int64) sdk.Dec {
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)

	_, lowerTickSqrtPrice, err := math.TickToSqrtPrice(lowerTick)
	s.Require().NoError(err)

	amountInBetween := math.CalcAmount0Delta(clPool.GetLiquidity(), lowerTickSqrtPrice, clPool.GetCurrentSqrtPrice(), true)
	return amountInBetween
}

func (s FunctionalIncentivesTestSuite) computePositionTicksFromConfig(clPoolId uint64, lowerTickTickSpacingsAway int64, upperTickTickSpacingsAway int64) (int64, int64) {
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)
	currentTick := clPool.GetCurrentTick()
	lowerTick := currentTick + int64(clPool.GetTickSpacing()*uint64(lowerTickTickSpacingsAway))
	if lowerTick < types.MinTick {
		lowerTick = types.MinTick
	}
	upperTick := currentTick + int64(clPool.GetTickSpacing()*uint64(upperTickTickSpacingsAway))
	if upperTick > types.MaxTick {
		upperTick = types.MaxTick
	}
	return lowerTick, upperTick
}

func (s FunctionalIncentivesTestSuite) createIncentiveRecord(incentiveRecordCreator sdk.AccAddress, clPoolId uint64, remainingCoin sdk.Coin, emissionRate sdk.Dec, startTime time.Time, uptimeIncentivized time.Duration) uint64 {
	// Get incentive records before creating a new one
	incentiveRecords, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId)
	s.Require().NoError(err)
	incentiveRecordCountBefore := len(incentiveRecords)

	// Fund creator
	s.FundAcc(incentiveRecordCreator, sdk.NewCoins(remainingCoin))

	// Create incentive record
	incentiveRecord := types.IncentiveRecord{
		PoolId: clPoolId,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromCoin(remainingCoin),
			EmissionRate:  emissionRate, // 1 per second
			StartTime:     startTime,
		},
		MinUptime: uptimeIncentivized,
	}
	incentiveRecord, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, clPoolId, incentiveRecordCreator, remainingCoin, incentiveRecord.IncentiveRecordBody.EmissionRate, incentiveRecord.IncentiveRecordBody.StartTime, incentiveRecord.MinUptime)
	s.Require().NoError(err)

	// Check that incentive record was created
	incentiveRecords, err = s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId)
	s.Require().NoError(err)

	s.Require().Equal(incentiveRecordCountBefore+1, len(incentiveRecords))

	return incentiveRecord.IncentiveId
}

func (s FunctionalIncentivesTestSuite) createPositions(positionConfigs []*positionConfig) {
	// from pool id to number of positions expected
	poolIdsMap := make(map[uint64]int64, 0)
	// map of new position ids added
	newPositionIdMap := make(map[uint64]struct{}, 0)

	for i, pos := range positionConfigs {
		var (
			positionId uint64
			liquidity  sdk.Dec
			lowerTick  int64
			upperTick  int64
			err        error
		)

		// Fund the position owner with desired amounts
		acceptableRoundingAmount := sdk.OneInt()
		fundCoins := make(sdk.Coins, len(pos.coins))
		for i, coin := range pos.coins {
			fundCoins[i] = sdk.NewCoin(coin.Denom, coin.Amount.Add(acceptableRoundingAmount))
		}
		s.FundAcc(pos.owner, fundCoins)

		if pos.isFullRange {
			// Create full range position.
			positionId, _, _, liquidity, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pos.poolId, pos.owner, pos.coins)
			s.Require().NoError(err, "failed creating full range position with index (%d)", i)

			lowerTick, upperTick = types.MinTick, types.MaxTick
		} else {
			lowerTick, upperTick = s.computePositionTicksFromConfig(pos.poolId, pos.lowerTickTickSpacingsAway, pos.upperTickTickSpacingsAway)

			// Create position with range specified relative to current tick.
			positionId, _, _, liquidity, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pos.poolId, pos.owner, pos.coins, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
			s.Require().NoError(err, "failed creating narrow range position with index (%d)", i)
		}

		positionConfigs[i] = &positionConfig{
			positionId: positionId,
			liquidity:  liquidity,
		}

		positionConfigs[i].lowerTick = lowerTick
		positionConfigs[i].upperTick = upperTick

		poolIdsMap[pos.poolId]++
		newPositionIdMap[positionId] = struct{}{}
	}

	// Get all positions and validate that they were created in the desired ranges
	poolIds := make([]uint64, 0, len(poolIdsMap))
	for poolId := range poolIdsMap {
		poolIds = append(poolIds, poolId)
	}
	sort.Slice(poolIds, func(i, j int) bool {
		return poolIds[i] < poolIds[j]
	})

	// Validate that the correct number of positions has been created
	for _, poolId := range poolIds {
		positions, err := s.App.ConcentratedLiquidityKeeper.GetAllPositionIdsForPoolId(s.Ctx, types.PositionPrefix, poolId)
		s.Require().NoError(err)

		// filter only new positions being added
		newPositionsForPool := osmoutils.Filter(func(positionId uint64) bool {
			_, ok := newPositionIdMap[positionId]
			return ok
		}, positions)

		s.Require().Equal(poolIdsMap[poolId], int64(len(newPositionsForPool)))

		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		currentTickForPool := pool.GetCurrentTick()

		s.T().Log("poolId", poolId, "currentTick", currentTickForPool)

		// filter position configs belonging to this pool
		positionConfigsForThisPoolId := osmoutils.Filter(func(position *positionConfig) bool {
			return position.poolId == poolId
		}, positionConfigs)

		for _, pos := range newPositionsForPool {
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, pos)
			s.Require().NoError(err)

			// confirm a valid configuration exists for this position
			matchingPositions := osmoutils.Filter(func(config *positionConfig) bool {
				return config.lowerTick == position.LowerTick && config.upperTick == position.UpperTick && position.Liquidity.Equal(config.liquidity)
			}, positionConfigsForThisPoolId)
			s.Require().GreaterOrEqual(1, len(matchingPositions))

			s.T().Log("position created", "lower tick", position.LowerTick, "upper tick", position.UpperTick, "liquidity", position.Liquidity)
		}
	}
}

func (s FunctionalIncentivesTestSuite) createIncentiveRecords(incentiveRecords []incentiveRecordConfig) {
	for _, incentiveRecordConfig := range incentiveRecords {
		_ = s.createIncentiveRecord(incentiveRecordConfig.creator, incentiveRecordConfig.poolId, incentiveRecordConfig.remainingCoin, incentiveRecordConfig.emissionRate, incentiveRecordConfig.startTime, incentiveRecordConfig.uptime)
	}
}

func (s FunctionalIncentivesTestSuite) swap(config swapConfig) {
	// This value should be set by this function
	s.Require().True(config.tokenIn.IsNil())

	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, config.poolId)
	s.Require().NoError(err)

	tokenZero := clPool.GetToken0()
	tokenOne := clPool.GetToken1()
	var (
		tokenInDenom  string
		tokenOutDenom string
	)
	if config.isZeroForOne {
		tokenInDenom = tokenZero
		tokenOutDenom = tokenOne
	} else {
		tokenInDenom = tokenOne
		tokenOutDenom = tokenZero
	}

	// estimate amount to swap in based on configuration
	var swapTokenIn sdk.Coin
	if config.isZeroForOne {
		amountInUntilNextTick := s.computeAmountInBetweenTicksZFO(config.poolId, config.boundTick)
		amountInAfterMultiplier := amountInUntilNextTick.Mul(config.amountInMultiplier).TruncateInt()
		swapTokenIn = sdk.NewCoin(tokenInDenom, amountInAfterMultiplier)
	}
	config.tokenIn = swapTokenIn

	s.FundAcc(config.swapper, sdk.NewCoins(swapTokenIn))

	s.T().Log("swapper", config.swapper, "tokenIn", swapTokenIn, "tokenOut", tokenOutDenom, "poolId", config.poolId, "boundTick", config.boundTick, "isZeroForOne", config.isZeroForOne)

	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, config.swapper, clPool, swapTokenIn, tokenOutDenom, sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().NoError(err)
}

func (s FunctionalIncentivesTestSuite) getCLCurrentLiquidity(clPoolId uint64) sdk.Dec {
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)
	currentTickLiquidity := clPool.GetLiquidity()

	return currentTickLiquidity
}

func (c *balancerLinkConfig) Apply(s *FunctionalIncentivesTestSuite) {
	s.balancerPoolIdLinkSetup(c)
}

func (c *positionsConfig) Apply(s *FunctionalIncentivesTestSuite) {
	s.createPositions(*c)
}

func (c *incentiveRecordsConfig) Apply(s *FunctionalIncentivesTestSuite) {
	s.createIncentiveRecords(*c)
}

func (c *swapConfig) Apply(s *FunctionalIncentivesTestSuite) {
	s.swap(*c)
}
