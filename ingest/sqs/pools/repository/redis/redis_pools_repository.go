package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type redisPoolsRepo struct {
	appCodec          codec.Codec
	repositoryManager mvc.TxManager
}

type poolTicks struct {
	poolID uint64
	Cmd    *redis.StringCmd
}

var (
	_ mvc.PoolsRepository = &redisPoolsRepo{}
)

const (
	poolsKey = "pools"
)

// NewRedisPoolsRepo will create an implementation of pools.Repository
func NewRedisPoolsRepo(appCodec codec.Codec, repositoryManager mvc.TxManager) mvc.PoolsRepository {
	return &redisPoolsRepo{
		appCodec:          appCodec,
		repositoryManager: repositoryManager,
	}
}

// GetAllPools implements mvc.PoolsRepository.
// Atomically reads all pools from Redis.
func (r *redisPoolsRepo) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	tx := r.repositoryManager.StartTx()

	sqsPoolMapByIDCmd, chainPoolMapByIDCmd, err := r.requestPoolsAtomically(ctx, tx, poolsKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	allPools, err := r.getPools(sqsPoolMapByIDCmd.Val(), chainPoolMapByIDCmd.Val())
	if err != nil {
		return nil, err
	}

	// Sort by ID
	sort.Slice(allPools, func(i, j int) bool {
		return allPools[i].GetId() < allPools[j].GetId()
	})

	return allPools, nil
}

func (r *redisPoolsRepo) StorePools(ctx context.Context, tx mvc.Tx, pools []domain.PoolI) error {
	if err := r.addPoolsTx(ctx, tx, poolsKey, pools); err != nil {
		return err
	}

	return nil
}

func (r *redisPoolsRepo) ClearAllPools(ctx context.Context, tx mvc.Tx) error {
	// CFMM pools
	if err := r.deletePoolsTx(ctx, tx, poolsKey); err != nil {
		return err
	}

	// Concentrated pools
	if err := r.deletePoolsTx(ctx, tx, poolsKey); err != nil {
		return err
	}

	// Cosmwasm pools
	if err := r.deletePoolsTx(ctx, tx, poolsKey); err != nil {
		return err
	}
	return nil
}

func (r *redisPoolsRepo) requestPoolsAtomically(ctx context.Context, tx mvc.Tx, storeKey string) (sqsPoolMapByID *redis.MapStringStringCmd, chainPoolMapByID *redis.MapStringStringCmd, err error) {
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
func (r *redisPoolsRepo) addPoolsTx(ctx context.Context, tx mvc.Tx, storeKey string, pools []domain.PoolI) error {
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

		isConcentrated := pool.GetType() == poolmanagertypes.Concentrated

		// Write concentrated tick model
		if isConcentrated {
			tickModel, err := pool.GetTickModel()
			if err != nil {
				// Skip pool
				continue
			}

			serializedTickModel, err := json.Marshal(tickModel)
			if err != nil {
				return err
			}

			err = pipeliner.HSet(ctx, concentratedTicksModelKey(storeKey), pool.GetId(), serializedTickModel).Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// deletePoolsTx pipelines the deletion of the pools at a given storeKey to be executed atomically in a transaction.
func (r *redisPoolsRepo) deletePoolsTx(ctx context.Context, tx mvc.Tx, storeKey string) error {
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	// Note that we have 2x write and read amplification due to storage layout. We can optimize this later.
	_, err = pipeliner.HDel(ctx, sqsPoolModelKey(storeKey)).Result()
	if err != nil {
		return err
	}

	_, err = pipeliner.HDel(ctx, chainPoolModelKey(storeKey)).Result()
	if err != nil {
		return err
	}

	_, err = pipeliner.HDel(ctx, concentratedTicksModelKey(storeKey)).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetTickModelForPools implements mvc.PoolsRepository.
// CONTRACT: pools must be concentrated
func (r *redisPoolsRepo) GetTickModelForPools(ctx context.Context, pools []uint64) (map[uint64]domain.TickModel, error) {
	tx := r.repositoryManager.StartTx()

	redixTx, err := tx.AsRedisTx()
	if err != nil {
		return nil, err
	}

	pipeliner, err := redixTx.GetPipeliner(ctx)
	if err != nil {
		return nil, err
	}

	poolTickData := make([]poolTicks, 0, len(pools))
	for _, poolID := range pools {
		stringCmd := pipeliner.HGet(ctx, concentratedTicksModelKey(poolsKey), strconv.FormatUint(poolID, 10))
		poolTickData = append(poolTickData, poolTicks{
			poolID: poolID,
			Cmd:    stringCmd,
		})
	}

	if err := tx.Exec(ctx); err != nil {
		return nil, err
	}

	result := make(map[uint64]domain.TickModel, len(poolTickData))

	for _, tickCmdData := range poolTickData {
		var tickData domain.TickModel
		err = json.Unmarshal([]byte(tickCmdData.Cmd.Val()), &tickData)
		if err != nil {
			return nil, err
		}
		result[tickCmdData.poolID] = tickData
	}

	return result, nil
}

// getTicksMapByIdCmd returns a map of tick models by pool ID.
// Uses transaction to ensure atomicity.
func getTicksMapByIdCmd(ctx context.Context, tx mvc.Tx) (*redis.MapStringStringCmd, error) {
	if !tx.IsActive() {
		return nil, fmt.Errorf("tx is inactive")
	}
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return nil, err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return nil, err
	}

	ticksMapByIDCmd := pipeliner.HGetAll(ctx, concentratedTicksModelKey(poolsKey))
	return ticksMapByIDCmd, nil
}

func sqsPoolModelKey(storeKey string) string {
	return fmt.Sprintf("%s/sqs", storeKey)
}

func chainPoolModelKey(storeKey string) string {
	return fmt.Sprintf("%s/chain", storeKey)
}

func concentratedTicksModelKey(storeKey string) string {
	return fmt.Sprintf("%s/ticks", storeKey)
}
