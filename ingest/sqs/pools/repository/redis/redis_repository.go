package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/redis/go-redis/v9"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type redisPoolsRepo struct {
	appCodec codec.Codec
	client   *redis.Client
}

var _ domain.PoolsRepository = &redisPoolsRepo{}

const (
	cfmmPoolKey         = "cfmmPool"
	concentratedPoolKey = "concentratedPool"
	cosmWasmPoolKey     = "cosmWasmPool"
)

// NewRedisPoolsRepo will create an implementation of pools.Repository
func NewRedisPoolsRepo(appCodec codec.Codec, client *redis.Client) domain.PoolsRepository {
	return &redisPoolsRepo{
		appCodec: appCodec,
		client:   client,
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
	sqsPoolMapByID := r.client.HGetAll(ctx, sqsPoolModelKey(storeKey)).Val()
	chainPoolMapByID := r.client.HGetAll(ctx, chainPoolModelKey(storeKey)).Val()

	if len(sqsPoolMapByID) != len(chainPoolMapByID) {
		return nil, fmt.Errorf("pools count mismatch: sqsPoolMapByID: %d, chainPoolMapByID: %d", len(sqsPoolMapByID), len(chainPoolMapByID))
	}

	pools := make([]domain.PoolI, 0, len(sqsPoolMapByID))
	for poolIDKeyStr, sqsPoolModelBytes := range sqsPoolMapByID {

		pool := &domain.PoolWrapper{
			SQSModel: domain.SQSPool{},
		}

		err := json.Unmarshal([]byte(sqsPoolModelBytes), &pool.SQSModel)
		if err != nil {
			fmt.Println(sqsPoolModelBytes)
			return nil, err
		}

		chainPoolModelBytes, ok := chainPoolMapByID[poolIDKeyStr]
		if !ok {
			return nil, fmt.Errorf("pool ID %s not found in chainPoolMapByID", poolIDKeyStr)
		}

		err = r.appCodec.UnmarshalInterfaceJSON([]byte(chainPoolModelBytes), &pool.ChainModel)
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
		serializedSQSPoolModel, err := json.Marshal(pool.GetSQSPoolModel())
		if err != nil {
			return err
		}

		serializedChainPoolModel, err := r.appCodec.MarshalInterfaceJSON(pool.GetUnderlyingPool())
		if err != nil {
			return err
		}

		// Note that we have 2x write and read amplification due to storage layout. We can optimize this later.
		err = r.client.HSet(ctx, sqsPoolModelKey(storeKey), pool.GetId(), serializedSQSPoolModel).Err()
		if err != nil {
			return err
		}

		err = r.client.HSet(ctx, chainPoolModelKey(storeKey), pool.GetId(), serializedChainPoolModel).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func sqsPoolModelKey(storeKey string) string {
	return fmt.Sprintf("%s/sqs", storeKey)
}

func chainPoolModelKey(storeKey string) string {
	return fmt.Sprintf("%s/chain", storeKey)
}
