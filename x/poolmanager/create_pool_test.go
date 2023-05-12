package poolmanager_test

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	stableswap "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (suite *KeeperTestSuite) TestPoolCreationFee() {
	params := suite.App.PoolManagerKeeper.GetParams(suite.Ctx)

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	tests := []struct {
		name            string
		poolCreationFee sdk.Coins
		msg             balancer.MsgCreateBalancerPool
		expectPass      bool
	}{
		{
			name:            "no pool creation fee for default asset pool",
			poolCreationFee: sdk.Coins{},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "nil pool creation fee on basic pool",
			poolCreationFee: nil,
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "attempt pool creation without sufficient funds for fees",
			poolCreationFee: sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(10000))},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		gammKeeper := suite.App.GAMMKeeper
		distributionKeeper := suite.App.DistrKeeper
		bankKeeper := suite.App.BankKeeper
		poolmanagerKeeper := suite.App.PoolManagerKeeper

		// set pool creation fee
		poolmanagerKeeper.SetParams(suite.Ctx, types.Params{
			PoolCreationFee: test.poolCreationFee,
		})

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		suite.Require().NoError(err, "test: %v", test.name)
		suite.FundAcc(sender, apptesting.DefaultAcctFunds)

		// note starting balances for community fee pool and pool creator account
		feePoolBalBeforeNewPool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
		senderBalBeforeNewPool := bankKeeper.GetAllBalances(suite.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := poolmanagerKeeper.CreatePool(suite.Ctx, test.msg)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := gammKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(gammtypes.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", gammtypes.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
			suite.Require().Equal(feePool, feePoolBalBeforeNewPool.Add(sdk.NewDecCoinsFromCoins(test.poolCreationFee...)...))
			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(gammtypes.GetPoolShareDenom(pool.GetId()), gammtypes.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := bankKeeper.GetAllBalances(suite.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(test.poolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			suite.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity := gammKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}
}

// TestCreatePool tests that all possible pools are created correctly.
func (suite *KeeperTestSuite) TestCreatePool() {
	var (
		validBalancerPoolMsg = balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.NewPoolParams(sdk.ZeroDec(), sdk.ZeroDec(), nil), []balancer.PoolAsset{
			{
				Token:  sdk.NewCoin(foo, defaultInitPoolAmount),
				Weight: sdk.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(bar, defaultInitPoolAmount),
				Weight: sdk.NewInt(1),
			},
		}, "")

		invalidBalancerPoolMsg = balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.NewPoolParams(sdk.ZeroDec(), sdk.NewDecWithPrec(1, 2), nil), []balancer.PoolAsset{
			{
				Token:  sdk.NewCoin(foo, defaultInitPoolAmount),
				Weight: sdk.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(bar, defaultInitPoolAmount),
				Weight: sdk.NewInt(1),
			},
		}, "")

		DefaultStableswapLiquidity = sdk.NewCoins(
			sdk.NewCoin(foo, defaultInitPoolAmount),
			sdk.NewCoin(bar, defaultInitPoolAmount),
		)

		validStableswapPoolMsg = stableswap.NewMsgCreateStableswapPool(suite.TestAccs[0], stableswap.PoolParams{SwapFee: sdk.NewDec(0), ExitFee: sdk.NewDec(0)}, DefaultStableswapLiquidity, []uint64{}, "")

		invalidStableswapPoolMsg = stableswap.NewMsgCreateStableswapPool(suite.TestAccs[0], stableswap.PoolParams{SwapFee: sdk.NewDec(0), ExitFee: sdk.NewDecWithPrec(1, 2)}, DefaultStableswapLiquidity, []uint64{}, "")

		validConcentratedPoolMsg = clmodel.NewMsgCreateConcentratedPool(suite.TestAccs[0], foo, bar, 1, defaultPoolSwapFee)

		defaultFundAmount = sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2))))
	)

	tests := []struct {
		name                                 string
		creatorFundAmount                    sdk.Coins
		isPermissionlessPoolCreationDisabled bool
		msg                                  types.CreatePoolMsg
		expectedModuleType                   reflect.Type
		expectError                          bool
	}{
		{
			name:               "first balancer pool - success",
			creatorFundAmount:  defaultFundAmount,
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "second balancer pool - success",
			creatorFundAmount:  defaultFundAmount,
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "stableswap pool - success",
			creatorFundAmount:  defaultFundAmount,
			msg:                validStableswapPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "concentrated pool - success",
			creatorFundAmount:  defaultFundAmount,
			msg:                validConcentratedPoolMsg,
			expectedModuleType: concentratedKeeperType,
		},
		{
			name:               "error: balancer pool with non zero exit fee",
			creatorFundAmount:  defaultFundAmount,
			msg:                invalidBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
			expectError:        true,
		},
		{
			name:               "error: stableswap pool with non zero exit fee",
			creatorFundAmount:  defaultFundAmount,
			msg:                invalidStableswapPoolMsg,
			expectedModuleType: gammKeeperType,
			expectError:        true,
		},
		{
			name:                                 "error: pool creation is disabled for concentrated pool via param",
			creatorFundAmount:                    defaultFundAmount,
			isPermissionlessPoolCreationDisabled: true,
			msg:                                  validConcentratedPoolMsg,
			expectedModuleType:                   concentratedKeeperType,
			expectError:                          true,
		},
	}

	for i, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

			if tc.isPermissionlessPoolCreationDisabled {
				params := suite.App.ConcentratedLiquidityKeeper.GetParams(suite.Ctx)
				params.IsPermissionlessPoolCreationEnabled = false
				suite.App.ConcentratedLiquidityKeeper.SetParams(suite.Ctx, params)
			}

			poolmanagerKeeper := suite.App.PoolManagerKeeper
			ctx := suite.Ctx

			poolCreationFee := poolmanagerKeeper.GetParams(suite.Ctx).PoolCreationFee
			suite.FundAcc(suite.TestAccs[0], append(tc.creatorFundAmount, poolCreationFee...))

			poolId, err := poolmanagerKeeper.CreatePool(ctx, tc.msg)

			if tc.expectError {
				suite.Require().Error(err)
				return
			}

			// Validate pool.
			suite.Require().NoError(err)
			suite.Require().Equal(uint64(i+1), poolId)

			// Validate that mapping pool id -> module type has been persisted.
			swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedModuleType, reflect.TypeOf(swapModule))
		})
	}
}

// Tests that only poolmanager as a pool creator can create a pool via CreatePoolZeroLiquidityNoCreationFee
func (suite *KeeperTestSuite) TestCreatePoolZeroLiquidityNoCreationFee() {
	suite.SetupTest()

	poolManagerModuleAcc := suite.App.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)

	withCreator := func(msg clmodel.MsgCreateConcentratedPool, address sdk.AccAddress) clmodel.MsgCreateConcentratedPool {
		msg.Sender = address.String()
		return msg
	}

	balancerPoolMsg := balancer.NewMsgCreateBalancerPool(poolManagerModuleAcc.GetAddress(), balancer.NewPoolParams(sdk.ZeroDec(), sdk.ZeroDec(), nil), []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(foo, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
		{
			Token:  sdk.NewCoin(bar, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
	}, "")

	concentratedPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolManagerModuleAcc.GetAddress(), foo, bar, 1, defaultPoolSwapFee)

	tests := []struct {
		name               string
		msg                types.CreatePoolMsg
		expectedModuleType reflect.Type
		expectError        error
	}{
		{
			name:               "pool manager creator for concentrated pool - success",
			msg:                concentratedPoolMsg,
			expectedModuleType: concentratedKeeperType,
		},
		{
			name:        "creator is not pool manager - failure",
			msg:         withCreator(concentratedPoolMsg, suite.TestAccs[0]),
			expectError: types.InvalidPoolCreatorError{CreatorAddresss: suite.TestAccs[0].String()},
		},
		{
			name:        "balancer pool with pool manager creator - error, wrong pool",
			msg:         balancerPoolMsg,
			expectError: types.InvalidPoolTypeError{PoolType: types.Balancer},
		},
	}

	for i, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

			poolmanagerKeeper := suite.App.PoolManagerKeeper
			ctx := suite.Ctx

			// Note: this is necessary for gauge creation in the after pool created hook.
			// There is a check requiring positive supply existing on-chain.
			suite.MintCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.OneInt())))

			pool, err := poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, tc.msg)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}

			// Validate pool.
			suite.Require().NoError(err)
			suite.Require().Equal(uint64(i+1), pool.GetId())

			// Validate that mapping pool id -> module type has been persisted.
			swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, pool.GetId())
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedModuleType, reflect.TypeOf(swapModule))
		})
	}
}

func (suite *KeeperTestSuite) TestSetAndGetAllPoolRoutes() {
	tests := []struct {
		name         string
		preSetRoutes []types.ModuleRoute
	}{
		{
			name:         "no routes",
			preSetRoutes: []types.ModuleRoute{},
		},
		{
			name: "only balancer",
			preSetRoutes: []types.ModuleRoute{
				{
					PoolType: types.Balancer,
					PoolId:   1,
				},
			},
		},
		{
			name: "two balancer",
			preSetRoutes: []types.ModuleRoute{
				{
					PoolType: types.Balancer,
					PoolId:   1,
				},
				{
					PoolType: types.Balancer,
					PoolId:   2,
				},
			},
		},
		{
			name: "all supported pools",
			preSetRoutes: []types.ModuleRoute{
				{
					PoolType: types.Balancer,
					PoolId:   1,
				},
				{
					PoolType: types.Stableswap,
					PoolId:   2,
				},
				{
					PoolType: types.Concentrated,
					PoolId:   3,
				},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

			suite.Setup()
			poolManagerKeeper := suite.App.PoolManagerKeeper

			for _, preSetRoute := range tc.preSetRoutes {
				poolManagerKeeper.SetPoolRoute(suite.Ctx, preSetRoute.PoolId, preSetRoute.PoolType)
			}

			moduleRoutes := poolManagerKeeper.GetAllPoolRoutes(suite.Ctx)

			// Validate.
			suite.Require().Len(moduleRoutes, len(tc.preSetRoutes))
			suite.Require().EqualValues(tc.preSetRoutes, moduleRoutes)
		})
	}
}

func (suite *KeeperTestSuite) TestGetNextPoolIdAndIncrement() {
	tests := []struct {
		name               string
		expectedNextPoolId uint64
	}{
		{
			name:               "small next pool ID",
			expectedNextPoolId: 2,
		},
		{
			name:               "large next pool ID",
			expectedNextPoolId: 2999999,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc
			suite.Setup()

			suite.App.PoolManagerKeeper.SetNextPoolId(suite.Ctx, tc.expectedNextPoolId)
			nextPoolId := suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx)
			suite.Require().Equal(tc.expectedNextPoolId, nextPoolId)

			// System under test.
			nextPoolId = suite.App.PoolManagerKeeper.GetNextPoolIdAndIncrement(suite.Ctx)
			suite.Require().Equal(tc.expectedNextPoolId, nextPoolId)
			suite.Require().Equal(tc.expectedNextPoolId+1, suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx))
		})
	}
}

func (suite *KeeperTestSuite) TestValidateCreatedPool() {
	tests := []struct {
		name          string
		poolId        uint64
		pool          types.PoolI
		expectedError error
	}{
		{
			name:   "pool ID 1",
			poolId: 1,
			pool: &balancer.Pool{
				Address: types.NewPoolAddress(1).String(),
				Id:      1,
			},
		},
		{
			name:   "pool ID 309",
			poolId: 309,
			pool: &balancer.Pool{
				Address: types.NewPoolAddress(309).String(),
				Id:      309,
			},
		},
		{
			name:   "error: unexpected ID",
			poolId: 1,
			pool: &balancer.Pool{
				Address: types.NewPoolAddress(1).String(),
				Id:      2,
			},
			expectedError: types.IncorrectPoolIdError{ExpectedPoolId: 1, ActualPoolId: 2},
		},
		{
			name:   "error: unexpected address",
			poolId: 2,
			pool: &balancer.Pool{
				Address: types.NewPoolAddress(1).String(),
				Id:      2,
			},
			expectedError: types.IncorrectPoolAddressError{ExpectedPoolAddress: types.NewPoolAddress(2).String(), ActualPoolAddress: types.NewPoolAddress(1).String()},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc
			suite.Setup()

			// System under test.
			err := suite.App.PoolManagerKeeper.ValidateCreatedPool(suite.Ctx, tc.poolId, tc.pool)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
		})
	}
}
