package poolmanager_test

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	cwmodel "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	stableswap "github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func (s *KeeperTestSuite) TestPoolCreationFee() {
	params := s.App.PoolManagerKeeper.GetParams(s.Ctx)

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
			msg: balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 2),
				ExitFee: osmomath.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "nil pool creation fee on basic pool",
			poolCreationFee: nil,
			msg: balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 2),
				ExitFee: osmomath.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "attempt pool creation without sufficient funds for fees",
			poolCreationFee: sdk.Coins{sdk.NewCoin("atom", osmomath.NewInt(10000))},
			msg: balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 2),
				ExitFee: osmomath.ZeroDec(),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()
		gammKeeper := s.App.GAMMKeeper
		distributionKeeper := s.App.DistrKeeper
		bankKeeper := s.App.BankKeeper
		poolmanagerKeeper := s.App.PoolManagerKeeper

		// set pool creation fee
		newParams := params
		newParams.PoolCreationFee = test.poolCreationFee
		poolmanagerKeeper.SetParams(s.Ctx, newParams)

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		s.Require().NoError(err, "test: %v", test.name)
		s.FundAcc(sender, apptesting.DefaultAcctFunds)

		// note starting balances for community fee pool and pool creator account
		feePoolBalBeforeNewPoolStruct, err := distributionKeeper.FeePool.Get(s.Ctx)
		s.Require().NoError(err, "test: %v", test.name)
		feePoolBalBeforeNewPool := feePoolBalBeforeNewPoolStruct.CommunityPool
		senderBalBeforeNewPool := bankKeeper.GetAllBalances(s.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := poolmanagerKeeper.CreatePool(s.Ctx, test.msg)

		if test.expectPass {
			s.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := gammKeeper.GetPoolAndPoke(s.Ctx, poolId)
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal(gammtypes.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", gammtypes.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePoolStruct, err := distributionKeeper.FeePool.Get(s.Ctx)
			s.Require().NoError(err, "test: %v", test.name)
			feePool := feePoolStruct.CommunityPool
			s.Require().Equal(feePool, feePoolBalBeforeNewPool.Add(sdk.NewDecCoinsFromCoins(test.poolCreationFee...)...))
			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(gammtypes.GetPoolShareDenom(pool.GetId()), gammtypes.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := bankKeeper.GetAllBalances(s.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(test.poolCreationFee...).Sub(expectedPoolTokens...).Add(expectedPoolShares)
			s.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity, err := gammKeeper.GetTotalLiquidity(s.Ctx)
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			s.Require().Error(err, "test: %v", test.name)
		}
	}
}

// TestCreatePool tests that all possible pools are created correctly.
func (s *KeeperTestSuite) TestCreatePool() {
	var (
		validBalancerPoolMsg = balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.NewPoolParams(osmomath.ZeroDec(), osmomath.ZeroDec(), nil), []balancer.PoolAsset{
			{
				Token:  sdk.NewCoin(FOO, defaultInitPoolAmount),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(BAR, defaultInitPoolAmount),
				Weight: osmomath.NewInt(1),
			},
		}, "")

		invalidBalancerPoolMsg = balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.NewPoolParams(osmomath.ZeroDec(), osmomath.NewDecWithPrec(1, 2), nil), []balancer.PoolAsset{
			{
				Token:  sdk.NewCoin(FOO, defaultInitPoolAmount),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(BAR, defaultInitPoolAmount),
				Weight: osmomath.NewInt(1),
			},
		}, "")

		DefaultStableswapLiquidity = sdk.NewCoins(
			sdk.NewCoin(FOO, defaultInitPoolAmount),
			sdk.NewCoin(BAR, defaultInitPoolAmount),
		)

		validStableswapPoolMsg = stableswap.NewMsgCreateStableswapPool(s.TestAccs[0], stableswap.PoolParams{SwapFee: osmomath.NewDec(0), ExitFee: osmomath.NewDec(0)}, DefaultStableswapLiquidity, []uint64{}, "")

		invalidStableswapPoolMsg = stableswap.NewMsgCreateStableswapPool(s.TestAccs[0], stableswap.PoolParams{SwapFee: osmomath.NewDec(0), ExitFee: osmomath.NewDecWithPrec(1, 2)}, DefaultStableswapLiquidity, []uint64{}, "")

		validConcentratedPoolMsg = clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], FOO, BAR, 1, defaultPoolSpreadFactor)

		validTransmuterCodeId = uint64(1)
		validCWPoolMsg        = cwmodel.NewMsgCreateCosmWasmPool(validTransmuterCodeId, s.TestAccs[0], s.GetDefaultTransmuterInstantiateMsgBytes())

		defaultFundAmount = sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount.Mul(osmomath.NewInt(2))), sdk.NewCoin(BAR, defaultInitPoolAmount.Mul(osmomath.NewInt(2))))
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
			name:               "cosmwasm pool - success",
			creatorFundAmount:  defaultFundAmount,
			msg:                validCWPoolMsg,
			expectedModuleType: cosmwasmKeeperType,
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
		s.Run(tc.name, func() {
			if tc.isPermissionlessPoolCreationDisabled {
				params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				params.IsPermissionlessPoolCreationEnabled = false
				s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)
			}

			if tc.expectedModuleType == cosmwasmKeeperType {
				codeId := s.StoreCosmWasmPoolContractCode(apptesting.TransmuterContractName)
				s.Require().Equal(validTransmuterCodeId, codeId)
				s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, codeId)
			}

			poolmanagerKeeper := s.App.PoolManagerKeeper
			ctx := s.Ctx

			poolCreationFee := poolmanagerKeeper.GetParams(s.Ctx).PoolCreationFee
			s.FundAcc(s.TestAccs[0], append(tc.creatorFundAmount, poolCreationFee...))

			poolId, err := poolmanagerKeeper.CreatePool(ctx, tc.msg)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			// Validate pool.
			s.Require().NoError(err)
			s.Require().Equal(uint64(i+1), poolId)

			// Validate that mapping pool id -> module type has been persisted.
			swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, poolId)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedModuleType, reflect.TypeOf(swapModule))
		})
	}
}

// Tests that only poolmanager as a pool creator can create a pool via CreatePoolZeroLiquidityNoCreationFee
func (s *KeeperTestSuite) TestCreatePoolZeroLiquidityNoCreationFee() {
	poolManagerModuleAcc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)

	withCreator := func(msg clmodel.MsgCreateConcentratedPool, address sdk.AccAddress) clmodel.MsgCreateConcentratedPool {
		msg.Sender = address.String()
		return msg
	}

	balancerPoolMsg := balancer.NewMsgCreateBalancerPool(poolManagerModuleAcc.GetAddress(), balancer.NewPoolParams(osmomath.ZeroDec(), osmomath.ZeroDec(), nil), []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(FOO, defaultInitPoolAmount),
			Weight: osmomath.NewInt(1),
		},
		{
			Token:  sdk.NewCoin(BAR, defaultInitPoolAmount),
			Weight: osmomath.NewInt(1),
		},
	}, "")

	concentratedPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolManagerModuleAcc.GetAddress(), FOO, BAR, 1, defaultPoolSpreadFactor)

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
			msg:         withCreator(concentratedPoolMsg, s.TestAccs[0]),
			expectError: types.InvalidPoolCreatorError{CreatorAddresss: s.TestAccs[0].String()},
		},
		{
			name:        "balancer pool with pool manager creator - error, wrong pool",
			msg:         balancerPoolMsg,
			expectError: types.InvalidPoolTypeError{PoolType: types.Balancer},
		},
	}

	for i, tc := range tests {
		s.Run(tc.name, func() {
			poolmanagerKeeper := s.App.PoolManagerKeeper
			ctx := s.Ctx

			// Note: this is necessary for gauge creation in the after pool created hook.
			// There is a check requiring positive supply existing on-chain.
			s.MintCoins(sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.OneInt())))

			pool, err := poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, tc.msg)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}

			// Validate pool.
			s.Require().NoError(err)
			s.Require().Equal(uint64(i+1), pool.GetId())

			// Validate that mapping pool id -> module type has been persisted.
			swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, pool.GetId())
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedModuleType, reflect.TypeOf(swapModule))
		})
	}
}

func (s *KeeperTestSuite) TestSetAndGetAllPoolRoutes() {
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
				{
					PoolType: types.CosmWasm,
					PoolId:   4,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Setup()
			poolManagerKeeper := s.App.PoolManagerKeeper

			for _, preSetRoute := range tc.preSetRoutes {
				poolManagerKeeper.SetPoolRoute(s.Ctx, preSetRoute.PoolId, preSetRoute.PoolType)
			}

			moduleRoutes := poolManagerKeeper.GetAllPoolRoutes(s.Ctx)

			// Validate.
			s.Require().Len(moduleRoutes, len(tc.preSetRoutes))
			s.Require().EqualValues(tc.preSetRoutes, moduleRoutes)
		})
	}
}

func (s *KeeperTestSuite) TestGetNextPoolIdAndIncrement() {
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
		s.Run(tc.name, func() {
			s.Setup()

			s.App.PoolManagerKeeper.SetNextPoolId(s.Ctx, tc.expectedNextPoolId)
			nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
			s.Require().Equal(tc.expectedNextPoolId, nextPoolId)

			// System under test.
			nextPoolId = s.App.PoolManagerKeeper.GetNextPoolIdAndIncrement(s.Ctx)
			s.Require().Equal(tc.expectedNextPoolId, nextPoolId)
			s.Require().Equal(tc.expectedNextPoolId+1, s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx))
		})
	}
}

func (s *KeeperTestSuite) TestValidateCreatedPool() {
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
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Setup()

			// System under test.
			err := s.App.PoolManagerKeeper.ValidateCreatedPool(s.Ctx, tc.poolId, tc.pool)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
		})
	}
}
