package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	doesNotExistDenom = "nodenom"
	// Max positive int64.
	int64Max = int64(^uint64(0) >> 1)
)

// TestSwapExactAmountIn_Events tests that events are correctly emitted
// when calling SwapExactAmountIn.
func (s *KeeperTestSuite) TestSwapExactAmountIn_Events() {
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
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountIn(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountIn{
				Sender:            s.TestAccs[0].String(),
				Routes:            tc.routes,
				TokenIn:           tc.tokenIn,
				TokenOutMinAmount: tc.tokenOutMinAmount,
			})

			if !tc.expectError {
				s.NoError(err)
				s.NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestSwapExactAmountOut_Events tests that events are correctly emitted
// when calling SwapExactAmountOut.
func (s *KeeperTestSuite) TestSwapExactAmountOut_Events() {
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
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountOut(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountOut{
				Sender:           s.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOut:         tc.tokenOut,
				TokenInMaxAmount: tc.tokenInMaxAmount,
			})

			if !tc.expectError {
				s.NoError(err)
				s.NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestJoinPool_Events tests that events are correctly emitted
// when calling JoinPool.
func (s *KeeperTestSuite) TestJoinPool_Events() {
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
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.JoinPool(sdk.WrapSDKContext(ctx), &types.MsgJoinPool{
				Sender:         s.TestAccs[0].String(),
				PoolId:         tc.poolId,
				ShareOutAmount: tc.shareOutAmount,
				TokenInMaxs:    tc.tokenInMaxs,
			})

			if !tc.expectError {
				s.Require().NoError(err)
				s.Require().NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, tc.expectedAddLiquidityEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestExitPool_Events tests that events are correctly emitted
// when calling ExitPool.
func (s *KeeperTestSuite) TestExitPool_Events() {
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
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			sender := s.TestAccs[0].String()

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
			s.Require().NoError(err)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))

			// System under test.
			response, err := msgServer.ExitPool(sdk.WrapSDKContext(ctx), &types.MsgExitPool{
				Sender:        sender,
				PoolId:        tc.poolId,
				ShareInAmount: joinPoolResponse.ShareOutAmount,
				TokenOutMins:  tc.tokenOutMins,
			})

			if !tc.expectError {
				s.Require().NoError(err)
				s.Require().NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, tc.expectedRemoveLiquidityEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

func (s *KeeperTestSuite) TestMsgMigrateShares_Events() {
	defaultAccount := s.TestAccs[0]
	defaultGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("100000000000000000000").RoundInt())
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200000000000)), sdk.NewCoin("usdc", sdk.NewInt(200000000000)))

	type param struct {
		sender                sdk.AccAddress
		sharesToMigrateDenom  string
		sharesToMigrateAmount sdk.Int
	}

	tests := []struct {
		name                       string
		param                      param
		sharesToCreate             sdk.Int
		expectedMigrateShareEvents int
		expectedMessageEvents      int
		expectError                bool
	}{
		{
			name: "migrate all of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:             defaultGammShares.Amount,
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      4, // 1 create pool, 1 exitPool, 1 createPosition, 1 migrateShares.
		},
		{
			name: "migrate half of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
			},
			sharesToCreate:             defaultGammShares.Amount,
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      4, // 1 create pool, 1 exitPool, 1 createPosition, 1 migrateShares.
		},
		{
			name: "double the created shares, migrate 1/4 of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
			},
			sharesToCreate:             defaultGammShares.Amount.Mul(sdk.NewInt(2)),
			expectedMigrateShareEvents: 1,
			expectedMessageEvents:      4, // 1 create pool, 1 exitPool, 1 createPosition, 1 migrateShares.
		},
		{
			name: "error: attempt to migrate shares from non-existent pool",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  "gamm/pool/1000",
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate: defaultGammShares.Amount,
			expectError:    true,
		},
		{
			name: "error: attempt to migrate more shares than the user has",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Add(sdk.NewInt(1)),
			},
			sharesToCreate:        defaultGammShares.Amount,
			expectedMessageEvents: 1, // 1 create pool
			expectError:           true,
		},
	}

	for _, test := range tests {
		s.SetupTest()
		msgServer := keeper.NewBalancerMsgServerImpl(s.App.GAMMKeeper)

		// Prepare both balancer and concentrated pools
		s.FundAcc(test.param.sender, defaultAccountFunds)
		balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
		clPool := s.PrepareConcentratedPool()

		// Set up migration records
		record := types.BalancerToConcentratedPoolLink{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()}
		err := s.App.GAMMKeeper.ReplaceMigrationRecords(s.Ctx, []types.BalancerToConcentratedPoolLink{record})
		s.Require().NoError(err)

		// Join gamm pool to create gamm shares directed in the test case
		_, _, err = s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, test.param.sender, balancerPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		s.Require().NoError(err)

		// Create migrate message
		sharesToMigrate := sdk.NewCoin(test.param.sharesToMigrateDenom, test.param.sharesToMigrateAmount)
		msg := &balancer.MsgMigrateSharesToFullRangeConcentratedPosition{
			Sender:          test.param.sender.String(),
			SharesToMigrate: sharesToMigrate,
		}

		// Reset event counts to 0 by creating a new manager.
		s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
		s.Require().Equal(0, len(s.Ctx.EventManager().Events()))

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		response, err := msgServer.MigrateSharesToFullRangeConcentratedPosition(sdk.WrapSDKContext(s.Ctx), msg)

		if !test.expectError {
			s.NoError(err)
			s.NotNil(response)
			s.AssertEventEmitted(s.Ctx, types.TypeEvtMigrateShares, test.expectedMigrateShareEvents)
			s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, test.expectedMessageEvents)
		} else {
			s.Require().Error(err)
			s.Require().Nil(response)
		}
	}
}
