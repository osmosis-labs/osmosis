package domain

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

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
