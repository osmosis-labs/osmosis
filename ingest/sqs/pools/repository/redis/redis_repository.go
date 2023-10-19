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
func (r *redisPoolsRepo) GetAllCFMM(ctx context.Context) ([]domain.PoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, cfmmPoolKey).Val()

	pools := make([]domain.PoolI, 0, len(poolMapByID))
	for key, v := range poolMapByID {
		var pool domain.PoolI
		if strings.HasPrefix(key, balancerPrefix) {
			pool = &domain.Pool{}
		} else if strings.HasPrefix(key, stableswapPrefix) {
			pool = &domain.Pool{}
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
func (r *redisPoolsRepo) GetAllConcentrated(ctx context.Context) ([]domain.PoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, concentratedPoolKey).Val()

	pools := make([]domain.PoolI, 0, len(poolMapByID))
	for _, v := range poolMapByID {
		pool := &domain.Pool{}
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
func (r *redisPoolsRepo) GetAllCosmWasm(ctx context.Context) ([]domain.PoolI, error) {
	// TODO: use generics to reduce code duplication stemming from these methods.
	poolMapByID := r.client.HGetAll(ctx, cosmWasmPoolKey).Val()

	pools := make([]domain.PoolI, 0, len(poolMapByID))
	for _, v := range poolMapByID {
		pool := &domain.Pool{}
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

// StoreCFMM implements domain.PoolsRepository.
func (r *redisPoolsRepo) StoreCFMM(ctx context.Context, pools []domain.PoolI) (err error) {
	for _, pool := range pools {
		var (
			serializedPool []byte
			poolType       = pool.GetType()
		)
		if poolType == poolmanagertypes.Balancer {
			balancerPool, ok := pool.(*domain.Pool)
			if !ok {
				return domain.InvalidPoolTypeError{PoolType: int32(poolType)}
			}

			serializedPool, err = json.Marshal(balancerPool)
			if err != nil {
				return err
			}
		} else if poolType == poolmanagertypes.Stableswap {
			stableswapPool, ok := pool.(*domain.Pool)
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
func (r *redisPoolsRepo) StoreConcentrated(ctx context.Context, pools []domain.PoolI) error {
	for _, pool := range pools {
		concentratedPool, ok := pool.(*domain.Pool)
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
func (r *redisPoolsRepo) StoreCosmWasm(ctx context.Context, pools []domain.PoolI) error {
	for _, pool := range pools {
		cosmWasmPool, ok := pool.(*domain.Pool)
		if !ok {
			return domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
		}

		serializedPool, err := json.Marshal(cosmWasmPool)
		if err != nil {
			return err
		}

		err = r.client.HSet(ctx, cosmWasmPoolKey, cosmWasmPool.GetId(), serializedPool).Err()
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
