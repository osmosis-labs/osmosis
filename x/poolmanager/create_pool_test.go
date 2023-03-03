package poolmanager_test

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
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
		msg             balancertypes.MsgCreateBalancerPool
		expectPass      bool
	}{
		{
			name:            "no pool creation fee for default asset pool",
			poolCreationFee: sdk.Coins{},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "nil pool creation fee on basic pool",
			poolCreationFee: nil,
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "attempt pool creation without sufficient funds for fees",
			poolCreationFee: sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(10000))},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
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

	validBalancerPoolMsg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.NewPoolParams(sdk.ZeroDec(), sdk.ZeroDec(), nil), []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(foo, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
		{
			Token:  sdk.NewCoin(bar, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
	}, "")

	validConcentratedPoolMsg := clmodel.NewMsgCreateConcentratedPool(suite.TestAccs[0], foo, bar, 1, DefaultExponentAtPriceOne, defaultPoolSwapFee)

	tests := []struct {
		name               string
		creatorFundAmount  sdk.Coins
		msg                types.CreatePoolMsg
		expectedModuleType reflect.Type
		expectError        bool
	}{
		{
			name:               "first balancer pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "second balancer pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "concentrated pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validConcentratedPoolMsg,
			expectedModuleType: concentratedKeeperType,
		},
		// TODO: add stableswap test
		// TODO: add concentrated-liquidity test
		// TODO: cover errors and edge cases
	}

	for i, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

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

func (suite *KeeperTestSuite) TestGetAllModuleRoutes() {
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
