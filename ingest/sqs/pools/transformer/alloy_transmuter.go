package poolstransformer

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

// updateAlloyTrasmuterInfo updates cosmwasmPoolModel with alloyed transmuter specific info.
// - It queries alloyed transmuter contract asset configs and share denom, the construct
// `AlloyTransmuterData`. Share denom for alloyed transmuter is the alloyed denom.
// - append the alloyed denom to pool denoms.
func (pi *poolTransformer) updateAlloyTrasmuterInfo(
	ctx sdk.Context,
	poolId uint64,
	contractAddress sdk.AccAddress,
	cosmWasmPoolModel *sqsdomain.CosmWasmPoolModel,
	poolDenoms *[]string,
) error {
	assetConfigs, err := alloyedTransmuterListAssetConfig(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	// share denom of alloyed transmuter pool is an alloyed denom
	alloyedDenom, err := alloyedTransmuterGetShareDenom(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	// append alloyed denom to denoms
	*poolDenoms = append(*poolDenoms, alloyedDenom)

	cosmWasmPoolModel.Data.AlloyTransmuter = &sqsdomain.AlloyTransmuterData{
		AlloyedDenom: alloyedDenom,
		AssetConfigs: assetConfigs,
	}

	return nil
}

// alloyedTransmuterListAssetConfig queries the asset configs of the alloyed transmuter contract.
func alloyedTransmuterListAssetConfig(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) ([]sqsdomain.TransmuterAssetConfig, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(`{"list_asset_configs":{}}`))
	if err != nil {
		return nil, fmt.Errorf(
			"error querying list_asset_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}
	var assetConfigsResponse struct {
		AssetConfigs []sqsdomain.TransmuterAssetConfig `json:"asset_configs"`
	}

	if err := json.Unmarshal(bz, &assetConfigsResponse); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling asset_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return assetConfigsResponse.AssetConfigs, nil
}

// alloyedTransmuterGetShareDenom queries the share denom of the alloyed transmuter contract.
func alloyedTransmuterGetShareDenom(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (string, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(`{"get_share_denom":{}}`))
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
