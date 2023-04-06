package keeper_test

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

func (suite *KeeperTestSuite) TestBackRunEvent() {
	testcases := map[string]struct {
		pool                     keeper.SwapToBackrun
		remainingTxPoolPoints    uint64
		remainingBlockPoolPoints uint64
		profit                   sdk.Int
		tokenOutAmount           sdk.Int
		inputCoin                sdk.Coin
	}{
		"basic valid": {
			pool: keeper.SwapToBackrun{
				PoolId:        1,
				TokenInDenom:  "uosmo",
				TokenOutDenom: "uatom",
			},
			remainingTxPoolPoints:    100,
			remainingBlockPoolPoints: 100,
			profit:                   sdk.NewInt(100),
			tokenOutAmount:           sdk.NewInt(100),
			inputCoin:                sdk.NewCoin("uosmo", sdk.NewInt(100)),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvent := sdk.NewEvent(
				types.TypeEvtBackrun,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyUserDenomIn, tc.pool.TokenInDenom),
				sdk.NewAttribute(types.AttributeKeyUserDenomOut, tc.pool.TokenOutDenom),
				sdk.NewAttribute(types.AttributeKeyTxPoolPointsRemaining, strconv.FormatUint(tc.remainingTxPoolPoints, 10)),
				sdk.NewAttribute(types.AttributeKeyBlockPoolPointsRemaining, strconv.FormatUint(tc.remainingBlockPoolPoints, 10)),
			)

			actualEvent, err := suite.App.ProtoRevKeeper.CreateBackrunEvent(suite.Ctx, tc.pool, tc.remainingTxPoolPoints)
			suite.Require().NoError(err)

			suite.Equal(expectedEvent, actualEvent)

			// Append the extra attributes added in the EmitBackrunEvent function
			expectedUpdatedEvent := expectedEvent.AppendAttributes(
				sdk.NewAttribute(types.AttributeKeyProtorevProfit, tc.profit.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevAmountIn, tc.inputCoin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevAmountOut, tc.tokenOutAmount.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevArbDenom, tc.inputCoin.Denom),
			)

			keeper.EmitBackrunEvent(suite.Ctx, actualEvent, tc.inputCoin, tc.profit, tc.tokenOutAmount)

			// Get last event emitted and ensure it is the expected event
			actualUpdatedEvent := suite.Ctx.EventManager().Events()[len(suite.Ctx.EventManager().Events())-1]
			suite.Equal(expectedUpdatedEvent, actualUpdatedEvent)
		})
	}
}
