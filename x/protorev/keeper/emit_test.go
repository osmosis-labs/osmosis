package keeper_test

import (
	"encoding/hex"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v20/x/protorev/types"
)

func (s *KeeperTestSuite) TestBackRunEvent() {
	testcases := map[string]struct {
		pool                     keeper.SwapToBackrun
		remainingTxPoolPoints    uint64
		remainingBlockPoolPoints uint64
		profit                   osmomath.Int
		tokenOutAmount           osmomath.Int
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
			profit:                   osmomath.NewInt(100),
			tokenOutAmount:           osmomath.NewInt(100),
			inputCoin:                sdk.NewCoin("uosmo", osmomath.NewInt(100)),
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			expectedEvent := sdk.NewEvent(
				types.TypeEvtBackrun,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyTxHash, strings.ToUpper(hex.EncodeToString(tmhash.Sum(s.Ctx.TxBytes())))),
				sdk.NewAttribute(types.AttributeKeyUserPoolId, strconv.FormatUint(tc.pool.PoolId, 10)),
				sdk.NewAttribute(types.AttributeKeyUserDenomIn, tc.pool.TokenInDenom),
				sdk.NewAttribute(types.AttributeKeyUserDenomOut, tc.pool.TokenOutDenom),
				sdk.NewAttribute(types.AttributeKeyTxPoolPointsRemaining, strconv.FormatUint(tc.remainingTxPoolPoints, 10)),
				sdk.NewAttribute(types.AttributeKeyBlockPoolPointsRemaining, strconv.FormatUint(tc.remainingBlockPoolPoints, 10)),
				sdk.NewAttribute(types.AttributeKeyProtorevProfit, tc.profit.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevAmountIn, tc.inputCoin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevAmountOut, tc.tokenOutAmount.String()),
				sdk.NewAttribute(types.AttributeKeyProtorevArbDenom, tc.inputCoin.Denom),
			)

			keeper.EmitBackrunEvent(s.Ctx, tc.pool, tc.inputCoin, tc.profit, tc.tokenOutAmount, tc.remainingTxPoolPoints, tc.remainingBlockPoolPoints)

			// Get last event emitted and ensure it is the expected event
			actualEvent := s.Ctx.EventManager().Events()[len(s.Ctx.EventManager().Events())-1]
			s.Equal(expectedEvent, actualEvent)
		})
	}
}
