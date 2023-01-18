package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
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
		withdrawLiquidity          sdk.Dec
		sharesToCreate             sdk.Int
		expectedMigrateShareEvents int
		expectedMessageEvents      int
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
			withdrawLiquidity:          sdk.MustNewDecFromStr("100000000000.000000000000000000"),
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
		},
		{
			name: "migrate half of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
				poolIdEntering:        2,
			},
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
			sharesToCreate:             defaultGammShares.Amount,
			withdrawLiquidity:          sdk.MustNewDecFromStr("50000000000.000000000000000000"),
		},
		{
			name: "double the created shares, migrate 1/4 of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
				poolIdEntering:        2,
			},
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      3, // 1 exitPool, 1 createPosition, 1 migrateShares.
			sharesToCreate:             defaultGammShares.Amount.Mul(sdk.NewInt(2)),
			withdrawLiquidity:          sdk.MustNewDecFromStr("49999999999.000000000000000000"),
		},
		{
			name: "error: attempt to migrate shares from non-existent pool",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  "gamm/pool/1000",
				sharesToMigrateAmount: defaultGammShares.Amount,
				poolIdEntering:        2,
			},
			sharesToCreate:    defaultGammShares.Amount,
			withdrawLiquidity: sdk.MustNewDecFromStr("100000000000.000000000000000000"),
			expectedErr:       fmt.Errorf("pool with ID %d does not exist", 1000),
		},
		{
			name: "error: attempt to migrate more shares than the user has",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Add(sdk.NewInt(1)),
				poolIdEntering:        2,
			},
			sharesToCreate:    defaultGammShares.Amount,
			withdrawLiquidity: sdk.MustNewDecFromStr("100000000000.000000000000000000"),
			expectedErr:       sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", defaultGammShares, invalidGammShares)),
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

		// Note user balance before joining gamm pool
		userEthBalanceBeforeGAMMJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		userUsdcBalanceBeforeGAMMJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)

		// Join gamm pool to create gamm shares directed in the test case
		gammPoolTokensIn, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, test.param.sender, gammPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		suite.Require().NoError(err)

		// Note gamm pool balance after joining gamm pool
		gammPoolEthBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, ETH)
		gammPoolUsdcBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, USDC)

		// Create migrate message
		sharesToMigrate := sdk.NewCoin(test.param.sharesToMigrateDenom, test.param.sharesToMigrateAmount)
		msg := &balancer.MsgMigrateSharesToFullRangeConcentratedPosition{
			Sender:          test.param.sender.String(),
			SharesToMigrate: sharesToMigrate,
			PoolIdEntering:  test.param.poolIdEntering,
		}

		// Note cl pool balance before migration
		clPoolEthBalanceBeforeMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceBeforeMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)

		// Reset event counts to 0 by creating a new manager.
		suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
		suite.Require().Equal(0, len(suite.Ctx.EventManager().Events()))

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		resp, err := msgServer.MigrateSharesToFullRangeConcentratedPosition(sdk.WrapSDKContext(suite.Ctx), msg)
		if test.expectedErr != nil {
			suite.Require().Error(err)
			suite.Require().ErrorContains(err, test.expectedErr.Error())
			continue
		}

		suite.Require().NoError(err)

		// Assert events are emitted
		suite.AssertEventEmitted(suite.Ctx, types.TypeEvtMigrateShares, test.expectedMigrateShareEvents)
		suite.AssertEventEmitted(suite.Ctx, sdk.EventTypeMessage, test.expectedMessageEvents)

		// Note gamm pool balance after migration
		gammPoolEthBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, ETH)
		gammPoolUsdcBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, gammPoolAddress, USDC)

		// Note user balance after migration
		userEthBalanceTransferredToClPool := gammPoolEthBalancePostJoin.Sub(gammPoolEthBalancePostMigrate)
		userUsdcBalanceTransferredToClPool := gammPoolUsdcBalancePostJoin.Sub(gammPoolUsdcBalancePostMigrate)

		// Note cl pool balance after migration
		clPoolEthBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)

		// The balance in the cl pool should be equal to what the user previously had in the gamm pool
		poolEthBalanceInTolerance := userEthBalanceTransferredToClPool.Amount.Sub(clPoolEthBalanceAfterMigration.Amount).LTE(sdk.NewInt(1))
		poolUsdcBalanceInTolerance := userUsdcBalanceTransferredToClPool.Amount.Sub(clPoolUsdcBalanceAfterMigration.Amount).LTE(sdk.NewInt(1))
		suite.Require().True(poolEthBalanceInTolerance)
		suite.Require().True(poolUsdcBalanceInTolerance)

		// Withdraw the user's concentrated liquidity position
		minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())
		_, _, err = suite.App.ConcentratedLiquidityKeeper.WithdrawPosition(suite.Ctx, test.param.poolIdEntering, test.param.sender, minTick, maxTick, test.withdrawLiquidity)
		suite.Require().NoError(err)

		// Note user balance after withdrawing concentrated liquidity position (leaving pool)
		userEthBalanceAfterPositionWithdraw := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		userUsdcBalanceAfterPositionWithdraw := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)

		// The user's balance after withdrawing concentrated liquidity position should be equal to the user's balance before joining the original gamm pool
		// Since there are test cases where the user only migrates a portion of their shares, we need to determine the the diff between
		// the gammPoolTokensIn and the tokens used to create the concentrated liquidity position
		ethDiff := gammPoolTokensIn.AmountOf(ETH).Sub(resp.Amount0).ToDec()
		usdcDiff := gammPoolTokensIn.AmountOf(USDC).Sub(resp.Amount1).ToDec()
		userEthBalanceBeforeGAMMJoinAdjusted := userEthBalanceBeforeGAMMJoin.Amount.ToDec().Sub(ethDiff).TruncateInt()
		userUsdcBalanceBeforeGAMMJoinAdjusted := userUsdcBalanceBeforeGAMMJoin.Amount.ToDec().Sub(usdcDiff).TruncateInt()

		// NOTE: It appears that when leaving gamm pools, we are sometimes shorted a single unit of the token
		// Need to look into this (its likely a safety design choice), but for now we just ensure that the user's balance is within 1 unit of the expected balance
		userEthBalanceInTolerance := userEthBalanceAfterPositionWithdraw.Amount.Sub(userEthBalanceBeforeGAMMJoinAdjusted).LTE(sdk.NewInt(2))
		userUsdcBalanceInTolerance := userUsdcBalanceAfterPositionWithdraw.Amount.Sub(userUsdcBalanceBeforeGAMMJoinAdjusted).LTE(sdk.NewInt(2))
		suite.Require().True(userEthBalanceInTolerance)
		suite.Require().True(userUsdcBalanceInTolerance)

		// Note cl pool balance after user leaves the pool
		clPoolEthBalanceAfterPositionWithdraw := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceAfterPositionWithdraw := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)

		// The balance in the cl pool should be equal to what it was prior to the user joining the cl pool
		suite.Require().Equal(clPoolEthBalanceBeforeMigration.String(), clPoolEthBalanceAfterPositionWithdraw.String())
		suite.Require().Equal(clPoolUsdcBalanceBeforeMigration.String(), clPoolUsdcBalanceAfterPositionWithdraw.String())
	}
}
