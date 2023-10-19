package redis

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
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

// NewRedisPoolsRepo will create an implementation of pools.Repository
func NewRedisPoolsRepo(client *redis.Client) domain.PoolsRepository {
	return &redisPoolsRepo{
		client: client,
	}
}

// GetAllCFMM implements domain.PoolsRepository.
// Returns balancer and stableswap pools sorted by ID.
func (r *redisPoolsRepo) GetAllCFMM(ctx context.Context) ([]domain.PoolI, error) {
	return r.getPools(ctx, cfmmPoolKey)
}

// GetAllConcentrated implements domain.PoolsRepository.
// Returns concentrated pools sorted by ID.
func (r *redisPoolsRepo) GetAllConcentrated(ctx context.Context) ([]domain.PoolI, error) {
	return r.getPools(ctx, concentratedPoolKey)
}

// GetAllCosmWasm implements domain.PoolsRepository.
// Returns cosmwasm pools sorted by ID.
func (r *redisPoolsRepo) GetAllCosmWasm(ctx context.Context) ([]domain.PoolI, error) {
	return r.getPools(ctx, cosmWasmPoolKey)
}

// StoreCFMM implements domain.PoolsRepository.
// CONTRACT: all pools are either balancer or stableswap.
// This method does not perform any validation.
func (r *redisPoolsRepo) StoreCFMM(ctx context.Context, pools []domain.PoolI) (err error) {
	return r.storePools(ctx, cfmmPoolKey, pools)
}

// StoreConcentrated implements domain.PoolsRepository.
// CONTRACT: all pools are concentrated.
// This method does not perform any validation.
func (r *redisPoolsRepo) StoreConcentrated(ctx context.Context, pools []domain.PoolI) error {
	return r.storePools(ctx, concentratedPoolKey, pools)
}

// StoreCosmWasm implements domain.PoolsRepository.
// CONTRACT: all pools are cosmwasm.
// This method does not perform any validation.
func (r *redisPoolsRepo) StoreCosmWasm(ctx context.Context, pools []domain.PoolI) error {
	return r.storePools(ctx, cosmWasmPoolKey, pools)
}

// getPools returns pools from Redis by storeKey.
func (r *redisPoolsRepo) getPools(ctx context.Context, storeKey string) ([]domain.PoolI, error) {
	poolMapByID := r.client.HGetAll(ctx, storeKey).Val()

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

// storePools stores pools in Redis by storeKey.
func (r *redisPoolsRepo) storePools(ctx context.Context, storeKey string, pools []domain.PoolI) error {
	for _, pool := range pools {
		serializedPool, err := json.Marshal(pool)
		if err != nil {
			return err
		}

		err = r.client.HSet(ctx, storeKey, pool.GetId(), serializedPool).Err()
		if err != nil {
			return err
		}
	}
	return nil
}
