package domain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// Ingester is an interface that defines the methods for the ingester.
// Ingester ingests data into a sink.
type Ingester interface {
	// ProcessAllBlockData processes the block and ingests data into a sink.
	// Returns error if the ingester fails to ingest data.
	// Also returns cwpools, which is used to create the initial address to pool mapping.
	ProcessAllBlockData(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)

	// ProcessChangedBlockData processes only the pools that were changed in the block.
	ProcessChangedBlockData(ctx sdk.Context, changedPools BlockPools) error
}

// PoolsTransformer is an interface that defines the methods for the pool transformer
type PoolsTransformer interface {
	// Transform processes the pool state, returning pools instrumented with all the necessary chain data.
	// Additionally, returns the taker fee map for every pool denom pair.
	// Returns error if the transformer fails to process pool data.
	Transform(ctx sdk.Context, blockPools BlockPools) ([]sqsdomain.PoolI, sqsdomain.TakerFeeMap, error)
}

// BlockPools contains the pools to be ingested in a block.
type BlockPools struct {
	// ConcentratedPools are the concentrated pools to be ingested.
	ConcentratedPools []poolmanagertypes.PoolI
	// ConcentratedPoolIDTickChange is the map of pool ID to tick change for concentrated pools.
	// We use these pool IDs to append concentrated pools with all ticks at the end of the block.
	ConcentratedPoolIDTickChange map[uint64]struct{}
	// CosmWasmPools are the CosmWasm pools to be ingested.
	CosmWasmPools []poolmanagertypes.PoolI
	// CFMMPools are the CFMM pools to be ingested.
	CFMMPools []poolmanagertypes.PoolI
}

// SQSGRPClient is an interface that defines the methods for the graceful SQS GRPC client.
// It handles graceful connection management. So that, if a GRPC ingest method returns status.Unavailable,
// the GRPC client will reset the connection and attempt to recreate it before retrying the ingest method.
type SQSGRPClient interface {
	// PushData pushes the height, pools and taker fee data to SQS via GRPC.
	// Returns error if the GRPC client fails to push data.
	// On status.Unavailable, it closes the connection and attempts to re-establish it during the next GRPC call.
	// Note: while there are built-in mechanisms to handle retry such as exponential backoff, they are no suitable for our context.
	// In our context, we would rather continue attempting to repush the data in the next block instead of blocking the system.
	PushData(ctx context.Context, height uint64, pools []sqsdomain.PoolI, takerFeesMap sqsdomain.TakerFeeMap) error
}
