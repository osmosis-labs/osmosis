package sqs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain/repository"

	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

const sqsIngesterName = "sidecar-query-server"

var _ ingest.Ingester = &sqsIngester{}

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	txManager         repository.TxManager
	poolsIngester     domain.AtomicIngester
	chainInfoIngester domain.AtomicIngester
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester, chainInfoIngester domain.AtomicIngester, txManager repository.TxManager) ingest.Ingester {
	return &sqsIngester{
		txManager:         txManager,
		chainInfoIngester: chainInfoIngester,
		poolsIngester:     poolsIngester,
	}
}

// ProcessBlock implements ingest.Ingester.
func (i *sqsIngester) ProcessBlock(ctx sdk.Context) error {
	// Start atomic transaction
	tx := i.txManager.StartTx()

	goCtx := sdk.WrapSDKContext(ctx)

	// Process block by reading and writing data and ingesting data into sinks
	if err := i.poolsIngester.ProcessBlock(ctx, tx); err != nil {
		return err
	}

	// Process block by reading and writing data and ingesting data into sinks
	if err := i.chainInfoIngester.ProcessBlock(ctx, tx); err != nil {
		return err
	}

	// Flush all writes atomically
	return tx.Exec(goCtx)
}

// GetName implements ingest.Ingester.
func (*sqsIngester) GetName() string {
	return sqsIngesterName
}
