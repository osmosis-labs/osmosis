package ingest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
)

// IngestManager is an interface that defines the methods for the ingest manager.
// Ingest manager handles the processing of blocks and ingesting data into various sinks
// tha are defined by the Ingester interface.
type IngestManager interface {
	// ProcessBlock processes the block and ingests data into various sinks.
	// Must never panic. If panic occurs, it is silently logged and ignored.
	// If the ingester returns an error, it is silently logged and ignored.
	ProcessBlock(ctx sdk.Context)
	// SetIngester sets the ingester.
	// Note: In the future, we may expand this to support multiple ingesters.
	SetIngester(ingester Ingester)
}

// Ingester is an interface that defines the methods for the ingester.
// Ingester ingests data into a sink.
type Ingester interface {
	// ProcessBlock processes the block and ingests data into a sink.
	// Returns error if the ingester fails to ingest data.
	ProcessBlock(ctx sdk.Context) error
}

// AtomicIngester is an interface that defines the methods for the atomic ingester.
// It processes a block by writing data into a transaction.
// The caller must call Exec on the transaction to flush data to sink.
type AtomicIngester interface {
	// ProcessBlock processes the block by writing data into a transaction.
	// Returns error if fails to process.
	// It does not flush data to sink. The caller must call Exec on the transaction
	ProcessBlock(ctx sdk.Context, tx mvc.Tx) error

	SetLogger(log.Logger)
}

// ingesterImpl is an implementation of IngesterManager.
type ingestManagerImpl struct {
	ingester Ingester
}

var _ IngestManager = &ingestManagerImpl{}

// NewIngestManager creates a new IngestManager.
func NewIngestManager() IngestManager {
	return &ingestManagerImpl{
		ingester: nil,
	}
}

// ProcessBlock implements IngestManager.
func (im *ingestManagerImpl) ProcessBlock(ctx sdk.Context) {
	defer func() {
		if r := recover(); r != nil {
			// Panics are silently logged and ignored.
			ctx.Logger().Error("panic while processing block during ingest", "err", r)
		}
	}()

	// Ingester must be set in the app. If not, we do nothing.
	if im.ingester != nil {
		if err := im.ingester.ProcessBlock(ctx); err != nil {
			// The error is silently logged and ignored.
			ctx.Logger().Error("error processing block during ingest", "err", err)
		}
	}
}

// SetIngester implements IngestManager.
func (im *ingestManagerImpl) SetIngester(ingester Ingester) {
	im.ingester = ingester
}
