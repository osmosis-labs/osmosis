package keeper

import (
	"encoding/hex"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// EmitBackrunEvent updates and emits a backrunEvent
func EmitBackrunEvent(ctx sdk.Context, backrunEvent sdk.Event, inputCoin sdk.Coin, profit, tokenOutAmount sdk.Int) {
	// Update the backrun event and add it to the context
	backrunEvent = backrunEvent.AppendAttributes(
		sdk.NewAttribute(types.AttributeKeyProtorevProfit, profit.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevAmountIn, inputCoin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevAmountOut, tokenOutAmount.String()),
		sdk.NewAttribute(types.AttributeKeyProtorevArbDenom, inputCoin.Denom),
	)
	ctx.EventManager().EmitEvent(backrunEvent)
}

// CreateBackrunEvent creates a backrun event to be emitted if the trade is executed successfully
func (k Keeper) CreateBackrunEvent(ctx sdk.Context, pool SwapToBackrun, remainingTxPoolPoints uint64) (sdk.Event, error) {
	// Get pool points remaning in block
	remainingBlockPoolPoints, err := k.remainingPoolPointsForBlock(ctx)
	if err != nil {
		return sdk.Event{}, err
	}
	// Get tx hash
	txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(ctx.TxBytes())))
	// Create backrun event to be emitted if the trade is executed successfully
	return sdk.NewEvent(
		types.TypeEvtBackrun,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeyTxHash, txHash),
		sdk.NewAttribute(types.AttributeKeyUserPoolId, strconv.FormatUint(pool.PoolId, 10)),
		sdk.NewAttribute(types.AttributeKeyUserDenomIn, pool.TokenInDenom),
		sdk.NewAttribute(types.AttributeKeyUserDenomOut, pool.TokenOutDenom),
		sdk.NewAttribute(types.AttributeKeyTxPoolPointsRemaining, strconv.FormatUint(remainingTxPoolPoints, 10)),
		sdk.NewAttribute(types.AttributeKeyBlockPoolPointsRemaining, strconv.FormatUint(remainingBlockPoolPoints, 10)),
	), nil
}

// RemainingPoolPointsForBlock calculates the number of pool points that can be consumed in the current block.
func (k Keeper) remainingPoolPointsForBlock(ctx sdk.Context) (uint64, error) {
	maxPoolPointsPerBlock, err := k.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return 0, err
	}

	currentPoolPointCount, err := k.GetPointCountForBlock(ctx)
	if err != nil {
		return 0, err
	}

	return maxPoolPointsPerBlock - currentPoolPointCount, nil
}
