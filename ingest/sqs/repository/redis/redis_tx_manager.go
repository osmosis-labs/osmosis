package redis

import (
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
)

// RedisTxManager is a structure encapsulating creation of atomic transactions.
type RedisTxManager struct {
	client *redis.Client
}

var (
	_ domain.TxManager = &RedisTxManager{}
)

// NewTxManager creates a new TxManager.
func NewTxManager(redisClient *redis.Client) domain.TxManager {
	return &RedisTxManager{
		client: redisClient,
	}
}

// StartTx implements mvc.AtomicRepositoryManager.
func (rm *RedisTxManager) StartTx() domain.Tx {
	return domain.NewRedisTx(rm.client.TxPipeline())
}
