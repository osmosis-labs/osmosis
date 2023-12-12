package parsing

import (
	"fmt"
	"os"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/json"
	concentratedmodel "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/model"
	cosmwasmpoolmodel "github.com/osmosis-labs/osmosis/v21/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// SerializedPool is a struct that is used to serialize a pool to JSON.
type SerializedPool struct {
	Type      poolmanagertypes.PoolType `json:"type"`
	ChainPool json.RawMessage           `json:"data"`
	SQSModel  domain.SQSPool            `json:"sqs_model"`
	TickModel *domain.TickModel         `json:"tick_model,omitempty"`
}

// StorePools stores the pools to a file.
func StorePools(actualPools []domain.PoolI, tickModelMap map[uint64]domain.TickModel, poolsFile string) error {
	_, err := os.Stat(poolsFile)
	if os.IsNotExist(err) {
		file, err := os.Create(poolsFile)
		if err != nil {
			return err
		}
		defer file.Close()

		pools := make([]json.RawMessage, 0, len(actualPools))

		for _, pool := range actualPools {
			if pool.GetType() == poolmanagertypes.Concentrated {
				tickModel, ok := tickModelMap[pool.GetId()]
				if !ok {
					return fmt.Errorf("no tick model in map %s", domain.ConcentratedTickModelNotSetError{
						PoolId: pool.GetId(),
					})
				}
				if err := pool.SetTickModel(&tickModel); err != nil {
					return err
				}
			}

			poolData, err := MarshalPool(pool)
			if err != nil {
				return err
			}

			pools = append(pools, poolData)
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

// StoreTakerFees stores the taker fees to a file.
func StoreTakerFees(takerFeesFile string, takerFeesMap domain.TakerFeeMap) error {
	_, err := os.Stat(takerFeesFile)
	if os.IsNotExist(err) {
		file, err := os.Create(takerFeesFile)
		if err != nil {
			return err
		}
		defer file.Close()

		takerFeesJSON, err := json.Marshal(takerFeesMap)
		if err != nil {
			return err
		}

		_, err = file.Write(takerFeesJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadPools reads the pools from a file and returns them
func ReadPools(poolsFile string) ([]domain.PoolI, map[uint64]domain.TickModel, error) {
	poolBytes, err := os.ReadFile(poolsFile)
	if err != nil {
		return nil, nil, err
	}

	var serializedPools []SerializedPool
	err = json.Unmarshal(poolBytes, &serializedPools)
	if err != nil {
		return nil, nil, err
	}

	actualPools := make([]domain.PoolI, 0, len(serializedPools))

	tickMap := make(map[uint64]domain.TickModel)

	for _, pool := range serializedPools {
		poolWrapper, err := UnmarshalPool(pool)
		if err != nil {
			return nil, nil, err
		}

		if poolWrapper.GetType() == poolmanagertypes.Concentrated {
			tickMap[poolWrapper.GetId()] = *pool.TickModel
		}

		actualPools = append(actualPools, poolWrapper)
	}

	return actualPools, tickMap, nil
}

// ReadTakerFees reads the taker fees from a file and returns them
func ReadTakerFees(takerFeeFileName string) (domain.TakerFeeMap, error) {
	takerFeeBytes, err := os.ReadFile(takerFeeFileName)
	if err != nil {
		return nil, err
	}

	takerFeeMap := domain.TakerFeeMap{}
	err = json.Unmarshal(takerFeeBytes, &takerFeeMap)
	if err != nil {
		return nil, err
	}

	return takerFeeMap, nil
}

// MarshalPool marshals a pool to JSON.
func MarshalPool(pool domain.PoolI) (json.RawMessage, error) {
	poolType := pool.GetType()

	underlyingPool := pool.GetUnderlyingPool()

	chainPoolBz, err := json.Marshal(underlyingPool)
	if err != nil {
		return nil, err
	}

	var tickModel *domain.TickModel
	if poolType == poolmanagertypes.Concentrated {
		tickModel, err = pool.GetTickModel()
		if err != nil {
			return nil, err
		}
	}

	serializedPool := SerializedPool{
		Type:      poolType,
		ChainPool: chainPoolBz,
		SQSModel:  pool.GetSQSPoolModel(),
		TickModel: tickModel,
	}

	poolData, err := json.Marshal(serializedPool)
	if err != nil {
		return nil, err
	}

	return poolData, nil
}

// UnmarshalPool unmarshals a pool from JSON.
func UnmarshalPool(serializedPool SerializedPool) (domain.PoolI, error) {
	var (
		chainModel poolmanagertypes.PoolI
	)

	switch serializedPool.Type {
	case poolmanagertypes.Concentrated:
		var concentratedPool concentratedmodel.Pool
		err := json.Unmarshal(serializedPool.ChainPool, &concentratedPool)
		if err != nil {
			return nil, err
		}
		chainModel = &concentratedPool
	case poolmanagertypes.CosmWasm:
		var transmuterPool cosmwasmpoolmodel.CosmWasmPool
		err := json.Unmarshal(serializedPool.ChainPool, &transmuterPool)
		if err != nil {
			return nil, err
		}
		chainModel = &transmuterPool
	case poolmanagertypes.Stableswap:
		var balancerPool stableswap.Pool
		err := json.Unmarshal(serializedPool.ChainPool, &balancerPool)
		if err != nil {
			return nil, err
		}
		chainModel = &balancerPool
	case poolmanagertypes.Balancer:
		var balancerPool balancer.Pool
		err := json.Unmarshal(serializedPool.ChainPool, &balancerPool)
		if err != nil {
			return nil, err
		}
		chainModel = &balancerPool
	default:
		return nil, domain.InvalidPoolTypeError{PoolType: int32(serializedPool.Type)}
	}

	poolWrapper := domain.PoolWrapper{
		ChainModel: chainModel,
		SQSModel:   serializedPool.SQSModel,
		TickModel:  serializedPool.TickModel,
	}

	return &poolWrapper, nil
}
