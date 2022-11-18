package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type lpTest struct {
	poolId          uint64
	owner           sdk.AccAddress
	currentTick     sdk.Int
	lowerTick       int64
	upperTick       int64
	currentSqrtP    sdk.Dec
	amount0Desired  sdk.Int
	amount0Expected sdk.Int
	amount1Desired  sdk.Int
	amount1Expected sdk.Int
	liquidityAmount sdk.Dec
	expectedError   error
}

var (
	denom0 = "eth"
	denom1 = "usdc"
)

func (s *KeeperTestSuite) TestCreatePosition() {

	tests := map[string]lpTest{
		"happy path": {
			poolId:          1,
			currentTick:     sdk.NewInt(85176),
			lowerTick:       int64(84222),
			upperTick:       int64(86129),
			currentSqrtP:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			amount0Desired:  sdk.NewInt(1000000),                            // 1 eth
			amount0Expected: sdk.NewInt(998587),                             // 0.998587 eth
			amount1Desired:  sdk.NewInt(5000000000),                         // 5000 usdc
			amount1Expected: sdk.NewInt(5000000000),                         // 5000 usdc
			liquidityAmount: sdk.MustNewDecFromStr("1517818840.967515822610790519"),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, tc.poolId, denom0, denom1, tc.currentSqrtP, tc.currentTick)

			asset0, asset1, liquidityCreated, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.amount0Desired, tc.amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.amount0Expected.String(), asset0.String())
			s.Require().Equal(tc.amount1Expected.String(), asset1.String())
			s.Require().Equal(tc.liquidityAmount.String(), liquidityCreated.String())

			// check position state
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.liquidityAmount.String(), position.Liquidity.String())

			// check tick state
			lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, tc.poolId, tc.lowerTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.liquidityAmount.String(), lowerTickInfo.LiquidityGross.String())
			s.Require().Equal(tc.liquidityAmount.String(), lowerTickInfo.LiquidityNet.String())

			upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, tc.poolId, tc.upperTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.liquidityAmount.String(), upperTickInfo.LiquidityGross.String())
			s.Require().Equal(tc.liquidityAmount.Neg().String(), upperTickInfo.LiquidityNet.String())
		})

	}
}

func (s *KeeperTestSuite) TestWithdrawPosition() {

	// mergeConfigs merges every desired non-zero field from overwrite
	// into dst. dst is mutated due to being a pointer.
	mergeConfigs := func(dst *lpTest, overwrite *lpTest) {
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
		}
	}

	tests := map[string]struct {
		setupConfig lpTest
		// when this is set, it ovewrites the setupConfig
		// and gives the overwritten configuration to
		// the system under test.
		sutConfigOverwrite *lpTest
	}{
		"basic test for active position": {
			// setup parameters for creation a pool and position.
			setupConfig: lpTest{
				poolId:          1,
				currentSqrtP:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
				currentTick:     sdk.NewInt(85176),
				lowerTick:       int64(84222),
				upperTick:       int64(86129),
				amount0Desired:  sdk.NewInt(1000000),    // 1 eth,
				amount1Desired:  sdk.NewInt(5000000000), // 5000 usdc
				liquidityAmount: sdk.MustNewDecFromStr("1517818840.967515822610790519"),
			},

			// system under test parameters
			// for withdrawing a position.
			sutConfigOverwrite: &lpTest{
				amount0Expected: sdk.NewInt(998587),     // 0.998587 eth
				amount1Expected: sdk.NewInt(5000000000), // 5000 usdc
			},
		},
		// no position created
		// invalid pool id
		// invalid ticks
		// liquidityAmount higher than available
		// full liquidity amount
		// liquidity amount lower than available
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()
			ctx := s.Ctx
			concentratedLiquidityKeeper := s.App.ConcentratedLiquidityKeeper

			owner := s.TestAccs[0]

			// Setup.
			_, err := concentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, tc.setupConfig.poolId, denom0, denom1, tc.setupConfig.currentSqrtP, tc.setupConfig.currentTick)
			s.Require().NoError(err)

			_, _, _, err = concentratedLiquidityKeeper.CreatePosition(ctx, tc.setupConfig.poolId, owner, tc.setupConfig.amount0Desired, tc.setupConfig.amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), tc.setupConfig.lowerTick, tc.setupConfig.upperTick)
			s.Require().NoError(err)

			config := tc.setupConfig
			mergeConfigs(&config, tc.sutConfigOverwrite)

			// System under test.
			amtDenom0, amtDenom1, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, config.poolId, owner, config.lowerTick, config.upperTick, config.liquidityAmount)

			if config.expectedError != nil {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(config.amount0Expected.String(), amtDenom0.String())
			s.Require().Equal(config.amount1Expected.String(), amtDenom1.String())
		})
	}
}
