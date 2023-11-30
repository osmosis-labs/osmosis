package ingest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IngestManager is an interface that defines the methods for the ingest manager.
// Ingest manager handles the processing of blocks and ingesting data into various sinks
// tha are defined by the Ingester interface.
type IngestManager interface {
	// RegisterIngester registers an ingester.
	RegisterIngester(ingester Ingester)
	// ProcessBlock processes the block and ingests data into various sinks.
	// Must never panic. If panic occurs, it is silently logged and ignored.
	// If the ingester returns an error, it is silently logged and ignored.
	ProcessBlock(ctx sdk.Context)
}

// Ingester is an interface that defines the methods for the ingester.
// Ingester ingests data into a sink.
type Ingester interface {
	// ProcessBlock processes the block and ingests data into a sink.
	// Returns error if the ingester fails to ingest data.
	ProcessBlock(ctx sdk.Context) error

	GetName() string
}

// ingesterImpl is an implementation of IngesterManager.
type ingestManagerImpl struct {
	ingesters []Ingester
}

var _ IngestManager = &ingestManagerImpl{}

// NewIngestManager creates a new IngestManager.
func NewIngestManager() IngestManager {
	return &ingestManagerImpl{
		ingesters: []Ingester{},
	}
}

// RegisterIngester implements IngestManager.
func (im *ingestManagerImpl) RegisterIngester(ingester Ingester) {
	im.ingesters = append(im.ingesters, ingester)
}

// ProcessBlock implements IngestManager.
func (im *ingestManagerImpl) ProcessBlock(ctx sdk.Context) {
	defer func() {
		if r := recover(); r != nil {
			// Panics are silently logged and ignored.
			ctx.Logger().Error("panic while processing block during ingest", "err", r)
		}
	}()

	// Ingesters must be set in the app. If not, we do nothing.
	for _, ingester := range im.ingesters {
		if err := ingester.ProcessBlock(ctx); err != nil {
			// The error is silently logged and ignored.
			ctx.Logger().Error("error processing block during ingest", "err", err, "ingester", ingester.GetName())
		}
	}
}
