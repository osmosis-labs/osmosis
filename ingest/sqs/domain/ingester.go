package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain/repository"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v24/x/poolmanager/types"
)

// Ingester is an interface that defines the methods for the ingester.
// Ingester ingests data into a sink.
type Ingester interface {
	// ProcessAllBlockData processes the block and ingests data into a sink.
	// Returns error if the ingester fails to ingest data.
	ProcessAllBlockData(ctx sdk.Context) error

	// ProcessChangedBlockData processes only the pools that were changed in the block.
	ProcessChangedBlockData(ctx sdk.Context, changedPools BlockPools) error
}

// PoolIngester is an interface that defines the methods for the pool ingester.
type PoolIngester interface {
	// ProcessPoolState processes the pool state and ingests data into a sink.
	// It appends all updates into a transaction for atomic commit at the end of the block.
	// Returns error if the ingester fails to process pool data.
	ProcessPoolState(ctx sdk.Context, tx repository.Tx, blockPools BlockPools) error
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
