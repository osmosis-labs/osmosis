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
	appCodec          codec.Codec
	client            *redis.Client
	repositoryManager domain.TxManager
}

var (
	_ domain.PoolsRepository = &redisPoolsRepo{}
)

const (
	cfmmPoolKey         = "cfmmPool"
	concentratedPoolKey = "concentratedPool"
	cosmWasmPoolKey     = "cosmWasmPool"
)

// NewRedisPoolsRepo will create an implementation of pools.Repository
func NewRedisPoolsRepo(appCodec codec.Codec, repositoryManager domain.TxManager) domain.PoolsRepository {
	return &redisPoolsRepo{
		appCodec:          appCodec,
		repositoryManager: repositoryManager,
	}
}

// GetAllCFMM implements domain.PoolsRepository.
// Returns balancer and stableswap pools sorted by ID.
func (r *redisPoolsRepo) GetAllCFMM(ctx context.Context) ([]domain.PoolI, error) {
	tx := r.repositoryManager.StartTx()

	sqsPoolMapByIDCmd, chainPoolMapByIDCmd, err := r.requestPoolsAtomically(ctx, tx, cfmmPoolKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	sqsPoolMapByID := sqsPoolMapByIDCmd.Val()
	chainPoolMapByID := chainPoolMapByIDCmd.Val()

	return r.getPools(sqsPoolMapByID, chainPoolMapByID)
}

// GetAllConcentrated implements domain.PoolsRepository.
// Returns concentrated pools sorted by ID.
func (r *redisPoolsRepo) GetAllConcentrated(ctx context.Context) ([]domain.PoolI, error) {
	tx := r.repositoryManager.StartTx()

	sqsPoolMapByIDCmd, chainPoolMapByIDCmd, err := r.requestPoolsAtomically(ctx, tx, concentratedPoolKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	sqsPoolMapByID := sqsPoolMapByIDCmd.Val()
	chainPoolMapByID := chainPoolMapByIDCmd.Val()

	return r.getPools(sqsPoolMapByID, chainPoolMapByID)
}

// GetAllCosmWasm implements domain.PoolsRepository.
// Returns cosmwasm pools sorted by ID.
func (r *redisPoolsRepo) GetAllCosmWasm(ctx context.Context) ([]domain.PoolI, error) {
	tx := r.repositoryManager.StartTx()

	sqsPoolMapByIDCmd, chainPoolMapByIDCmd, err := r.requestPoolsAtomically(ctx, tx, cosmWasmPoolKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	sqsPoolMapByID := sqsPoolMapByIDCmd.Val()
	chainPoolMapByID := chainPoolMapByIDCmd.Val()

	return r.getPools(sqsPoolMapByID, chainPoolMapByID)
}

// GetAllPools implements domain.PoolsRepository.
// Atomically reads all pools from Redis.
func (r *redisPoolsRepo) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	tx := r.repositoryManager.StartTx()

	sqsPoolMapByIDCmdCFMM, chainPoolMapByIDCmdCFMM, err := r.requestPoolsAtomically(ctx, tx, cfmmPoolKey)
	if err != nil {
		return nil, err
	}

	sqsPoolMapByIDCmdConcentrated, chainPoolMapByIDCmdConcentrated, err := r.requestPoolsAtomically(ctx, tx, concentratedPoolKey)
	if err != nil {
		return nil, err
	}

	sqsPoolMapByIDCmdCosmwasm, chainPoolMapByIDCmdCosmwasm, err := r.requestPoolsAtomically(ctx, tx, cosmWasmPoolKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	cfmmPools, err := r.getPools(sqsPoolMapByIDCmdCFMM.Val(), chainPoolMapByIDCmdCFMM.Val())
	if err != nil {
		return nil, err
	}

	concentratedPools, err := r.getPools(sqsPoolMapByIDCmdConcentrated.Val(), chainPoolMapByIDCmdConcentrated.Val())
	if err != nil {
		return nil, err
	}

	cosmwasmPools, err := r.getPools(sqsPoolMapByIDCmdCosmwasm.Val(), chainPoolMapByIDCmdCosmwasm.Val())
	if err != nil {
		return nil, err
	}

	allPools := make([]domain.PoolI, 0, len(cfmmPools)+len(concentratedPools)+len(cosmwasmPools))

	allPools = append(allPools, cfmmPools...)
	allPools = append(allPools, concentratedPools...)
	allPools = append(allPools, cosmwasmPools...)

	// Sort by ID
	sort.Slice(allPools, func(i, j int) bool {
		return allPools[i].GetId() < allPools[j].GetId()
	})

	return allPools, nil
}

func (r *redisPoolsRepo) StorePools(ctx context.Context, tx domain.Tx, cfmmPools []domain.PoolI, concentratedPools []domain.PoolI, cosmwasmPools []domain.PoolI) error {

	if err := r.addCFMMPoolsTx(ctx, tx, cfmmPools); err != nil {
		return err
	}

	if err := r.addConcentratedPoolsTx(ctx, tx, concentratedPools); err != nil {
		return err
	}

	if err := r.addCosmwasmPoolsTx(ctx, tx, cosmwasmPools); err != nil {
		return err
	}

	return nil
}

// addCFMMPoolsTx pipelines the given CFMM pools at the given storeKey to be executed atomically in a transaction.
// CONTRACT: all pools are CFMM.
// This method does not perform any validation.
func (r *redisPoolsRepo) addCFMMPoolsTx(ctx context.Context, tx domain.Tx, pools []domain.PoolI) (err error) {
	return r.addPoolsTx(ctx, tx, cfmmPoolKey, pools)
}

// addConcentratedPoolsTx pipelines the given concentrated pools at the given storeKey to be executed atomically in a transaction.
// CONTRACT: all pools are concentrated.
// This method does not perform any validation.
func (r *redisPoolsRepo) addConcentratedPoolsTx(ctx context.Context, tx domain.Tx, pools []domain.PoolI) error {
	return r.addPoolsTx(ctx, tx, concentratedPoolKey, pools)
}

// addCosmWasmPoolsTx pipelines the given cosmwasm pools at the given storeKey to be executed atomically in a transaction.
// CONTRACT: all pools are cosmwasm.
// This method does not perform any validation.
func (r *redisPoolsRepo) addCosmwasmPoolsTx(ctx context.Context, tx domain.Tx, pools []domain.PoolI) error {
	return r.addPoolsTx(ctx, tx, cosmWasmPoolKey, pools)
}

func (r *redisPoolsRepo) requestPoolsAtomically(ctx context.Context, tx domain.Tx, storeKey string) (sqsPoolMapByID *redis.MapStringStringCmd, chainPoolMapByID *redis.MapStringStringCmd, err error) {
	if !tx.IsActive() {
		return nil, nil, fmt.Errorf("tx is inactive")
	}

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return nil, nil, err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return nil, nil, err
	}

	sqsPoolMapByID = pipeliner.HGetAll(ctx, sqsPoolModelKey(storeKey))
	chainPoolMapByID = pipeliner.HGetAll(ctx, chainPoolModelKey(storeKey))

	return sqsPoolMapByID, chainPoolMapByID, nil
}

// getPools returns pools from Redis by storeKey.
func (r *redisPoolsRepo) getPools(sqsPoolMapByID, chainPoolMapByID map[string]string) ([]domain.PoolI, error) {
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

// addPoolsTx pipelines the given pools at the given storeKey to be executed atomically in a transaction.
func (r *redisPoolsRepo) addPoolsTx(ctx context.Context, tx domain.Tx, storeKey string, pools []domain.PoolI) error {

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

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
		err = pipeliner.HSet(ctx, sqsPoolModelKey(storeKey), pool.GetId(), serializedSQSPoolModel).Err()
		if err != nil {
			return err
		}

		err = pipeliner.HSet(ctx, chainPoolModelKey(storeKey), pool.GetId(), serializedChainPoolModel).Err()
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
