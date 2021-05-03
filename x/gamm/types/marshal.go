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

func (pa PoolAccount) getAllPoolAssetPretty() []poolAssetPretty {
	prettyPoolAssets := make([]poolAssetPretty, len(pa.PoolAssets))
	for i, v := range pa.PoolAssets {
		prettyPoolAssets[i] = poolAssetPretty{
			Weight: sdk.NewDecFromInt(v.Weight).QuoInt64(GuaranteedWeightPrecision),
			Token:  v.Token,
		}
	}
	return prettyPoolAssets
}

func uglifyPrettyPoolAssets(prettyAssets []poolAssetPretty) []PoolAsset {
	// don't be mean to the standard assets D:
	uglyPoolAssets := make([]PoolAsset, len(prettyAssets))
	for i, v := range prettyAssets {
		uglyPoolAssets[i] = PoolAsset{
			Weight: v.Weight.MulInt64(GuaranteedWeightPrecision).RoundInt(),
			Token:  v.Token,
		}
	}
	return uglyPoolAssets
}

type poolAccountPretty struct {
	Address            sdk.AccAddress    `json:"address" yaml:"address"`
	PubKey             string            `json:"public_key" yaml:"public_key"`
	AccountNumber      uint64            `json:"account_number" yaml:"account_number"`
	Sequence           uint64            `json:"sequence" yaml:"sequence"`
	Id                 uint64            `json:"id" yaml:"id"`
	PoolParams         PoolParams        `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string            `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Int           `json:"total_weight" yaml:"total_weight"`
	TotalShare         sdk.Coin          `json:"total_share" yaml:"total_share"`
	PoolAssets         []poolAssetPretty `json:"pool_assets" yaml:"pool_assets"`
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

	prettyPoolAssets := pa.getAllPoolAssetPretty()

	bs, err := yaml.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        pa.TotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         prettyPoolAssets,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// MarshalJSON returns the JSON representation of a PoolAccount.
func (pa PoolAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	prettyPoolAssets := pa.getAllPoolAssetPretty()

	return json.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        pa.TotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         prettyPoolAssets,
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
	pa.TotalWeight = alias.TotalWeight
	pa.TotalShare = alias.TotalShare
	pa.PoolAssets = uglifyPrettyPoolAssets(alias.PoolAssets)

	return nil
}
