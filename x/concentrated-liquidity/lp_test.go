package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type lpTest struct {
	poolId                            uint64
	currentTick                       sdk.Int
	lowerTick                         int64
	upperTick                         int64
	joinTime                          time.Time
	positionId                        uint64
	underlyingLockId                  uint64
	currentSqrtP                      sdk.Dec
	tokensProvided                    sdk.Coins
	customTokensProvided              bool
	amount0Minimum                    sdk.Int
	amount0Expected                   sdk.Int
	amount1Minimum                    sdk.Int
	amount1Expected                   sdk.Int
	liquidityAmount                   sdk.Dec
	tickSpacing                       uint64
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
		tokensProvided:                    DefaultCoins,
		customTokensProvided:              false,
		amount0Minimum:                    sdk.ZeroInt(),
		amount0Expected:                   DefaultAmt0Expected,
		amount1Minimum:                    sdk.ZeroInt(),
		amount1Expected:                   DefaultAmt1Expected,
		liquidityAmount:                   DefaultLiquidityAmt,
		tickSpacing:                       DefaultTickSpacing,
		joinTime:                          DefaultJoinTime,
		positionId:                        1,
		underlyingLockId:                  0,

		preSetChargeFee: oneEth,
		// in this setup lower tick < current tick < upper tick
		// the fee accumulator for ticks <= current tick are updated.
		expectedFeeGrowthOutsideLower: cl.EmptyCoins,
		// as a result, the upper tick is not updated.
		expectedFeeGrowthOutsideUpper: cl.EmptyCoins,
	}

	roundingError = sdk.OneInt()

	positionCases = map[string]lpTest{
		"base case": {
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"create a position with non default tick spacing (10) with ticks that fall into tick spacing requirements": {
			tickSpacing:                   10,
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick < current tick -> both tick's fee accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick + 100),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick < current tick -> the fee is not charged so tick accumulators are unset": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick + 100),

			preSetChargeFee:               sdk.NewDecCoin(ETH, sdk.ZeroInt()), // zero fee
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"current tick < lower tick < upper tick -> both tick's fee accumulators are unitilialized": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultLowerTick - 100),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
		},
		"lower tick < upper tick == current tick -> both tick's fee accumulators are updated with one eth": {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			currentTick: sdk.NewInt(DefaultUpperTick),

			preSetChargeFee:               oneEth,
			expectedFeeGrowthOutsideLower: oneEthCoins,

			// Rounding up in favor of the pool.
			amount0Expected: DefaultAmt0Expected.Add(roundingError),
			amount1Expected: DefaultAmt1Expected,
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
			amount0Minimum: sdk.NewInt(-1),
			expectedError:  types.NotPositiveRequireAmountError{Amount: sdk.NewInt(-1).String()},
		},
		"error: amount1 min is negative": {
			amount1Minimum: sdk.NewInt(-1),
			expectedError:  types.NotPositiveRequireAmountError{Amount: sdk.NewInt(-1).String()},
		},
		"error: amount of token 0 is smaller than minimum; should fail and not update state": {
			amount0Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)),
			// Add one since rounding up in favor of the pool.
			expectedError: types.InsufficientLiquidityCreatedError{Actual: baseCase.amount0Expected.Add(roundingError), Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)), IsTokenZero: true},
		},
		"error: amount of token 1 is smaller than minimum; should fail and not update state": {
			amount1Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2)),

			expectedError: types.InsufficientLiquidityCreatedError{Actual: baseCase.amount1Expected, Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2))},
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
			expectedError:        types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: first position cannot have a zero amount for denom1": {
			customTokensProvided: true,
			tokensProvided:       sdk.NewCoins(DefaultCoin0),
			expectedError:        types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: sdk.ZeroInt()},
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

			// Fund account to pay for the pool creation fee.
			s.FundAcc(s.TestAccs[0], PoolCreationFee)

			// Create a CL pool with custom tickSpacing
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, sdk.ZeroDec()))
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
				s.SetupPosition(1, s.TestAccs[1], DefaultCoins, tc.lowerTick, tc.upperTick, DefaultJoinTime)
				expectedNumCreatePositionEvents += 1
			}

			expectedLiquidityCreated := tc.liquidityAmount
			if tc.isNotFirstPositionWithSameAccount {
				// Since this is a second position with the same parameters,
				// we expect to create half of the final liquidity amount.
				expectedLiquidityCreated = tc.liquidityAmount.QuoInt64(2)

				s.SetupPosition(1, s.TestAccs[0], DefaultCoins, tc.lowerTick, tc.upperTick, DefaultJoinTime)
				expectedNumCreatePositionEvents += 1
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], DefaultCoins)

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())

			// System under test.
			positionId, asset0, asset1, liquidityCreated, joinTime, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.tokensProvided, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)

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

			hasPosition := clKeeper.HasPosition(s.Ctx, tc.positionId)
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
	nonOwner := s.TestAccs[1]

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		createPositionOverwrite bool
		timeElapsed             time.Duration
		createLockLocked        bool
		createLockUnlocking     bool
		createLockUnlocked      bool
		withdrawWithNonOwner    bool
	}{
		"base case: withdraw full liquidity amount": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998976 eth
				// Note: subtracting one due to truncations in favor of the pool when withdrawing.
				amount1Expected: baseCase.amount1Expected.Sub(sdk.OneInt()), // 5000 usdc
			},
			timeElapsed: defaultTimeElapsed,
		},
		"withdraw full liquidity amount with underlying lock that has finished unlocking": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				// Note: subtracting one due to truncations in favor of the pool when withdrawing.
				amount0Expected:  DefaultAmt0.Sub(sdk.OneInt()),
				amount1Expected:  DefaultAmt1.Sub(sdk.OneInt()),
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
			},
			createLockUnlocked: true,
			timeElapsed:        defaultTimeElapsed,
		},
		"error: withdraw full liquidity amount but still locked": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
				expectedError:    types.LockNotMatureError{PositionId: 1, LockId: 1},
			},
			createLockLocked: true,
			timeElapsed:      defaultTimeElapsed,
		},
		"error: withdraw full liquidity amount but still unlocking": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount:  FullRangeLiquidityAmt,
				underlyingLockId: 1,
				expectedError:    types.LockNotMatureError{PositionId: 1, LockId: 1},
			},
			createLockUnlocking: true,
			timeElapsed:         defaultTimeElapsed,
		},
		"withdraw partial liquidity amount": {
			setupConfig: baseCase,
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.QuoRoundUp(sdk.NewDec(2)),
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
				amount1Expected: baseCase.amount1Expected.Sub(sdk.OneInt()), // 5000 usdc
			},
			timeElapsed: 0,
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
				liquidityAmount: baseCase.liquidityAmount.Add(sdk.OneDec()), // 1 more than available
				expectedError:   types.InsufficientLiquidityError{Actual: baseCase.liquidityAmount.Add(sdk.OneDec()), Available: baseCase.liquidityAmount},
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
		// TODO: test with custom amounts that potentially lead to truncations.
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			var (
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				liquidityCreated            = sdk.ZeroDec()
				owner                       = s.TestAccs[0]
				tc                          = tc
				config                      = *tc.setupConfig
				sutConfigOverwrite          = *tc.sutConfigOverwrite
				err                         error
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, &sutConfigOverwrite)

			// If a setupConfig is provided, use it to create a pool and position.
			pool := s.PrepareConcentratedPool()
			fundCoins := config.tokensProvided
			s.FundAcc(owner, fundCoins)

			// Create a position from the parameters in the test case.
			if tc.createLockLocked {
				_, _, _, liquidityCreated, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed)
				s.Require().NoError(err)
			} else if tc.createLockUnlocking {
				_, _, _, liquidityCreated, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed+time.Hour)
				s.Require().NoError(err)
			} else if tc.createLockUnlocked {
				_, _, _, liquidityCreated, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed-time.Hour)
				s.Require().NoError(err)
			} else {
				_, _, _, liquidityCreated, _, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), owner, config.tokensProvided, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			// Set mock listener to make sure that is is called when desired.
			// It must be set after test position creation so that we do not record the call
			// for initial position update.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(tc.timeElapsed))

			// Set global fee growth to 1 ETH and charge the fee to the pool.
			globalFeeGrowth := sdk.NewDecCoin(ETH, sdk.NewInt(1))
			err = concentratedLiquidityKeeper.ChargeFee(s.Ctx, pool.GetId(), globalFeeGrowth)
			s.Require().NoError(err)

			// Add global uptime growth
			err = addToUptimeAccums(s.Ctx, pool.GetId(), concentratedLiquidityKeeper, defaultUptimeGrowth)
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

			communityPoolBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName))

			// Set expected incentives and fund pool with appropriate amount
			expectedIncentivesClaimed = expectedIncentivesFromUptimeGrowth(defaultUptimeGrowth, liquidityCreated, tc.timeElapsed, defaultMultiplier)

			// Fund full amount since forfeited incentives for the last position are sent to the community pool.
			expectedFullIncentivesFromAllUptimes := expectedIncentivesFromUptimeGrowth(defaultUptimeGrowth, liquidityCreated, types.SupportedUptimes[len(types.SupportedUptimes)-1], defaultMultiplier)
			s.FundAcc(pool.GetIncentivesAddress(), expectedFullIncentivesFromAllUptimes)

			// Note the pool and owner balances before collecting fees.
			poolBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			incentivesBalanceBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetIncentivesAddress())
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)

			expectedPoolBalanceDelta := expectedFeesClaimed.Add(sdk.NewCoin(ETH, config.amount0Expected.Abs())).Add(sdk.NewCoin(USDC, config.amount1Expected.Abs()))

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
			poolBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			incentivesBalanceAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetIncentivesAddress())
			ownerBalancerAfterCollect := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)
			communityPoolBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName))

			expectedOwnerBalanceDelta := expectedPoolBalanceDelta.Add(expectedIncentivesClaimed...)
			actualOwnerBalancerDelta := ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect)

			communityPoolBalanceDelta := communityPoolBalanceAfter.Sub(communityPoolBalanceBefore)

			actualIncentivesClaimed := incentivesBalanceBeforeCollect.Sub(incentivesBalanceAfterCollect).Sub(communityPoolBalanceDelta)

			s.Require().Equal(expectedPoolBalanceDelta.String(), poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).String())

			// TODO: Investigate why full range liquidity positions are slightly under claiming incentives here
			// https://github.com/osmosis-labs/osmosis/issues/4897
			errTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: sdk.NewDec(3),
				RoundingDir:       osmomath.RoundDown,
			}

			s.Require().NotEmpty(expectedOwnerBalanceDelta)
			for _, coin := range expectedOwnerBalanceDelta {
				expected := expectedOwnerBalanceDelta.AmountOf(coin.Denom)
				actual := actualOwnerBalancerDelta.AmountOf(coin.Denom)
				s.Require().Equal(0, errTolerance.Compare(expected, actual), fmt.Sprintf("expected %s, actual %s", expected, actual))
			}

			if tc.timeElapsed > 0 {
				s.Require().NotEmpty(expectedIncentivesClaimed)
			}
			for _, coin := range expectedIncentivesClaimed {
				expected := expectedIncentivesClaimed.AmountOf(coin.Denom)
				actual := actualIncentivesClaimed.AmountOf(coin.Denom)
				s.Require().Equal(0, errTolerance.Compare(expected, actual), fmt.Sprintf("expected %s, actual %s", expected, actual))
			}

			if expectedRemainingLiquidity.IsZero() {
				// Check that the positionLiquidity was deleted.
				positionLiquidity, err := concentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, config.positionId)
				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: config.positionId})
				s.Require().Equal(sdk.Dec{}, positionLiquidity)
			} else {
				// Check that the position was updated.
				s.validatePositionUpdate(s.Ctx, config.positionId, expectedRemainingLiquidity)
			}

			// Check tick state.
			s.validateTickUpdates(s.Ctx, config.poolId, owner, config.lowerTick, config.upperTick, expectedRemainingLiquidity, config.expectedFeeGrowthOutsideLower, config.expectedFeeGrowthOutsideUpper)

			// Validate event emitted.
			s.AssertEventEmitted(s.Ctx, types.TypeEvtWithdrawPosition, 1)

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

			// Dumb sanity-check that creating a position with the same liquidity amount after fully removing it does not error.
			// This is to be more thoroughly tested separately.
			if expectedRemainingLiquidity.IsZero() {
				// Add one USDC because we withdraw one less than originally funded due to truncation in favor of the pool.
				s.FundAcc(owner, sdk.NewCoins(sdk.NewCoin(USDC, sdk.OneInt())))
				_, _, _, _, _, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), owner, config.tokensProvided, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddToPosition() {
	defaultTimeElapsed := time.Hour * 24
	roundingError := sdk.OneInt()
	invalidSender := s.TestAccs[2]

	// These amounts are set based on the actual amounts passed in as inputs
	// to create position in the default config case (prior to rounding). We use them as
	// a reference to test rounding behavior when adding to positions.
	amount0PerfectRatio := sdk.NewInt(998977)
	amount1PerfectRatio := sdk.NewInt(5000000000)

	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it overwrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite      *lpTest
		timeElapsed             time.Duration
		createPositionOverwrite bool
		createLockLocked        bool
		createLockUnlocking     bool
		createLockUnlocked      bool
		lastPositionInPool      bool
		senderNotOwner          bool

		amount0ToAdd sdk.Int
		amount1ToAdd sdk.Int
	}{
		"add base amount to existing liquidity with perfect ratio": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: sdk.NewInt(9999998816),
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
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio.QuoRaw(2)),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: sdk.NewInt(7499995358),
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

		// Error catching

		"error: attempt to add to a position with underlying lock that has finished unlocking": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Add(amount0PerfectRatio).Sub(roundingError),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				amount1Expected: sdk.NewInt(9999998816),
				expectedError:   types.PositionSuperfluidStakedError{PositionId: uint64(1)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockUnlocked: true,
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
				amount1Expected: sdk.NewInt(9999998816),

				expectedError: types.PositionSuperfluidStakedError{PositionId: uint64(1)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockLocked: true,
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
				amount1Expected: sdk.NewInt(9999998816),

				expectedError: types.PositionSuperfluidStakedError{PositionId: uint64(1)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: amount0PerfectRatio,
			amount1ToAdd: amount1PerfectRatio,

			createLockUnlocking: true,
		},
		"error: final amount less than original amount": {
			setupConfig: baseCase,

			// system under test parameters
			sutConfigOverwrite: &lpTest{
				amount0Expected: amount0PerfectRatio.Sub(roundingError),
				// Since we round on the other the asset when we withdraw, asset0 turns into the bottleneck and
				// thus we cannot use the full amount of asset1. We calculate the below using the following formula and rounding up:
				// amount1 = L * (sqrtPriceUpper - sqrtPriceLower)
				// https://www.wolframalpha.com/input?i=3035764327.860030912175533748+*+%2870.710678118654752440+-+67.416615162732695594%29
				expectedError: types.InsufficientLiquidityCreatedError{Actual: sdk.NewInt(4999996906), Minimum: baseCase.tokensProvided.AmountOf(DefaultCoin1.Denom).Sub(roundingError)},
			},
			timeElapsed:  defaultTimeElapsed,
			amount0ToAdd: sdk.ZeroInt(),
			amount1ToAdd: sdk.ZeroInt(),
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
		"error: attempt to add negative asset0 to position": {
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
		"error: attempt to add negative asset1 to position": {
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
				tc                          = tc
				config                      = *tc.setupConfig
				sutConfigOverwrite          = *tc.sutConfigOverwrite
				err                         error
			)

			// If specific configs are provided in the test case, overwrite the config with those values.
			mergeConfigs(&config, &sutConfigOverwrite)

			// If a setupConfig is provided, use it to create a pool and position.
			pool := s.PrepareConcentratedPool()
			fundCoins := config.tokensProvided
			if tc.amount0ToAdd.IsPositive() && tc.amount1ToAdd.IsPositive() {
				fundCoins = fundCoins.Add(sdk.NewCoins(sdk.NewCoin(ETH, tc.amount0ToAdd), sdk.NewCoin(USDC, tc.amount1ToAdd))...)
			}
			s.FundAcc(owner, fundCoins)

			// Create a position from the parameters in the test case.
			var amount0Initial, amount1Initial sdk.Int
			if tc.createLockLocked {
				_, amount0Initial, amount1Initial, _, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed)
				s.Require().NoError(err)
			} else if tc.createLockUnlocking {
				_, amount0Initial, amount1Initial, _, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed+time.Hour)
				s.Require().NoError(err)
			} else if tc.createLockUnlocked {
				_, amount0Initial, amount1Initial, _, _, _, err = concentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), owner, fundCoins, tc.timeElapsed-time.Hour)
				s.Require().NoError(err)
			} else {
				_, amount0Initial, amount1Initial, _, _, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), owner, config.tokensProvided, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}
			preSendBalanceSender := s.App.BankKeeper.GetAllBalances(s.Ctx, owner)

			if !tc.lastPositionInPool {
				s.FundAcc(s.TestAccs[1], fundCoins)
				_, _, _, _, _, err = concentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[1], config.tokensProvided, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			sender := owner
			if tc.senderNotOwner {
				sender = invalidSender
			}

			// --- System under test ---
			newPosId, newAmt0, newAmt1, err := concentratedLiquidityKeeper.AddToPosition(s.Ctx, sender, config.positionId, tc.amount0ToAdd, tc.amount1ToAdd)
			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(sdk.Int{}, newAmt0)
				s.Require().Equal(sdk.Int{}, newAmt1)
				s.Require().Equal(uint64(0), newPosId)
				s.Require().ErrorContains(err, config.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), newAmt0.String())
			s.Require().Equal(config.amount1Expected.String(), newAmt1.String())

			// We expect the position ID to be 3 since we have two setup positions
			s.Require().Equal(uint64(3), newPosId)

			// Ensure balances were deducted by the correct amounts
			// Note that we subtract rounding error from the initial amount of
			// both assets since both are truncated upon withdrawal (so there is at least one
			// unit of each left in the pool).
			postSendBalanceSender := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)
			s.Require().Equal(
				sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), config.amount0Expected.Sub(amount0Initial.Sub(roundingError))), sdk.NewCoin(pool.GetToken1(), config.amount1Expected.Sub(amount1Initial.Sub(roundingError)))),
				preSendBalanceSender.Sub(postSendBalanceSender),
			)
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
		if !overwrite.expectedFeeGrowthOutsideLower.IsEqual(sdk.DecCoins{}) {
			dst.expectedFeeGrowthOutsideLower = overwrite.expectedFeeGrowthOutsideLower
		}
		if !overwrite.expectedFeeGrowthOutsideUpper.IsEqual(sdk.DecCoins{}) {
			dst.expectedFeeGrowthOutsideUpper = overwrite.expectedFeeGrowthOutsideUpper
		}
		if overwrite.positionId != 0 {
			dst.positionId = overwrite.positionId
		}
		if overwrite.underlyingLockId != 0 {
			dst.underlyingLockId = overwrite.underlyingLockId
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
			expectedErr: InsufficientFundsError,
		},
		"only asset1 is greater than sender has, position creation (user to pool)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(100000000000000)),
			expectedErr: InsufficientFundsError,
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
			expectedErr: InsufficientFundsError,
		},
		"only asset1 is greater than sender has, withdraw (pool to user)": {
			coin0:       sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1:       sdk.NewCoin("usdc", sdk.NewInt(100000000000000)),
			poolToUser:  true,
			expectedErr: InsufficientFundsError,
		},
		"asset0 is negative - error": {
			coin0: sdk.Coin{Denom: "eth", Amount: sdk.NewInt(1000000).Neg()},
			coin1: sdk.NewCoin("usdc", sdk.NewInt(1000000)),

			expectedErr: types.Amount0IsNegativeError{Amount0: sdk.NewInt(1000000).Neg()},
		},
		"asset1 is negative - error": {
			coin0: sdk.NewCoin("eth", sdk.NewInt(1000000)),
			coin1: sdk.Coin{Denom: "usdc", Amount: sdk.NewInt(1000000).Neg()},

			expectedErr: types.Amount1IsNegativeError{Amount1: sdk.NewInt(1000000).Neg()},
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
			concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
			if !ok {
				s.FailNow("poolI is not a ConcentratedPoolExtension")
			}

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
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErr.Error())
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
			expectedPositionLiquidity: sdk.ZeroDec(),
			expectedTickLiquidity:     sdk.ZeroDec(),
			expectedPoolLiquidity:     sdk.ZeroDec(),
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
			_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(
				s.Ctx,
				1,
				s.TestAccs[0],
				DefaultCoins,
				sdk.ZeroInt(), sdk.ZeroInt(),
				DefaultLowerTick, DefaultUpperTick,
			)
			s.Require().NoError(err)

			// explicitly make update time different to ensure that the pool is updated with last liqudity update.
			expectedUpdateTime := tc.joinTime.Add(time.Second)
			s.Ctx = s.Ctx.WithBlockTime(expectedUpdateTime)

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
			)

			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Equal(sdk.Int{}, actualAmount0)
				s.Require().Equal(sdk.Int{}, actualAmount1)
			} else {
				s.Require().NoError(err)

				var (
					expectedAmount0 sdk.Dec
					expectedAmount1 sdk.Dec
				)

				// For the context of this test case, we are not testing the calculation of the amounts
				// As a result, whenever non-default values are expected, we estimate them using the internal CalcActualAmounts function
				if tc.amount0Expected.IsNil() || tc.amount1Expected.IsNil() {
					pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.poolId)
					s.Require().NoError(err)

					expectedAmount0, expectedAmount1, err = pool.CalcActualAmounts(s.Ctx, tc.lowerTick, tc.upperTick, tc.liquidityDelta)
					s.Require().NoError(err)
				} else {
					expectedAmount0 = tc.amount0Expected.ToDec()
					expectedAmount1 = tc.amount1Expected.ToDec()
				}

				s.Require().Equal(expectedAmount0.TruncateInt().String(), actualAmount0.String())
				s.Require().Equal(expectedAmount1.TruncateInt().String(), actualAmount1.String())

				// validate if position has been properly updated
				s.validatePositionUpdate(s.Ctx, tc.positionId, tc.expectedPositionLiquidity)
				s.validateTickUpdates(s.Ctx, tc.poolId, s.TestAccs[tc.ownerIndex], tc.lowerTick, tc.upperTick, tc.expectedTickLiquidity, cl.EmptyCoins, cl.EmptyCoins)

				// validate if pool liquidity has been updated properly
				poolI, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
				if !ok {
					s.FailNow("poolI is not a ConcentratedPoolExtension")
				}
				s.Require().Equal(tc.expectedPoolLiquidity, concentratedPool.GetLiquidity())

				// Test that liquidity update time was succesfully changed.
				s.Require().Equal(expectedUpdateTime, poolI.GetLastLiquidityUpdate())
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitializeInitialPositionForPool() {
	sqrt := func(x int64) sdk.Dec {
		sqrt, err := sdk.NewDec(x).ApproxSqrt()
		s.Require().NoError(err)
		return sqrt
	}

	type sendTest struct {
		amount0Desired        sdk.Int
		amount1Desired        sdk.Int
		tickSpacing           uint64
		expectedCurrSqrtPrice sdk.Dec
		expectedTick          sdk.Int
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
			amount0Desired:        sdk.OneInt(),
			amount1Desired:        sdk.NewInt(100_000_050),
			tickSpacing:           DefaultTickSpacing,
			expectedCurrSqrtPrice: sqrt(100_000_050),
			expectedTick:          sdk.NewInt(72000000),
		},
		"100_000_051 and tick spacing 100, price level where curr sqrt price does not translate to allowed tick (assumes exponent at price one of -6 and tick spacing of 100)": {
			amount0Desired:        sdk.OneInt(),
			amount1Desired:        sdk.NewInt(100_000_051),
			tickSpacing:           DefaultTickSpacing,
			expectedCurrSqrtPrice: sqrt(100_000_051),
			expectedTick:          sdk.NewInt(72000000),
		},
		"100_000_051 and tick spacing 1, price level where curr sqrt price translates to allowed tick (assumes exponent at price one of -6 and tick spacing of 1)": {
			amount0Desired:        sdk.OneInt(),
			amount1Desired:        sdk.NewInt(100_000_051),
			tickSpacing:           1,
			expectedCurrSqrtPrice: sqrt(100_000_051),
			expectedTick:          sdk.NewInt(72000001),
		},
		"error: amount0Desired is zero": {
			amount0Desired: sdk.ZeroInt(),
			amount1Desired: DefaultAmt1,
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: DefaultAmt1},
		},
		"error: amount1Desired is zero": {
			amount0Desired: DefaultAmt0,
			amount1Desired: sdk.ZeroInt(),
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: DefaultAmt0, Amount1: sdk.ZeroInt()},
		},
		"error: both amount0Desired and amount01Desired is zero": {
			amount0Desired: sdk.ZeroInt(),
			amount1Desired: sdk.ZeroInt(),
			tickSpacing:    DefaultTickSpacing,
			expectedError:  types.InitialLiquidityZeroError{Amount0: sdk.ZeroInt(), Amount1: sdk.ZeroInt()},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// create a CL pool
			pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, sdk.ZeroDec())

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
				s.Require().Equal(tc.expectedTick.String(), pool.GetCurrentTick().String())
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
			poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, tc.tickSpacing, sdk.ZeroDec()))
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
				s.SetupPosition(1, s.TestAccs[1], DefaultCoins, tc.lowerTick, tc.upperTick, DefaultJoinTime)
			}

			// Fund test account and create the desired position
			s.FundAcc(s.TestAccs[0], DefaultCoins)

			// Note user and pool account balances before create position is called
			userBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePrePositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// System under test.
			positionId, amtDenom0CreatePosition, amtDenom1CreatePosition, liquidityCreated, _, err := clKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.tokensProvided, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)

			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime.Add(time.Hour * 24))
			amtDenom0WithdrawPosition, amtDenom1WithdrawPosition, err := clKeeper.WithdrawPosition(s.Ctx, s.TestAccs[0], positionId, liquidityCreated)
			s.Require().NoError(err)

			// INVARIANTS

			// 1. amount for denom0 and denom1 upon creating and withdraw position should be same
			// Note: subtracting one because create position rounds in favor of the pool.
			s.Require().Equal(amtDenom0CreatePosition.Sub(sdk.OneInt()).String(), amtDenom0WithdrawPosition.String())
			s.Require().Equal(amtDenom1CreatePosition.Sub(sdk.OneInt()).String(), amtDenom1WithdrawPosition.String())

			// 2. user balance and pool balance after creating / withdrawing position should be same
			userBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalancePostPositionCreation := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// Note: subtracting one since position creation rounds in favor of the pool.
			s.Require().Equal(userBalancePrePositionCreation.AmountOf(ETH).Sub(sdk.OneInt()).String(), userBalancePostPositionCreation.AmountOf(ETH).String())
			s.Require().Equal(userBalancePrePositionCreation.AmountOf(USDC).Sub(sdk.OneInt()).String(), userBalancePostPositionCreation.AmountOf(USDC).String())

			// Note: adding one since withdrawal rounds in favor of the pool.
			s.Require().Equal(poolBalancePrePositionCreation.AmountOf(ETH).Add(roundingError).String(), poolBalancePostPositionCreation.AmountOf(ETH).String())
			s.Require().Equal(poolBalancePrePositionCreation.AmountOf(USDC).Add(roundingError).String(), poolBalancePostPositionCreation.AmountOf(USDC).String())

			// 3. Check that position's liquidity was deleted
			positionLiquidity, err := clKeeper.GetPositionLiquidity(s.Ctx, tc.positionId)
			s.Require().Error(err)
			s.Require().ErrorAs(err, &types.PositionIdNotFoundError{PositionId: tc.positionId})
			s.Require().Equal(sdk.Dec{}, positionLiquidity)

			// 4. Check that pool has come back to original state

			liquidityAfter, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, poolID)
			s.Require().NoError(err)

			s.Require().NoError(err)

			// Note: one ends up remaining due to rounding in favor of the pool.
			s.Require().Equal(liquidityBefore.AmountOf(ETH).Add(roundingError).String(), liquidityAfter.AmountOf(ETH).String())
			s.Require().Equal(liquidityBefore.AmountOf(USDC).Add(roundingError).String(), liquidityAfter.AmountOf(USDC).String())
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
				positionId         uint64
				concentratedLockId uint64
				err                error
			)
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// create a CL pool and fund account
			pool := s.PrepareConcentratedPool()
			coinsToFund := sdk.NewCoins(DefaultCoin0, DefaultCoin1)
			s.FundAcc(s.TestAccs[0], coinsToFund)

			if tc.unlockingPosition {
				positionId, _, _, _, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund, tc.remainingLockDuration)
				s.Require().NoError(err)
			} else if tc.lockedPosition {
				positionId, _, _, _, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund, tc.remainingLockDuration)
				s.Require().NoError(err)
			} else {
				positionId, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coinsToFund)
				s.Require().NoError(err)
			}

			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
			s.Require().NoError(err)

			// Increment block time by a second to ensure test cases with zero lock duration are in the past
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))

			// System under test
			lockIsMature, _ := s.App.ConcentratedLiquidityKeeper.IsLockMature(s.Ctx, concentratedLockId)

			if tc.expectedLockIsMature {
				s.Require().True(lockIsMature)
			} else {
				s.Require().False(lockIsMature)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidatePositionUpdateById() {
	tests := map[string]struct {
		positionId           uint64
		updateInitiatorIndex int
		lowerTickGiven       int64
		upperTickGiven       int64
		liquidityDeltaGiven  sdk.Dec
		joinTimeGiven        time.Time
		poolIdGiven          uint64
		expectError          error
	}{
		"valid update - adding liquidity": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"valid update - removing liquidity": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  sdk.OneDec().Neg(), // negative
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"valid update - does not exist yet": {
			positionId:           DefaultPositionId + 2,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
		},
		"error: attempted to remove too much liquidty": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  DefaultLiquidityAmt.Add(sdk.OneDec()).Neg(),
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
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.PositionOwnerMismatchError{},
		},
		"error: lower tick mismatch": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick + 1,
			upperTickGiven:       DefaultUpperTick,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: upper tick mismatch": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: invalid join time": {
			positionId:           DefaultPositionId,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.LowerTickMismatchError{},
		},
		"error: pool id mismatch": {
			positionId:           DefaultPositionId + 1,
			updateInitiatorIndex: 0,
			lowerTickGiven:       DefaultLowerTick,
			upperTickGiven:       DefaultUpperTick + 1,
			liquidityDeltaGiven:  sdk.OneDec(),
			joinTimeGiven:        DefaultJoinTime,
			poolIdGiven:          defaultPoolId,
			expectError:          types.PositionsNotInSamePoolError{},
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
