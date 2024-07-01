package poolstransformer

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"
)

const (
	listAssetConfigsQueryString = `{"list_asset_configs":{}}`
	getShareDenomQueryString    = `{"get_share_denom":{}}`
)

// updateAlloyTransmuterInfo updates cosmwasmPoolModel with alloyed transmuter specific info.
// - It queries alloyed transmuter contract asset configs and share denom, the construct
// `AlloyTransmuterData`. Share denom for alloyed transmuter is the alloyed denom.
// - append the alloyed denom to pool denoms.
func (pi *poolTransformer) updateAlloyTransmuterInfo(
	ctx sdk.Context,
	poolId uint64,
	contractAddress sdk.AccAddress,
	cosmWasmPoolModel *sqscosmwasmpool.CosmWasmPoolModel,
	poolDenoms *[]string,
) error {
	assetConfigs, err := alloyTransmuterListAssetConfig(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	// share denom of alloy transmuter pool is an alloyed denom
	alloyedDenom, err := alloyTransmuterGetShareDenom(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	// append alloyed denom to denoms
	*poolDenoms = append(*poolDenoms, alloyedDenom)

	cosmWasmPoolModel.Data.AlloyTransmuter = &sqscosmwasmpool.AlloyTransmuterData{
		AlloyedDenom: alloyedDenom,
		AssetConfigs: assetConfigs,
	}

	return nil
}

// alloyTransmuterListAssetConfig queries the asset configs of the alloyed transmuter contract.
func alloyTransmuterListAssetConfig(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) ([]sqscosmwasmpool.TransmuterAssetConfig, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(listAssetConfigsQueryString))
	if err != nil {
		return nil, fmt.Errorf(
			"error querying list_asset_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}
	var assetConfigsResponse struct {
		AssetConfigs []sqscosmwasmpool.TransmuterAssetConfig `json:"asset_configs"`
	}

	if err := json.Unmarshal(bz, &assetConfigsResponse); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling asset_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return assetConfigsResponse.AssetConfigs, nil
}

// alloyTransmuterGetShareDenom queries the share denom of the alloyed transmuter contract.
func alloyTransmuterGetShareDenom(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (string, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(getShareDenomQueryString))
	if err != nil {
		return "", fmt.Errorf(
			"error querying get_share_denom for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	var getShareDenomResponse struct {
		ShareDenom string `json:"share_denom"`
	}

	if err := json.Unmarshal(bz, &getShareDenomResponse); err != nil {
		return "", fmt.Errorf(
			"error unmarshalling share_denom for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return getShareDenomResponse.ShareDenom, nil
}
