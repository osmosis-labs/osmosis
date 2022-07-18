package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v10/wasmbinding/bindings"
)

func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.OsmosisQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "osmosis query")
		}

		switch {
		case contractQuery.FullDenom != nil:
			creator := contractQuery.FullDenom.CreatorAddr
			subdenom := contractQuery.FullDenom.Subdenom

			fullDenom, err := GetFullDenom(creator, subdenom)
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

		case contractQuery.DenomAdmin != nil:
			res, err := qp.GetDenomAdmin(ctx, contractQuery.DenomAdmin.Subdenom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal DenomAdminResponse response: %w", err)
			}

			return bz, nil

		case contractQuery.PoolState != nil:
			poolId := contractQuery.PoolState.PoolId

			state, err := qp.GetPoolState(ctx, poolId)
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

		case contractQuery.SpotPrice != nil:
			spotPrice, err := qp.GetSpotPrice(ctx, contractQuery.SpotPrice)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo spot price query")
			}

			res := bindings.SpotPriceResponse{Price: spotPrice.String()}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo spot price query response")
			}

			return bz, nil

		case contractQuery.EstimateSwap != nil:
			swapAmount, err := qp.EstimateSwap(ctx, contractQuery.EstimateSwap)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo estimate swap query")
			}

			res := bindings.EstimatePriceResponse{Amount: *swapAmount}
			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "osmo estimate swap query response")
			}

			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown osmosis query variant"}
		}
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
