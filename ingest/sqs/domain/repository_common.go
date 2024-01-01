package domain

import (
	"context"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllPools atomically reads and returns all on-chain pools sorted by ID.
	// Note that this does NOT return tick models for the concentrated pools
	GetAllPools(context.Context) ([]PoolI, error)

	// GetPools atomically reads and returns the pools with the given IDs.
	// Note that this does NOT return tick models for the concentrated pools
	GetPools(ctx context.Context, poolIDs map[uint64]struct{}) (map[uint64]PoolI, error)

	GetTickModelForPools(ctx context.Context, pools []uint64) (map[uint64]TickModel, error)

	// StorePools atomically stores the given pools.
	StorePools(ctx context.Context, tx Tx, pools []PoolI) error
	// ClearAllPools atomically clears all pools.
	ClearAllPools(ctx context.Context, tx Tx) error
}

// RouterRepository represent the router's repository contract
type RouterRepository interface {
	GetTakerFee(ctx context.Context, denom0, denom1 string) (osmomath.Dec, error)
	GetAllTakerFees(ctx context.Context) (TakerFeeMap, error)
	SetTakerFee(ctx context.Context, tx Tx, denom0, denom1 string, takerFee osmomath.Dec) error
}

// ChainInfoRepository represents the contract for a repository handling chain information
type ChainInfoRepository interface {
	// StoreLatestHeight stores the latest blockchain height
	StoreLatestHeight(ctx context.Context, tx Tx, height uint64) error

	// GetLatestHeight retrieves the latest blockchain height
	GetLatestHeight(ctx context.Context) (uint64, error)

	// GetLatestHeightRetrievalTime retrieves the latest blockchain height retrieval time.
	GetLatestHeightRetrievalTime(ctx context.Context) (time.Time, error)

	// StoreLatestHeightRetrievalTime stores the latest blockchain height retrieval time.
	StoreLatestHeightRetrievalTime(ctx context.Context, time time.Time) error
}

// Tx defines an interface for atomic transaction.
type Tx interface {
	// Exec executes the transaction.
	// Returns an error if transaction is not in progress.
	Exec(context.Context) error

	// IsActive returns true if transaction is in progress.
	IsActive() bool

	// AsRedisTx returns a redis transaction.
	// Returns an error if this is not a redis transaction.
	AsRedisTx() (*RedisTx, error)

	// ClearAll clears all data. Returns an error if any.
	ClearAll(ctx context.Context) error
}

// RedisTx is a redis transaction.
type RedisTx struct {
	pipeliner redis.Pipeliner
}

// IsActive implements Tx.
func (rt *RedisTx) IsActive() bool {
	return rt.pipeliner != nil
}

func NewRedisTx(pipeliner redis.Pipeliner) *RedisTx {
	return &RedisTx{
		pipeliner: pipeliner,
	}
}

// Exec implements Tx.
func (rt *RedisTx) Exec(ctx context.Context) error {
	_, err := rt.pipeliner.Exec(ctx)
	rt.pipeliner = nil
	return err
}

// GetPipeliner returns a redis pipeliner for the current transaction.
// Returns an error if transaction is not in progress.
func (rt *RedisTx) GetPipeliner(ctx context.Context) (redis.Pipeliner, error) {
	if !rt.IsActive() {
		return nil, errors.New("no tx in progress")
	}

	return rt.pipeliner, nil
}

// ClearAll implements Tx.
func (rt *RedisTx) ClearAll(ctx context.Context) error {
	// TODO: can we make async flush here?
	flushCmd := rt.pipeliner.FlushAll(ctx)

	_, err := flushCmd.Result()
	if err != nil {
		return err
	}

	return nil
}

// AsRedisTx implements Tx.
func (rt *RedisTx) AsRedisTx() (*RedisTx, error) {
	return rt, nil
}

var _ Tx = &RedisTx{}

// TxManager defines an interface for atomic transaction manager.
type TxManager interface {
	// StartTx starts a new atomic transaction.
	StartTx() Tx
}

// AtomicIngester is an interface that defines the methods for the atomic ingester.
// It processes a block by writing data into a transaction.
// The caller must call Exec on the transaction to flush data to sink.
type AtomicIngester interface {
	// ProcessBlock processes the block by writing data into a transaction.
	// Returns error if fails to process.
	// It does not flush data to sink. The caller must call Exec on the transaction
	ProcessBlock(ctx sdk.Context, tx Tx) error
}
