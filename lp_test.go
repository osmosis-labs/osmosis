package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

type lpTest struct {
	poolId          uint64
	owner           sdk.AccAddress
	currentTick     sdk.Int
	lowerTick       int64
	upperTick       int64
	currentSqrtP    sdk.Dec
	amount0Desired  sdk.Int
	amount0Minimum  sdk.Int
	amount0Expected sdk.Int
	amount1Desired  sdk.Int
	amount1Minimum  sdk.Int
	amount1Expected sdk.Int
	liquidityAmount sdk.Dec
	expectedError   error
}

var (
	denom0   = "eth"
	denom1   = "usdc"
	baseCase = &lpTest{
		poolId:          1,
		currentTick:     sdk.NewInt(85176),
		lowerTick:       int64(84222),
		upperTick:       int64(86129),
		currentSqrtP:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
		amount0Desired:  sdk.NewInt(1000000),                            // 1 eth
		amount0Minimum:  sdk.ZeroInt(),
		amount0Expected: sdk.NewInt(998587),     // 0.998587 eth
		amount1Desired:  sdk.NewInt(5000000000), // 5000 usdc
		amount1Minimum:  sdk.ZeroInt(),
		amount1Expected: sdk.NewInt(5000000000), // 5000 usdc
		liquidityAmount: sdk.MustNewDecFromStr("1517818840.967515822610790519"),
	}
)

func (s *KeeperTestSuite) TestCreatePosition() {
	tests := map[string]lpTest{
		// "base case": *baseCase,
		"amount of token 0 is smaller than minimum; should fail and not update state": {
			amount0Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)),
			expectedError:  types.InsufficientLiquidityCreatedError{Actual: baseCase.amount0Expected, Minimum: baseCase.amount0Expected.Mul(sdk.NewInt(2)), IsTokenZero: true},
		},
		"amount of token 1 is smaller than minimum; should fail and not update state": {
			amount1Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2)),
			expectedError:  types.InsufficientLiquidityCreatedError{Actual: baseCase.amount1Expected, Minimum: baseCase.amount1Expected.Mul(sdk.NewInt(2))},
		},
		"error: invalid tick": {
			lowerTick:     types.MaxTick + 1,
			expectedError: types.InvalidTickError{Tick: types.MaxTick + 1, IsLower: true},
		},
		// TODO: add more tests
		// - custom hand-picked values
		// - error edge cases
		// - think of overflows
		// - think of truncations
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			// Merge tc with baseCase and update tc
			// to the merged result. This is done
			// to reduce the amount of boilerplate
			// in test cases.
			baseConfigCopy := *baseCase
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			_, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, tc.poolId, denom0, denom1, tc.currentSqrtP, tc.currentTick)
			s.Require().NoError(err)

			asset0, asset1, liquidityCreated, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.amount0Desired, tc.amount1Desired, tc.amount0Minimum, tc.amount1Minimum, tc.lowerTick, tc.upperTick)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(asset0, sdk.Int{})
				s.Require().Equal(asset1, sdk.Int{})
				s.Require().ErrorAs(err, &tc.expectedError)

				// make sure that position is not created
				_, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick)

				s.Require().Error(err)
				s.Require().ErrorAs(err, &types.PositionNotFoundError{PoolId: tc.poolId, LowerTick: tc.lowerTick, UpperTick: tc.upperTick})
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.amount0Expected.String(), asset0.String())
			s.Require().Equal(tc.amount1Expected.String(), asset1.String())
			s.Require().Equal(tc.liquidityAmount.String(), liquidityCreated.String())

			// check position state
			s.validatePositionUpdate(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, tc.liquidityAmount)

			// check tick state
			s.validateTickUpdates(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, tc.liquidityAmount)
		})
	}
}

func (s *KeeperTestSuite) TestWithdrawPosition() {
	tests := map[string]struct {
		setupConfig *lpTest
		// when this is set, it ovewrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite *lpTest
	}{
		"base case: withdraw full liquidity amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				amount0Expected: baseCase.amount0Expected, // 0.998587 eth
				amount1Expected: baseCase.amount1Expected, // 5000 usdc
			},
		},
		"withdraw partial liquidity amount": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.QuoInt64(2),

				amount0Expected: baseCase.amount0Expected.QuoRaw(2),                   // 0.4992935 / 2 eth
				amount1Expected: baseCase.amount1Expected.QuoRaw(2).Sub(sdk.OneInt()), // 2499 usdc, one is lost due to truncation
			},
		},
		"error: no position created": {
			// setup parameters for creation a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				lowerTick:     -1, // valid tick at which no position exists
				expectedError: types.PositionNotFoundError{PoolId: 1, LowerTick: -1, UpperTick: 86129},
			},
		},
		"error: pool id for pool that does not exist": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				poolId:        2, // does not exist
				expectedError: types.PoolNotFoundError{PoolId: 2},
			},
		},
		"error: invalid tick given": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				lowerTick:     types.MaxTick + 1, // invalid tick
				expectedError: types.InvalidTickError{Tick: types.MaxTick + 1, IsLower: true},
			},
		},
		"error: insufficient liqudity": {
			// setup parameters for creating a pool and position.
			setupConfig: baseCase,

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				liquidityAmount: baseCase.liquidityAmount.Add(sdk.OneDec()), // 1 more than available
				expectedError:   types.InsufficientLiquidityError{Actual: baseCase.liquidityAmount.Add(sdk.OneDec()), Available: baseCase.liquidityAmount},
			},
		},
		// TODO: test with custom amounts that potentially lead to truncations.
	}

	for name, tc := range tests {
		s.Run(name, func() {
			var (
				tc                 = tc
				config             = *tc.setupConfig
				sutConfigOverwrite = *tc.sutConfigOverwrite
			)

			s.SetupTest()

			var (
				ctx                         = s.Ctx
				concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
				liquidityCreated            = sdk.ZeroDec()
				owner                       = s.TestAccs[0]
			)

			// Setup.
			if tc.setupConfig != nil {
				_, err := concentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, config.poolId, denom0, denom1, config.currentSqrtP, config.currentTick)
				s.Require().NoError(err)

				_, _, liquidityCreated, err = concentratedLiquidityKeeper.CreatePosition(ctx, config.poolId, owner, config.amount0Desired, config.amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), config.lowerTick, config.upperTick)
				s.Require().NoError(err)
			}

			mergeConfigs(&config, &sutConfigOverwrite)

			expectedRemainingLiquidity := liquidityCreated.Sub(config.liquidityAmount)

			// System under test.
			amtDenom0, amtDenom1, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, config.poolId, owner, config.lowerTick, config.upperTick, config.liquidityAmount)

			if config.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(amtDenom0, sdk.Int{})
				s.Require().Equal(amtDenom1, sdk.Int{})
				s.Require().ErrorAs(err, &config.expectedError)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), amtDenom0.String())
			s.Require().Equal(config.amount1Expected.String(), amtDenom1.String())

			// Check that the position was updated.
			s.validatePositionUpdate(ctx, config.poolId, owner, config.lowerTick, config.upperTick, expectedRemainingLiquidity)

			// check tick state
			s.validateTickUpdates(ctx, config.poolId, owner, config.lowerTick, config.upperTick, expectedRemainingLiquidity)
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
		if !overwrite.amount0Expected.IsNil() {
			dst.amount0Expected = overwrite.amount0Expected
		}
		if !overwrite.amount1Expected.IsNil() {
			dst.amount1Expected = overwrite.amount1Expected
		}
		if overwrite.expectedError != nil {
			dst.expectedError = overwrite.expectedError
		}
		if !overwrite.amount0Minimum.IsNil() {
			dst.amount0Minimum = overwrite.amount0Minimum
		}
		if !overwrite.amount1Minimum.IsNil() {
			dst.amount1Minimum = overwrite.amount1Minimum
		}
	}
}
