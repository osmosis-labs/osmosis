package concentrated_liquidity_test

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type lpTest struct {
	poolId                            uint64
	currentTick                       int64
	lowerTick                         int64
	expectedLowerTick                 int64
	upperTick                         int64
	expectedUpperTick                 int64
	joinTime                          time.Time
	positionId                        uint64
	underlyingLockId                  uint64
	currentSqrtP                      osmomath.BigDec
	tokensProvided                    sdk.Coins
	customTokensProvided              bool
	amount0Minimum                    osmomath.Int
	amount0Expected                   osmomath.Int
	amount1Minimum                    osmomath.Int
	amount1Expected                   osmomath.Int
	liquidityAmount                   osmomath.Dec
	tickSpacing                       uint64
	isNotFirstPosition                bool
	isNotFirstPositionWithSameAccount bool
	expectedError                     error

	// spread reward related fields
	preSetChargeSpreadRewards              sdk.DecCoin
	expectedSpreadRewardGrowthOutsideLower sdk.DecCoins
	expectedSpreadRewardGrowthOutsideUpper sdk.DecCoins
}

var (
	baseCase = &lpTest{
		isNotFirstPosition:                false,
		isNotFirstPositionWithSameAccount: false,
		poolId:                            1,
		currentTick:                       DefaultCurrTick,
		lowerTick:                         DefaultLowerTick,
		upperTick:                         DefaultUpperTick,
		currentSqrtP:                      DefaultCurrSqrtPrice,
		tokensProvided:                    DefaultCoins,
		customTokensProvided:              false,
		amount0Minimum:                    osmomath.ZeroInt(),
		amount0Expected:                   DefaultAmt0Expected,
		amount1Minimum:                    osmomath.ZeroInt(),
		amount1Expected:                   DefaultAmt1Expected,
		liquidityAmount:                   DefaultLiquidityAmt,
		tickSpacing:                       DefaultTickSpacing,
		joinTime:                          DefaultJoinTime,
		positionId:                        1,
		underlyingLockId:                  0,

		preSetChargeSpreadRewards: oneEth,
		// in this setup lower tick < current tick < upper tick
		// the spread reward accumulator for ticks <= current tick are updated.
		expectedSpreadRewardGrowthOutsideLower: cl.EmptyCoins,
		// as a result, the upper tick is not updated.
		expectedSpreadRewardGrowthOutsideUpper: cl.EmptyCoins,
	}

	errToleranceOneRoundDown = osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.OneDec(),
		RoundingDir:       osmomath.RoundDown,
	}

	roundingError = osmomath.OneInt()

	positionCases = map[string]lpTest{
		"base case": {
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"create a position with non default tick spacing (10) with ticks that fall into tick spacing requirements": {
			tickSpacing:                            10,
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick < current tick -> both tick's spread reward accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: DefaultUpperTick + 100,

			preSetChargeSpreadRewards:              oneEth,
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick < current tick -> the spread reward is not charged so tick accumulators are unset": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: DefaultUpperTick + 100,

			preSetChargeSpreadRewards:              sdk.NewDecCoin(ETH, osmomath.ZeroInt()), // zero spread reward
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"current tick < lower tick < upper tick -> both tick's spread reward accumulators are unitilialized": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: DefaultLowerTick - 100,

			preSetChargeSpreadRewards:              oneEth,
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick == current tick -> both tick's spread reward accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: DefaultUpperTick,

			preSetChargeSpreadRewards:              oneEth,
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"second position: lower tick < upper tick == current tick -> both tick's spread reward accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: DefaultUpperTick,

			isNotFirstPositionWithSameAccount: true,
			positionId:                        2,

			liquidityAmount:                        baseCase.liquidityAmount.MulInt64(2),
			preSetChargeSpreadRewards:              oneEth,
			expectedSpreadRewardGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
	}
)

func (s *KeeperTestSuite) TestCreatePosition() {
	tests := map[string]lpTest{
		"error: non-existent pool": {
			poolId:        2,
			expectedError: types.PoolNotFoundError{PoolId: 2},
		},
		"error: lower tick out of bounds": {
			lowerTick:     DefaultMinTick - 100,
			expectedError: types.InvalidTickError{Tick: DefaultMinTick - 100, IsLower: true, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		"error: upper tick out of bounds": {
			upperTick:     DefaultMaxTick + 100,
			expectedError: types.InvalidTickError{Tick: DefaultMaxTick + 100, IsLower: false, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		"error: upper tick is below the lower tick, but both are in bounds": {
			lowerTick:     500,
			upperTick:     400,
			expectedError: types.InvalidLowerUpperTickError{LowerTick: 500, UpperTick: 400},
		},
		"error: amount0 min is negative": {
			amount0Minimum: osmomath.NewInt(-1),
			expectedError:  types.NotPositiveRequireAmountError{Amount: osmomath.NewInt(-1).String()},
		},
		"error: amount1 min is negative": {
			amount1Minimum: osmomath.NewInt(-1),
			expectedError:  types.NotPositiveRequireAmountError{Amount: osmomath.NewInt(-1).String()},
		},
		"error: amount of token 0 is smaller than minimum; should fail and not update state": {
			amount0Minimum: baseCase.amount0Expected.Mul(osmomath.NewInt(2)),
			// Add one since rounding up in favor of the pool.
			expectedError: types.InsufficientLiquidityCreatedError{Actual: baseCase.amount0Expected.Add(roundingError), Minimum: baseCase.amount0Expected.Mul(osmomath.NewInt(2)), IsTokenZero: true},
		},
		"error: amount of token 1 is smaller than minimum; should fail and not update state": {
			amount1Minimum: baseCase.amount1Expected.Mul(osmomath.NewInt(2)),

			expectedError: types.InsufficientLiquidityCreatedError{Actual: baseCase.amount1Expected, Minimum: baseCase.amount1Expected.Mul(osmomath.NewInt(2))},
		},
		"error: a non first position with zero amount desired for both denoms should fail liquidity delta check": {
			isNotFirstPosition:   true,
			customTokensProvided: true,
			tokensProvided:       sdk.Coins{},
			expectedError:        errors.New("cannot create a position with zero amounts of both pool tokens"),
		},
		"error: attempt to use and upper and lower tick that are not divisible by tick spacing": {
			lowerTick:     int64(305451),
			upperTick:     int64(315001),
			tickSpacing:   10,
			expectedError: types.TickSpacingError{TickSpacing: 10, LowerTick: int64(305451), UpperTick: int64(315001)},
		},
		"error: first position cannot have a zero amount for denom0": {
			customTokensProvided: true,
			tokensProvided:       sdk.NewCoins(DefaultCoin1),
			expectedError:        types.InitialLiquidityZeroError{Amount0: osmomath.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: first position cannot have a zero amount for denom1": {
			customTokensProvided: true,
			tokensProvided:       sdk.NewCoins(DefaultCoin0),
			expectedError:        types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: osmomath.ZeroInt()},
		},
		"error: first position cannot have a zero amount for both denom0 and denom1": {
			customTokensProvided: true,
			tokensProvided:       sdk.Coins{},
			expectedError:        errors.New("cannot create a position with zero amounts of both pool tokens"),
		},
		// TODO: add more tests
		// - custom hand-picked values
		// - think of overflows
		// - think of truncations
	}

	// add test cases for different positions
	for name, test := range positionCases {
		tests[name] = test
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// Merge tc with baseCase and update tc to the merged result. This is done to reduce the amount of boilerplate in test cases.
			baseConfigCopy := *baseCase
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			// Fund account to pay for the pool creation spread reward.
			s.FundAcc(s.TestAccs[0], PoolCreationFee)

			// Create a CL pool with custom tickSpacing
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, osmomath.ZeroDec()))
			s.Require().NoError(err)

			// Set mock listener to make sure that is is called when desired.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			pool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
			s.Require().NoError(err)

			// Pre-set spread reward growth accumulator
			if !tc.preSetChargeSpreadRewards.IsZero() {
				s.AddToSpreadRewardAccumulator(poolID, tc.preSetChargeSpreadRewards)
			}

			expectedNumCreatePositionEvents := 1

			// If we want to test a non-first position, we create a first position with a separate account
			if tc.isNotFirstPosition {
				s.SetupPosition(1, s.TestAccs[1], DefaultCoins, tc.lowerTick, tc.upperTick, false)
				expectedNumCreatePositionEvents += 1
			}

			expectedLiquidityCreated := tc.liquidityAmount
			if tc.isNotFirstPositionWithSameAccount {
				// Since this is a second position with the same parameters,
				// we expect to create half of the final liquidity amount.
				expectedLiquidityCreated = tc.liquidityAmount.QuoInt64(2)

				s.SetupPosition(1, s.TestAccs[0], DefaultCoins, tc.lowerTick, tc.upperTick, false)
				expectedNumCreatePositionEvents += 1
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], DefaultCoins)

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())

			// System under test.
			positionData, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.tokensProvided, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)

			var (
				positionId       = positionData.ID
				newLowerTick     = positionData.LowerTick
				newUpperTick     = positionData.UpperTick
				liquidityCreated = positionData.Liquidity
				asset0           = positionData.Amount0
				asset1           = positionData.Amount1
			)

			// Note user and pool account balances to compare after create position is called
			userBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())

			// If we expect an error, make sure:
			// - some error was emitted
			// - asset0 and asset1 that were calculated from create position is nil
			// - the error emitted was the expected error
			// - account balances are untouched
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(asset0, osmomath.Int{})
				s.Require().Equal(asset1, osmomath.Int{})
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Check account balances
				s.Require().Equal(userBalancePrePositionCreation.String(), userBalancePostPositionCreation.String())
				s.Require().Equal(poolBalancePrePositionCreation.String(), poolBalancePostPositionCreation.String())

				// Redundantly ensure that liquidity was not created
				liquidity, err := clKeeper.GetPositionLiquidity(s.Ctx, positionId)
				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: positionId})
				s.Require().Equal(osmomath.Dec{}, liquidity)
				return
			}

			// Else, check that we had no error from creating the position, and that the liquidity and assets that were returned are expected
			s.Require().NoError(err)
			s.Require().Equal(tc.positionId, positionId)
			s.Require().Equal(expectedLiquidityCreated.String(), liquidityCreated.String())
			s.Require().Equal(tc.amount0Expected.String(), asset0.String())
			s.Require().Equal(tc.amount1Expected.String(), asset1.String())
			if tc.expectedLowerTick != 0 {
				s.Require().Equal(tc.expectedLowerTick, newLowerTick)
				tc.lowerTick = newLowerTick
			}
			if tc.expectedUpperTick != 0 {
				s.Require().Equal(tc.expectedUpperTick, newUpperTick)
				tc.upperTick = newUpperTick
			}

			// Check account balances
			s.Require().Equal(userBalancePrePositionCreation.Sub(sdk.NewCoins(sdk.NewCoin(ETH, asset0), (sdk.NewCoin(USDC, asset1)))...).String(), userBalancePostPositionCreation.String())
			s.Require().Equal(poolBalancePrePositionCreation.Add(sdk.NewCoin(ETH, asset0), (sdk.NewCoin(USDC, asset1))).String(), poolBalancePostPositionCreation.String())

			hasPosition := clKeeper.HasPosition(s.Ctx, tc.positionId)
			s.Require().True(hasPosition)

			// Check position state
			s.validatePositionUpdate(s.Ctx, positionId, expectedLiquidityCreated)

			s.validatePositionSpreadRewardAccUpdate(s.Ctx, tc.poolId, positionId, expectedLiquidityCreated)

			// Upscale accumulator values
			tc.expectedSpreadRewardGrowthOutsideLower = tc.expectedSpreadRewardGrowthOutsideLower.MulDecTruncate(cl.PerUnitLiqScalingFactor)
			tc.expectedSpreadRewardGrowthOutsideUpper = tc.expectedSpreadRewardGrowthOutsideUpper.MulDecTruncate(cl.PerUnitLiqScalingFactor)

			// Check tick state
			s.validateTickUpdates(tc.poolId, tc.lowerTick, tc.upperTick, tc.liquidityAmount, tc.expectedSpreadRewardGrowthOutsideLower, tc.expectedSpreadRewardGrowthOutsideUpper)

			// Validate events emitted.
			s.AssertEventEmitted(s.Ctx, types.TypeEvtCreatePosition, expectedNumCreatePositionEvents)

			// Validate that listeners were called the desired number of times
			expectedAfterInitialPoolPositionCreatedCallCount := 0
			if !tc.isNotFirstPosition {
				// We want the hook to be called only for the very first position in the pool.
				// Such position initializes current sqrt price and tick. As a result,
				// we want the hook to run for the purposes of creating twap records.
				// On any subsequent update, adding liquidity does not change the price.
				// Therefore, we do not have to call this hook.
				expectedAfterInitialPoolPositionCreatedCallCount = 1
			}
			s.validateListenerCallCount(0, expectedAfterInitialPoolPositionCreatedCallCount, 0, 0)
		})
	}
}

type lockState int

const (
	nolock lockState = iota
	locked
	unlocking
	unlocked
)

func (s *KeeperTestSuite) createPositionWithLockState(ls lockState, poolId uint64, owner sdk.AccAddress, providedCoins sdk.Coins, dur time.Duration) (uint64, osmomath.Dec) {
	var (
		positionData          cl.CreatePositionData
		fullRangePositionData cltypes.CreateFullRangePositionData
		err                   error
	)

	if ls == locked {
		fullRangePositionData, _, err = s.Clk.CreateFullRangePositionLocked(s.Ctx, poolId, owner, providedCoins, dur)
	} else if ls == unlocking {
		fullRangePositionData, _, err = s.Clk.CreateFullRangePositionUnlocking(s.Ctx, poolId, owner, providedCoins, dur+time.Hour)
	} else if ls == unlocked {
		fullRangePositionData, _, err = s.Clk.CreateFullRangePositionUnlocking(s.Ctx, poolId, owner, providedCoins, dur-time.Hour)
	} else {
		positionData, err = s.Clk.CreatePosition(s.Ctx, poolId, owner, providedCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)
		return positionData.ID, positionData.Liquidity
	}
	// full range case
	s.Require().NoError(err)
	return fullRangePositionData.ID, fullRangePositionData.Liquidity
}

func (s *KeeperTestSuite) TestWithdrawPosition() {
	defaultTimeElapsed := time.Hour * 24
	uptimeHelper := getExpectedUptimes()
	defaultUptimeGrowth := uptimeHelper.hundredTokensMultiDenom
	DefaultJoinTime := s.Ctx.BlockTime()
	nonOwner := s.TestAccs[1]

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		createPositionOverwrite bool
		timeElapsed             time.Duration
		createLockState         lockState
		withdrawWithNonOwner    bool
		isFullLiquidityWithdraw bool
	}{
		"base case: withdraw full liquidity amount": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998976 eth
				// Note: subtracting one due to truncations in favor of the pool when withdrawing.
				amount1Expected: baseCase.amount1Expected.Sub(osmomath.OneInt()), // 5000 usdc
			},
			timeElapsed:             defaultTimeElapsed,
			isFullLiquidityWithdraw: true,
		},
		"withdraw full liquidity amount with underlying lock that has finished unlocking": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				// Note: subtracting one due to truncations in favor of the pool when withdrawing.
				// amount0Expected = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
				// Where:
				// * liquidity = FullRangeLiquidityAmt
				// * sqrtPriceB = MaxSqrtPrice
				// * sqrtPriceA = DefaultCurrSqrtPrice
				// Exact calculation: https://www.wolframalpha.com/input?i=70710678.118654752940000000+*+%2810000000000000000000.000000000000000000+-+70.710678118654752440%29+%2F+%2810000000000000000000.000000000000000000+*+70.710678118654752440%29
				amount0Expected: osmomath.NewInt(999999),
				// amount1Expected = liq * (sqrtPriceB - sqrtPriceA)
				// Where:
				// * liquidity = FullRangeLiquidityAmt
				// * sqrtPriceB = DefaultCurrSqrtPrice
				// * sqrtPriceA = MinSqrtPrice
				// Exact calculation: https://www.wolframalpha.com/input?i=70710678.118654752940000000+*+%2870.710678118654752440+-+0.000001000000000000%29
				amount1Expected:  osmomath.NewInt(4999999929),
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
			},
			createLockState:         unlocked,
			timeElapsed:             defaultTimeElapsed,
			isFullLiquidityWithdraw: true,
		},
		"error: withdraw full liquidity amount but still locked": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
				expectedError:    types.LockNotMatureError{PositionId: 1, LockId: 1},
			},
			createLockState: locked,
			timeElapsed:     defaultTimeElapsed,
		},
		"error: withdraw full liquidity amount but still unlocking": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
				expectedError:    types.LockNotMatureError{PositionId: 1, LockId: 1},
			},
			createLockState: unlocking,
			timeElapsed:     defaultTimeElapsed,
		},
		"withdraw partial liquidity amount": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.QuoRoundUp(osmomath.NewDec(2)),
				amount0Expected: baseCase.amount0Expected.QuoRaw(2), // 0.499488
				amount1Expected: baseCase.amount1Expected.QuoRaw(2), // 2500 usdc
			},
			timeElapsed: defaultTimeElapsed,
		},
		"withdraw full liquidity amount, forfeit incentives": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998976 eth
				// Note: subtracting one due to truncations in favor of the pool when withdrawing.
				amount1Expected: baseCase.amount1Expected.Sub(osmomath.OneInt()), // 5000 usdc
			},
			timeElapsed:             0,
			isFullLiquidityWithdraw: true,
		},
		"error: no position created": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				lowerTick:     -1, // valid tick at which no position exists
				positionId:    DefaultPositionId + 1,
				expectedError: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
			},
			timeElapsed: defaultTimeElapsed,
		},
		"error: insufficient liquidity": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.Add(osmomath.OneDec()), // 1 more than available
				expectedError:   types.InsufficientLiquidityError{Actual: baseCase.liquidityAmount.Add(osmomath.OneDec()), Available: baseCase.liquidityAmount},
			},
			timeElapsed: defaultTimeElapsed,
		},
		"error: try withdrawing negative liquidity": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.Sub(baseCase.liquidityAmount.Mul(osmomath.NewDec(2))),
				expectedError:   types.InsufficientLiquidityError{Actual: baseCase.liquidityAmount.Sub(baseCase.liquidityAmount.Mul(osmomath.NewDec(2))), Available: baseCase.liquidityAmount},
			},
			timeElapsed: defaultTimeElapsed,
		},
		"error: attempt to withdraw a position that does not belong to the caller": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				expectedError: types.NotPositionOwnerError{PositionId: 1, Address: nonOwner.String()},
			},
			timeElapsed:          defaultTimeElapsed,
			withdrawWithNonOwner: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			var (
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				liquidityCreated            = osmomath.ZeroDec()
				owner                       = s.TestAccs[0]
				config                      = *tc.setupConfig
				err                         error
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, tc.sutConfigOverwrite)

			// If a setupConfig is provided, use it to create a pool and position.
			pool := s.PrepareConcentratedPool()
			fundCoins := config.tokensProvided
			s.FundAcc(owner, fundCoins)

			_, liquidityCreated = s.createPositionWithLockState(tc.createLockState, pool.GetId(), owner, fundCoins, tc.timeElapsed)

			// Set mock listener to make sure that is is called when desired.
			// It must be set after test position creation so that we do not record the call
			// for initial position update.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))
			store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))

			// Set global spread reward growth to 1 ETH and charge the spread reward to the pool.
			globalSpreadRewardGrowth := sdk.NewDecCoin(ETH, osmomath.NewInt(1))
			s.AddToSpreadRewardAccumulator(pool.GetId(), globalSpreadRewardGrowth)

			// Add global uptime growth
			err = addToUptimeAccums(s.Ctx, pool.GetId(), concentratedLiquidityKeeper, defaultUptimeGrowth)
			s.Require().NoError(err)

			// Determine the liquidity expected to remain after the withdraw.
			expectedRemainingLiquidity := liquidityCreated.Sub(config.liquidityAmount)

			expectedSpreadRewardsClaimed := sdk.NewCoins()
			expectedIncentivesClaimed := sdk.NewCoins()

			// Set the expected spread rewards claimed to the amount of liquidity created since the global spread reward growth is 1.
			// Fund the pool account with the expected spread rewards claimed.
			if expectedRemainingLiquidity.IsZero() {
				expectedSpreadRewardsClaimed = expectedSpreadRewardsClaimed.Add(sdk.NewCoin(ETH, liquidityCreated.TruncateInt()))
				s.FundAcc(pool.GetSpreadRewardsAddress(), expectedSpreadRewardsClaimed)
			}

			// Set expected incentives and fund pool with appropriate amount
			expectedIncentivesClaimed = expectedIncentivesFromUptimeGrowth(defaultUptimeGrowth, liquidityCreated, tc.timeElapsed, defaultMultiplier)

			// Fund full amount since forfeited incentives for the last position are sent to the community pool.
			largestSupportedUptime := s.Clk.GetLargestSupportedUptimeDuration(s.Ctx)
			expectedFullIncentivesFromAllUptimes := expectedIncentivesFromUptimeGrowth(defaultUptimeGrowth, liquidityCreated, largestSupportedUptime, defaultMultiplier)
			s.FundAcc(pool.GetIncentivesAddress(), expectedFullIncentivesFromAllUptimes)

			// Note the pool and owner balances before withdrawal of the position.
			poolBalanceBeforeWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			poolSpreadRewardBalanceBeforeWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetSpreadRewardsAddress())
			incentivesBalanceBeforeWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetIncentivesAddress())
			ownerBalancerBeforeWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)

			expectedPoolBalanceDelta := sdk.NewCoins(sdk.NewCoin(ETH, config.amount0Expected.Abs()), sdk.NewCoin(USDC, config.amount1Expected.Abs()))

			var withdrawAccount sdk.AccAddress
			if tc.withdrawWithNonOwner {
				withdrawAccount = nonOwner
			} else {
				withdrawAccount = owner
			}

			// System under test.
			amtDenom0, amtDenom1, err := concentratedLiquidityKeeper.WithdrawPosition(s.Ctx, withdrawAccount, config.positionId, config.liquidityAmount)
			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(amtDenom0, osmomath.Int{})
				s.Require().Equal(amtDenom1, osmomath.Int{})
				s.Require().ErrorContains(err, config.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), amtDenom0.String())
			s.Require().Equal(config.amount1Expected.String(), amtDenom1.String())

			// If the remaining liquidity is zero, all spread rewards and incentives should be collected and the position should be deleted.
			// Check if all spread rewards and incentives were collected.
			poolBalanceAfterWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			poolSpreadRewardBalanceAfterWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetSpreadRewardsAddress())
			incentivesBalanceAfterWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetIncentivesAddress())
			ownerBalancerAfterWithdraw := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)

			// If the position was the last in the pool, we expect it to receive full incentives since there
			// is nobody to forfeit to
			updatedPool, err := concentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)
			if updatedPool.GetLiquidity().LTE(osmomath.OneDec()) {
				expectedIncentivesClaimed = expectedFullIncentivesFromAllUptimes
			}

			// owner should only have tokens equivalent to the delta balance of the pool
			expectedOwnerBalanceDelta := expectedPoolBalanceDelta.Add(expectedIncentivesClaimed...).Add(expectedSpreadRewardsClaimed...)
			actualOwnerBalancerDelta := ownerBalancerAfterWithdraw.Sub(ownerBalancerBeforeWithdraw...)
			actualIncentivesClaimed := incentivesBalanceBeforeWithdraw.Sub(incentivesBalanceAfterWithdraw...)

			s.Require().Equal(expectedPoolBalanceDelta.String(), poolBalanceBeforeWithdraw.Sub(poolBalanceAfterWithdraw...).String())
			s.Require().NotEmpty(expectedOwnerBalanceDelta)
			for _, coin := range expectedOwnerBalanceDelta {
				expected := expectedOwnerBalanceDelta.AmountOf(coin.Denom)
				actual := actualOwnerBalancerDelta.AmountOf(coin.Denom)
				s.Require().True(expected.Equal(actual))
			}

			if tc.timeElapsed > 0 {
				s.Require().NotEmpty(expectedIncentivesClaimed)
			}
			for _, coin := range expectedIncentivesClaimed {
				expected := expectedIncentivesClaimed.AmountOf(coin.Denom)
				actual := actualIncentivesClaimed.AmountOf(coin.Denom)
				s.Require().True(expected.Equal(actual))
			}

			s.Require().Equal(poolSpreadRewardBalanceBeforeWithdraw.Sub(poolSpreadRewardBalanceAfterWithdraw...).String(), expectedSpreadRewardsClaimed.String())

			// if the position's expected remaining liquidity is equal to zero, we check if all state
			// have been correctly deleted.
			if expectedRemainingLiquidity.IsZero() {
				// Check that the position was deleted.
				position, err := concentratedLiquidityKeeper.GetPosition(s.Ctx, config.positionId)
				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: config.positionId})
				s.Require().Equal(clmodel.Position{}, position)

				// Since the positionLiquidity is deleted, retrieving it should return an error.
				positionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, config.positionId)
				s.Require().Error(err)
				s.Require().ErrorIs(err, types.PositionIdNotFoundError{PositionId: config.positionId})
				s.Require().Equal(osmomath.Dec{}, positionLiquidity)

				// check underlying stores were correctly deleted
				emptyPositionStruct := clmodel.Position{}
				positionIdToPositionKey := types.KeyPositionId(config.positionId)
				osmoutils.Get(store, positionIdToPositionKey, &position)
				s.Require().Equal(model.Position{}, emptyPositionStruct)

				// Retrieve the position ID from the store via owner/poolId key and compare to expected values.
				ownerPoolIdToPositionIdKey := types.KeyAddressPoolIdPositionId(s.TestAccs[0], defaultPoolId, DefaultPositionId)
				positionIdBytes := store.Get(ownerPoolIdToPositionIdKey)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position ID from the store via poolId key and compare to expected values.
				poolIdtoPositionIdKey := types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
				positionIdBytes = store.Get(poolIdtoPositionIdKey)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
				positionIdToLockIdKey := types.KeyPositionIdForLock(DefaultPositionId)
				underlyingLockIdBytes := store.Get(positionIdToLockIdKey)
				s.Require().Nil(underlyingLockIdBytes)

				// Retrieve the lock ID to position ID mapping from the store and compare to expected values.
				lockIdToPositionIdKey := types.KeyLockIdForPositionId(config.underlyingLockId)
				positionIdBytes = store.Get(lockIdToPositionIdKey)
				s.Require().Nil(positionIdBytes)

				// ensure that the lock is still there if there was lock that was existing before the withdraw process
				if tc.createLockState != nolock {
					_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, 1)
					s.Require().NoError(err)
				}
			} else {
				// Check that the position was updated.
				s.validatePositionUpdate(s.Ctx, config.positionId, expectedRemainingLiquidity)
			}

			// Check that ticks were removed if liquidity is fully withdrawn.
			lowerTickValue := store.Get(types.KeyTick(defaultPoolId, config.lowerTick))
			upperTickValue := store.Get(types.KeyTick(defaultPoolId, config.upperTick))
			if tc.isFullLiquidityWithdraw {
				s.Require().Nil(lowerTickValue)
				s.Require().Nil(upperTickValue)
			} else {
				s.Require().NotNil(lowerTickValue)
				s.Require().NotNil(upperTickValue)
			}

			// Check tick state.
			s.validateTickUpdates(config.poolId, config.lowerTick, config.upperTick, expectedRemainingLiquidity, config.expectedSpreadRewardGrowthOutsideLower, config.expectedSpreadRewardGrowthOutsideUpper)

			// Validate event emitted.
			s.AssertEventEmitted(s.Ctx, types.TypeEvtWithdrawPosition, 1)

			// Validate that listeners were called the desired number of times
			expectedAfterLastPoolPositionRemovedCallCount := 0
			if expectedRemainingLiquidity.IsZero() {
				// We want the hook to be called only when the last position (liquidity) is removed.
				// Not having any liquidity in the pool implies not having a valid sqrt price and tick. As a result,
				// we want the hook to run for the purposes of updating twap records.
				// Upon reading liquidity (recreating positions) to such pool, AfterInitialPoolPositionCreatedCallCount
				// will be called. Hence, updating twap with valid latest spot price.
				expectedAfterLastPoolPositionRemovedCallCount = 1
			}
			s.validateListenerCallCount(0, 0, expectedAfterLastPoolPositionRemovedCallCount, 0)

			// Dumb sanity-check that creating a position with the same liquidity amount after fully removing it does not error.
			// This is to be more thoroughly tested separately.
			if expectedRemainingLiquidity.IsZero() {
				// Add one USDC because we withdraw one less than originally funded due to truncation in favor of the pool.
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(USDC, osmomath.OneInt())))
				_, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), owner, config.tokensProvided, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddToPosition() {
	defaultTimeElapsed := time.Hour * 24
	invalidSender := s.TestAccs[2]

	// These amounts are set based on the actual amounts passed in as inputs
	// to create position in the default config case (prior to rounding). We use them as
	// a reference to test rounding behavior when adding to positions.
	amount0PerfectRatio := osmomath.NewInt(998977)
	amount1PerfectRatio := osmomath.NewInt(5000000000)

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		timeElapsed             time.Duration
		createPositionOverwrite bool
		createLockState         lockState
		lastPositionInPool      bool
		senderNotOwner          bool

		amount0ToAdd osmomath.Int
		amount1ToAdd osmomath.Int
	}{
		"add base amount to existing liquidity with perfect ratio": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: DefaultAmt0Expected.Add(amount0PerfectRatio),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: osmomath.NewInt(9999998816),
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,
		},
		"add base amount to existing liquidity with perfect ratio (rounding error added back in)": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio),
				// Subtract rounding error due to truncation after perfect join (asset0's truncation
				// leaves it on the amount above since we added the error upfront)
				amount1Expected: amount1PerfectRatio.Add(amount1PerfectRatio).Sub(roundingError),
			},
			timeElapsed: defaultTimeElapsed,
			// We add back in the rounding error for this test case to demonstrate that rounding pushes us off the boundary
			// in the previous test case
			amount0ToAdd: amount0PerfectRatio.Add(roundingError),
			amount1ToAdd: amount1PerfectRatio,
		},
		"add partial liquidity amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: DefaultAmt0Expected.Add(amount0PerfectRatio.QuoRaw(2)),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: osmomath.NewInt(7499995358),
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio.QuoRaw(2),
			amount1ToAdd: amount1PerfectRatio.QuoRaw(2),
		},
		"add partial liquidity amount (with rounding error added back in)": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio.QuoRaw(2)),
				// We add back in the rounding error for this test case to demonstrate that rounding pushes us off the boundary
				// in the previous test case
				amount1Expected: amount1PerfectRatio.Add(amount1PerfectRatio.QuoRaw(2)).Sub(roundingError),
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio.QuoRaw(2).Add(roundingError),
			amount1ToAdd: amount1PerfectRatio.QuoRaw(2),
		},
		"Add to a position with underlying lock that has finished unlocking": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				// 1998976eth (amount withdrawn with rounded down amounts) + 998977(token amount in)
				amount0Expected: osmomath.NewInt(2997953),
				// tokens Provided for token1 is 9999999999 (amount withdrawn) + 5000000000 = 14999999999usdc.
				// We calculate calc amount1 by using the following equation:
				// liq * (sqrtPriceB - sqrtPriceA), where liq is equal to the original joined liq + added liq, sqrtPriceB is current sqrt price, and sqrtPriceA is min sqrt price.
				// Note that these numbers were calculated using `GetLiquidityFromAmounts` and `TickToSqrtPrice` and thus assume correctness of those functions.
				// https://www.wolframalpha.com/input?i=212041526.154556192317664016+*+%2870.728769315114743566+-+0.000001000000000000%29
				amount1Expected: osmomath.NewInt(14997435977),
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockState: unlocked,
		},
		"error: attempt to add to a position with underlying lock that is still locked": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio).Sub(roundingError),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: osmomath.NewInt(9999998816),

				expectedError: types.PositionSuperfluidStakedError{PositionId: uint64(1)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockState: locked,
		},
		"error: attempt to add negative amounts for both assets to position": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.NegativeAmountAddedError{PositionId: 1, Asset0Amount: amount0PerfectRatio.Neg(), Asset1Amount: amount1PerfectRatio.Neg()},
			},
			lastPositionInPool: true,
			timeElapsed:        defaultTimeElapsed,
			amount0ToAdd:       amount0PerfectRatio.Neg(),
			amount1ToAdd:       amount1PerfectRatio.Neg(),
		},
		"error: attempt to add negative amounts for amount0": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.NegativeAmountAddedError{PositionId: 1, Asset0Amount: amount0PerfectRatio.Neg(), Asset1Amount: amount1PerfectRatio},
			},
			lastPositionInPool: true,
			timeElapsed:        defaultTimeElapsed,
			amount0ToAdd:       amount0PerfectRatio.Neg(),
			amount1ToAdd:       amount1PerfectRatio,
		},
		"error: attempt to add negative amounts for amount1": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.NegativeAmountAddedError{PositionId: 1, Asset0Amount: amount0PerfectRatio, Asset1Amount: amount1PerfectRatio.Neg()},
			},
			lastPositionInPool: true,
			timeElapsed:        defaultTimeElapsed,
			amount0ToAdd:       amount0PerfectRatio,
			amount1ToAdd:       amount1PerfectRatio.Neg(),
		},
		"error: both amounts are zero": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.ErrZeroLiquidity,
			},
			lastPositionInPool: true,
			timeElapsed:        defaultTimeElapsed,
			amount0ToAdd:       osmomath.ZeroInt(),
			amount1ToAdd:       osmomath.ZeroInt(),
		},
		"error: attempt to add to a position with underlying lock that is unlocking": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio).Sub(roundingError),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: osmomath.NewInt(9999998816),

				expectedError: types.PositionSuperfluidStakedError{PositionId: uint64(1)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockState: unlocking,
		},
		"error: no position created": {
			// setup parameters for creation a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				lowerTick:     -1, // valid tick at which no position exists
				positionId:    DefaultPositionId + 3,
				expectedError: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 3},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,
		},
		"error: attempt to add to last position in pool": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
			},
			lastPositionInPool: true,
			timeElapsed:        defaultTimeElapsed,
			amount0ToAdd:       amount0PerfectRatio,
			amount1ToAdd:       amount1PerfectRatio,
		},
		"error: not position owner": {
			// setup parameters for creating a pool and position.
			setupConfig:    baseCase,
			senderNotOwner: true,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				expectedError: types.NotPositionOwnerError{PositionId: 1, Address: invalidSender.String()},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,
		},
		"error: minimum amount 0 is less than actual amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Minimum: osmomath.NewInt(1000000),
				expectedError: types.InsufficientLiquidityCreatedError{
					Actual: osmomath.NewInt(1997954).Sub(roundingError),
					//  minimum amount we have input becomes default amt 0 expected (from original position withdraw) + 1000000 (input)
					Minimum:     DefaultAmt0Expected.Add(osmomath.NewInt(1000000)),
					IsTokenZero: true,
				},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,
		},
		"error: minimum amount 1 is less than actual amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount1Minimum: osmomath.NewInt(10000000000),
				expectedError: types.InsufficientLiquidityCreatedError{
					Actual: osmomath.NewInt(9999998816),
					// minimum amount we have input becomes default amt 1 expected (from original position withdraw) + 10000000000 (input) - 1 (rounding)
					Minimum:     DefaultAmt1Expected.Add(osmomath.NewInt(10000000000)).Sub(osmomath.OneInt()),
					IsTokenZero: false,
				},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// --- Setup ---
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			var (
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				owner                       = s.TestAccs[0]
				config                      = *tc.setupConfig
				err                         error
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, tc.sutConfigOverwrite)

			// If a setupConfig is provided, use it to create a pool and position.
			pool := s.PrepareConcentratedPool()
			fundCoins, lockCoins := config.tokensProvided, config.tokensProvided
			// Fund tokens that is used to create initial position
			if tc.amount0ToAdd.IsPositive() && tc.amount1ToAdd.IsPositive() {
				fundCoins = fundCoins.Add(sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd), sdk.NewCoin(USDC, tc.amount1ToAdd))...)
				if tc.createLockState != nolock {
					lockCoins = fundCoins
				}
			}
			s.FundAcc(owner, fundCoins)

			// Create a position from the parameters in the test case.
			positionId, _ := s.createPositionWithLockState(tc.createLockState, pool.GetId(), owner, lockCoins, tc.timeElapsed)

			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))
			preBalanceToken0 := s.App.BankKeeper.GetBalance(s.Ctx, owner, pool.GetToken0())

			if !tc.lastPositionInPool {
				s.FundAcc(s.TestAccs[1], fundCoins)
				_, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[1], config.tokensProvided, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			sender := owner
			if tc.senderNotOwner {
				sender = invalidSender
			}

			// now we fund the sender account again for the amount0ToAdd and amount1ToAdd coins.
			// only fund coins if the amount is non-negative or else test would panic here
			if !tc.amount0ToAdd.IsNegative() {
				s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd)))
			}
			if !tc.amount1ToAdd.IsNegative() {
				s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin(USDC, tc.amount1ToAdd)))
			}

			// --- System under test ---
			newPosId, newAmt0, newAmt1, err := concentratedLiquidityKeeper.AddToPosition(s.Ctx, sender, config.positionId, tc.amount0ToAdd, tc.amount1ToAdd, config.amount0Minimum, config.amount1Minimum)
			// config.amount0Minimum
			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(osmomath.Int{}, newAmt0)
				s.Require().Equal(osmomath.Int{}, newAmt1)
				s.Require().Equal(uint64(0), newPosId)
				s.Require().ErrorContains(err, config.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), newAmt0.String())
			s.Require().Equal(config.amount1Expected.String(), newAmt1.String())

			// We expect the position ID to be 3 since we have two setup positions
			s.Require().Equal(uint64(3), newPosId)

			expectedAmount1Delta := osmomath.ZeroInt()

			// delta amount1 only exists if the actual amount from addToPosition is not equivalent to tokens provided.
			// delta amount1 is calculated via (amount1 to create initial position) + (amount1 added to position) - (actual amount 1)
			if fundCoins.AmountOf(pool.GetToken1()).Add(tc.amount1ToAdd).Sub(newAmt1).GT(osmomath.ZeroInt()) {
				expectedAmount1Delta = config.tokensProvided.AmountOf(pool.GetToken1()).Add(tc.amount1ToAdd).Sub(newAmt1)
			}

			postBalanceToken0 := s.App.BankKeeper.GetBalance(s.Ctx, sender, pool.GetToken0())
			postBalanceToken1 := s.App.BankKeeper.GetBalance(s.Ctx, sender, pool.GetToken1())

			osmoassert.Equal(s.T(), errToleranceOneRoundDown, preBalanceToken0.Amount, postBalanceToken0.Amount)
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, expectedAmount1Delta, postBalanceToken1.Amount.Sub(tc.amount1ToAdd))

			// now check that old position id has been successfully deleted
			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
			s.Require().Error(err)
		})
	}
}

func (s *KeeperTestSuite) TestSingleSidedAddToPosition() {
	defaultTimeElapsed := time.Hour * 24

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		timeElapsed             time.Duration
		createPositionOverwrite bool
		createLockState         lockState
		lastPositionInPool      bool
		senderNotOwner          bool

		amount0ToAdd osmomath.Int
		amount1ToAdd osmomath.Int
	}{
		"single sided amount0 add": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				// calculated with x/concentrated-liquidity/python/clmath.py
				// The input values are taken from debugger assuming the rest of the system is correct:
				// sqrtPriceLowerTick = Decimal("1.000049998750062497000000000000000000")
				// sqrtPriceUpperTick = Decimal("1.000099995000499938000000000000000000")
				// liquidity = Decimal("20004500137.498290928785113714000000000000000000")
				// calc_amount_zero_delta(liquidity, sqrtPriceLowerTick, sqrtPriceUpperTick, False)
				// Decimal('999999.999999999999999999999957642595723576')
				// The value above gets rounded down to DefaultAmt0.Sub(osmomath.OneInt()). Then, we add DefaultAmt0.
				amount0Expected: DefaultAmt0.Sub(osmomath.OneInt()).Add(DefaultAmt0),
				amount1Expected: osmomath.ZeroInt(),
				// current tick is 0, so create the position completely above it
				lowerTick: 100,
				upperTick: 200,
			},
			amount0ToAdd: DefaultAmt0,
			amount1ToAdd: osmomath.ZeroInt(),
		},
		"single sided amount1 add": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: osmomath.ZeroInt(),
				amount1Expected: DefaultAmt1.Add(DefaultAmt1).Sub(osmomath.NewInt(1)),
				// current tick is 0, so create the position completely below it
				lowerTick: -200,
				upperTick: -100,
			},
			amount0ToAdd: osmomath.ZeroInt(),
			amount1ToAdd: DefaultAmt1,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// --- Setup ---
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			var (
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				owner                       = s.TestAccs[1]
				config                      = *tc.setupConfig
				err                         error
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, tc.sutConfigOverwrite)

			// Create a pool and position with a full range position.
			pool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(ETH, USDC)
			fundCoins := config.tokensProvided

			// Fund tokens that is used to create the owners initial position
			if !tc.amount0ToAdd.IsZero() {
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd.Add(osmomath.NewInt(1)))))
			}
			if !tc.amount1ToAdd.IsZero() {
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(USDC, tc.amount1ToAdd.Add(osmomath.NewInt(1)))))
			}

			// Create a position from the parameters in the test case.
			testCoins := sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd), sdk.NewCoin(USDC, tc.amount1ToAdd))
			positionData, err := s.Clk.CreatePosition(s.Ctx, pool.GetId(), owner, testCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), config.lowerTick, config.upperTick)
			s.Require().NoError(err)

			// Move the block time forward
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultTimeElapsed))

			// Check the position owner's balance before the add.
			preBalanceToken0 := s.App.BankKeeper.GetBalance(s.Ctx, owner, pool.GetToken0())
			preBalanceToken1 := s.App.BankKeeper.GetBalance(s.Ctx, owner, pool.GetToken1())

			if !tc.lastPositionInPool {
				s.FundAcc(s.TestAccs[0], fundCoins)
				_, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], config.tokensProvided, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			// now we fund the owner account again for the amount0ToAdd and amount1ToAdd coins.
			// only fund coins if the amount is non-negative or else test would panic here
			if !tc.amount0ToAdd.IsNegative() {
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd.Add(osmomath.NewInt(1)))))
			}
			if !tc.amount1ToAdd.IsNegative() {
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(USDC, tc.amount1ToAdd.Add(osmomath.NewInt(1)))))
			}

			// --- System under test ---
			newPosId, newAmt0, newAmt1, err := concentratedLiquidityKeeper.AddToPosition(s.Ctx, owner, positionData.ID, tc.amount0ToAdd, tc.amount1ToAdd, config.amount0Minimum, config.amount1Minimum)
			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(osmomath.Int{}, newAmt0)
				s.Require().Equal(osmomath.Int{}, newAmt1)
				s.Require().Equal(uint64(0), newPosId)
				s.Require().ErrorContains(err, config.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), newAmt0.String())
			s.Require().Equal(config.amount1Expected.String(), newAmt1.String())

			// We expect the position ID to be 4 since we have three setup positions (initial full range position, owner position, and lastPositionInPool position)
			s.Require().Equal(uint64(4), newPosId)

			postBalanceToken0 := s.App.BankKeeper.GetBalance(s.Ctx, owner, pool.GetToken0())
			postBalanceToken1 := s.App.BankKeeper.GetBalance(s.Ctx, owner, pool.GetToken1())

			// Ensure that we utilized all the tokens we funded the account with when adding to the position.
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, postBalanceToken0.Amount.Sub(preBalanceToken0.Amount), osmomath.ZeroInt())
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, postBalanceToken1.Amount.Sub(preBalanceToken1.Amount), osmomath.ZeroInt())

			// now check that old position id has been successfully deleted
			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionData.ID)
			s.Require().Error(err)
		})
	}
}

// mergeConfigs merges every desired non-zero field from overwrite
// into dst. dst is mutated due to being a pointer.
func mergeConfigs(dst *lpTest, overwrite *lpTest) {
	if overwrite != nil {
		if overwrite.poolId != 0 {
			dst.poolId = overwrite.poolId
		}
		if overwrite.lowerTick != 0 {
			dst.lowerTick = overwrite.lowerTick
		}
		if overwrite.upperTick != 0 {
			dst.upperTick = overwrite.upperTick
		}

		overwiteTokens := false
		for _, coin := range overwrite.tokensProvided {
			if coin.Amount.IsPositive() {
				overwiteTokens = true
			}
		}
		if overwiteTokens {
			dst.tokensProvided = overwrite.tokensProvided
		}
		if !overwrite.liquidityAmount.IsNil() {
			dst.liquidityAmount = overwrite.liquidityAmount
		}
		if !overwrite.amount0Minimum.IsNil() {
			dst.amount0Minimum = overwrite.amount0Minimum
		}
		if overwrite.customTokensProvided {
			dst.tokensProvided = overwrite.tokensProvided
		}
		if !overwrite.amount0Expected.IsNil() {
			dst.amount0Expected = overwrite.amount0Expected
		}
		if !overwrite.amount1Minimum.IsNil() {
			dst.amount1Minimum = overwrite.amount1Minimum
		}
		if !overwrite.amount1Expected.IsNil() {
			dst.amount1Expected = overwrite.amount1Expected
		}
		if overwrite.expectedError != nil {
			dst.expectedError = overwrite.expectedError
		}
		if overwrite.tickSpacing != 0 {
			dst.tickSpacing = overwrite.tickSpacing
		}
		if overwrite.isNotFirstPosition != false {
			dst.isNotFirstPosition = overwrite.isNotFirstPosition
		}
		if overwrite.isNotFirstPositionWithSameAccount {
			dst.isNotFirstPositionWithSameAccount = overwrite.isNotFirstPositionWithSameAccount
		}
		if !overwrite.joinTime.IsZero() {
			dst.joinTime = overwrite.joinTime
		}
		if !overwrite.expectedSpreadRewardGrowthOutsideLower.Equal(sdk.DecCoins{}) {
			dst.expectedSpreadRewardGrowthOutsideLower = overwrite.expectedSpreadRewardGrowthOutsideLower
		}
		if !overwrite.expectedSpreadRewardGrowthOutsideUpper.Equal(sdk.DecCoins{}) {
			dst.expectedSpreadRewardGrowthOutsideUpper = overwrite.expectedSpreadRewardGrowthOutsideUpper
		}
		if overwrite.positionId != 0 {
			dst.positionId = overwrite.positionId
		}
		if overwrite.underlyingLockId != 0 {
			dst.underlyingLockId = overwrite.underlyingLockId
		}
		if overwrite.expectedLowerTick != 0 {
			dst.expectedLowerTick = overwrite.expectedLowerTick
		}
		if overwrite.expectedUpperTick != 0 {
			dst.expectedUpperTick = overwrite.expectedUpperTick
		}
	}
}

func (s *KeeperTestSuite) TestSendCoinsBetweenPoolAndUser() {
	type sendTest struct {
		coin0       sdk.Coin
		coin1       sdk.Coin
		poolToUser  bool
		expectedErr error
	}
	tests := map[string]sendTest{
		"asset0 and asset1 are positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
		},
		"only asset0 is positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", osmomath.NewInt(0)),
		},
		"only asset1 is positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", osmomath.NewInt(0)),
			coin1: sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
		},
		"only asset0 is greater than sender has, position creation (user to pool)": {
			coin0:       sdk.NewCoin("eth", osmomath.NewInt(100000000000000)),
			coin1:       sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
			expectedErr: InsufficientFundsError,
		},
		"only asset1 is greater than sender has, position creation (user to pool)": {
			coin0:       sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", osmomath.NewInt(100000000000000)),
			expectedErr: InsufficientFundsError,
		},
		"asset0 and asset1 are positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1:      sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
			poolToUser: true,
		},
		"only asset0 is positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1:      sdk.NewCoin("usdc", osmomath.NewInt(0)),
			poolToUser: true,
		},
		"only asset1 is positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", osmomath.NewInt(0)),
			coin1:      sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
			poolToUser: true,
		},
		"only asset0 is greater than sender has, withdraw (pool to user)": {
			coin0:       sdk.NewCoin("eth", osmomath.NewInt(100000000000000)),
			coin1:       sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
			poolToUser:  true,
			expectedErr: InsufficientFundsError,
		},
		"only asset1 is greater than sender has, withdraw (pool to user)": {
			coin0:       sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", osmomath.NewInt(100000000000000)),
			poolToUser:  true,
			expectedErr: InsufficientFundsError,
		},
		"asset0 is negative - error": {
			coin0: sdk.Coin{Denom: "eth", Amount: osmomath.NewInt(1000000).Neg()},
			coin1: sdk.NewCoin("usdc", osmomath.NewInt(1000000)),

			expectedErr: types.Amount0IsNegativeError{Amount0: osmomath.NewInt(1000000).Neg()},
		},
		"asset1 is negative - error": {
			coin0: sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1: sdk.Coin{Denom: "usdc", Amount: osmomath.NewInt(1000000).Neg()},

			expectedErr: types.Amount1IsNegativeError{Amount1: osmomath.NewInt(1000000).Neg()},
		},
		"asset0 is zero - passes": {
			coin0: sdk.NewCoin("eth", osmomath.ZeroInt()),
			coin1: sdk.NewCoin("usdc", osmomath.NewInt(1000000)),
		},
		"asset1 is zero - passes": {
			coin0: sdk.NewCoin("eth", osmomath.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", osmomath.ZeroInt()),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// create a CL pool
			s.PrepareConcentratedPool()

			// store pool interface
			poolI, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, 1)
			s.Require().NoError(err)
			concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
			if !ok {
				s.FailNow("poolI is not a ConcentratedPoolExtension")
			}

			// fund pool address and user address
			s.FundAcc(poolI.GetAddress(), sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(10000000000000)), sdk.NewCoin("usdc", osmomath.NewInt(1000000000000))))
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(10000000000000)), sdk.NewCoin("usdc", osmomath.NewInt(1000000000000))))

			// set sender and receiver based on test case
			sender := s.TestAccs[0]
			receiver := concentratedPool.GetAddress()
			if tc.poolToUser {
				sender = concentratedPool.GetAddress()
				receiver = s.TestAccs[0]
			}

			// store pre send balance of sender and receiver
			preSendBalanceSender := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)
			preSendBalanceReceiver := s.App.BankKeeper.GetAllBalances(s.Ctx, receiver)

			// system under test
			err = s.App.ConcentratedLiquidityKeeper.SendCoinsBetweenPoolAndUser(s.Ctx, concentratedPool.GetToken0(), concentratedPool.GetToken1(), tc.coin0.Amount, tc.coin1.Amount, sender, receiver)

			// store post send balance of sender and receiver
			postSendBalanceSender := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)
			postSendBalanceReceiver := s.App.BankKeeper.GetAllBalances(s.Ctx, receiver)

			// check error if expected
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				return
			}

			// otherwise, ensure balances are added/deducted appropriately
			expectedPostSendBalanceSender := preSendBalanceSender.Sub(sdk.NewCoins(tc.coin0, tc.coin1)...)
			expectedPostSendBalanceReceiver := preSendBalanceReceiver.Add(tc.coin0, tc.coin1)

			s.Require().NoError(err)
			s.Require().Equal(expectedPostSendBalanceSender.String(), postSendBalanceSender.String())
			s.Require().Equal(expectedPostSendBalanceReceiver.String(), postSendBalanceReceiver.String())
		})
	}
}

func (s *KeeperTestSuite) TestUpdatePosition() {
	type updatePositionTest struct {
		poolId                    uint64
		ownerIndex                int
		lowerTick                 int64
		upperTick                 int64
		joinTime                  time.Time
		positionId                uint64
		liquidityDelta            osmomath.Dec
		amount0Expected           osmomath.Int
		amount1Expected           osmomath.Int
		expectedPositionLiquidity osmomath.Dec
		expectedTickLiquidity     osmomath.Dec
		expectedPoolLiquidity     osmomath.Dec
		numPositions              int
		expectedError             bool
	}

	tests := map[string]updatePositionTest{
		"update existing position with positive amount": {
			poolId:         1,
			ownerIndex:     0,
			lowerTick:      DefaultLowerTick,
			upperTick:      DefaultUpperTick,
			joinTime:       DefaultJoinTime,
			positionId:     DefaultPositionId,
			liquidityDelta: DefaultLiquidityAmt,
			// Note: rounds up in favor of the pool.
			amount0Expected:           DefaultAmt0Expected.Add(roundingError),
			amount1Expected:           DefaultAmt1Expected,
			expectedPositionLiquidity: DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
			expectedTickLiquidity:     DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
			expectedPoolLiquidity:     DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
			numPositions:              1,
			expectedError:             false,
		},
		"update existing position with negative amount": {
			poolId:          1,
			ownerIndex:      0,
			lowerTick:       DefaultLowerTick,
			upperTick:       DefaultUpperTick,
			joinTime:        DefaultJoinTime,
			positionId:      DefaultPositionId,
			liquidityDelta:  DefaultLiquidityAmt.Neg(), // negative
			amount0Expected: DefaultAmt0Expected.Neg(),
			// Note: rounds down in favor of the pool (compared to the positive case which rounds up).
			amount1Expected:           DefaultAmt1Expected.Sub(roundingError).Neg(),
			expectedPositionLiquidity: osmomath.ZeroDec(),
			expectedTickLiquidity:     osmomath.ZeroDec(),
			expectedPoolLiquidity:     osmomath.ZeroDec(),
			numPositions:              1,
			expectedError:             false,
		},
		"error: attempting to create two same position and update them": {
			poolId:          1,
			ownerIndex:      0,
			lowerTick:       DefaultLowerTick,
			upperTick:       DefaultUpperTick,
			joinTime:        DefaultJoinTime,
			liquidityDelta:  DefaultLiquidityAmt.Neg(),
			amount0Expected: DefaultAmt0Expected.Neg(),
			amount1Expected: DefaultAmt1Expected.Neg(),
			numPositions:    2,
			expectedError:   true,
		},
		"error - update existing position with negative amount (more than liquidity provided)": {
			poolId:         1,
			ownerIndex:     0,
			lowerTick:      DefaultLowerTick,
			upperTick:      DefaultUpperTick,
			joinTime:       DefaultJoinTime,
			liquidityDelta: DefaultLiquidityAmt.Neg().Mul(osmomath.NewDec(2)),
			numPositions:   1,
			expectedError:  true,
		},
		"error: different tick range from the existing position": {
			positionId:     DefaultPositionId,
			poolId:         1,
			ownerIndex:     0,
			lowerTick:      DefaultUpperTick + 1,
			upperTick:      DefaultUpperTick + 100,
			joinTime:       DefaultJoinTime,
			liquidityDelta: DefaultLiquidityAmt,
			numPositions:   1,
			expectedError:  true,
		},
		"error: invalid pool id": {
			poolId:         2,
			ownerIndex:     0,
			lowerTick:      DefaultLowerTick,
			upperTick:      DefaultUpperTick,
			joinTime:       DefaultJoinTime,
			liquidityDelta: DefaultLiquidityAmt,
			numPositions:   1,
			expectedError:  true,
		},
		"error: invalid owner": {
			poolId:          1,
			ownerIndex:      1, // using a different address makes this a new position
			lowerTick:       DefaultLowerTick,
			upperTick:       DefaultUpperTick,
			joinTime:        DefaultJoinTime,
			liquidityDelta:  DefaultLiquidityAmt,
			numPositions:    1,
			amount0Expected: DefaultAmt0Expected,
			amount1Expected: DefaultAmt1Expected,
			expectedError:   true,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			s.Ctx = s.Ctx.WithBlockTime(time.Unix(0, 0))

			// create a CL pool
			s.PrepareConcentratedPool()

			// to ensure that the position's join time is set to the desired value
			s.Ctx = s.Ctx.WithBlockTime(tc.joinTime)

			// create position
			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], DefaultCoins)
			_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(
				s.Ctx,
				1,
				s.TestAccs[0],
				DefaultCoins,
				osmomath.ZeroInt(), osmomath.ZeroInt(),
				DefaultLowerTick, DefaultUpperTick,
			)
			s.Require().NoError(err)

			// explicitly make update time different to ensure that the pool is updated with last liquidity update.
			expectedUpdateTime := tc.joinTime.Add(time.Second)
			s.Ctx = s.Ctx.WithBlockTime(expectedUpdateTime)

			// system under test
			updateData, err := s.App.ConcentratedLiquidityKeeper.UpdatePosition(
				s.Ctx,
				tc.poolId,
				s.TestAccs[tc.ownerIndex],
				tc.lowerTick,
				tc.upperTick,
				tc.liquidityDelta,
				tc.joinTime,
				tc.positionId,
			)

			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Equal(osmomath.Int{}, updateData.Amount0)
				s.Require().Equal(osmomath.Int{}, updateData.Amount1)
			} else {
				s.Require().NoError(err)

				if tc.liquidityDelta.Equal(DefaultLiquidityAmt.Neg()) {
					s.Require().True(updateData.LowerTickIsEmpty)
					s.Require().True(updateData.UpperTickIsEmpty)
				} else {
					s.Require().False(updateData.LowerTickIsEmpty)
					s.Require().False(updateData.UpperTickIsEmpty)
				}

				var (
					expectedAmount0 osmomath.Dec
					expectedAmount1 osmomath.Dec
				)

				// For the context of this test case, we are not testing the calculation of the amounts
				// As a result, whenever non-default values are expected, we estimate them using the internal CalcActualAmounts function
				if tc.amount0Expected.IsNil() || tc.amount1Expected.IsNil() {
					pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.poolId)
					s.Require().NoError(err)

					expectedAmount0, expectedAmount1, err = pool.CalcActualAmounts(s.Ctx, tc.lowerTick, tc.upperTick, tc.liquidityDelta)
					s.Require().NoError(err)
				} else {
					expectedAmount0 = tc.amount0Expected.ToLegacyDec()
					expectedAmount1 = tc.amount1Expected.ToLegacyDec()
				}

				s.Require().Equal(expectedAmount0.TruncateInt().String(), updateData.Amount0.String())
				s.Require().Equal(expectedAmount1.TruncateInt().String(), updateData.Amount1.String())

				// validate if position has been properly updated
				s.validatePositionUpdate(s.Ctx, tc.positionId, tc.expectedPositionLiquidity)
				s.validateTickUpdates(tc.poolId, tc.lowerTick, tc.upperTick, tc.expectedTickLiquidity, cl.EmptyCoins, cl.EmptyCoins)

				// validate if pool liquidity has been updated properly
				poolI, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
				if !ok {
					s.FailNow("poolI is not a ConcentratedPoolExtension")
				}
				s.Require().Equal(tc.expectedPoolLiquidity, concentratedPool.GetLiquidity())

				// Test that liquidity update time was successfully changed.
				s.Require().Equal(expectedUpdateTime, poolI.GetLastLiquidityUpdate())
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitializeInitialPositionForPool() {
	sqrt := func(x int64) osmomath.BigDec {
		sqrt, err := osmomath.MonotonicSqrt(osmomath.NewDec(x))
		s.Require().NoError(err)
		return osmomath.BigDecFromDecMut(sqrt)
	}

	type sendTest struct {
		amount0Desired        osmomath.Int
		amount1Desired        osmomath.Int
		tickSpacing           uint64
		expectedCurrSqrtPrice osmomath.BigDec
		expectedTick          int64
		expectedError         error
	}
	tests := map[string]sendTest{
		"happy path": {
			amount0Desired:        DefaultAmt0,
			amount1Desired:        DefaultAmt1,
			tickSpacing:           DefaultTickSpacing,
			expectedCurrSqrtPrice: DefaultCurrSqrtPrice,
			expectedTick:          DefaultCurrTick,
		},
		"100_000_050 and tick spacing 100, price level where curr sqrt price does not translate to allowed tick (assumes exponent at price one of -6 and tick spacing of 100)": {
			amount0Desired:        osmomath.OneInt(),
			amount1Desired:        osmomath.NewInt(100_000_050),
			tickSpacing:           DefaultTickSpacing,
			expectedCurrSqrtPrice: sqrt(100_000_050),
			expectedTick:          72000000,
		},
		"100_000_051 and tick spacing 100, price level where curr sqrt price does not translate to allowed tick (assumes exponent at price one of -6 and tick spacing of 100)": {
			amount0Desired:        osmomath.OneInt(),
			amount1Desired:        osmomath.NewInt(100_000_051),
			tickSpacing:           DefaultTickSpacing,
			expectedCurrSqrtPrice: sqrt(100_000_051),
			expectedTick:          72000000,
		},
		"100_000_051 and tick spacing 1, price level where curr sqrt price translates to allowed tick (assumes exponent at price one of -6 and tick spacing of 1)": {
			amount0Desired:        osmomath.OneInt(),
			amount1Desired:        osmomath.NewInt(100_000_051),
			tickSpacing:           1,
			expectedCurrSqrtPrice: sqrt(100_000_051),
			// We expect the returned tick to always be rounded down.
			// In this case, tick 72000000 corresponds to 100_000_000,
			// while 72000001 corresponds to 100_000_100.
			expectedTick: 72000000,
		},
		"error: amount0Desired is zero": {
			amount0Desired: osmomath.ZeroInt(),
			amount1Desired: DefaultAmt1,
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: osmomath.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: amount1Desired is zero": {
			amount0Desired: DefaultAmt0,
			amount1Desired: osmomath.ZeroInt(),
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: osmomath.ZeroInt()},
		},
		"error: both amount0Desired and amount01Desired is zero": {
			amount0Desired: osmomath.ZeroInt(),
			amount1Desired: osmomath.ZeroInt(),
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: osmomath.ZeroInt(), Amount1: osmomath.ZeroInt()},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// create a CL pool
			pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, osmomath.ZeroDec())

			// System under test
			err := s.App.ConcentratedLiquidityKeeper.InitializeInitialPositionForPool(s.Ctx, pool, tc.amount0Desired, tc.amount1Desired)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
			} else {
				s.Require().NoError(err)

				pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedCurrSqrtPrice.String(), pool.GetCurrentSqrtPrice().String())
				s.Require().Equal(tc.expectedTick, pool.GetCurrentTick())
			}
		})
	}
}

func (s *KeeperTestSuite) TestInverseRelation_CreatePosition_WithdrawPosition() {
	var (
		errToleranceOneRoundUp = osmomath.ErrTolerance{
			AdditiveTolerance: osmomath.OneDec(),
			RoundingDir:       osmomath.RoundUp,
		}
	)
	tests := makeTests(positionCases)

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
			// Merge tc with baseCase and update tc to the merged result. This is done to reduce the amount of boilerplate in test cases.
			baseConfigCopy := *baseCase
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			clKeeper := s.App.ConcentratedLiquidityKeeper

			// Fund account to pay for the pool creation spread reward.
			s.FundAcc(s.TestAccs[0], PoolCreationFee)

			// Create a CL pool with custom tickSpacing
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, osmomath.ZeroDec()))
			s.Require().NoError(err)
			poolBefore, err := clKeeper.GetPool(s.Ctx, poolID)
			s.Require().NoError(err)

			liquidityBefore, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, poolID)
			s.Require().NoError(err)

			// Pre-set spread reward growth accumulator
			if !tc.preSetChargeSpreadRewards.IsZero() {
				s.AddToSpreadRewardAccumulator(1, tc.preSetChargeSpreadRewards)
			}

			// If we want to test a non-first position, we create a first position with a separate account
			if tc.isNotFirstPosition {
				s.SetupPosition(1, s.TestAccs[1], DefaultCoins, tc.lowerTick, tc.upperTick, false)
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], DefaultCoins)

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// System under test.
			positionData, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.tokensProvided, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)

			var (
				positionId              = positionData.ID
				newLowerTick            = positionData.LowerTick
				newUpperTick            = positionData.UpperTick
				liquidityCreated        = positionData.Liquidity
				amtDenom0CreatePosition = positionData.Amount0
				amtDenom1CreatePosition = positionData.Amount1
			)

			if tc.expectedLowerTick != 0 {
				s.Require().Equal(tc.expectedLowerTick, newLowerTick)
				tc.lowerTick = newLowerTick
			}
			if tc.expectedUpperTick != 0 {
				s.Require().Equal(tc.expectedUpperTick, newUpperTick)
				tc.upperTick = newUpperTick
			}

			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime.Add(time.Hour * 24))
			amtDenom0WithdrawPosition, amtDenom1WithdrawPosition, err := clKeeper.WithdrawPosition(s.Ctx, s.TestAccs[0], positionId, liquidityCreated)
			s.Require().NoError(err)

			// INVARIANTS

			// 1. amount for denom0 and denom1 upon creating and withdraw position should be same
			// Note: round down because create position rounds in favor of the pool.
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, amtDenom0CreatePosition, amtDenom0WithdrawPosition)
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, amtDenom1CreatePosition, amtDenom1WithdrawPosition)

			// 2. user balance and pool balance after creating / withdrawing position should be same
			userBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// Note: rounding down since position creation rounds in favor of the pool.
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, userBalancePrePositionCreation.AmountOf(ETH), userBalancePostPositionCreation.AmountOf(ETH))
			osmoassert.Equal(s.T(), errToleranceOneRoundDown, userBalancePrePositionCreation.AmountOf(USDC), userBalancePostPositionCreation.AmountOf(USDC))

			// Note: rounding up since withdrawal rounds in favor of the pool.
			osmoassert.Equal(s.T(), errToleranceOneRoundUp, poolBalancePrePositionCreation.AmountOf(ETH), poolBalancePostPositionCreation.AmountOf(ETH))
			osmoassert.Equal(s.T(), errToleranceOneRoundUp, poolBalancePrePositionCreation.AmountOf(USDC), poolBalancePostPositionCreation.AmountOf(USDC))

			// 3. Check that position's liquidity was deleted
			positionLiquidity, err := clKeeper.GetPositionLiquidity(s.Ctx, tc.positionId)
			s.Require().Error(err)
			s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: tc.positionId})
			s.Require().Equal(osmomath.Dec{}, positionLiquidity)

			// 4. Check that pool has come back to original state

			liquidityAfter, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, poolID)
			s.Require().NoError(err)

			// Note: one ends up remaining due to rounding in favor of the pool.
			osmoassert.Equal(s.T(), errToleranceOneRoundUp, liquidityBefore.AmountOf(ETH), liquidityAfter.AmountOf(ETH))
			osmoassert.Equal(s.T(), errToleranceOneRoundUp, liquidityBefore.AmountOf(USDC), liquidityAfter.AmountOf(USDC))
		})
	}
}

func (s *KeeperTestSuite) TestUninitializePool() {
	tests := map[string]struct {
		poolId       uint64
		hasPositions bool
		expectError  error
	}{
		"valid uninitialization": {
			poolId: defaultPoolId,
		},
		"error: pool does not exist": {
			poolId:      defaultPoolId + 1,
			expectError: types.PoolNotFoundError{PoolId: defaultPoolId + 1},
		},
		"error: attempted to uninitialize pool with liquidity": {
			poolId:       defaultPoolId,
			hasPositions: true,
			expectError:  types.UninitializedPoolWithLiquidityError{PoolId: defaultPoolId},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			pool := s.PrepareConcentratedPool()

			if tc.hasPositions {
				s.SetupDefaultPosition(pool.GetId())
			}

			err := clKeeper.InitializeInitialPositionForPool(s.Ctx, pool, DefaultAmt0, DefaultAmt1)
			s.Require().NoError(err)

			err = clKeeper.UninitializePool(s.Ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectError)
				return
			}
			s.Require().NoError(err)

			// get pool and confirm that sqrt price and tick were reset to zero
			pool, err = clKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			actualSqrtPrice := pool.GetCurrentSqrtPrice()
			actualTick := pool.GetCurrentTick()
			s.Require().Equal(osmomath.ZeroBigDec(), actualSqrtPrice)
			s.Require().Equal(int64(0), actualTick)
		})
	}
}

func (s *KeeperTestSuite) TestIsLockMature() {
	type sendTest struct {
		remainingLockDuration time.Duration
		unlockingPosition     bool
		lockedPosition        bool
		expectedLockIsMature  bool
	}
	tests := map[string]sendTest{
		"lock does not exist": {
			remainingLockDuration: 0,
			expectedLockIsMature:  true,
		},
		"unlocked": {
			remainingLockDuration: 0,
			unlockingPosition:     true,
			expectedLockIsMature:  true,
		},
		"unlocking": {
			remainingLockDuration: 1 * time.Hour,
			unlockingPosition:     true,
			expectedLockIsMature:  false,
		},
		"locked": {
			remainingLockDuration: 1 * time.Hour,
			lockedPosition:        true,
			expectedLockIsMature:  false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			var (
				concentratedLockId uint64
				positionData       types.CreateFullRangePositionData
				err                error
			)
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// create a CL pool and fund account
			pool := s.PrepareConcentratedPool()
			coinsToFund := sdk.NewCoins(DefaultCoin0, DefaultCoin1)
			s.FundAcc(s.TestAccs[0], coinsToFund)

			if tc.unlockingPosition {
				positionData, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund, tc.remainingLockDuration)
			} else if tc.lockedPosition {
				positionData, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund, tc.remainingLockDuration)
			} else {
				positionData, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund)
			}
			s.Require().NoError(err)

			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionData.ID)
			s.Require().NoError(err)

			// Increment block time by a second to ensure test cases with zero lock duration are in the past
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))

			// System under test
			lockIsMature, _ := s.App.ConcentratedLiquidityKeeper.IsLockMature(s.Ctx, concentratedLockId)

			s.Require().Equal(tc.expectedLockIsMature, lockIsMature)
		})
	}
}

func (s *KeeperTestSuite) TestValidatePositionUpdateById() {
	tests := map[string]struct {
		positionId              uint64
		updateInitiatorIndex    int
		lowerTickGiven          int64
		upperTickGiven          int64
		liquidityDeltaGiven     osmomath.Dec
		joinTimeGiven           time.Time
		poolIdGiven             uint64
		modifyPositionLiquidity bool
		expectError             error
	}{
		"valid update - adding liquidity": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"valid update - removing liquidity": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  osmomath.OneDec().Neg(), // negative
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"valid update - does not exist yet": {
			positionId:           DefaultPositionId + 2,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"liquidity position is less than liquidity delta (non negative)": {
			positionId:              DefaultPositionId,
			updateInitiatorIndex:    0,
			lowerTickGiven:          DefaultLowerTick,
			upperTickGiven:          DefaultUpperTick,
			liquidityDeltaGiven:     osmomath.NewDec(2), // non negative
			joinTimeGiven:           DefaultJoinTime,
			modifyPositionLiquidity: true, // modifies position to have less liquidity than liquidity delta
			poolIdGiven:             defaultPoolId,
		},
		"error: attempted to remove too much liquidty": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  DefaultLiquidityAmt.Add(osmomath.OneDec()).Neg(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LiquidityWithdrawalError{},
		},
		"error: invalid position id": {
			positionId:  0,
			expectError: types.ErrZeroPositionId,
		},
		"error: not an owner": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 1,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.PositionOwnerMismatchError{},
		},
		"error: lower tick mismatch": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick + 1,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: upper tick mismatch": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: invalid join time": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: pool id mismatch": {
			positionId:           DefaultPositionId + 1,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  osmomath.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.PositionsNotInSamePoolError{},
		},
		"error: liquidity position is less than liquidity delta (negative)": {
			positionId:              DefaultPositionId,
			updateInitiatorIndex:    0,
			lowerTickGiven:          DefaultLowerTick,
			upperTickGiven:          DefaultUpperTick,
			liquidityDeltaGiven:     osmomath.NewDec(2).Neg(),
			joinTimeGiven:           DefaultJoinTime,
			modifyPositionLiquidity: true, // modifies position to have less liquidity than liquidity delta
			poolIdGiven:             defaultPoolId,
			expectError:             types.LiquidityWithdrawalError{},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			clKeeper := s.App.ConcentratedLiquidityKeeper

			// Fund test accounts
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(DefaultCoin0, DefaultCoin1))

			// Create two pools
			poolOne := s.PrepareConcentratedPool()
			poolTwo := s.PrepareConcentratedPool()

			// Create a position in pool one from account 0
			s.SetupDefaultPositionAcc(poolOne.GetId(), s.TestAccs[0])
			// Create a position in pool two from account 1
			s.SetupDefaultPositionAcc(poolTwo.GetId(), s.TestAccs[1])

			updateInitiator := s.TestAccs[tc.updateInitiatorIndex]

			if tc.modifyPositionLiquidity {
				position, err := s.Clk.GetPosition(s.Ctx, tc.positionId)
				s.Require().NoError(err)
				owner, err := sdk.AccAddressFromBech32(position.Address)
				s.Require().NoError(err)
				s.Clk.SetPosition(s.Ctx, defaultPoolId, owner, position.LowerTick, position.UpperTick, position.JoinTime, osmomath.OneDec(), position.PositionId, 0)
			}

			err := clKeeper.ValidatePositionUpdateById(s.Ctx, tc.positionId, updateInitiator, tc.lowerTickGiven, tc.upperTickGiven, tc.liquidityDeltaGiven, tc.joinTimeGiven, tc.poolIdGiven)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectError)
				return
			}
			s.Require().NoError(err)
		})
	}
}
