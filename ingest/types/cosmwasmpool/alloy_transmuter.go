package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	ALLOY_TRANSMUTER_CONTRACT_NAME               = "crates.io:transmuter"
	ALLOY_TRANSMUTER_MIN_CONTRACT_VERSION        = "3.0.0"
	ALLOY_TRANSMUTER_CONTRACT_VERSION_CONSTRAINT = ">= " + ALLOY_TRANSMUTER_MIN_CONTRACT_VERSION
)

func (model *CosmWasmPoolModel) IsAlloyTransmuter() bool {
	return model.ContractInfo.Matches(
		ALLOY_TRANSMUTER_CONTRACT_NAME,
		mustParseSemverConstraint(ALLOY_TRANSMUTER_CONTRACT_VERSION_CONSTRAINT),
	)
}

// Tranmuter Alloyed Data, since v3.0.0
// AssetConfigs is a list of denom and normalization factor pairs including the alloyed denom.
type AlloyTransmuterData struct {
	AlloyedDenom       string                  `json:"alloyed_denom"`
	AssetConfigs       []TransmuterAssetConfig `json:"asset_configs"`
	RebalancingConfigs RebalancingConfigs      `json:"rebalancing_configs"`
	// AssetGroups is a map of group label to list of denoms
	// Since: transmuter v4.0.0
	AssetGroups map[string]AssetGroup `json:"asset_groups"`

	// IncentivePoolBalances is a list of coin balances for incentive pools
	// Since: transmuter v4.0.0
	IncentivePoolBalances []sdk.Coin      `json:"incentive_pool_balances"`
	PreComputedData       PrecomputedData `json:"precomputed_data"`
}

// PrecomputedData for the alloyed pool.
type PrecomputedData struct {
	// StdNormFactor is the standard normalization factor for the pool.
	// It is computed as the LCM of all the normalization factors of the assets in the pool.
	// This is used for computing asset weights for checking rate limiting.
	StdNormFactor osmomath.Int `json:"std_norm_factor"`

	// NormalizationScalingFactors is the scaling factor for each asset in the pool.
	// Each index corresponds to the asset at the same index in the AssetConfigs.
	// This is used for computing asset weights for checking rate limiting.
	NormalizationScalingFactors map[string]osmomath.Int `json:"normalization_scaling_factors"`
}

// Configuration for each asset in the transmuter pool
type TransmuterAssetConfig struct {
	// Denom of the asset
	Denom string `json:"denom"`

	// Normalization factor for the asset.
	// [more info](https://github.com/osmosis-labs/transmuter/tree/v3.0.0?tab=readme-ov-file#normalization-factors)
	NormalizationFactor osmomath.Int `json:"normalization_factor"`
}

// AlloyedRateLimiter is a struct that contains the rate limiter configuration for the alloyed pool.
type AlloyedRateLimiter struct {
	StaticLimiterByDenomMap map[string]StaticLimiter `json:"static_limiters"`
	ChangeLimiterByDenomMap map[string]ChangeLimiter `json:"change_limiters"`
}

// GetStaticLimiter returns the StaticLimiter for the given denom.
func (limiter *AlloyedRateLimiter) GetStaticLimiter(denom string) (StaticLimiter, bool) {
	staticLimiter, ok := limiter.StaticLimiterByDenomMap[denom]
	return staticLimiter, ok
}

// GetChangeLimiter returns the ChangeLimiter for the given denom.
func (limiter *AlloyedRateLimiter) GetChangeLimiter(denom string) (ChangeLimiter, bool) {
	changeLimiter, ok := limiter.ChangeLimiterByDenomMap[denom]
	return changeLimiter, ok
}

// StaticLimiter represents a static rate limiter configuration.
type StaticLimiter struct {
	UpperLimit string `json:"upper_limit"`
}

// WindowConfig represents the configuration for a rate limiter window.
type WindowConfig struct {
	WindowSize    uint64 `json:"window_size"`
	DivisionCount uint64 `json:"division_count"`
}

// Division represents a time division with its associated values.
type Division struct {
	// StartedAt is the time when the division is marked as started (Unix timestamp).
	StartedAt int64 `json:"started_at"`

	// UpdatedAt is the time when the division was last updated (Unix timestamp).
	UpdatedAt int64 `json:"updated_at"`

	// LatestValue is the latest value that gets updated (represented as a decimal string).
	LatestValue string `json:"latest_value"`

	// Integral is the sum of each updated value * elapsed time since last update (represented as a decimal string).
	Integral string `json:"integral"`
}

// ChangeLimiter represents a change rate limiter configuration.
type ChangeLimiter struct {
	Divisions      []Division   `json:"divisions"`
	LatestValue    string       `json:"latest_value"`
	WindowConfig   WindowConfig `json:"window_config"`
	BoundaryOffset string       `json:"boundary_offset"`
}

// RebalancingConfig represents the rebalancing configuration for an asset.
type RebalancingConfig struct {
	IdealUpper             string `json:"ideal_upper"`
	IdealLower             string `json:"ideal_lower"`
	CriticalUpper          string `json:"critical_upper"`
	CriticalLower          string `json:"critical_lower"`
	Limit                  string `json:"limit"`
	AdjustmentRateStrained string `json:"adjustment_rate_strained"`
	AdjustmentRateCritical string `json:"adjustment_rate_critical"`
}

// RebalancingConfigs is a struct that contains the rebalancing configurations for the alloyed pool.
// Since: transmuter v4.0.0
type RebalancingConfigs map[string]RebalancingConfig

// AssetGroup is a struct that contains the asset group configuration for the alloyed pool.
type AssetGroup struct {
	Denoms      []string `json:"denoms"`
	IsCorrupted bool     `json:"is_corrupted"`
}
