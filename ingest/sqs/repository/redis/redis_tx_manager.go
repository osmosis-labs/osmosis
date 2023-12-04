package redis

import (
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

// RedisTxManager is a structure encapsulating creation of atomic transactions.
type RedisTxManager struct {
	client *redis.Client
}

var (
	_ mvc.TxManager = &RedisTxManager{}
)

// NewTxManager creates a new TxManager.
func NewTxManager(redisClient *redis.Client) mvc.TxManager {
	return &RedisTxManager{
		client: redisClient,
	}
}

// StartTx implements mvc.AtomicRepositoryManager.
func (rm *RedisTxManager) StartTx() mvc.Tx {
	return mvc.NewRedisTx(rm.client.TxPipeline())
}
