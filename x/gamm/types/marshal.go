package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

type poolAssetPretty struct {
	Token  sdk.Coin `json:"token" yaml:"token"`
	Weight sdk.Dec  `json:"weight" yaml:"weight"`
}

func (asset PoolAsset) prettify() poolAssetPretty {
	return poolAssetPretty{
		Weight: sdk.NewDecFromInt(asset.Weight).QuoInt64(GuaranteedWeightPrecision),
		Token:  asset.Token,
	}
}

// D: at name
// func (asset poolAssetPretty) uglify() PoolAsset {
// 	return PoolAsset{
// 		Weight: asset.Weight.MulInt64(GuaranteedWeightPrecision).RoundInt(),
// 		Token:  asset.Token,
// 	}
// }

// MarshalYAML returns the YAML representation of a PoolAsset.
// This is assumed to not be called on a stand-alone instance, so it removes the first marshalled line.
func (pa PoolAsset) MarshalYAML() (interface{}, error) {
	bz, err := yaml.Marshal(pa.prettify())
	if err != nil {
		return nil, err
	}
	s := string(bz)
	return s, nil
}

type poolPretty struct {
	Address            sdk.AccAddress `json:"address" yaml:"address"`
	Id                 uint64         `json:"id" yaml:"id"`
	PoolParams         PoolParams     `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string         `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Dec        `json:"total_weight" yaml:"total_weight"`
	TotalShares        sdk.Coin       `json:"total_shares" yaml:"total_shares"`
	PoolAssets         []PoolAsset    `json:"pool_assets" yaml:"pool_assets"`
}

//nolint:forcetypeassert
func (pa Pool) String() string {
	out, _ := pa.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a Pool.
func (pa Pool) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(pa.TotalWeight).QuoInt64(GuaranteedWeightPrecision)

	bz, err := yaml.Marshal(poolPretty{
		Address:            accAddr,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        decTotalWeight,
		TotalShares:        pa.TotalShares,
		PoolAssets:         pa.PoolAssets,
	})
	if err != nil {
		return nil, err
	}

	return string(bz), nil
}

// MarshalJSON returns the JSON representation of a Pool.
func (pa Pool) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(pa.TotalWeight)

	return json.Marshal(poolPretty{
		Address:            accAddr,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        decTotalWeight,
		TotalShares:        pa.TotalShares,
		PoolAssets:         pa.PoolAssets,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a Pool.
func (pa *Pool) UnmarshalJSON(bz []byte) error {
	var alias poolPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	pa.Address = alias.Address.String()
	pa.Id = alias.Id
	pa.PoolParams = alias.PoolParams
	pa.FuturePoolGovernor = alias.FuturePoolGovernor
	pa.TotalWeight = alias.TotalWeight.RoundInt()
	pa.TotalShares = alias.TotalShares
	pa.PoolAssets = alias.PoolAssets

	return nil
}
