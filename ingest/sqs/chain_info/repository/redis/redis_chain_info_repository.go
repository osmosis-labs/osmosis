package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/json"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

type chainInfoRepo struct {
	repositoryManager mvc.TxManager
}

// TimeWrapper is a wrapper for time.Time to allow for JSON marshalling
type TimeWrapper struct {
	Time time.Time `json:"time"`
}

const (
	latestHeightKey     = "latestHeight"
	latestHeightField   = "height"
	latestHeightTimeKey = "timeLatestHeight"
)

// NewChainInfoRepo creates a new repository for chain information
func NewChainInfoRepo(repositoryManager mvc.TxManager) *chainInfoRepo {
	return &chainInfoRepo{
		repositoryManager: repositoryManager,
	}
}

// StoreLatestHeight stores the latest blockchain height into Redis
func (r *chainInfoRepo) StoreLatestHeight(ctx context.Context, tx mvc.Tx, height uint64) error {
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	heightStr := strconv.FormatUint(height, 10)
	// Use HSet for storing the latest height
	cmd := pipeliner.HSet(ctx, latestHeightKey, latestHeightField, heightStr)
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

// GetLatestHeight retrieves the latest blockchain height from Redis
func (r *chainInfoRepo) GetLatestHeight(ctx context.Context) (uint64, error) {
	tx := r.repositoryManager.StartTx()
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return 0, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return 0, err
	}

	// Use HGet for getting the latest height
	heightCmd := pipeliner.HGet(ctx, latestHeightKey, latestHeightField)

	if err := tx.Exec(ctx); err != nil {
		return 0, err
	}

	heightStr := heightCmd.Val()
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing height from Redis: %v", err)
	}

	return height, nil
}

// GetLatestHeightRetrievalTime implements mvc.ChainInfoRepository.
func (r *chainInfoRepo) GetLatestHeightRetrievalTime(ctx context.Context) (time.Time, error) {
	tx := r.repositoryManager.StartTx()
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return time.Time{}, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return time.Time{}, err
	}

	cmd := pipeliner.Get(ctx, latestHeightTimeKey)

	if err := tx.Exec(ctx); err != nil {
		return time.Time{}, err
	}

	heightStr := cmd.Val()

	var timeWrapper TimeWrapper
	if err := json.Unmarshal([]byte(heightStr), &timeWrapper); err != nil {
		return time.Time{}, err
	}

	return timeWrapper.Time, nil
}

// StoreLatestHeightRetrievalTime implements mvc.ChainInfoRepository.
func (r *chainInfoRepo) StoreLatestHeightRetrievalTime(ctx context.Context, time time.Time) error {
	tx := r.repositoryManager.StartTx()
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	timeWrapper := TimeWrapper{
		Time: time.UTC(), // always in UTC
	}

	bz, err := json.Marshal(timeWrapper)
	if err != nil {
		return err
	}

	cmd := pipeliner.Set(ctx, latestHeightTimeKey, bz, 0)

	if err := tx.Exec(ctx); err != nil {
		return err
	}

	err = cmd.Err()
	if err != nil {
		return err
	}

	return nil
}
