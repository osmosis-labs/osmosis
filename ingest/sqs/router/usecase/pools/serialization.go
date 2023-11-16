package pools

import (
	"encoding/json"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type routableSerializedPool struct {
	PoolData      []byte       `json:"pool_wrapper"`
	TokenOutDenom string       `json:"token_out_denom"`
	TakerFee      osmomath.Dec `json:"taker_fee"`
}

type SerializedPoolByType struct {
	PoolData []byte                    `json:"pool_wrapper"`
	PoolType poolmanagertypes.PoolType `json:"pool_type"`
}

// // UnmarshalJSON implements json.Unmarshaler.
// func (r *RoutablePoolSerialized) UnmarshalJSON([]byte) error {
// 	r.UnmarshalJSON()
// }

// var _ json.Unmarshaler = &RoutablePoolSerialized{}

func (r *SerializedPoolByType) Unmarshal(bz []byte) (domain.RoutablePool, error) {
	var routablePool domain.RoutablePool
	switch r.PoolType {
	case poolmanagertypes.Concentrated:
		var concentratedPool routableConcentratedPoolImpl
		err := json.Unmarshal(r.PoolData, &concentratedPool)
		if err != nil {
			return nil, err
		}
		routablePool = &concentratedPool
	case poolmanagertypes.CosmWasm:
		var transmuterPool routableTransmuterPoolImpl
		err := json.Unmarshal(r.PoolData, &transmuterPool)
		if err != nil {
			return nil, err
		}
		routablePool = &transmuterPool
	case poolmanagertypes.Balancer:
		fallthrough
	case poolmanagertypes.Stableswap:
		var cfmmPool routableCFMMPoolImpl
		err := json.Unmarshal(r.PoolData, &cfmmPool)
		if err != nil {
			return nil, err
		}
		routablePool = &cfmmPool
	default:
		return nil, domain.InvalidPoolTypeError{PoolType: int32(r.PoolType)}
	}

	return routablePool, nil
}
