package poolstransformer

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sqscosmwasmpool "github.com/osmosis-labs/osmosis/v30/ingest/types/cosmwasmpool"

	commondomain "github.com/osmosis-labs/osmosis/v30/ingest/common/domain"
)

const (
	listAssetConfigsQueryString       = `{"list_asset_configs":{}}`
	getShareDenomQueryString          = `{"get_share_denom":{}}`
	listLimitersQueryString           = `{"list_limiters":{}}`
	listRebalancingConfigsQueryString = `{"list_rebalancing_configs":{}}`
	listAssetGroupsQueryString        = `{"list_asset_groups":{}}`
	incentivePoolBalancesQueryString  = `{"get_incentive_pool_balances":{}}`
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
	rebalancingConfigs, err := alloyedTransmuterListRebalancingConfigs(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		rateLimiterData, err := alloyTransmuterListLimiters(ctx, pi.wasmKeeper, poolId, contractAddress)
		if err != nil {
			return err
		}
		rebalancingConfigs = make(map[string]sqscosmwasmpool.RebalancingConfig)

		for denom, rebalancingConfig := range rateLimiterData.StaticLimiterByDenomMap {
			rebalancingConfigs["denom::"+denom] = sqscosmwasmpool.RebalancingConfig{
				Limit: rebalancingConfig.UpperLimit,
			}
		}
	}

	// append alloyed denom to denoms
	*poolDenoms = append(*poolDenoms, alloyedDenom)

	// Attempt to fetch asset groups (v4+). If query fails (e.g., v3), ignore.
	assetGroups := make(map[string]sqscosmwasmpool.AssetGroup)
	if groups, err := alloyedTransmuterListAssetGroups(ctx, pi.wasmKeeper, poolId, contractAddress); err == nil {
		assetGroups = groups
	}

	// Attempt to fetch incentive pool balances (v4+). If query fails (e.g., v3), ignore.
	incentivePoolBalances := []sdk.Coin{}
	if balances, err := alloyedTransmuterIncentivePoolBalances(ctx, pi.wasmKeeper, poolId, contractAddress); err == nil {
		incentivePoolBalances = balances
	}

	cosmWasmPoolModel.Data.AlloyTransmuter = &sqscosmwasmpool.AlloyTransmuterData{
		AlloyedDenom:          alloyedDenom,
		AssetConfigs:          assetConfigs,
		RebalancingConfigs:    rebalancingConfigs,
		AssetGroups:           assetGroups,
		IncentivePoolBalances: incentivePoolBalances,
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

type listAssetGroupsResponse struct {
	AssetGroups map[string]sqscosmwasmpool.AssetGroup `json:"asset_groups"`
}

// alloyedTransmuterListAssetGroups queries the asset groups of the alloyed transmuter contract.
// Since: transmuter v4.0.0
func alloyedTransmuterListAssetGroups(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (map[string]sqscosmwasmpool.AssetGroup, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(listAssetGroupsQueryString))
	if err != nil {
		return nil, fmt.Errorf(
			"error querying list_asset_groups for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	var resp listAssetGroupsResponse
	if err := json.Unmarshal(bz, &resp); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling asset_groups for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return resp.AssetGroups, nil
}

// alloyTransmuterListLimiters queries the limiters of the alloyed transmuter contract.
// Deprecated: transmuter v4.0.0
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

type listRebalancingConfigsResponse struct {
	RebalancingConfigs [][]json.RawMessage `json:"rebalancing_configs"`
}

// alloyedTransmuterListRebalancingConfigs queries the rebalancing configs of the alloyed transmuter contract.
// Since: transmuter v4.0.0
func alloyedTransmuterListRebalancingConfigs(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (sqscosmwasmpool.RebalancingConfigs, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(listRebalancingConfigsQueryString))
	if err != nil {
		return sqscosmwasmpool.RebalancingConfigs{}, fmt.Errorf(
			"error querying list_rebalancing_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	var rebalancingConfigsResponse listRebalancingConfigsResponse

	if err := json.Unmarshal(bz, &rebalancingConfigsResponse); err != nil {
		return sqscosmwasmpool.RebalancingConfigs{}, fmt.Errorf(
			"error unmarshalling rebalancing_configs for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	rebalancingConfigs := make(sqscosmwasmpool.RebalancingConfigs)

	for _, config := range rebalancingConfigsResponse.RebalancingConfigs {
		// Each config is a tuple of [scope, RebalancingConfig]
		if len(config) < 2 {
			ctx.Logger().Error("Error parsing rebalancing config", "pool_id", poolId)
			continue
		}

		// First element is the scope
		var scope string
		if err := json.Unmarshal(config[0], &scope); err != nil {
			ctx.Logger().Error("Error unmarshalling scope from rebalancing config", "pool_id", poolId, "error", err)
			continue
		}

		// Second element is the RebalancingConfig
		var rebalancingConfig sqscosmwasmpool.RebalancingConfig
		if err := json.Unmarshal(config[1], &rebalancingConfig); err != nil {
			ctx.Logger().Error("Error unmarshalling rebalancing config", "pool_id", poolId, "scope", scope, "error", err)
			continue
		}

		rebalancingConfigs[scope] = rebalancingConfig
	}

	return rebalancingConfigs, nil
}

type incentivePoolBalancesResponse struct {
	Balances []sdk.Coin `json:"balances"`
}

// alloyedTransmuterIncentivePoolBalances queries the incentive pool balances of the alloyed transmuter contract.
// Since: transmuter v4.0.0
func alloyedTransmuterIncentivePoolBalances(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) ([]sdk.Coin, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(incentivePoolBalancesQueryString))
	if err != nil {
		return nil, fmt.Errorf(
			"error querying incentive_pool_balances for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	var resp incentivePoolBalancesResponse
	if err := json.Unmarshal(bz, &resp); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling incentive_pool_balances for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return resp.Balances, nil
}
