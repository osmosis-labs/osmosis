package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (s *KeeperTestSuite) TestOrderInitialPoolDenoms() {
	denom0, denom1, err := types.OrderInitialPoolDenoms("axel", "osmo")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "axel")
	s.Require().Equal(denom1, "osmo")

	denom0, denom1, err = types.OrderInitialPoolDenoms("usdc", "eth")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "eth")
	s.Require().Equal(denom1, "usdc")

	denom0, denom1, err = types.OrderInitialPoolDenoms("usdc", "usdc")
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestInitializePool() {
	// Create a valid PoolI from a valid ConcentratedPoolExtension
	validConcentratedPool := s.PrepareConcentratedPool()
	validPoolI := validConcentratedPool.(poolmanagertypes.PoolI)

	// Create a concentrated liquidity pool with invalid tick spacing
	invalidTickSpacing := uint64(25)
	invalidTickSpacingConcentratedPool, err := clmodel.NewConcentratedLiquidityPool(2, ETH, USDC, invalidTickSpacing, DefaultExponentAtPriceOne, DefaultZeroSwapFee)

	// Create a concentrated liquidity pool with invalid swap fee
	invalidSwapFee := sdk.MustNewDecFromStr("0.1")
	invalidSwapFeeConcentratedPool, err := clmodel.NewConcentratedLiquidityPool(3, ETH, USDC, DefaultTickSpacing, DefaultExponentAtPriceOne, invalidSwapFee)
	s.Require().NoError(err)

	// Create an invalid PoolI that doesn't implement ConcentratedPoolExtension
	var invalidPoolI poolmanagertypes.PoolI

	validCreatorAddress := sdk.AccAddress([]byte("addr1---------------"))

	tests := []struct {
		name           string
		poolI          poolmanagertypes.PoolI
		creatorAddress sdk.AccAddress
		expectedErr    error
	}{
		{
			name:           "Happy path",
			poolI:          validPoolI,
			creatorAddress: validCreatorAddress,
		},
		{
			name:           "Wrong pool type: empty pool interface that doesn't implement ConcentratedPoolExtension",
			poolI:          invalidPoolI,
			creatorAddress: validCreatorAddress,
			expectedErr:    fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", invalidPoolI),
		},
		{
			name:           "Invalid tick spacing",
			poolI:          &invalidTickSpacingConcentratedPool,
			creatorAddress: validCreatorAddress,
			expectedErr:    fmt.Errorf("invalid tick spacing. Got %d", invalidTickSpacing),
		},
		{
			name:           "Invalid swap fee",
			poolI:          &invalidSwapFeeConcentratedPool,
			creatorAddress: validCreatorAddress,
			expectedErr:    fmt.Errorf("invalid swap fee. Got %d", invalidSwapFee),
		},
		// We cannot test
		// We don't check creator address because we don't mint anything when making concentrated liquidity pools

	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Method under test.
			err := s.App.ConcentratedLiquidityKeeper.InitializePool(s.Ctx, test.poolI, test.creatorAddress)

			if test.expectedErr == nil {
				// Ensure no error is returned
				s.Require().NoError(err)

				// ensure that fee accumulator has been properly initialized
				feeAccumulator, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, test.poolI.GetId())
				s.Require().NoError(err)
				s.Require().Equal(sdk.DecCoins(nil), feeAccumulator.GetValue())

				// Ensure that uptime accumulators have been properly initialized
				uptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, test.poolI.GetId())
				s.Require().NoError(err)
				s.Require().Equal(len(types.SupportedUptimes), len(uptimeAccumulators))
				for _, uptimeAccumulator := range uptimeAccumulators {
					s.Require().Equal(cl.EmptyCoins, uptimeAccumulator.GetValue())
				}
			} else {
				// Ensure specified error is returned
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetPoolById() {
	tests := []struct {
		name        string
		poolId      uint64
		expectedErr error
	}{
		{
			name:   "Get existing pool",
			poolId: validPoolId,
		},
		{
			name:        "Get non-existing pool",
			poolId:      2,
			expectedErr: types.PoolNotFoundError{PoolId: 2},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Create default CL pool
			pool := s.PrepareConcentratedPool()

			// Get pool defined in test case
			getPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, test.poolId)

			if test.expectedErr == nil {
				// Ensure no error is returned
				s.Require().NoError(err)

				// Ensure that pool returned matches the default pool attributes
				s.Require().Equal(pool.GetId(), getPool.GetId())
				s.Require().Equal(pool.GetAddress(), getPool.GetAddress())
				s.Require().Equal(pool.GetCurrentSqrtPrice(), getPool.GetCurrentSqrtPrice())
				s.Require().Equal(pool.GetCurrentTick(), getPool.GetCurrentTick())
				s.Require().Equal(pool.GetLiquidity(), getPool.GetLiquidity())
			} else {
				// Ensure specified error is returned
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)

				// Check that GetPoolById returns a nil pool object due to error
				s.Require().Nil(getPool)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPoolExists() {
	s.SetupTest()

	// Create default CL pool
	pool := s.PrepareConcentratedPool()

	// Check that the pool exists
	poolExists := s.App.ConcentratedLiquidityKeeper.PoolExists(s.Ctx, pool.GetId())
	s.Require().True(poolExists)

	// try checking for a non-existent pool
	poolExists = s.App.ConcentratedLiquidityKeeper.PoolExists(s.Ctx, 2)

	// ensure that this returns false
	s.Require().False(poolExists)
}

func (s *KeeperTestSuite) TestConvertConcentratedToPoolInterface() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()

	// Ensure no error occurs when converting to PoolInterface
	_, err := cl.ConvertConcentratedToPoolInterface(concentratedPool)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestPoolIToConcentratedPool() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()
	poolI := concentratedPool.(poolmanagertypes.PoolI)

	// Ensure no error occurs when converting to ConcentratedPool
	_, err := cl.ConvertPoolInterfaceToConcentrated(poolI)
	s.Require().NoError(err)

	// Create a default stableswap pool
	stableswapPoolID := s.PrepareBasicStableswapPool()
	stableswapPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, stableswapPoolID)
	s.Require().NoError(err)

	// Ensure error occurs when converting to ConcentratedPool
	_, err = cl.ConvertPoolInterfaceToConcentrated(stableswapPool)
	s.Require().Error(err)
	s.Require().ErrorContains(err, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", stableswapPool).Error())
}

func (s *KeeperTestSuite) TestGetPoolDenoms() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()

	// Get denoms from pool
	denoms, err := s.App.ConcentratedLiquidityKeeper.GetPoolDenoms(s.Ctx, concentratedPool.GetId())
	s.Require().NoError(err)

	// Ensure denoms match
	s.Require().Equal([]string{ETH, USDC}, denoms)

	// try getting denoms from a non-existent pool
	_, err = s.App.ConcentratedLiquidityKeeper.GetPoolDenoms(s.Ctx, 2)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestCalculateSpotPrice() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()

	// should error when price is zero
	spotPrice, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, concentratedPool.GetId(), ETH, USDC)
	s.Require().Error(err)
	s.Require().Equal(sdk.Dec{}, spotPrice)

	// set up default position to have proper spot price
	s.SetupDefaultPosition(defaultPoolId)

	spotPriceBaseUSDC, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, concentratedPool.GetId(), ETH, USDC)
	s.Require().NoError(err)
	s.Require().Equal(spotPriceBaseUSDC, DefaultCurrSqrtPrice.Power(2))

	// test that we have correct values for reversed quote asset and base asset
	spotPriceBaseETH, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, concentratedPool.GetId(), USDC, ETH)
	s.Require().NoError(err)
	s.Require().Equal(spotPriceBaseETH, sdk.OneDec().Quo(DefaultCurrSqrtPrice.Power(2)))

	// try getting spot price from a non-existent pool
	spotPrice, err = s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, concentratedPool.GetId()+1, USDC, ETH)
	s.Require().Error(err)
	s.Require().True(spotPrice.IsNil())
}

func (s *KeeperTestSuite) TestValidateSwapFee() {
	tests := []struct {
		name        string
		swapFee     sdk.Dec
		expectValid bool
	}{
		{
			name:        "Valid swap fee",
			swapFee:     sdk.MustNewDecFromStr("0.003"),
			expectValid: true,
		},
		{
			name:        "Invalid swap fee",
			swapFee:     sdk.MustNewDecFromStr("0.5"),
			expectValid: false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Method under test.
			params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
			isValid := s.App.ConcentratedLiquidityKeeper.ValidateSwapFee(s.Ctx, params, test.swapFee)

			if test.expectValid {
				s.Require().True(isValid)
			} else {
				s.Require().False(isValid)
			}
		})
	}
}
