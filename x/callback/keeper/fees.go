package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EstimateCallbackFees returns the fees that will be charged for registering a callback at the given block height
// The returned value is in the order of:
// 1. Future reservation fees
// 2. Block reservation fees
// 3. Transaction fees
// 4. Errors, if any
func (k Keeper) EstimateCallbackFees(ctx sdk.Context, blockHeight int64) (sdk.Coin, sdk.Coin, sdk.Coin, error) {
	if blockHeight <= ctx.BlockHeight() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, status.Errorf(codes.InvalidArgument, "block height %d is not in the future", blockHeight)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	// Calculates the fees based on how far in the future the callback is registered
	futureReservationThreshold := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if blockHeight > futureReservationThreshold {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, status.Errorf(codes.OutOfRange, "block height %d is too far in the future. max block height callback can be registered at %d", blockHeight, futureReservationThreshold)
	}
	// futureReservationFeeMultiplies * (requestBlockHeight - currentBlockHeight)
	futureReservationFeesAmount := params.FutureReservationFeeMultiplier.MulInt64(blockHeight - ctx.BlockHeight())

	// Calculates the fees based on how many callbacks are registered at the given block height
	callbacksForHeight, err := k.GetCallbacksByHeight(ctx, blockHeight)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, status.Errorf(codes.NotFound, "could not fetch callbacks for given height: %s", err.Error())
	}
	totalCallbacks := len(callbacksForHeight)
	if totalCallbacks >= int(params.MaxBlockReservationLimit) {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, status.Errorf(codes.OutOfRange, "block height %d has reached max reservation limit", blockHeight)
	}
	// blockReservationFeeMultiplier * totalCallbacksRegistered
	blockReservationFeesAmount := params.BlockReservationFeeMultiplier.MulInt64(int64(totalCallbacks))

	// Calculates the fees based on the max gas limit of the callback and current price of gas
	transactionFee := k.CalculateTransactionFees(ctx, params.GetCallbackGasLimit(), params.GetMinPriceOfGas())
	futureReservationFee := sdk.NewCoin(transactionFee.Denom, futureReservationFeesAmount.RoundInt())
	blockReservationFee := sdk.NewCoin(transactionFee.Denom, blockReservationFeesAmount.RoundInt())
	return futureReservationFee, blockReservationFee, transactionFee, nil
}

func (k Keeper) CalculateTransactionFees(_ sdk.Context, gasAmount uint64, minPriceOfGas sdk.Coin) sdk.Coin {
	transactionFeeAmount := minPriceOfGas.Amount.MulRaw(int64(gasAmount))
	transactionFee := sdk.NewCoin(minPriceOfGas.Denom, transactionFeeAmount)
	return transactionFee
}
