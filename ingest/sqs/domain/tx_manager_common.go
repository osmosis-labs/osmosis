package domain

import (
	"github.com/redis/go-redis/v9"
)

// RedisTxManager is a structure encapsulating creation of atomic transactions.
type RedisTxManager struct {
	client *redis.Client
}

var (
	_ TxManager = &RedisTxManager{}
)

// NewTxManager creates a new TxManager.
func NewTxManager(redisClient *redis.Client) TxManager {
	return &RedisTxManager{
		client: redisClient,
	}
}

// StartTx implements mvc.AtomicRepositoryManager.
func (rm *RedisTxManager) StartTx() Tx {
	return NewRedisTx(rm.client.TxPipeline())
}
