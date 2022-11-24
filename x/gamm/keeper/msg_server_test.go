package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

const (
	// Max positive int64.
	int64Max = int64(^uint64(0) >> 1)
)

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
