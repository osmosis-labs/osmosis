package concentrated_liquidity_test

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type lpTest struct {
	poolId                            uint64
	owner                             sdk.AccAddress
	currentTick                       sdk.Int
	lowerTick                         int64
	upperTick                         int64
	joinTime                          time.Time
	positionId                        uint64
	currentSqrtP                      sdk.Dec
	amount0Desired                    sdk.Int
	amount0Minimum                    sdk.Int
	amount0Expected                   sdk.Int
	amount1Desired                    sdk.Int
	amount1Minimum                    sdk.Int
	amount1Expected                   sdk.Int
	liquidityAmount                   sdk.Dec
	tickSpacing                       uint64
	exponentAtPriceOne                sdk.Int
	isNotFirstPosition                bool
	isNotFirstPositionWithSameAccount bool
	expectedError                     error

	// fee related fields
	preSetChargeFee               sdk.DecCoin
	expectedFeeGrowthOutsideLower sdk.DecCoins
	expectedFeeGrowthOutsideUpper sdk.DecCoins
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
		amount0Desired:                    DefaultAmt0,
		amount0Minimum:                    sdk.ZeroInt(),
		amount0Expected:                   DefaultAmt0Expected,
		amount1Desired:                    DefaultAmt1,
		amount1Minimum:                    sdk.ZeroInt(),
		amount1Expected:                   DefaultAmt1Expected,
		liquidityAmount:                   DefaultLiquidityAmt,
		tickSpacing:                       DefaultTickSpacing,
		exponentAtPriceOne:                DefaultExponentAtPriceOne,
		joinTime:                          DefaultJoinTime,
		positionId:                        1,

		preSetChargeFee: oneEth,
		// in this setup lower tick < current tick < upper tick
		// the fee accumulator for ticks <= current tick are updated.
		expectedFeeGrowthOutsideLower: cl.EmptyCoins,
		// as a result, the upper tick is not updated.
		expectedFeeGrowthOutsideUpper: cl.EmptyCoins,
	}

	positionCases = map[string]lpTest{
		"base case": {
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"create a position with non default tick spacing (10) with ticks that fall into tick spacing requirements": {
			tickSpacing:                   10,
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"lower tick < upper tick < current tick -> both tick's fee accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick + 1),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"lower tick < upper tick < current tick -> the fee is not charged so tick accumulators are unset": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick + 1),

			preSetChargeFee:               sdk.NewDecCoin(ETH, sdk.ZeroInt()), // zero fee
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"current tick < lower tick < upper tick -> both tick's fee accumulators are unitilialized": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultLowerTick - 1),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"lower tick < upper tick == current tick -> both tick's fee accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,
		},
		"second position: lower tick < upper tick == current tick -> both tick's fee accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick),

			isNotFirstPositionWithSameAccount: true,
			positionId:                        2,

			liquidityAmount:               baseCase.liquidityAmount.MulInt64(2),
			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,
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
			lowerTick:     DefaultMinTick - 1,
			expectedError: types.InvalidTickError{Tick: DefaultMinTick - 1, IsLower: true, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		"error: upper tick out of bounds": {
			upperTick:     DefaultMaxTick + 1,
			expectedError: types.InvalidTickError{Tick: DefaultMaxTick + 1, IsLower: false, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		"error: upper tick is below the lower tick, but both are in bounds": {
			lowerTick:     50,
			upperTick:     40,
			expectedError: types.InvalidLowerUpperTickError{LowerTick: 50, UpperTick: 40},
		},
		"error: amount of token 0 is smaller than minimum; should fail and not update state": {
			amount0Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)),
			expectedError:  types.InsufficientLiquidityCreatedError{Actual: baseCase.amount0Expected, Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)), IsTokenZero: true},
		},
		"error: amount of token 1 is smaller than minimum; should fail and not update state": {
			amount1Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2)),
			expectedError:  types.InsufficientLiquidityCreatedError{Actual: baseCase.amount1Expected, Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2))},
		},
		"error: a non first position with zero amount desired for both denoms should fail liquidity delta check": {
			isNotFirstPosition: true,
			amount0Desired:     sdk.ZeroInt(),
			amount1Desired:     sdk.ZeroInt(),
			expectedError:      errors.New("liquidityDelta calculated equals zero"),
		},
		"error: attempt to use and upper and lower tick that are not divisible by tick spacing": {
			lowerTick:     int64(305451),
			upperTick:     int64(315001),
			tickSpacing:   10,
			expectedError: types.TickSpacingError{TickSpacing: 10, LowerTick: int64(305451), UpperTick: int64(315001)},
		},
		"error: first position cannot have a zero amount for denom0": {
			amount0Desired: sdk.ZeroInt(),
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: first position cannot have a zero amount for denom1": {
			amount1Desired: sdk.ZeroInt(),
			expectedError:  types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: sdk.ZeroInt()},
		},
		"error: first position cannot have a zero amount for both denom0 and denom1": {
			amount1Desired: sdk.ZeroInt(),
			amount0Desired: sdk.ZeroInt(),
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: sdk.ZeroInt()},
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

			// Merge tc with baseCase and update tc to the merged result. This is done to reduce the amount of boilerplate in test cases.
			baseConfigCopy := *baseCase
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			clKeeper := s.App.ConcentratedLiquidityKeeper

			// Fund account to pay for the pool creation fee.
			s.FundAcc(s.TestAccs[0], PoolCreationFee)

			// Create a CL pool with custom tickSpacing
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, tc.exponentAtPriceOne, sdk.ZeroDec()))
			s.Require().NoError(err)

			// Set mock listener to make sure that is is called when desired.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			pool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
			s.Require().NoError(err)

			// Pre-set fee growth accumulator
			if !tc.preSetChargeFee.IsZero() {
				err = clKeeper.ChargeFee(s.Ctx, 1, tc.preSetChargeFee)
				s.Require().NoError(err)
			}

			expectedNumCreatePositionEvents := 1

			// If we want to test a non-first position, we create a first position with a separate account
			if tc.isNotFirstPosition {
				s.SetupPosition(1, s.TestAccs[1], DefaultCoin0, DefaultCoin1, tc.lowerTick, tc.upperTick, DefaultJoinTime)
				expectedNumCreatePositionEvents += 1
			}

			expectedLiquidityCreated := tc.liquidityAmount
			if tc.isNotFirstPositionWithSameAccount {
				// Since this is a second position with the same parameters,
				// we expect to create half of the final liquidity amount.
				expectedLiquidityCreated = tc.liquidityAmount.QuoInt64(2)

				s.SetupPosition(1, s.TestAccs[0], DefaultCoin0, DefaultCoin1, tc.lowerTick, tc.upperTick, DefaultJoinTime)
				expectedNumCreatePositionEvents += 1
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())

			// System under test.
			positionId, asset0, asset1, liquidityCreated, joinTime, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.amount0Desired, tc.amount1Desired, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)

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
				s.Require().Equal(asset0, sdk.Int{})
				s.Require().Equal(asset1, sdk.Int{})
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Check account balances
				s.Require().Equal(userBalancePrePositionCreation.String(), userBalancePostPositionCreation.String())
				s.Require().Equal(poolBalancePrePositionCreation.String(), poolBalancePostPositionCreation.String())

				// Redundantly ensure that liquidity was not created
				liquidity, err := clKeeper.GetPositionLiquidity(s.Ctx, positionId)
				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: positionId})
				s.Require().Equal(sdk.Dec{}, liquidity)
				return
			}

			// Else, check that we had no error from creating the position, and that the liquidity and assets that were returned are expected
			s.Require().NoError(err)
			s.Require().Equal(tc.positionId, positionId)
			s.Require().Equal(tc.amount0Expected.String(), asset0.String())
			s.Require().Equal(tc.amount1Expected.String(), asset1.String())
			s.Require().Equal(expectedLiquidityCreated.String(), liquidityCreated.String())
			s.Require().Equal(s.Ctx.BlockTime(), joinTime)

			// Check account balances
			s.Require().Equal(userBalancePrePositionCreation.Sub(sdk.NewCoins(sdk.NewCoin(ETH, asset0), (sdk.NewCoin(USDC, asset1)))).String(), userBalancePostPositionCreation.String())
			s.Require().Equal(poolBalancePrePositionCreation.Add(sdk.NewCoin(ETH, asset0), (sdk.NewCoin(USDC, asset1))).String(), poolBalancePostPositionCreation.String())

			hasPosition := clKeeper.HasFullPosition(s.Ctx, tc.positionId)
			s.Require().True(hasPosition)

			// Check position state
			s.validatePositionUpdate(s.Ctx, positionId, expectedLiquidityCreated)

			s.validatePositionFeeAccUpdate(s.Ctx, tc.poolId, positionId, expectedLiquidityCreated)

			// Check tick state
			s.validateTickUpdates(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, tc.liquidityAmount, tc.expectedFeeGrowthOutsideLower, tc.expectedFeeGrowthOutsideUpper)

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

func (s *KeeperTestSuite) TestWithdrawPosition() {
	defaultTimeElapsed := time.Hour * 24
	uptimeHelper := getExpectedUptimes()
	defaultUptimeGrowth := uptimeHelper.hundredTokensMultiDenom
	DefaultJoinTime := s.Ctx.BlockTime()

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		createPositionOverwrite bool
		timeElapsed             time.Duration
	}{
		"base case: withdraw full liquidity amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998976 eth
				amount1Expected: baseCase.amount1Expected, // 5000 usdc
			},
			timeElapsed: defaultTimeElapsed,
		},
		"withdraw partial liquidity amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.QuoRoundUp(sdk.NewDec(2)),
				amount0Expected: baseCase.amount0Expected.QuoRaw(2), // 0.499488
				amount1Expected: baseCase.amount1Expected.QuoRaw(2), // 2500 usdc
			},
			timeElapsed: defaultTimeElapsed,
		},
		"withdraw full liquidity amount, forfeit incentives": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998976 eth
				amount1Expected: baseCase.amount1Expected, // 5000 usdc
			},
			timeElapsed: 0,
		},
		"error: no position created": {
			// setup parameters for creation a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				lowerTick:     -1, // valid tick at which no position exists
				positionId:    DefaultPositionId + 1,
				expectedError: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
			},
			timeElapsed: defaultTimeElapsed,
		},
		"error: insufficient liquidity": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.Add(sdk.OneDec()), // 1 more than available
				expectedError:   types.InsufficientLiquidityError{Actual: baseCase.liquidityAmount.Add(sdk.OneDec()), Available: baseCase.liquidityAmount},
			},
			timeElapsed: defaultTimeElapsed,
		},
		// TODO: test with custom amounts that potentially lead to truncations.
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			var (
				ctx                         = s.Ctx
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				liquidityCreated            = sdk.ZeroDec()
				owner                       = s.TestAccs[0]
				tc                          = tc
				config                      = *tc.setupConfig
				sutConfigOverwrite          = *tc.sutConfigOverwrite
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, &sutConfigOverwrite)

			// If a setupConfig is provided, use it to create a pool and position.
			pool := s.PrepareConcentratedPool()
			s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10000000000000)), sdk.NewCoin(USDC, sdk.NewInt(1000000000000))))

			// Create a position from the parameters in the test case.
			_, _, _, liquidityCreated, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, config.amount0Desired, config.amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// Set mock listener to make sure that is is called when desired.
			// It must be set after test position creation so that we do not record the call
			// for initial position update.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			ctx = ctx.WithBlockTime(ctx.BlockTime().Add(tc.timeElapsed))

			// Set global fee growth to 1 ETH and charge the fee to the pool.
			globalFeeGrowth := sdk.NewDecCoin(ETH, sdk.NewInt(1))
			err = concentratedLiquidityKeeper.ChargeFee(ctx, pool.GetId(), globalFeeGrowth)
			s.Require().NoError(err)

			// Add global uptime growth
			err = addToUptimeAccums(ctx, pool.GetId(), concentratedLiquidityKeeper, defaultUptimeGrowth)
			s.Require().NoError(err)

			// Determine the liquidity expected to remain after the withdraw.
			expectedRemainingLiquidity := liquidityCreated.Sub(config.liquidityAmount)

			expectedFeesClaimed := sdk.NewCoins()
			expectedIncentivesClaimed := sdk.NewCoins()
			// Set the expected fees claimed to the amount of liquidity created since the global fee growth is 1.
			// Fund the pool account with the expected fees claimed.
			if expectedRemainingLiquidity.IsZero() {
				expectedFeesClaimed = expectedFeesClaimed.Add(sdk.NewCoin(ETH, liquidityCreated.TruncateInt()))
				s.FundAcc(pool.GetAddress(), expectedFeesClaimed)
			}

			// Set expected incentives and fund pool with appropriate amount
			expectedIncentivesClaimed = expectedIncentivesFromUptimeGrowth(defaultUptimeGrowth, liquidityCreated, tc.timeElapsed, sdk.OneInt())
			s.FundAcc(pool.GetIncentivesAddress(), expectedIncentivesClaimed)

			// Note the pool and owner balances before collecting fees.
			poolBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, pool.GetAddress())
			incentivesBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, pool.GetIncentivesAddress())
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetAllBalances(ctx, owner)

			expectedPoolBalanceDelta := expectedFeesClaimed.Add(sdk.NewCoin(ETH, config.amount0Expected.Abs())).Add(sdk.NewCoin(USDC, config.amount1Expected.Abs()))

			// System under test.
			amtDenom0, amtDenom1, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, config.positionId, config.liquidityAmount)
			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(amtDenom0, sdk.Int{})
				s.Require().Equal(amtDenom1, sdk.Int{})
				s.Require().ErrorContains(err, config.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), amtDenom0.String())
			s.Require().Equal(config.amount1Expected.String(), amtDenom1.String())

			// If the remaining liquidity is zero, all fees and incentives should be collected and the position should be deleted.
			// Check if all fees and incentives were collected.
			poolBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, pool.GetAddress())
			incentivesBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, pool.GetIncentivesAddress())
			ownerBalancerAfterCollect := s.App.BankKeeper.GetAllBalances(ctx, owner)

			s.Require().Equal(expectedPoolBalanceDelta.String(), poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).String())
			s.Require().Equal(expectedPoolBalanceDelta.Add(expectedIncentivesClaimed...).String(), ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).String())
			s.Require().Equal(expectedIncentivesClaimed.String(), incentivesBalanceBeforeCollect.Sub(incentivesBalanceAfterCollect).String())

			if expectedRemainingLiquidity.IsZero() {
				// Check that the positionLiquidity was deleted.
				positionLiquidity, err := concentratedLiquidityKeeper.GetPositionLiquidity(ctx, config.positionId)
				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: config.positionId})
				s.Require().Equal(sdk.Dec{}, positionLiquidity)
			} else {
				// Check that the position was updated.
				s.validatePositionUpdate(ctx, config.positionId, expectedRemainingLiquidity)
			}

			// Check tick state.
			s.validateTickUpdates(ctx, config.poolId, owner, config.lowerTick, config.upperTick, expectedRemainingLiquidity, config.expectedFeeGrowthOutsideLower, config.expectedFeeGrowthOutsideUpper)

			// Validate event emitted.
			s.AssertEventEmitted(ctx, types.TypeEvtWithdrawPosition, 1)

			// Validate that listeners were called the desired number of times
			expectedAfterLastPoolPositionRemovedCallCount := 0
			if expectedRemainingLiquidity.IsZero() {
				// We want the hook to be called only when the last position (liquidity) is removed.
				// Not having any liquidity in the pool implies not having a valid sqrt price and tick. As a result,
				// we want the hook to run for the purposes of updating twap records.
				// Upon readding liquidity (recreating positions) to such pool, AfterInitialPoolPositionCreatedCallCount
				// will be called. Hence, updating twap with valid latest spot price.
				expectedAfterLastPoolPositionRemovedCallCount = 1
			}
			s.validateListenerCallCount(0, 0, expectedAfterLastPoolPositionRemovedCallCount, 0)
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
		if !overwrite.liquidityAmount.IsNil() {
			dst.liquidityAmount = overwrite.liquidityAmount
		}
		if !overwrite.amount0Minimum.IsNil() {
			dst.amount0Minimum = overwrite.amount0Minimum
		}
		if !overwrite.amount0Desired.IsNil() {
			dst.amount0Desired = overwrite.amount0Desired
		}
		if !overwrite.amount0Expected.IsNil() {
			dst.amount0Expected = overwrite.amount0Expected
		}
		if !overwrite.amount1Minimum.IsNil() {
			dst.amount1Minimum = overwrite.amount1Minimum
		}
		if !overwrite.amount1Desired.IsNil() {
			dst.amount1Desired = overwrite.amount1Desired
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
		if !overwrite.expectedFeeGrowthOutsideLower.IsEqual(sdk.DecCoins{}) {
			dst.expectedFeeGrowthOutsideLower = overwrite.expectedFeeGrowthOutsideLower
		}
		if !overwrite.expectedFeeGrowthOutsideUpper.IsEqual(sdk.DecCoins{}) {
			dst.expectedFeeGrowthOutsideUpper = overwrite.expectedFeeGrowthOutsideUpper
		}
		if overwrite.positionId != 0 {
			dst.positionId = overwrite.positionId
		}
	}
}

func (s *KeeperTestSuite) TestSendCoinsBetweenPoolAndUser() {
	type sendTest struct {
		coin0       sdk.Coin
		coin1       sdk.Coin
		poolToUser  bool
		expectError bool
	}
	tests := map[string]sendTest{
		"asset0 and asset1 are positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", sdk.NewInt(1000000)),
		},
		"only asset0 is positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", sdk.NewInt(0)),
		},
		"only asset1 is positive, position creation (user to pool)": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(0)),
			coin1: sdk.NewCoin("usdc", sdk.NewInt(1000000)),
		},
		"only asset0 is greater than sender has, position creation (user to pool)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(100000000000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(1000000)),
			expectError: true,
		},
		"only asset1 is greater than sender has, position creation (user to pool)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(100000000000000)),
			expectError: true,
		},
		"asset0 and asset1 are positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:      sdk.NewCoin("usdc", sdk.NewInt(1000000)),
			poolToUser: true,
		},
		"only asset0 is positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:      sdk.NewCoin("usdc", sdk.NewInt(0)),
			poolToUser: true,
		},
		"only asset1 is positive, withdraw (pool to user)": {
			coin0:      sdk.NewCoin("eth", sdk.NewInt(0)),
			coin1:      sdk.NewCoin("usdc", sdk.NewInt(1000000)),
			poolToUser: true,
		},
		"only asset0 is greater than sender has, withdraw (pool to user)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(100000000000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(1000000)),
			poolToUser:  true,
			expectError: true,
		},
		"only asset1 is greater than sender has, withdraw (pool to user)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(100000000000000)),
			poolToUser:  true,
			expectError: true,
		},
		"asset0 is negative - error": {
			coin0: sdk.Coin{Denom: "eth", Amount: sdk.NewInt(1000000).Neg()},
			coin1: sdk.NewCoin("usdc", sdk.NewInt(1000000)),

			expectError: true,
		},
		"asset1 is negative - error": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1: sdk.Coin{Denom: "usdc", Amount: sdk.NewInt(1000000).Neg()},

			expectError: true,
		},
		"asset0 is zero - passes": {
			coin0: sdk.NewCoin("eth", sdk.ZeroInt()),
			coin1: sdk.NewCoin("usdc", sdk.NewInt(1000000)),
		},
		"asset1 is zero - passes": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1: sdk.NewCoin("usdc", sdk.ZeroInt()),
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
			concentratedPool := poolI.(types.ConcentratedPoolExtension)

			// fund pool address and user address
			s.FundAcc(poolI.GetAddress(), sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))

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
			if tc.expectError {
				s.Require().Error(err)
				return
			}

			// otherwise, ensure balances are added/deducted appropriately
			expectedPostSendBalanceSender := preSendBalanceSender.Sub(sdk.NewCoins(tc.coin0, tc.coin1))
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
		liquidityDelta            sdk.Dec
		amount0Expected           sdk.Int
		amount1Expected           sdk.Int
		expectedPositionLiquidity sdk.Dec
		expectedTickLiquidity     sdk.Dec
		expectedPoolLiquidity     sdk.Dec
		numPositions              int
		expectedError             bool
	}

	tests := map[string]updatePositionTest{
		"update existing position with positive amount": {
			poolId:                    1,
			ownerIndex:                0,
			lowerTick:                 DefaultLowerTick,
			upperTick:                 DefaultUpperTick,
			joinTime:                  DefaultJoinTime,
			positionId:                1,
			liquidityDelta:            DefaultLiquidityAmt,
			amount0Expected:           DefaultAmt0Expected,
			amount1Expected:           DefaultAmt1Expected,
			expectedPositionLiquidity: DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
			expectedTickLiquidity:     DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
			expectedPoolLiquidity:     DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
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
			liquidityDelta: DefaultLiquidityAmt.Neg().Mul(sdk.NewDec(2)),
			numPositions:   1,
			expectedError:  true,
		},
		"try updating with ticks outside existing position's tick range - error because fee accumulator is uninitialized": {
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
		"new position when calling update position - error because fee accumulator is not initialized": {
			poolId:         1,
			ownerIndex:     1, // using a different address makes this a new position
			lowerTick:      DefaultLowerTick,
			upperTick:      DefaultUpperTick,
			joinTime:       DefaultJoinTime,
			liquidityDelta: DefaultLiquidityAmt,
			numPositions:   1,
			expectedError:  true,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// create a CL pool
			s.PrepareConcentratedPool()

			// create position
			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(
				s.Ctx,
				1,
				s.TestAccs[0],
				DefaultAmt0, DefaultAmt1,
				sdk.ZeroInt(), sdk.ZeroInt(),
				DefaultLowerTick, DefaultUpperTick,
			)
			s.Require().NoError(err)

			// system under test
			actualAmount0, actualAmount1, err := s.App.ConcentratedLiquidityKeeper.UpdatePosition(
				s.Ctx,
				tc.poolId,
				s.TestAccs[tc.ownerIndex],
				tc.lowerTick,
				tc.upperTick,
				tc.liquidityDelta,
				tc.joinTime,
				tc.positionId,
				DefaultUnderlyingLockId,
			)

			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Equal(sdk.Int{}, actualAmount0)
				s.Require().Equal(sdk.Int{}, actualAmount1)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(actualAmount0, tc.amount0Expected)
				s.Require().Equal(actualAmount1, tc.amount1Expected)

				// validate if position has been properly updated
				s.validatePositionUpdate(s.Ctx, tc.positionId, tc.expectedPositionLiquidity)
				s.validateTickUpdates(s.Ctx, tc.poolId, s.TestAccs[tc.ownerIndex], tc.lowerTick, tc.upperTick, tc.expectedTickLiquidity, cl.EmptyCoins, cl.EmptyCoins)

				// validate if pool liquidity has been updated properly
				poolI, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				concentratedPool := poolI.(types.ConcentratedPoolExtension)
				s.Require().Equal(tc.expectedPoolLiquidity, concentratedPool.GetLiquidity())

			}
		})
	}
}

func (s *KeeperTestSuite) TestInitializeInitialPositionForPool() {
	type sendTest struct {
		amount0Desired sdk.Int
		amount1Desired sdk.Int
		expectedError  error
	}
	tests := map[string]sendTest{
		"happy path": {
			amount0Desired: DefaultAmt0,
			amount1Desired: DefaultAmt1,
		},
		"error: amount0Desired is zero": {
			amount0Desired: sdk.ZeroInt(),
			amount1Desired: DefaultAmt1,
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: amount1Desired is zero": {
			amount0Desired: DefaultAmt0,
			amount1Desired: sdk.ZeroInt(),
			expectedError:  types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: sdk.ZeroInt()},
		},
		"error: both amount0Desired and amount01Desired is zero": {
			amount0Desired: sdk.ZeroInt(),
			amount1Desired: sdk.ZeroInt(),
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: sdk.ZeroInt()},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// create a CL pool
			pool := s.PrepareConcentratedPool()

			// System under test
			err := s.App.ConcentratedLiquidityKeeper.InitializeInitialPositionForPool(s.Ctx, pool, tc.amount0Desired, tc.amount1Desired)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestInverseRelation_CreatePosition_WithdrawPosition() {

	tests := map[string]lpTest{}

	// add test cases for different positions
	for name, test := range positionCases {
		tests[name] = test
	}

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

			// Fund account to pay for the pool creation fee.
			s.FundAcc(s.TestAccs[0], PoolCreationFee)

			// Create a CL pool with custom tickSpacing
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, tc.exponentAtPriceOne, sdk.ZeroDec()))
			s.Require().NoError(err)
			poolBefore, err := clKeeper.GetPool(s.Ctx, poolID)
			s.Require().NoError(err)

			liquidityBefore, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, poolID)
			s.Require().NoError(err)

			// Pre-set fee growth accumulator
			if !tc.preSetChargeFee.IsZero() {
				err = clKeeper.ChargeFee(s.Ctx, 1, tc.preSetChargeFee)
				s.Require().NoError(err)
			}

			// If we want to test a non-first position, we create a first position with a separate account
			if tc.isNotFirstPosition {
				s.SetupPosition(1, s.TestAccs[1], DefaultCoin0, DefaultCoin1, tc.lowerTick, tc.upperTick, DefaultJoinTime)
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// System under test.
			positionId, amtDenom0CreatePosition, amtDenom1CreatePosition, liquidityCreated, _, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.amount0Desired, tc.amount1Desired, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)

			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime.Add(time.Hour * 24))
			amtDenom0WithdrawPosition, amtDenom1WithdrawPosition, err := clKeeper.WithdrawPosition(s.Ctx, s.TestAccs[0], positionId, liquidityCreated)
			s.Require().NoError(err)

			// INVARIANTS

			// 1. amount for denom0 and denom1 upon creating and withdraw position should be same
			s.Require().Equal(amtDenom0CreatePosition, amtDenom0WithdrawPosition)
			s.Require().Equal(amtDenom1CreatePosition, amtDenom1WithdrawPosition)

			// 2. user balance and pool balance after creating / withdrawing position should be same
			userBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())
			s.Require().Equal(userBalancePrePositionCreation, userBalancePostPositionCreation)
			s.Require().Equal(poolBalancePrePositionCreation, poolBalancePostPositionCreation)

			// 3. Check that position's liquidity was deleted
			positionLiquidity, err := clKeeper.GetPositionLiquidity(s.Ctx, tc.positionId)
			s.Require().Error(err)
			s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: tc.positionId})
			s.Require().Equal(sdk.Dec{}, positionLiquidity)

			// 4. Check that pool has come back to original state

			liquidityAfter, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, poolID)
			s.Require().NoError(err)

			s.Require().NoError(err)
			s.Require().Equal(liquidityBefore, liquidityAfter)
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
		"error: attempted to unitialize pool with liquidity": {
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
			s.Require().Equal(sdk.ZeroDec(), actualSqrtPrice)
			s.Require().Equal(sdk.ZeroInt(), actualTick)
		})
	}
}
