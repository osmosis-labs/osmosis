package keeper

import (
	"encoding/hex"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// EmitBackrunEvent updates and emits a backrunEvent
func EmitBackrunEvent(ctx sdk.Context, pool SwapToBackrun, inputCoin sdk.Coin, profit, tokenOutAmount osmomath.Int, remainingTxPoolPoints, remainingBlockPoolPoints uint64) {
	// Get tx hash
	txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(ctx.TxBytes())))
	// Update the backrun event and add it to the context
	backrunEvent := sdk.NewEvent(
		types.TypeEvtBackrun,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeyTxHash, txHash),
		sdk.NewAttribute(types.AttributeKeyUserPoolId, strconv.FormatUint(pool.PoolId, 10)),
		sdk.NewAttribute(types.AttributeKeyUserDenomIn, pool.TokenInDenom),
		sdk.NewAttribute(types.AttributeKeyUserDenomOut, pool.TokenOutDenom),
		sdk.NewAttribute(types.AttributeKeyTxPoolPointsRemaining, strconv.FormatUint(remainingTxPoolPoints, 10)),
		sdk.NewAttribute(types.AttributeKeyBlockPoolPointsRemaining, strconv.FormatUint(remainingBlockPoolPoints, 10)),
		sdk.NewAttribute(types.AttributeKeyProtorevProfit, profit.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevAmountIn, inputCoin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevAmountOut, tokenOutAmount.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevArbDenom, inputCoin.Denom),
	)
	ctx.EventManager().EmitEvent(backrunEvent)
}
