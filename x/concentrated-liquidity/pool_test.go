package concentrated_liquidity_test

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	sftypes "github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
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
	invalidSpreadFactor := sdk.MustNewDecFromStr("0.1")
	invalidSpreadFactorConcentratedPool, err := clmodel.NewConcentratedLiquidityPool(3, ETH, USDC, DefaultTickSpacing, invalidSpreadFactor)
	s.Require().NoError(err)

	// Create an invalid PoolI that doesn't implement ConcentratedPoolExtension
	invalidPoolId := s.PrepareBalancerPool()
	invalidPoolI, err := s.App.GAMMKeeper.GetPool(s.Ctx, invalidPoolId)
	s.Require().NoError(err)

	validCreatorAddress := s.TestAccs[0]

	tests := []struct {
		name                      string
		poolI                     poolmanagertypes.PoolI
		authorizedDenomsOverwrite []string
		creatorAddress            sdk.AccAddress
		expectedErr               error
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
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			if len(test.authorizedDenomsOverwrite) > 0 {
				params := s.App.PoolManagerKeeper.GetParams(s.Ctx)
				params.AuthorizedQuoteDenoms = test.authorizedDenomsOverwrite
				s.App.PoolManagerKeeper.SetParams(s.Ctx, params)
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
	s.Require().Equal(sdk.Dec{}, spotPrice)

	// set up default position to have proper spot price
	s.SetupDefaultPosition(defaultPoolId)

	// ETH is token0 so its price will be the DefaultCurrSqrtPrice squared
	spotPriceBaseETH, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId, USDC, ETH)
	s.Require().NoError(err)
	s.Require().Equal(spotPriceBaseETH, DefaultCurrSqrtPrice.PowerInteger(2).SDKDec())

	// test that we have correct values for reversed quote asset and base asset
	spotPriceBaseUSDC, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, poolId, ETH, USDC)
	s.Require().NoError(err)
	s.Require().Equal(spotPriceBaseUSDC, osmomath.OneDec().Quo(DefaultCurrSqrtPrice.PowerInteger(2)).SDKDec())

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
		spreadFactor sdk.Dec
		expectValid  bool
	}{
		{
			name:         "Valid spread factor",
			spreadFactor: params.AuthorizedSpreadFactors[0],
			expectValid:  true,
		},
		{
			name:         "Invalid spread factor",
			spreadFactor: params.AuthorizedSpreadFactors[0].Add(sdk.SmallestDec()),
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
		CurrentTickLiquidity: sdk.ZeroDec(),
		Token0:               ETH,
		Token1:               USDC,
		CurrentSqrtPrice:     osmomath.OneDec(),
		CurrentTick:          0,
		TickSpacing:          DefaultTickSpacing,
		ExponentAtPriceOne:   -6,
		SpreadFactor:         sdk.MustNewDecFromStr("0.003"),
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
			name:                      "error: cant create position whose lower tick is not divisible by new tick spacing",
			poolIdToTickSpacingRecord: []types.PoolIdToTickSpacingRecord{{PoolId: 1, NewTickSpacing: 10}},
			position:                  positionRange{lowerTick: -95, upperTick: 100},
			expectedCreatePositionErr: types.TickSpacingError{TickSpacing: 10, LowerTick: -95, UpperTick: 100},
		},
		{
			name:                      "error: cant create position whose upper tick is not divisible by new tick spacing",
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
			_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), -100, 100)
			s.Require().NoError(err)

			// Attempt to create a position that is not divisible by the tick spacing
			_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), test.position.lowerTick, test.position.upperTick)
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
			_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), test.position.lowerTick, test.position.upperTick)
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
		defaultPoolCoinOne = sdk.NewCoin(USDC, sdk.OneInt())
		defaultPoolCoinTwo = sdk.NewCoin(ETH, sdk.NewInt(2))
		nonPoolCool        = sdk.NewCoin("uosmo", sdk.NewInt(3))

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
			expectedResult: sdk.Coins(nil),
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
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
		}
		defaultBondDenomAsset balancer.PoolAsset = balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)),
		}
		defaultPoolAssets []balancer.PoolAsset = []balancer.PoolAsset{defaultFooAsset, defaultBondDenomAsset}
		defaultAddress                         = s.TestAccs[0]
		defaultFunds                           = sdk.NewCoins(defaultPoolAssets[0].Token, sdk.NewCoin("stake", sdk.NewInt(5000000000)))
		defaultBlockTime                       = time.Unix(1, 1).UTC()
		defaultLockedAmt                       = sdk.NewCoins(sdk.NewCoin("cl/pool/1", sdk.NewInt(10000)))
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
						Liquidity:  sdk.MustNewDecFromStr("10000.000000000000001000"),
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
