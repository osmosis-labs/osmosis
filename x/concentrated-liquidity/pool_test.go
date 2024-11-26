package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	sftypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (s *KeeperTestSuite) TestInitializePool() {
	// Create a valid PoolI from a valid ConcentratedPoolExtension
	validConcentratedPool := s.PrepareConcentratedPool()
	validPoolI, ok := validConcentratedPool.(poolmanagertypes.PoolI)
	s.Require().True(ok)

	// Create a concentrated liquidity pool with unauthorized tick spacing
	invalidTickSpacing := uint64(25)
	invalidTickSpacingConcentratedPool, err := clmodel.NewConcentratedLiquidityPool(2, ETH, USDC, invalidTickSpacing, DefaultZeroSpreadFactor)
	s.Require().NoError(err)

	// Create a concentrated liquidity pool with unauthorized spread factor
	invalidSpreadFactor := osmomath.MustNewDecFromStr("0.1")
	invalidSpreadFactorConcentratedPool, err := clmodel.NewConcentratedLiquidityPool(3, ETH, USDC, DefaultTickSpacing, invalidSpreadFactor)
	s.Require().NoError(err)

	// Create an invalid PoolI that doesn't implement ConcentratedPoolExtension
	invalidPoolId := s.PrepareBalancerPool()
	invalidPoolI, err := s.App.GAMMKeeper.GetPool(s.Ctx, invalidPoolId)
	s.Require().NoError(err)

	validCreatorAddress := s.TestAccs[0]

	poolmanagerModuleAccount := s.App.AccountKeeper.GetModuleAccount(s.Ctx, poolmanagertypes.ModuleName).GetAddress()

	tests := []struct {
		name                               string
		poolI                              poolmanagertypes.PoolI
		authorizedDenomsOverwrite          []string
		unrestrictedPoolCreatorWhitelist   []string
		permissionlessPoolCreationDisabled bool
		creatorAddress                     sdk.AccAddress
		expectedErr                        error
	}{
		{
			name:           "Happy path",
			poolI:          validPoolI,
			creatorAddress: validCreatorAddress,
		},
		{
			name:                               "Permissionless pool creation disabled",
			poolI:                              validPoolI,
			permissionlessPoolCreationDisabled: true,
			creatorAddress:                     validCreatorAddress,
			expectedErr:                        types.ErrPermissionlessPoolCreationDisabled,
		},
		{
			name:                               "bypass disabled permissionless pool creation check because poolmanager module account",
			poolI:                              validPoolI,
			permissionlessPoolCreationDisabled: true,
			creatorAddress:                     poolmanagerModuleAccount,
		},
		{
			name:                               "bypass disabled permissionless pool creation check because of whitelisted bypass",
			poolI:                              validPoolI,
			permissionlessPoolCreationDisabled: true,
			creatorAddress:                     validCreatorAddress,
			unrestrictedPoolCreatorWhitelist:   []string{validCreatorAddress.String()},
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
			expectedErr:    types.UnauthorizedTickSpacingError{ProvidedTickSpacing: invalidTickSpacing, AuthorizedTickSpacings: s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx).AuthorizedTickSpacing},
		},
		{
			name:           "Invalid spread factor",
			poolI:          &invalidSpreadFactorConcentratedPool,
			creatorAddress: validCreatorAddress,
			expectedErr:    types.UnauthorizedSpreadFactorError{ProvidedSpreadFactor: invalidSpreadFactor, AuthorizedSpreadFactors: s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx).AuthorizedSpreadFactors},
		},
		{
			name:  "unauthorized quote denom",
			poolI: validPoolI,
			// this flag overwrites the default authorized quote denoms
			// so that the test case fails.
			authorizedDenomsOverwrite: []string{"otherDenom"},
			creatorAddress:            validCreatorAddress,
			expectedErr:               types.UnauthorizedQuoteDenomError{ProvidedQuoteDenom: USDC, AuthorizedQuoteDenoms: []string{"otherDenom"}},
		},
		{
			name:                      "bypass unauthorized quote denom check because poolmanager module account",
			poolI:                     validPoolI,
			authorizedDenomsOverwrite: []string{"otherDenom"},
			// despite the quote denom not being authorized, will still
			// pass because its coming from the poolmanager module account
			creatorAddress: poolmanagerModuleAccount,
		},
		{
			name:                      "bypass unauthorized quote denom check because of whitelisted bypass",
			poolI:                     validPoolI,
			authorizedDenomsOverwrite: []string{"otherDenom"},
			// despite the quote denom not being authorized, will still
			// pass because its coming from a whitelisted pool creator
			unrestrictedPoolCreatorWhitelist: []string{validCreatorAddress.String()},
			creatorAddress:                   validCreatorAddress,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			if test.permissionlessPoolCreationDisabled {
				params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				params.IsPermissionlessPoolCreationEnabled = false
				s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)
			}

			if len(test.authorizedDenomsOverwrite) > 0 {
				params := s.App.PoolManagerKeeper.GetParams(s.Ctx)
				params.AuthorizedQuoteDenoms = test.authorizedDenomsOverwrite
				s.App.PoolManagerKeeper.SetParams(s.Ctx, params)
			}

			if len(test.unrestrictedPoolCreatorWhitelist) > 0 {
				params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				params.UnrestrictedPoolCreatorWhitelist = test.unrestrictedPoolCreatorWhitelist
				s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)
			}

			s.setListenerMockOnConcentratedLiquidityKeeper()

			// Method under test.
			err := s.App.ConcentratedLiquidityKeeper.InitializePool(s.Ctx, test.poolI, test.creatorAddress)

			if test.expectedErr == nil {
				// Ensure no error is returned
				s.Require().NoError(err)

				// Ensure that fee accumulator has been properly initialized
				spreadRewardAccumulator, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, test.poolI.GetId())
				s.Require().NoError(err)
				s.Require().Equal(sdk.DecCoins(nil), spreadRewardAccumulator.GetValue())

				// Ensure that uptime accumulators have been properly initialized
				uptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, test.poolI.GetId())
				s.Require().NoError(err)
				s.Require().Equal(len(types.SupportedUptimes), len(uptimeAccumulators))
				for _, uptimeAccumulator := range uptimeAccumulators {
					s.Require().Equal(cl.EmptyCoins, uptimeAccumulator.GetValue())
				}

				s.validateListenerCallCount(1, 0, 0, 0)
			} else {
				// Ensure specified error is returned
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())

				// Ensure that fee accumulator has not been initialized
				_, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, test.poolI.GetId())
				s.Require().Error(err)

				// Ensure that uptime accumulators have not been initialized
				_, err = s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, test.poolI.GetId())
				s.Require().Error(err)

				s.validateListenerCallCount(0, 0, 0, 0)
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

func (s *KeeperTestSuite) TestAsPoolI() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()

	// Ensure no error occurs when converting to PoolInterface
	_, err := cl.AsPoolI(concentratedPool)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestPoolIToConcentratedPool() {
	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()
	poolI, ok := concentratedPool.(poolmanagertypes.PoolI)
	s.Require().True(ok)

	// Ensure no error occurs when converting to ConcentratedPool
	_, err := cl.AsConcentrated(poolI)
	s.Require().NoError(err)

	// Create a default stableswap pool
	stableswapPoolID := s.PrepareBasicStableswapPool()
	stableswapPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, stableswapPoolID)
	s.Require().NoError(err)

	// Ensure error occurs when converting to ConcentratedPool
	_, err = cl.AsConcentrated(stableswapPool)
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
	poolId := concentratedPool.GetId()

	// should error when price is zero
	spotPrice, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId, ETH, USDC)
	s.Require().Error(err)
	s.Require().ErrorAs(err, &types.NoSpotPriceWhenNoLiquidityError{PoolId: poolId})
	s.Require().Equal(osmomath.BigDec{}, spotPrice)

	// set up default position to have proper spot price
	s.SetupDefaultPosition(defaultPoolId)

	// ETH is token0 so its price will be the DefaultCurrSqrtPrice squared
	spotPriceBaseETH, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId, USDC, ETH)
	s.Require().NoError(err)
	// TODO: remove Dec truncation before https://github.com/osmosis-labs/osmosis/issues/5726 is complete
	// Currently exists for state-compatibility with v19.x
	s.Require().Equal(spotPriceBaseETH.Dec(), DefaultCurrSqrtPrice.PowerInteger(2).Dec())

	// test that we have correct values for reversed quote asset and base asset
	spotPriceBaseUSDC, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId, ETH, USDC)
	s.Require().NoError(err)
	// TODO: remove Dec truncation before https://github.com/osmosis-labs/osmosis/issues/5726 is complete
	// Currently exists for state-compatibility with v19.x
	s.Require().Equal(spotPriceBaseUSDC.Dec(), osmomath.OneBigDec().Quo(DefaultCurrSqrtPrice.PowerInteger(2)).Dec())

	// try getting spot price from a non-existent pool
	spotPrice, err = s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId+1, USDC, ETH)
	s.Require().Error(err)
	s.Require().True(spotPrice.IsNil())
}

func (s *KeeperTestSuite) TestValidateSpreadFactor() {
	s.SetupTest()
	params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	tests := []struct {
		name         string
		spreadFactor osmomath.Dec
		expectValid  bool
	}{
		{
			name:         "Valid spread factor",
			spreadFactor: params.AuthorizedSpreadFactors[0],
			expectValid:  true,
		},
		{
			name:         "Invalid spread factor",
			spreadFactor: params.AuthorizedSpreadFactors[0].Add(osmomath.SmallestDec()),
			expectValid:  false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Method under test.
			isValid := s.App.ConcentratedLiquidityKeeper.ValidateSpreadFactor(s.Ctx, params, test.spreadFactor)

			s.Require().Equal(test.expectValid, isValid)
		})
	}
}

func (s *KeeperTestSuite) TestValidateTickSpacing() {
	s.SetupTest()
	params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	tests := []struct {
		name        string
		tickSpacing uint64
		expectValid bool
	}{
		{
			name:        "Valid tick spacing",
			tickSpacing: params.AuthorizedTickSpacing[0],
			expectValid: true,
		},
		{
			name:        "Invalid tick spacing",
			tickSpacing: params.AuthorizedTickSpacing[0] + 1,
			expectValid: false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Method under test.
			isValid := s.App.ConcentratedLiquidityKeeper.ValidateTickSpacing(s.Ctx, params, test.tickSpacing)

			s.Require().Equal(test.expectValid, isValid)
		})
	}
}

func (s *KeeperTestSuite) TestSetPool() {
	var invalidPool types.ConcentratedPoolExtension
	validPool := clmodel.Pool{
		Address:              s.TestAccs[0].String(),
		IncentivesAddress:    s.TestAccs[1].String(),
		Id:                   1,
		CurrentTickLiquidity: osmomath.ZeroDec(),
		Token0:               ETH,
		Token1:               USDC,
		CurrentSqrtPrice:     osmomath.OneBigDec(),
		CurrentTick:          0,
		TickSpacing:          DefaultTickSpacing,
		ExponentAtPriceOne:   -6,
		SpreadFactor:         osmomath.MustNewDecFromStr("0.003"),
		LastLiquidityUpdate:  s.Ctx.BlockTime(),
	}
	tests := []struct {
		name          string
		pool          types.ConcentratedPoolExtension
		expectedError error
	}{
		{
			name: "happy path",
			pool: &validPool,
		},
		{
			name:          "invalidPool",
			pool:          invalidPool,
			expectedError: errors.New("invalid pool type when setting concentrated pool"),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Retrieving the pool by ID should return an error.
			retrievedPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, 1)
			s.Require().Error(err)
			s.Require().Nil(retrievedPool)

			// Method under test.
			err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, test.pool)
			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			// Retrieving the pool by ID should return the same pool.
			retrievedPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, test.pool.GetId())
			s.Require().NoError(err)
			s.Require().Equal(test.pool, retrievedPool)
		})
	}
}

func (s *KeeperTestSuite) TestValidateAuthorizedQuoteDenoms() {
	tests := []struct {
		name                  string
		quoteDenom            string
		authorizedQuoteDenoms []string
		expectValid           bool
	}{
		{
			name:                  "found - true",
			quoteDenom:            ETH,
			authorizedQuoteDenoms: []string{ETH, USDC},
			expectValid:           true,
		},
		{
			name:                  "not found - false",
			quoteDenom:            ETH,
			authorizedQuoteDenoms: []string{BAR, FOO},
			expectValid:           false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Method under test.
			isValid := cl.ValidateAuthorizedQuoteDenoms(s.Ctx, test.quoteDenom, test.authorizedQuoteDenoms)

			s.Require().Equal(test.expectValid, isValid)
		})
	}
}

func (s *KeeperTestSuite) TestDecreaseConcentratedPoolTickSpacing() {
	type positionRange struct {
		lowerTick int64
		upperTick int64
	}

	tests := []struct {
		name                       string
		poolIdToTickSpacingRecord  []types.PoolIdToTickSpacingRecord
		position                   positionRange
		expectedDecreaseSpacingErr error
		expectedCreatePositionErr  error
	}{
		{
			name:                      "happy path: tick spacing 100 -> 10",
			poolIdToTickSpacingRecord: []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 10}},
			position:                  positionRange{lowerTick: -10, upperTick: 10},
		},
		{
			name:                       "error: new tick spacing not authorized",
			poolIdToTickSpacingRecord:  []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 11}},
			position:                   positionRange{lowerTick: -10, upperTick: 10},
			expectedDecreaseSpacingErr: fmt.Errorf("tick spacing %d is not valid", 11),
		},
		{
			name:                       "error: new tick spacing higher than current",
			poolIdToTickSpacingRecord:  []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 1000}},
			position:                   positionRange{lowerTick: -10, upperTick: 10},
			expectedDecreaseSpacingErr: fmt.Errorf("tick spacing %d is not valid", 1000),
		},
		{
			name:                      "error: can't create position whose lower tick is not divisible by new tick spacing",
			poolIdToTickSpacingRecord: []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 10}},
			position:                  positionRange{lowerTick: -95, upperTick: 100},
			expectedCreatePositionErr: types.TickSpacingError{TickSpacing: 10, LowerTick: -95, UpperTick: 100},
		},
		{
			name:                      "error: can't create position whose upper tick is not divisible by new tick spacing",
			poolIdToTickSpacingRecord: []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 10}},
			position:                  positionRange{lowerTick: -100, upperTick: 95},
			expectedCreatePositionErr: types.TickSpacingError{TickSpacing: 10, LowerTick: -100, UpperTick: 95},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			owner := s.TestAccs[0]

			// Create OSMO <> USDC pool with tick spacing of 100
			concentratedPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(ETH, USDC)

			// Create a position in the pool that is divisible by the tick spacing
			_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), -100, 100)
			s.Require().NoError(err)

			// Attempt to create a position that is not divisible by the tick spacing
			_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), test.position.lowerTick, test.position.upperTick)
			s.Require().Error(err)

			// Alter the tick spacing of the pool
			err = s.App.ConcentratedLiquidityKeeper.DecreaseConcentratedPoolTickSpacing(s.Ctx, test.poolIdToTickSpacingRecord)
			if test.expectedDecreaseSpacingErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedDecreaseSpacingErr.Error())
				return
			}
			s.Require().NoError(err)

			// Attempt to create a position that was previously not divisible by the tick spacing but now is
			_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), test.position.lowerTick, test.position.upperTick)
			if test.expectedCreatePositionErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedCreatePositionErr.Error())
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestGetTotalPoolLiquidity() {
	var (
		defaultPoolCoinOne = sdk.NewCoin(USDC, osmomath.OneInt())
		defaultPoolCoinTwo = sdk.NewCoin(ETH, osmomath.NewInt(2))
		nonPoolCool        = sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(3))

		defaultCoins = sdk.NewCoins(defaultPoolCoinOne, defaultPoolCoinTwo)
	)

	tests := []struct {
		name           string
		poolId         uint64
		poolLiquidity  sdk.Coins
		expectedResult sdk.Coins
		expectedErr    error
	}{
		{
			name:           "valid with 2 coins",
			poolId:         defaultPoolId,
			poolLiquidity:  defaultCoins,
			expectedResult: defaultCoins,
		},
		{
			name:           "valid with 1 coin",
			poolId:         defaultPoolId,
			poolLiquidity:  sdk.NewCoins(defaultPoolCoinTwo),
			expectedResult: sdk.NewCoins(defaultPoolCoinTwo),
		},
		{
			// can only happen if someone sends extra tokens to pool
			// address. Should not occur in practice.
			name:           "valid with 3 coins",
			poolId:         defaultPoolId,
			poolLiquidity:  sdk.NewCoins(defaultPoolCoinTwo, defaultPoolCoinOne, nonPoolCool),
			expectedResult: defaultCoins,
		},
		{
			// this can happen if someone sends random dust to pool address.
			name:           "only non-pool coin - does not show up in result",
			poolId:         defaultPoolId,
			poolLiquidity:  sdk.NewCoins(nonPoolCool),
			expectedResult: sdk.Coins{},
		},
		{
			name:        "invalid pool id",
			poolId:      defaultPoolId + 1,
			expectedErr: types.PoolNotFoundError{PoolId: defaultPoolId + 1},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Create default CL pool
			pool := s.PrepareConcentratedPool()

			s.FundAcc(pool.GetAddress(), tc.poolLiquidity)

			// Get pool defined in test case
			actual, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, tc.poolId)

			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				s.Require().Nil(actual)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedResult, actual)
		})
	}
}

func (s *KeeperTestSuite) TestValidateTickSpacingUpdate() {
	tests := []struct {
		name                     string
		newTickSpacing           uint64
		expectedValidationResult bool
	}{
		{
			name:                     "happy case: reduce tick spacing to smaller tick",
			newTickSpacing:           1,
			expectedValidationResult: true,
		},
		{
			name:                     "validation fail: try reducing unauthorized tick spacing",
			newTickSpacing:           3,
			expectedValidationResult: false,
		},
		{
			name:                     "validation fail: try increasing tick spacing",
			newTickSpacing:           500,
			expectedValidationResult: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Create default CL pool
			// default pool tick spacing is 100.
			pool := s.PrepareConcentratedPool()

			params := types.DefaultParams()
			validationResult := s.App.ConcentratedLiquidityKeeper.ValidateTickSpacingUpdate(s.Ctx, pool, params, tc.newTickSpacing)
			if tc.expectedValidationResult {
				s.Require().True(validationResult)
			} else {
				s.Require().False(validationResult)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetUserUnbondingPositions() {
	var (
		defaultFooAsset balancer.PoolAsset = balancer.PoolAsset{
			Weight: osmomath.NewInt(100),
			Token:  sdk.NewCoin("foo", osmomath.NewInt(10000)),
		}
		defaultBondDenomAsset balancer.PoolAsset = balancer.PoolAsset{
			Weight: osmomath.NewInt(100),
			Token:  sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000)),
		}
		defaultPoolAssets []balancer.PoolAsset = []balancer.PoolAsset{defaultFooAsset, defaultBondDenomAsset}
		defaultAddress                         = s.TestAccs[0]
		defaultFunds                           = sdk.NewCoins(defaultPoolAssets[0].Token, sdk.NewCoin("stake", osmomath.NewInt(5000000000)))
		defaultBlockTime                       = time.Unix(1, 1).UTC()
		defaultLockedAmt                       = sdk.NewCoins(sdk.NewCoin("cl/pool/1", osmomath.NewInt(10000)))
	)

	tests := []struct {
		name           string
		address        sdk.AccAddress
		expectedResult []clmodel.PositionWithPeriodLock
		expectedErr    error
	}{
		{
			name:    "happy path",
			address: defaultAddress,
			expectedResult: []clmodel.PositionWithPeriodLock{
				{
					Position: clmodel.Position{
						PositionId: 3,
						Address:    defaultAddress.String(),
						PoolId:     1,
						LowerTick:  types.MinInitializedTick,
						UpperTick:  types.MaxTick,
						JoinTime:   defaultBlockTime,
						Liquidity:  osmomath.MustNewDecFromStr("10000.000000000000001000"),
					},
					Locks: lockuptypes.PeriodLock{

						ID:       2,
						Owner:    defaultAddress.String(),
						Duration: time.Hour,
						EndTime:  defaultBlockTime.Add(time.Hour),
						Coins:    defaultLockedAmt,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

			clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(defaultFunds[0].Denom, defaultFunds[1].Denom)
			clLockupDenom := types.GetConcentratedLockupDenomFromPoolId(clPool.GetId())
			err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, sftypes.SuperfluidAsset{
				Denom:     clLockupDenom,
				AssetType: sftypes.SuperfluidAssetTypeConcentratedShare,
			})
			s.Require().NoError(err)

			// Create 3 locked positions
			for i := 0; i < 3; i++ {
				_, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, defaultFunds, time.Hour)
				s.Require().NoError(err)
			}

			// The query should return nothing since none of the locks are unlocking
			positionsWithPeriodLock, err := s.App.ConcentratedLiquidityKeeper.GetUserUnbondingPositions(s.Ctx, tc.address)
			s.Require().NoError(err)
			s.Require().Nil(positionsWithPeriodLock)

			// Begin unlocking the second lock only
			lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, 2)
			s.Require().NoError(err)
			_, err = s.App.LockupKeeper.BeginUnlock(s.Ctx, 2, lock.Coins)
			s.Require().NoError(err)

			// The query should return the second lock only
			positionsWithPeriodLock, err = s.App.ConcentratedLiquidityKeeper.GetUserUnbondingPositions(s.Ctx, tc.address)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				s.Require().Nil(positionsWithPeriodLock)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedResult, positionsWithPeriodLock)
		})
	}
}

// This test validates scaling factor migration
// - Creates a pool to migration
// - Creates two positions at different block times
//   - Position 1: Zero accumulator and expected to receive incentives
//   - Position 2: Narrow position. Non-zero accumulator and not expected to receive incentives
//
// # For second position, perform a swap to cross one of the initialized ticks
//
// System under test: Migrates the pool
//
// - Ensures that the pool accumulator trackers are updated.
// - Ensure that the pool accumulator is updates
// - Ensure that the position accumulators are updated
// - Ensures that the position 1 receives incentives  but not position 2
func (s *KeeperTestSuite) TestMigrateIncentivesAccumulatorToScalingFactor() {
	const incentiveDenom = appparams.BaseCoinUnit

	var emissionRatePerSecDec = osmomath.OneDec()

	s.SetupTest()

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()
	poolID := concentratedPool.GetId()

	// Create position one
	// It has position accumulator snapshot of zero
	positionOneID, positionOneLiquidity := s.CreateFullRangePosition(concentratedPool, DefaultCoins)

	// Create incentive
	totalIncentiveAmount := sdk.NewCoin(incentiveDenom, osmomath.NewInt(1000000))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(totalIncentiveAmount))
	_, err := s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, poolID, s.TestAccs[0], totalIncentiveAmount, emissionRatePerSecDec, s.Ctx.BlockTime(), types.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	// Refetch pool
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolID)
	s.Require().NoError(err)
	currentTick := concentratedPool.GetCurrentTick()

	// Create position two (narrow)
	// It has non-zero position accumulator snapshot
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionDataTwo, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolID, s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), currentTick-100, currentTick+100)
	s.Require().NoError(err)
	positionTwoID := positionDataTwo.ID

	// Refetch pool
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolID)
	s.Require().NoError(err)

	// Cross next right tick to update the tick accumulator by swapping
	amtIn, _, _ := s.computeSwapAmounts(poolID, concentratedPool.GetCurrentSqrtPrice(), currentTick+100, false, false)
	s.swapOneForZeroRight(poolID, sdk.NewCoin(USDC, amtIn.Ceil().TruncateInt()))

	// Sync acccumulator
	err = s.App.ConcentratedLiquidityKeeper.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, poolID)
	s.Require().NoError(err)

	// Retrieve pool uptime accumulator
	uptimeAcc, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolID)
	s.Require().NoError(err)

	// Ensure that the accumulator has been properly initialized
	expectedInitialAccumulatorGrowth := sdk.NewDecCoins(sdk.NewDecCoinFromDec(incentiveDenom, osmomath.NewDec(60).MulMut(cl.PerUnitLiqScalingFactor).QuoTruncate(positionOneLiquidity)))
	s.Require().Equal(len(types.SupportedUptimes), len(uptimeAcc))
	s.Require().Equal(expectedInitialAccumulatorGrowth.String(), uptimeAcc[0].GetValue().String())

	// Get ticks before migration
	ticksBeforeMigration, err := s.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Get claimable amount for position one before the migration
	claimableIncentivesOneBeforeMigration, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, positionOneID)
	s.Require().NoError(err)

	// System under test.
	err = s.App.ConcentratedLiquidityKeeper.MigrateIncentivesAccumulatorToScalingFactor(s.Ctx, poolID)
	s.Require().NoError(err)

	// Ensure that the pool accumulator has been properly migrated
	expectedMigratedAccumulatorGrowth := expectedInitialAccumulatorGrowth.MulDecTruncate(cl.PerUnitLiqScalingFactor)
	updatedUptimeAcc, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolID)
	s.Require().NoError(err)
	s.Require().Equal(len(types.SupportedUptimes), len(updatedUptimeAcc))
	incentivizedUpdatedAccumulator := updatedUptimeAcc[0]
	s.Require().Equal(expectedMigratedAccumulatorGrowth.String(), incentivizedUpdatedAccumulator.GetValue().String())

	// Ensure that the ticks have been migrated
	ticksAfterMigration, err := s.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(s.Ctx, poolID)
	s.Require().NoError(err)

	s.Require().NotEmpty(ticksBeforeMigration)
	s.Require().Equal(len(ticksBeforeMigration), len(ticksAfterMigration))
	for i := range ticksBeforeMigration {
		// Validate that the tick uptime accumulator has been properly migrated
		s.Require().Equal(ticksBeforeMigration[i].Info.UptimeTrackers.List[0].UptimeGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor), ticksAfterMigration[i].Info.UptimeTrackers.List[0].UptimeGrowthOutside)
	}

	// Ensure that position 1 accumulator is not updated (zero)
	s.validateUptimePositionAccumulator(incentivizedUpdatedAccumulator, positionOneID, cl.EmptyCoins)

	// Rerun the same swap to get the same result for the incentive
	//
	positionOneCompareID, _ := s.CreateFullRangePosition(concentratedPool, DefaultCoins)

	// Create incentive
	totalIncentiveAmount = sdk.NewCoin(incentiveDenom, osmomath.NewInt(1000000))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(totalIncentiveAmount))
	_, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, poolID, s.TestAccs[0], totalIncentiveAmount, emissionRatePerSecDec, s.Ctx.BlockTime(), types.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	// Refetch pool
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolID)
	s.Require().NoError(err)

	// Cross next right tick to update the tick accumulator by swapping
	amtIn, _, _ = s.computeSwapAmounts(poolID, concentratedPool.GetCurrentSqrtPrice(), currentTick+100, false, false)
	s.swapOneForZeroRight(poolID, sdk.NewCoin(USDC, amtIn.Ceil().TruncateInt()))

	claimableIncentivesCompareOneAfterMigration, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, positionOneCompareID)

	// Do the same swap as before the migration to get the same result
	s.Require().Equal(claimableIncentivesCompareOneAfterMigration.String(), claimableIncentivesOneBeforeMigration.String())

	// Ensure that position 2 cannot claim any incentives
	s.validateClaimableIncentives(positionTwoID, sdk.NewCoins())
}

func (s *KeeperTestSuite) TestMigrateSpreadFactorAccumulatorToScalingFactor() {
	s.SetupTest()
	s.App.ConcentratedLiquidityKeeper.SetSpreadFactorPoolIDMigrationThreshold(s.Ctx, 1000)

	spreadRewardAccumValue := sdk.NewDecCoins(sdk.NewDecCoinFromDec(USDC, osmomath.MustNewDecFromStr("276701288297")))
	positionAccumValue := sdk.NewDecCoins(sdk.NewDecCoinFromDec(USDC, osmomath.MustNewDecFromStr("276701288297").Quo(osmomath.MustNewDecFromStr("2"))))

	// Create CL pool that will not be migrated
	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.MustNewDecFromStr("0.003"))
	poolIDNonMigrated := concentratedPool.GetId()

	// Create CL pool that will be migrated
	concentratedPool = s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.MustNewDecFromStr("0.003"))
	poolIDMigrated := concentratedPool.GetId()

	// Create a position in pool that will not be migrated
	poolNonMigrated, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolIDNonMigrated)
	s.Require().NoError(err)
	poolNonMigratedPositionID, _ := s.CreateFullRangePosition(poolNonMigrated, DefaultCoins)

	// Create a position in pool that will be migrated
	poolMigrated, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolIDMigrated)
	s.Require().NoError(err)
	poolMigratedPositionID, _ := s.CreateFullRangePosition(poolMigrated, DefaultCoins)

	// Manually set spread reward accumulator for pool that will not be migrated
	feeAccumulatorNonMigrated, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolIDNonMigrated)
	s.Require().NoError(err)
	feeAccumulatorNonMigrated.AddToAccumulator(spreadRewardAccumValue)

	// Manually set spread reward accumulator for position that will not be migrated
	nonMigratedPositionAccumulatorKey := types.KeySpreadRewardPositionAccumulator(poolNonMigratedPositionID)
	feeAccumulatorNonMigrated.SetPositionIntervalAccumulation(nonMigratedPositionAccumulatorKey, positionAccumValue)

	// Manually set spread reward accumulator for pool that will be migrated
	feeAccumulatorMigrated, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolIDMigrated)
	s.Require().NoError(err)
	feeAccumulatorMigrated.AddToAccumulator(spreadRewardAccumValue)

	// Manually set spread reward accumulator for position that will be migrated
	migratedPositionAccumulatorKey := types.KeySpreadRewardPositionAccumulator(poolMigratedPositionID)
	feeAccumulatorMigrated.SetPositionIntervalAccumulation(migratedPositionAccumulatorKey, positionAccumValue)

	// Non-migrated pool claim
	nonMigratedPoolBeforeUpgradeSpreadFactor, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, poolNonMigratedPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(nonMigratedPoolBeforeUpgradeSpreadFactor)

	// Migrated pool claim
	migratedPoolBeforeUpgradeSpreadFactor, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, poolMigratedPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(migratedPoolBeforeUpgradeSpreadFactor)

	// System under test.
	err = s.App.ConcentratedLiquidityKeeper.MigrateSpreadFactorAccumulatorToScalingFactor(s.Ctx, poolIDMigrated)
	s.Require().NoError(err)

	// Manually change the pool IDs list to the pool ID in the test
	types.MigratedSpreadFactorAccumulatorPoolIDsV25 = map[uint64]struct{}{}
	types.MigratedSpreadFactorAccumulatorPoolIDsV25[poolIDMigrated] = struct{}{}

	// Non-migrated pool: ensure that the claimable spread rewards are the same before and after migration
	nonMigratedPoolAfterUpgradeSpreadFactor, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, poolNonMigratedPositionID)
	s.Require().NoError(err)
	s.Require().Equal(nonMigratedPoolBeforeUpgradeSpreadFactor.String(), nonMigratedPoolAfterUpgradeSpreadFactor.String())

	// Migrated pool: ensure that the claimable spread rewards are the same before and after migration
	migratedPoolAfterUpgradeSpreadFactor, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, poolMigratedPositionID)
	s.Require().NoError(err)
	s.Require().Equal(migratedPoolBeforeUpgradeSpreadFactor.String(), migratedPoolAfterUpgradeSpreadFactor.String())

	// Position's accumulator for non migrated pool should not be updated
	feeAccumulatorNonMigrated, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolIDNonMigrated)
	s.Require().NoError(err)
	nonMigratedPositionAfterMigration, err := feeAccumulatorNonMigrated.GetPosition(nonMigratedPositionAccumulatorKey)
	s.Require().NoError(err)
	s.Require().Equal(positionAccumValue.String(), nonMigratedPositionAfterMigration.AccumValuePerShare.String())

	// Position's accumulator for migrated pool should be updated
	feeAccumulatorMigrated, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolIDMigrated)
	s.Require().NoError(err)
	migratedPositionAfterMigration, err := feeAccumulatorMigrated.GetPosition(migratedPositionAccumulatorKey)
	s.Require().NoError(err)
	s.Require().Equal(positionAccumValue.MulDecTruncate(cl.PerUnitLiqScalingFactor).String(), migratedPositionAfterMigration.AccumValuePerShare.String())
}

// Basic test to validate that positions are correctly returned for a pool
func (s *KeeperTestSuite) TestGetPositionIDsByPoolID() {
	s.SetupTest()

	const numPositionsToCreateFirstPool = 3

	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()
	poolID := concentratedPool.GetId()

	// Create second pool
	secondPool := s.PrepareConcentratedPool()

	positionIDs, err := s.App.ConcentratedLiquidityKeeper.GetPositionIDsByPoolID(s.Ctx, poolID)
	s.Require().NoError(err)

	s.Require().Equal([]uint64{}, positionIDs)

	// Create three positions
	for i := 0; i < numPositionsToCreateFirstPool; i++ {
		s.CreateFullRangePosition(concentratedPool, DefaultCoins)
	}

	// Create one position in second pool
	s.CreateFullRangePosition(secondPool, DefaultCoins)

	positionIDs, err = s.App.ConcentratedLiquidityKeeper.GetPositionIDsByPoolID(s.Ctx, poolID)
	s.Require().NoError(err)

	s.Require().Equal([]uint64{1, 2, 3}, positionIDs)

	positionIDs, err = s.App.ConcentratedLiquidityKeeper.GetPositionIDsByPoolID(s.Ctx, secondPool.GetId())
	s.Require().NoError(err)

	s.Require().Equal([]uint64{numPositionsToCreateFirstPool + 1}, positionIDs)
}
