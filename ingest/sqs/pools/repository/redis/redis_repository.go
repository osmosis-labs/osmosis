package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"

	concentrated "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/model"
	cosmwasmpool "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type redisPoolsRepo struct {
	client *redis.Client
}

var _ domain.PoolsRepository = &redisPoolsRepo{}

const (
	cfmmPoolKey         = "cfmmPool"
	concentratedPoolKey = "concentratedPool"
	cosmWasmPoolKey     = "cosmWasmPool"
)

var (
	balancerPrefix   = strconv.Itoa(int(poolmanagertypes.Balancer))
	stableswapPrefix = strconv.Itoa(int(poolmanagertypes.Stableswap))
)

// NewRedisPoolsRepo will create an implementation of pools.Repository
func NewRedisPoolsRepo(client *redis.Client) domain.PoolsRepository {
	return &redisPoolsRepo{
		client: client,
	}
}

// GetAllCFMM implements domain.PoolsRepository.
// Returns balancer and stableswap pools sorted by ID.
func (r *redisPoolsRepo) GetAllCFMM(ctx context.Context) ([]domain.CFMMPoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, cfmmPoolKey).Val()

	pools := make([]domain.CFMMPoolI, 0, len(poolMapByID))
	for key, v := range poolMapByID {
		var pool domain.CFMMPoolI
		if strings.HasPrefix(key, balancerPrefix) {
			pool = &balancer.Pool{}
		} else if strings.HasPrefix(key, stableswapPrefix) {
			pool = &stableswap.Pool{}
		} else {
			return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
		}

		err := json.Unmarshal([]byte(v), pool)
		if err != nil {
			return nil, err
		}

		pools = append(pools, pool)
	}

	// Sort by ID ascending.
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].GetId() < pools[j].GetId()
	})

	return pools, nil
}

// GetAllConcentrated implements domain.PoolsRepository.
// Returns concentrated pools sorted by ID.
func (r *redisPoolsRepo) GetAllConcentrated(ctx context.Context) ([]domain.ConcentratedPoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, concentratedPoolKey).Val()

	pools := make([]domain.ConcentratedPoolI, 0, len(poolMapByID))
	for _, v := range poolMapByID {
		pool := &concentrated.Pool{}
		err := json.Unmarshal([]byte(v), pool)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	// Sort by ID ascending.
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].GetId() < pools[j].GetId()
	})

	return pools, nil
}

// GetAllCosmWasm implements domain.PoolsRepository.
// Returns cosmwasm pools sorted by ID.
func (r *redisPoolsRepo) GetAllCosmWasm(ctx context.Context) ([]domain.CosmWasmPoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, cosmWasmPoolKey).Val()

	pools := make([]domain.CosmWasmPoolI, 0, len(poolMapByID))
	for _, v := range poolMapByID {
		pool := &cosmwasmpool.CosmWasmPool{}
		err := json.Unmarshal([]byte(v), pool)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	// Sort by ID ascending.
	sort.Slice(pools, func(i, j int) bool {
		// TODO: avoid casting after removing dependency on the chain types
		// Currently, CosmWasmPool does not implement this method and panics due
		// to serialization issues.
		// nolint: forcetypeassert
		poolI := pools[i].(*cosmwasmpool.CosmWasmPool)
		// nolint: forcetypeassert
		poolJ := pools[j].(*cosmwasmpool.CosmWasmPool)

		return poolI.PoolId < poolJ.PoolId
	})

	return pools, nil
}

// StoreCFMM implements domain.PoolsRepository.
func (r *redisPoolsRepo) StoreCFMM(ctx context.Context, pools []domain.CFMMPoolI) (err error) {
	for _, pool := range pools {
		var (
			serializedPool []byte
			poolType       = pool.GetType()
		)
		if poolType == poolmanagertypes.Balancer {
			balancerPool, ok := pool.(*balancer.Pool)
			if !ok {
				return domain.InvalidPoolTypeError{PoolType: int32(poolType)}
			}

			serializedPool, err = json.Marshal(balancerPool)
			if err != nil {
				return err
			}
		} else if poolType == poolmanagertypes.Stableswap {
			stableswapPool, ok := pool.(*stableswap.Pool)
			if !ok {
				return domain.InvalidPoolTypeError{PoolType: int32(poolType)}
			}

			serializedPool, err = json.Marshal(stableswapPool)
			if err != nil {
				return err
			}
		} else {
			return domain.InvalidPoolTypeError{PoolType: int32(poolType)}
		}

		err = r.client.HSet(ctx, cfmmPoolKey, cfmmKeyFromPoolTypeAndID(poolType, pool.GetId()), serializedPool).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreConcentrated implements domain.PoolsRepository.
func (r *redisPoolsRepo) StoreConcentrated(ctx context.Context, pools []domain.ConcentratedPoolI) error {
	for _, pool := range pools {
		concentratedPool, ok := pool.(*concentrated.Pool)
		if !ok {
			return domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
		}

		serializedPool, err := json.Marshal(concentratedPool)
		if err != nil {
			return err
		}

		err = r.client.HSet(ctx, concentratedPoolKey, pool.GetId(), serializedPool).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// StoreCosmWasm implements domain.PoolsRepository.
func (r *redisPoolsRepo) StoreCosmWasm(ctx context.Context, pools []domain.CosmWasmPoolI) error {
	for _, pool := range pools {
		cosmWasmPool, ok := pool.(*cosmwasmpool.CosmWasmPool)
		if !ok {
			return domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
		}

		serializedPool, err := json.Marshal(cosmWasmPool)
		if err != nil {
			return err
		}

		err = r.client.HSet(ctx, cosmWasmPoolKey, cosmWasmPool.PoolId, serializedPool).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// returns CFMM key from pool type and ID
// Note: can be optimized by avoiding the use of Sprintf
func cfmmKeyFromPoolTypeAndID(poolType poolmanagertypes.PoolType, ID uint64) string {
	return fmt.Sprintf("%d%d", poolType, ID)
}
