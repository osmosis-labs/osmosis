package wasm

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	bindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
)

func CustomQuerier(osmoKeeper *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.OsmosisQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "osmosis query")
		}

		if contractQuery.FullDenom != nil {
			contract := contractQuery.FullDenom.Contract
			subdenom := contractQuery.FullDenom.Subdenom

			fullDenom, err := GetFullDenom(contract, subdenom)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo full denom query")
			}

			res := bindings.FullDenomResponse{
				Denom: fullDenom,
			}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo full denom query response")
			}
			return bz, nil
		} else if contractQuery.PoolState != nil {
			poolId := contractQuery.PoolState.PoolId

			state, err := osmoKeeper.GetPoolState(ctx, poolId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo pool state query")
			}

			assets := ConvertSdkCoinsToWasmCoins(state.Assets)
			shares := ConvertSdkCoinToWasmCoin(state.Shares)

			res := bindings.PoolStateResponse{
				Assets: assets,
				Shares: shares,
			}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo pool state query response")
			}
			return bz, nil
		} else if contractQuery.SpotPrice != nil {
			spotPrice, err := osmoKeeper.GetSpotPrice(ctx, contractQuery.SpotPrice)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo spot price query")
			}

			res := bindings.SpotPriceResponse{Price: spotPrice.String()}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo spot price query response")
			}
			return bz, nil
		} else if contractQuery.EstimateSwap != nil {
			swapAmount, err := osmoKeeper.EstimateSwap(ctx, contractQuery.EstimateSwap)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo estimate swap query")
			}

			res := bindings.EstimatePriceResponse{Amount: *swapAmount}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo estimate swap query response")
			}
			return bz, nil
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown osmosis query variant"}
	}
}

// ConvertSdkCoinsToWasmCoins converts sdk type coins to wasm vm type coins
func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) wasmvmtypes.Coins {
	var toSend wasmvmtypes.Coins
	for _, coin := range coins {
		c := ConvertSdkCoinToWasmCoin(coin)
		toSend = append(toSend, c)
	}
	return toSend
}

// ConvertSdkCoinToWasmCoin converts a sdk type coin to a wasm vm type coin
func ConvertSdkCoinToWasmCoin(coin sdk.Coin) wasmvmtypes.Coin {
	return wasmvmtypes.Coin{
		Denom: coin.Denom,
		// Note: gamm tokens have 18 decimal places, so 10^22 is common, no longer in u64 range
		Amount: coin.Amount.String(),
	}
}
