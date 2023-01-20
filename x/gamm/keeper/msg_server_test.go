package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

const (
	doesNotExistDenom = "nodenom"
	// Max positive int64.
	int64Max = int64(^uint64(0) >> 1)
)

// TestSwapExactAmountIn_Events tests that events are correctly emitted
// when calling SwapExactAmountIn.
func (suite *KeeperTestSuite) TestSwapExactAmountIn_Events() {
	const (
		tokenInMinAmount = 1
		tokenIn          = 5
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountInRoute
		tokenIn               sdk.Coin
		tokenOutMinAmount     sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:            []poolmanagertypes.SwapAmountInRoute{},
			tokenIn:           sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount: sdk.NewInt(tokenInMinAmount),
			expectError:       true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
			},
			tokenIn:               sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount:     sdk.NewInt(tokenInMinAmount),
			expectedSwapEvents:    1,
			expectedMessageEvents: 3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"two hops": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
				{
					PoolId:        2,
					TokenOutDenom: "baz",
				},
			},
			tokenIn:               sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount:     sdk.NewInt(tokenInMinAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 5, // 1 gamm + 4 events emitted by other keeper methods.
		},
		"invalid - two hops, denom does not exist": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
				{
					PoolId:        2,
					TokenOutDenom: "baz",
				},
			},
			tokenIn:           sdk.NewCoin(doesNotExistDenom, sdk.NewInt(tokenIn)),
			tokenOutMinAmount: sdk.NewInt(tokenInMinAmount),
			expectError:       true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountIn(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountIn{
				Sender:            suite.TestAccs[0].String(),
				Routes:            tc.routes,
				TokenIn:           tc.tokenIn,
				TokenOutMinAmount: tc.tokenOutMinAmount,
			})

			if !tc.expectError {
				suite.NoError(err)
				suite.NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestSwapExactAmountOut_Events tests that events are correctly emitted
// when calling SwapExactAmountOut.
func (suite *KeeperTestSuite) TestSwapExactAmountOut_Events() {
	const (
		tokenInMaxAmount = int64Max
		tokenOut         = 5
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountOutRoute
		tokenOut              sdk.Coin
		tokenInMaxAmount      sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:           []poolmanagertypes.SwapAmountOutRoute{},
			tokenOut:         sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount: sdk.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
			},
			tokenOut:              sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount:      sdk.NewInt(tokenInMaxAmount),
			expectedSwapEvents:    1,
			expectedMessageEvents: 3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"two hops": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
				{
					PoolId:       2,
					TokenInDenom: "baz",
				},
			},
			tokenOut:              sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount:      sdk.NewInt(tokenInMaxAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 5, // 1 gamm + 4 events emitted by other keeper methods.
		},
		"invalid - two hops, denom does not exist": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
				{
					PoolId:       2,
					TokenInDenom: "baz",
				},
			},
			tokenOut:         sdk.NewCoin(doesNotExistDenom, sdk.NewInt(tokenOut)),
			tokenInMaxAmount: sdk.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountOut(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountOut{
				Sender:           suite.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOut:         tc.tokenOut,
				TokenInMaxAmount: tc.tokenInMaxAmount,
			})

			if !tc.expectError {
				suite.NoError(err)
				suite.NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestJoinPool_Events tests that events are correctly emitted
// when calling JoinPool.
func (suite *KeeperTestSuite) TestJoinPool_Events() {
	const (
		tokenInMaxAmount = int64Max
		shareOut         = 110
	)

	testcases := map[string]struct {
		poolId                     uint64
		shareOutAmount             sdk.Int
		tokenInMaxs                sdk.Coins
		expectError                bool
		expectedAddLiquidityEvents int
		expectedMessageEvents      int
	}{
		"successful join": {
			poolId:         1,
			shareOutAmount: sdk.NewInt(shareOut),
			tokenInMaxs: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("bar", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("baz", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("uosmo", sdk.NewInt(tokenInMaxAmount)),
			),
			expectedAddLiquidityEvents: 1,
			expectedMessageEvents:      3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"tokenInMaxs do not match all tokens in pool - invalid join": {
			poolId:         1,
			shareOutAmount: sdk.NewInt(shareOut),
			tokenInMaxs:    sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tokenInMaxAmount))),
			expectError:    true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.JoinPool(sdk.WrapSDKContext(ctx), &types.MsgJoinPool{
				Sender:         suite.TestAccs[0].String(),
				PoolId:         tc.poolId,
				ShareOutAmount: tc.shareOutAmount,
				TokenInMaxs:    tc.tokenInMaxs,
			})

			if !tc.expectError {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, tc.expectedAddLiquidityEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestExitPool_Events tests that events are correctly emitted
// when calling ExitPool.
func (suite *KeeperTestSuite) TestExitPool_Events() {
	const (
		tokenOutMinAmount = 1
		shareIn           = 110
	)

	testcases := map[string]struct {
		poolId                        uint64
		shareInAmount                 sdk.Int
		tokenOutMins                  sdk.Coins
		expectError                   bool
		expectedRemoveLiquidityEvents int
		expectedMessageEvents         int
	}{
		"successful exit": {
			poolId:                        1,
			shareInAmount:                 sdk.NewInt(shareIn),
			tokenOutMins:                  sdk.NewCoins(),
			expectedRemoveLiquidityEvents: 1,
			expectedMessageEvents:         3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"invalid tokenOutMins": {
			poolId:        1,
			shareInAmount: sdk.NewInt(shareIn),
			tokenOutMins:  sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tokenOutMinAmount))),
			expectError:   true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			sender := suite.TestAccs[0].String()

			// Pre-join pool to be able to ExitPool.
			joinPoolResponse, err := msgServer.JoinPool(sdk.WrapSDKContext(ctx), &types.MsgJoinPool{
				Sender:         sender,
				PoolId:         tc.poolId,
				ShareOutAmount: sdk.NewInt(shareIn),
				TokenInMaxs: sdk.NewCoins(
					sdk.NewCoin("foo", sdk.NewInt(int64Max)),
					sdk.NewCoin("bar", sdk.NewInt(int64Max)),
					sdk.NewCoin("baz", sdk.NewInt(int64Max)),
					sdk.NewCoin("uosmo", sdk.NewInt(int64Max)),
				),
			})
			suite.Require().NoError(err)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))

			// System under test.
			response, err := msgServer.ExitPool(sdk.WrapSDKContext(ctx), &types.MsgExitPool{
				Sender:        sender,
				PoolId:        tc.poolId,
				ShareInAmount: joinPoolResponse.ShareOutAmount,
				TokenOutMins:  tc.tokenOutMins,
			})

			if !tc.expectError {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, tc.expectedRemoveLiquidityEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

func (suite *KeeperTestSuite) TestMsgMigrateShares() {
	defaultAccount := suite.TestAccs[0]
	defaultGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("100000000000000000000").RoundInt())
	invalidGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("100000000000000000001").RoundInt())
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200000000000)), sdk.NewCoin("usdc", sdk.NewInt(200000000000)))
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(100),
		RoundingDir:       osmomath.RoundDown,
	}

	type param struct {
		sender                sdk.AccAddress
		sharesToMigrateDenom  string
		sharesToMigrateAmount sdk.Int
		poolIdEntering        uint64
	}

	tests := []struct {
		name                       string
		param                      param
		expectedErr                error
		sharesToCreate             sdk.Int
		expectedMigrateShareEvents int
		expectedMessageEvents      int
		expectedPosition           *model.Position
		errTolerance               osmomath.ErrTolerance
	}{
		{
			name: "migrate all of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
				poolIdEntering:        2,
			},
			sharesToCreate:             defaultGammShares.Amount,
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
			expectedPosition:           &model.Position{Liquidity: sdk.MustNewDecFromStr("100000000000.000000010000000000")},
			errTolerance:               defaultErrorTolerance,
		},
		{
			name: "migrate half of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
				poolIdEntering:        2,
			},
			sharesToCreate:             defaultGammShares.Amount,
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
			expectedPosition:           &model.Position{Liquidity: sdk.MustNewDecFromStr("50000000000.000000005000000000")},
			errTolerance:               defaultErrorTolerance,
		},
		{
			name: "double the created shares, migrate 1/4 of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
				poolIdEntering:        2,
			},
			sharesToCreate:             defaultGammShares.Amount.Mul(sdk.NewInt(2)),
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
			expectedPosition:           &model.Position{Liquidity: sdk.MustNewDecFromStr("49999999999.000000004999999999")},
			errTolerance:               defaultErrorTolerance,
		},
		{
			name: "error: attempt to migrate shares from non-existent pool",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  "gamm/pool/1000",
				sharesToMigrateAmount: defaultGammShares.Amount,
				poolIdEntering:        2,
			},
			sharesToCreate: defaultGammShares.Amount,
			expectedErr:    fmt.Errorf("pool with ID %d does not exist", 1000),
		},
		{
			name: "error: attempt to migrate shares to non-existent pool",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
				poolIdEntering:        3,
			},
			sharesToCreate: defaultGammShares.Amount,
			expectedErr:    cltypes.PoolNotFoundError{PoolId: 3},
		},
		{
			name: "error: attempt to migrate more shares than the user has",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Add(sdk.NewInt(1)),
				poolIdEntering:        2,
			},
			sharesToCreate: defaultGammShares.Amount,
			expectedErr:    sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", defaultGammShares, invalidGammShares)),
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		msgServer := keeper.NewBalancerMsgServerImpl(suite.App.GAMMKeeper)

		// Prepare both balancer and concentrated pools
		suite.FundAcc(test.param.sender, defaultAccountFunds)
		gammPoolId := suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
		gammPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, gammPoolId)
		suite.Require().NoError(err)
		clPool := suite.PrepareConcentratedPool()

		// Note gamm and cl pool addresses
		gammPoolAddress := gammPool.GetAddress()
		clPoolAddress := clPool.GetAddress()
		minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())

		// Join gamm pool to create gamm shares directed in the test case
		_, _, err = suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, test.param.sender, gammPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		suite.Require().NoError(err)

		// Note gamm pool balance after joining gamm pool
		gammPoolEthBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, ETH)
		gammPoolUsdcBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, USDC)

		// Note users gamm share balance after joining gamm pool
		userGammBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, "gamm/pool/1")

		// Create migrate message
		sharesToMigrate := sdk.NewCoin(test.param.sharesToMigrateDenom, test.param.sharesToMigrateAmount)
		msg := &balancer.MsgMigrateSharesToFullRangeConcentratedPosition{
			Sender:          test.param.sender.String(),
			SharesToMigrate: sharesToMigrate,
			PoolIdEntering:  test.param.poolIdEntering,
		}

		// Reset event counts to 0 by creating a new manager.
		suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
		suite.Require().Equal(0, len(suite.Ctx.EventManager().Events()))

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		resp, err := msgServer.MigrateSharesToFullRangeConcentratedPosition(sdk.WrapSDKContext(suite.Ctx), msg)
		if test.expectedErr != nil {
			suite.Require().Error(err)
			suite.Require().ErrorContains(err, test.expectedErr.Error())

			// Assure the user's gamm shares still exist
			userGammBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, "gamm/pool/1")
			suite.Require().Equal(userGammBalancePostJoin.String(), userGammBalanceAfterFailedMigration.String())

			// Assure cl pool has no balance after a failed migration.
			clPoolEthBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
			clPoolUsdcBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)
			suite.Require().Equal(sdk.NewInt(0), clPoolEthBalanceAfterFailedMigration.Amount)
			suite.Require().Equal(sdk.NewInt(0), clPoolUsdcBalanceAfterFailedMigration.Amount)

			// Assure the position was not created.
			_, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(suite.Ctx, clPool.GetId(), test.param.sender, minTick, maxTick)
			suite.Require().Error(err)
			continue
		}
		suite.Require().NoError(err)

		// Assure the expected position was created.
		position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(suite.Ctx, clPool.GetId(), test.param.sender, minTick, maxTick)
		suite.Require().NoError(err)
		suite.Require().Equal(test.expectedPosition, position)

		// Assert events are emitted
		suite.AssertEventEmitted(suite.Ctx, types.TypeEvtMigrateShares, test.expectedMigrateShareEvents)
		suite.AssertEventEmitted(suite.Ctx, sdk.EventTypeMessage, test.expectedMessageEvents)

		// Note gamm pool balance after migration
		gammPoolEthBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, ETH)
		gammPoolUsdcBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, USDC)

		// Note user amount transferred to cl pool from gamm pool
		userEthBalanceTransferredToClPool := gammPoolEthBalancePostJoin.Sub(gammPoolEthBalancePostMigrate)
		userUsdcBalanceTransferredToClPool := gammPoolUsdcBalancePostJoin.Sub(gammPoolUsdcBalancePostMigrate)

		// Note cl pool balance after migration
		clPoolEthBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)

		// The balance in the cl pool should be equal to what the user previously had in the gamm pool.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		suite.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, clPoolEthBalanceAfterMigration.Amount))
		suite.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, clPoolUsdcBalanceAfterMigration.Amount))

		// Assert user amount transferred to cl pool from gamm pool should be equal to the amount we migrated from the migrate message.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		suite.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, resp.Amount0))
		suite.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, resp.Amount1))
	}
}
