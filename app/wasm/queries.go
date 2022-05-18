package wasm

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmbindings "github.com/osmosis-labs/osmosis/v9/app/wasm/bindings"
	gammkeeper "github.com/osmosis-labs/osmosis/v9/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v9/x/gamm/types"
)

type QueryPlugin struct {
	gammKeeper *gammkeeper.Keeper
}

// NewQueryPlugin constructor
func NewQueryPlugin(
	gammK *gammkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		gammKeeper: gammK,
	}
}

func (qp QueryPlugin) GetPoolState(ctx sdk.Context, poolID uint64) (*wasmbindings.PoolAssets, error) {
	poolData, err := qp.gammKeeper.GetPoolAndPoke(ctx, poolID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get pool")
	}
	var poolAssets wasmbindings.PoolAssets
	poolAssets.Assets = poolData.GetTotalPoolLiquidity(ctx)
	poolAssets.Shares = sdk.Coin{
		Denom:  gammtypes.GetPoolShareDenom(poolID),
		Amount: poolData.GetTotalShares(),
	}
	return &poolAssets, nil
}

func (qp QueryPlugin) GetSpotPrice(ctx sdk.Context, spotPrice *wasmbindings.SpotPrice) (*sdk.Dec, error) {
	if spotPrice == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm spot price null"}
	}
	poolId := spotPrice.Swap.PoolId
	denomIn := spotPrice.Swap.DenomIn
	denomOut := spotPrice.Swap.DenomOut
	withSwapFee := spotPrice.WithSwapFee
	price, err := qp.gammKeeper.CalculateSpotPrice(ctx, poolId, denomIn, denomOut)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get spot price")
	}
	if withSwapFee {
		poolData, err := qp.gammKeeper.GetPoolAndPoke(ctx, poolId)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm get pool")
		}
		price = price.Mul(sdk.OneDec().Sub(poolData.GetSwapFee(ctx)))
	}
	return &price, nil
}

func (qp QueryPlugin) EstimateSwap(ctx sdk.Context, estimateSwap *wasmbindings.EstimateSwap) (*wasmbindings.SwapAmount, error) {
	if estimateSwap == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm estimate swap null"}
	}
	if err := sdk.ValidateDenom(estimateSwap.First.DenomIn); err != nil {
		return nil, sdkerrors.Wrap(err, "gamm estimate swap denom in")
	}
	if err := sdk.ValidateDenom(estimateSwap.First.DenomOut); err != nil {
		return nil, sdkerrors.Wrap(err, "gamm estimate swap denom out")
	}
	senderAddr, err := sdk.AccAddressFromBech32(estimateSwap.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm estimate swap sender address")
	}

	if estimateSwap.Amount == (wasmbindings.SwapAmount{}) {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm estimate swap empty swap"}
	}

	estimate, err := PerformSwap(qp.gammKeeper, ctx, senderAddr, estimateSwap.ToSwapMsg())
	return estimate, err
}
