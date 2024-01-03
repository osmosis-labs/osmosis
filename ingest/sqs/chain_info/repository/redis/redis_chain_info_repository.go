package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	cmd := pipeliner.HSet(ctx, latestHeightKey, latestHeightField, heightStr)
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

// GetLatestHeight retrieves the latest blockchain height from Redis
//
// N.B. sometimes the node gets stuck and does not make progress.
// However, it returns 200 OK for the status endpoint and claims to be not catching up.
// This has caused the healthcheck to pass with false positives in production.
// As a result, we need to keep track of the last seen height that chain ingester pushes into
// the Redis repository.
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
