package poolextractor

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/sqs/domain"
)

// poolExtractor is an abstraction that extracts pools from the chain.
type poolExtractor struct {
	keepers     commondomain.PoolExtractorKeepers
	poolTracker domain.BlockPoolUpdateTracker
}

// New creates a new pool extractor.
func New(keepers commondomain.PoolExtractorKeepers, poolTracker domain.BlockPoolUpdateTracker) commondomain.PoolExtractor {
	return &poolExtractor{
		keepers:     keepers,
		poolTracker: poolTracker,
	}
}

// ExtractAll implements commondomain.PoolExtractor.
func (p *poolExtractor) ExtractAll(ctx sdk.Context) (commondomain.BlockPools, map[uint64]commondomain.PoolCreation, error) {
	// If cold start, we process all the pools in the chain.
	// Get the pool IDs that were created in the block where cold start happens
	createdPoolIDs := p.poolTracker.GetCreatedPoolIDs()

	// Concentrated pools

	concentratedPools, err := p.keepers.ConcentratedKeeper.GetPools(ctx)
	if err != nil {
		return commondomain.BlockPools{}, nil, err
	}

	// CFMM pools

	cfmmPools, err := p.keepers.GammKeeper.GetPools(ctx)
	if err != nil {
		return commondomain.BlockPools{}, nil, err
	}

	// CosmWasm pools

	cosmWasmPools, err := p.keepers.CosmWasmPoolKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return commondomain.BlockPools{}, nil, err
	}

	// Generate the initial cwPool address to pool mapping
	for _, pool := range cosmWasmPools {
		p.poolTracker.TrackCosmWasmPoolsAddressToPoolMap(pool)
	}

	blockPools := commondomain.BlockPools{
		ConcentratedPools: concentratedPools,
		CosmWasmPools:     cosmWasmPools,
		CFMMPools:         cfmmPools,
	}

	return blockPools, createdPoolIDs, nil
}

// ExtractChanged implements commondomain.PoolExtractor.
func (p *poolExtractor) ExtractChanged(ctx sdk.Context) (commondomain.BlockPools, error) {
	// If not cold start, we only process the pools that were changed this block.
	concentratedPools := p.poolTracker.GetConcentratedPools()
	concentratedPoolIDTickChange := p.poolTracker.GetConcentratedPoolIDTickChange()
	cfmmPools := p.poolTracker.GetCFMMPools()
	cosmWasmPools := p.poolTracker.GetCosmWasmPools()

	changedBlockPools := commondomain.BlockPools{
		ConcentratedPools:            concentratedPools,
		ConcentratedPoolIDTickChange: concentratedPoolIDTickChange,
		CosmWasmPools:                cosmWasmPools,
		CFMMPools:                    cfmmPools,
	}

	poolIDsTracked := make(map[uint64]struct{}, len(changedBlockPools.ConcentratedPools))

	// Copy over the pools that were changed in the block
	for _, pool := range changedBlockPools.ConcentratedPools {
		changedBlockPools.ConcentratedPoolIDTickChange[pool.GetId()] = struct{}{}

		poolIDsTracked[pool.GetId()] = struct{}{}
	}

	// Update concentrated pools
	for poolID := range concentratedPoolIDTickChange {
		// Skip if the pool if it is already tracked
		if _, ok := poolIDsTracked[poolID]; ok {
			continue
		}

		pool, err := p.keepers.ConcentratedKeeper.GetConcentratedPoolById(ctx, poolID)
		if err != nil {
			return commondomain.BlockPools{}, err
		}

		changedBlockPools.ConcentratedPools = append(changedBlockPools.ConcentratedPools, pool)
	}

	return changedBlockPools, nil
}

// ExtractCreated implements commondomain.PoolExtractor.
func (p *poolExtractor) ExtractCreated(ctx sdk.Context) (commondomain.BlockPools, map[uint64]commondomain.PoolCreation, error) {
	changedPools, err := p.ExtractChanged(ctx)
	if err != nil {
		return commondomain.BlockPools{}, nil, err
	}

	createdPoolIDs := p.poolTracker.GetCreatedPoolIDs()

	result := commondomain.BlockPools{
		ConcentratedPoolIDTickChange: make(map[uint64]struct{}),
	}

	// Copy over the pools that were created in the block

	// CFMM
	for _, pool := range changedPools.CFMMPools {
		if _, ok := createdPoolIDs[pool.GetId()]; ok {
			result.CFMMPools = append(result.CFMMPools, pool)
		}
	}

	// CosmWasm
	for _, pool := range changedPools.CosmWasmPools {
		if _, ok := createdPoolIDs[pool.GetId()]; ok {
			result.CosmWasmPools = append(result.CosmWasmPools, pool)
		}
	}

	// Concentrated
	for _, pool := range changedPools.ConcentratedPools {
		if _, ok := createdPoolIDs[pool.GetId()]; ok {
			result.ConcentratedPools = append(result.ConcentratedPools, pool)
		}
	}

	// Concentrated ticks
	for poolID := range changedPools.ConcentratedPoolIDTickChange {
		if _, ok := createdPoolIDs[poolID]; ok {
			result.ConcentratedPoolIDTickChange[poolID] = struct{}{}
		}
	}

	return result, createdPoolIDs, nil
}
