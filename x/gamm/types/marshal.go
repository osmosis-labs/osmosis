package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gopkg.in/yaml.v2"
)

type poolAssetPretty struct {
	Token  types.Coin `json:"token" yaml:"token"`
	Weight sdk.Dec    `json:"weight" yaml:"weight"`
}

func (asset PoolAsset) prettify() poolAssetPretty {
	return poolAssetPretty{
		Weight: sdk.NewDecFromInt(asset.Weight).QuoInt64(GuaranteedWeightPrecision),
		Token:  asset.Token,
	}
}

func (asset poolAssetPretty) uglify() PoolAsset {
	return PoolAsset{
		Weight: asset.Weight.MulInt64(GuaranteedWeightPrecision).RoundInt(),
		Token:  asset.Token,
	}
}

// MarshalJSON returns the JSON representation of a PoolAsset.
func (pa PoolAsset) MarshalJSON() ([]byte, error) {
	return json.Marshal(pa.prettify())
}

// UnmarshalJSON unmarshals raw JSON bytes into a PoolAsset.
func (pa *PoolAsset) UnmarshalJSON(bz []byte) error {
	var alias poolAssetPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	validAsset := alias.uglify()

	pa.Weight = validAsset.Weight
	pa.Token = validAsset.Token

	return nil
}

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

type poolAccountPretty struct {
	Address            sdk.AccAddress `json:"address" yaml:"address"`
	PubKey             string         `json:"public_key" yaml:"public_key"`
	AccountNumber      uint64         `json:"account_number" yaml:"account_number"`
	Sequence           uint64         `json:"sequence" yaml:"sequence"`
	Id                 uint64         `json:"id" yaml:"id"`
	PoolParams         PoolParams     `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string         `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Dec        `json:"total_weight" yaml:"total_weight"`
	TotalShare         sdk.Coin       `json:"total_share" yaml:"total_share"`
	PoolAssets         []PoolAsset    `json:"pool_assets" yaml:"pool_assets"`
}

func (pa PoolAccount) String() string {
	out, _ := pa.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a PoolAccount.
func (pa PoolAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(pa.TotalWeight).QuoInt64(GuaranteedWeightPrecision)

	bz, err := yaml.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        decTotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         pa.PoolAssets,
	})

	if err != nil {
		return nil, err
	}

	return string(bz), nil
}

// MarshalJSON returns the JSON representation of a PoolAccount.
func (pa PoolAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(pa.TotalWeight).QuoInt64(GuaranteedWeightPrecision)

	return json.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        decTotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         pa.PoolAssets,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a PoolAccount.
func (pa *PoolAccount) UnmarshalJSON(bz []byte) error {
	var alias poolAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	pa.BaseAccount = authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	pa.Id = alias.Id
	pa.PoolParams = alias.PoolParams
	pa.FuturePoolGovernor = alias.FuturePoolGovernor
	pa.TotalWeight = alias.TotalWeight.MulInt64(GuaranteedWeightPrecision).RoundInt()
	pa.TotalShare = alias.TotalShare
	pa.PoolAssets = alias.PoolAssets

	return nil
}
