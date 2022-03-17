package wasm

import (
	"fmt"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	bindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
	"github.com/osmosis-labs/osmosis/v7/app/wasm/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
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

func (qp QueryPlugin) GetFullDenom(ctx sdk.Context, contract string, subDenom string) (*string, error) {
	// Address validation
	contractAddress, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "address from bech32")
	}
	err = sdk.VerifyAddressFormat(contractAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "verify address format")
	}
	// TODO: sub-denom validations
	// - sub denom length (min/max) checks
	// - sub denom chars
	// -
	// TODO: Confirm "cw" prefix
	fullDenom := fmt.Sprintf("cw/%s/%s", contract, subDenom)

	return &fullDenom, nil
}

func (qp QueryPlugin) GetPoolState(ctx sdk.Context, poolId uint64) (*types.PoolState, error) {
	poolData, err := qp.gammKeeper.GetPool(ctx, poolId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get pool")
	}
	var poolState types.PoolState
	poolAssets := poolData.GetAllPoolAssets()
	for _, asset := range poolAssets {
		poolState.Assets = append(poolState.Assets, asset.Token)
	}
	poolState.Shares = poolData.GetTotalShares()

	return &poolState, nil
}

func (qp QueryPlugin) GetSpotPrice(ctx sdk.Context, spotPrice *bindings.SpotPrice) (*sdk.Dec, error) {
	poolId := spotPrice.Swap.PoolId
	denomIn := spotPrice.Swap.DenomIn
	denomOut := spotPrice.Swap.DenomOut
	withSwapFee := spotPrice.WithSwapFee
	var price sdk.Dec
	var err error
	if withSwapFee {
		price, err = qp.gammKeeper.CalculateSpotPriceWithSwapFee(ctx, poolId, denomIn, denomOut)
	} else {
		price, err = qp.gammKeeper.CalculateSpotPrice(ctx, poolId, denomIn, denomOut)
	}
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get spot price")
	}
	return &price, nil
}

func (qp QueryPlugin) EstimatePrice(ctx sdk.Context, estimatePrice *bindings.EstimatePrice) (*bindings.SwapAmount, error) {
	sender := estimatePrice.Contract
	poolId := estimatePrice.First.PoolId
	denomIn := estimatePrice.First.DenomIn
	denomOut := estimatePrice.First.DenomOut
	route := estimatePrice.Route
	if len(route) != 0 {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "TODO: multi-hop swap price estimation"}
	}

	var senderAddress = sdk.AccAddress(sender)

	if estimatePrice.Amount.In != nil {
		tokenIn := sdk.NewCoin(denomIn, *estimatePrice.Amount.In)

		// Populate route
		var steps gammtypes.SwapAmountInRoutes
		firstStep := gammtypes.SwapAmountInRoute{
			PoolId:        poolId,
			TokenOutDenom: denomOut,
		}
		steps = append(steps, firstStep)

		tokenOutMinAmount := sdk.NewInt(1)
		estimatedAmount, err := qp.gammKeeper.MultihopSwapExactAmountIn(ctx, senderAddress, steps, tokenIn, tokenOutMinAmount)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm estimate price exact amount in")
		}
		return &bindings.SwapAmount{Out: &estimatedAmount}, nil
	} else if estimatePrice.Amount.Out != nil {
		tokenOut := sdk.NewCoin(denomOut, *estimatePrice.Amount.Out)

		// Populate route
		var steps gammtypes.SwapAmountOutRoutes
		firstStep := gammtypes.SwapAmountOutRoute{
			PoolId:       poolId,
			TokenInDenom: denomIn,
		}
		steps = append(steps, firstStep)

		tokenInMaxAmount := sdk.NewInt(math.MaxInt64)
		estimatedAmount, err := qp.gammKeeper.MultihopSwapExactAmountOut(ctx, senderAddress, steps, tokenInMaxAmount, tokenOut)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm estimate price exact amount out")
		}
		return &bindings.SwapAmount{In: &estimatedAmount}, nil
	}
	return nil, wasmvmtypes.UnsupportedRequest{Kind: "must support either EstimatePrice.Amount.In or EstimatePrice.Amount.Out"}
}
