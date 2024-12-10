package poolstransformer

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sqscosmwasmpool "github.com/osmosis-labs/osmosis/v28/ingest/types/cosmwasmpool"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
)

const (
	listAssetConfigsQueryString = `{"list_asset_configs":{}}`
	getShareDenomQueryString    = `{"get_share_denom":{}}`
	listLimitersQueryString     = `{"list_limiters":{}}`
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

	rateLimiterData, err := alloyTransmuterListLimiters(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	// append alloyed denom to denoms
	*poolDenoms = append(*poolDenoms, alloyedDenom)

	cosmWasmPoolModel.Data.AlloyTransmuter = &sqscosmwasmpool.AlloyTransmuterData{
		AlloyedDenom:      alloyedDenom,
		AssetConfigs:      assetConfigs,
		RateLimiterConfig: rateLimiterData,
	}

	return nil
}

// alloyTransmuterListAssetConfig queries the asset configs of the alloyed transmuter contract.
func alloyTransmuterListAssetConfig(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
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
	wasmKeeper commondomain.WasmKeeper,
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

type listLimitersResponse struct {
	Limiters [][2]json.RawMessage `json:"limiters"`
}

type LimiterInfo [2]string

type LimiterValue struct {
	StaticLimiter *sqscosmwasmpool.StaticLimiter `json:"static_limiter,omitempty"`
	ChangeLimiter *sqscosmwasmpool.ChangeLimiter `json:"change_limiter,omitempty"`
}

// alloyTransmuterListLimiters queries the limiters of the alloyed transmuter contract.
func alloyTransmuterListLimiters(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (sqscosmwasmpool.AlloyedRateLimiter, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(listLimitersQueryString))
	if err != nil {
		return sqscosmwasmpool.AlloyedRateLimiter{}, fmt.Errorf(
			"error querying list_limiters for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	listLimtersResponseData := listLimitersResponse{}

	if err := json.Unmarshal(bz, &listLimtersResponseData); err != nil {
		return sqscosmwasmpool.AlloyedRateLimiter{}, fmt.Errorf(
			"error unmarshalling limiters for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	staticLimiters := make(map[string]sqscosmwasmpool.StaticLimiter, 0)
	changeLimiters := make(map[string]sqscosmwasmpool.ChangeLimiter, 0)

	for _, limiterRaw := range listLimtersResponseData.Limiters {
		var info LimiterInfo
		err := json.Unmarshal(limiterRaw[0], &info)
		if err != nil {
			ctx.Logger().Error("Error unmarshaling info:", "err", err, "pool_id", poolId)
			continue
		}

		var value LimiterValue
		err = json.Unmarshal(limiterRaw[1], &value)
		if err != nil {
			ctx.Logger().Error("Error unmarshaling value:", "err", err, "pool_id", poolId)
			continue
		}

		// First element of info is the denom
		if len(info) < 1 {
			ctx.Logger().Error("Error parsing limiter info", "pool_id", poolId)
		}

		denom := info[0]

		if value.StaticLimiter != nil {
			staticLimiters[denom] = *value.StaticLimiter
		} else if value.ChangeLimiter != nil {
			changeLimiters[denom] = *value.ChangeLimiter
		}
	}

	return sqscosmwasmpool.AlloyedRateLimiter{
		StaticLimiterByDenomMap: staticLimiters,
		ChangeLimiterByDenomMap: changeLimiters,
	}, nil
}
