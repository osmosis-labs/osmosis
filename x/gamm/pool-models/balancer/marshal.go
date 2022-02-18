package balancer

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

type balancerPoolPretty struct {
	Address            sdk.AccAddress    `json:"address" yaml:"address"`
	Id                 uint64            `json:"id" yaml:"id"`
	PoolParams         PoolParams        `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string            `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Dec           `json:"total_weight" yaml:"total_weight"`
	TotalShares        sdk.Coin          `json:"total_shares" yaml:"total_shares"`
	PoolAssets         []types.PoolAsset `json:"pool_assets" yaml:"pool_assets"`
}

func (pa Pool) String() string {
	out, err := pa.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(out)
}

// MarshalJSON returns the JSON representation of a Pool.
func (pa Pool) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	decTotalWeight := sdk.NewDecFromInt(pa.TotalWeight)

	return json.Marshal(balancerPoolPretty{
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
	var alias balancerPoolPretty
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
