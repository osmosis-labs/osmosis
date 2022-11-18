package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestCreatePosition() {
	denom0 := "eth"
	denom1 := "usdc"

	tests := map[string]struct {
		poolId            uint64
		currentTick       sdk.Int
		lowerTick         int64
		upperTick         int64
		currentSqrtP      sdk.Dec
		amount0Desired    sdk.Int
		amount0Expected   sdk.Int
		amount1Desired    sdk.Int
		amount1Expected   sdk.Int
		expectedLiquidity sdk.Dec
	}{
		"happy path": {
			poolId:            1,
			currentTick:       sdk.NewInt(85176),
			lowerTick:         int64(84222),
			upperTick:         int64(86129),
			currentSqrtP:      sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			amount0Desired:    sdk.NewInt(1000000),                            // 1 eth
			amount0Expected:   sdk.NewInt(998587),                             // 0.998587 eth
			amount1Desired:    sdk.NewInt(5000000000),                         // 5000 usdc
			amount1Expected:   sdk.NewInt(5000000000),                         // 5000 usdc
			expectedLiquidity: sdk.MustNewDecFromStr("1517818840.967515822610790519"),
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
			s.Require().Equal(tc.expectedLiquidity.String(), liquidityCreated.String())

			// check position state
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedLiquidity.String(), position.Liquidity.String())

			// check tick state
			lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, tc.poolId, tc.lowerTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedLiquidity.String(), lowerTickInfo.LiquidityGross.String())
			s.Require().Equal(tc.expectedLiquidity.String(), lowerTickInfo.LiquidityNet.String())

			upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, tc.poolId, tc.upperTick)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedLiquidity.String(), upperTickInfo.LiquidityGross.String())
			s.Require().Equal(tc.expectedLiquidity.Neg().String(), upperTickInfo.LiquidityNet.String())
		})

	}
}

func (s *KeeperTestSuite) TestWithdrawPosition() {

	tests := map[string]struct {
		poolId       uint64
		denom0       string
		denom1       string
		currentSqrtP sdk.Dec
		currentTick  sdk.Int

		amount0Desired sdk.Int
		amount1Desired sdk.Int

		owner           sdk.AccAddress
		lowerTick       int64
		upperTick       int64
		liquidityAmount sdk.Dec

		expectedAmount0 sdk.Int
		expectedAmount1 sdk.Int
		expectError     bool
	}{
		"basic test for active position": {
			poolId: 1,
			denom0: "eth",
			denom1: "usdc",

			currentSqrtP: sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			currentTick:  sdk.NewInt(85176),

			amount0Desired: sdk.NewInt(1000000),    // 1 eth,
			amount1Desired: sdk.NewInt(5000000000), // 5000 usdc

			lowerTick: int64(84222),
			upperTick: int64(86129),

			liquidityAmount: sdk.MustNewDecFromStr("1517818840.967515822610790519"),

			expectedAmount0: sdk.NewInt(998587),     // 0.998587 eth
			expectedAmount1: sdk.NewInt(5000000000), // 5000 usdc
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
			s.SetupTest()
			ctx := s.Ctx
			concentratedLiquidityKeeper := s.App.ConcentratedLiquidityKeeper

			// Setup.
			_, err := concentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, tc.poolId, tc.denom0, tc.denom1, tc.currentSqrtP, tc.currentTick)
			s.Require().NoError(err)

			_, _, _, err = concentratedLiquidityKeeper.CreatePosition(ctx, tc.poolId, s.TestAccs[0], tc.amount0Desired, tc.amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), tc.lowerTick, tc.upperTick)
			s.Require().NoError(err)

			// System under test.
			amtDenom0, amtDenom1, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, tc.poolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, tc.liquidityAmount)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedAmount0.String(), amtDenom0.String())
			s.Require().Equal(tc.expectedAmount1.String(), amtDenom1.String())
		})
	}
}
