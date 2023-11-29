package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

type chainInfoRepo struct {
	repositoryManager mvc.TxManager
}

const (
	latestHeightKey   = "latestHeight"
	latestHeightField = "height"
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
	err = pipeliner.HSet(ctx, latestHeightKey, latestHeightField, heightStr).Err()
	if err != nil {
		return err
	}

	return tx.Exec(ctx)
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
