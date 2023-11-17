package routertesting

import (
	"encoding/json"
	"os"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type SerializedPool struct {
	Type poolmanagertypes.PoolType `json:"type"`
	Data json.RawMessage           `json:"data"`
}

const poolsFile = "pools.json"

func StorePools(actualPools []domain.PoolI) error {
	_, err := os.Stat(poolsFile)
	if os.IsNotExist(err) {
		file, err := os.Create(poolsFile)
		if err != nil {
			return err
		}
		defer file.Close()

		pools := make([]SerializedPool, 0, len(actualPools))

		for _, pool := range actualPools {
			poolType := pool.GetType()

			poolData, err := json.Marshal(pool)
			if err != nil {
				return err
			}

			pool := SerializedPool{
				Type: poolType,
				Data: poolData,
			}

			pools = append(pools, pool)
		}

		poolsJSON, err := json.Marshal(pools)
		if err != nil {
			return err
		}

		_, err = file.Write(poolsJSON)
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	return nil
}

func ReadPools() ([]domain.PoolI, error) {
	poolBytes, err := os.ReadFile(poolsFile)
	if err != nil {
		return nil, err
	}

	var serializedPools []SerializedPool
	err = json.Unmarshal(poolBytes, &serializedPools)
	if err != nil {
		return nil, err
	}

	actualPools := make([]domain.PoolI, 0, len(serializedPools))
	for _, pool := range serializedPools {
		switch pool.Type {
		case poolmanagertypes.Concentrated:
			var concentratedPool pools.RoutableConcentratedPoolImpl
			err := json.Unmarshal(pool.Data, &concentratedPool)
			if err != nil {
				return nil, err
			}
			actualPools = append(actualPools, &concentratedPool)
		case poolmanagertypes.CosmWasm:
			var transmuterPool pools.RoutableTransmuterPoolImpl
			err := json.Unmarshal(pool.Data, &transmuterPool)
			if err != nil {
				return nil, err
			}
			actualPools = append(actualPools, &transmuterPool)
		case poolmanagertypes.Stableswap:
			fallthrough
		case poolmanagertypes.Balancer:
			var cfmmPool pools.RoutableCFMMPoolImpl
			err := json.Unmarshal(pool.Data, &cfmmPool)
			if err != nil {
				return nil, err
			}
			actualPools = append(actualPools, &cfmmPool)
		default:
			return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.Type)}
		}
	}

	return actualPools, nil
}
