package poolextractor

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
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
func (p *poolExtractor) ExtractAll(ctx sdk.Context) (commondomain.BlockPools, error) {
	// Concentrated pools

	concentratedPools, err := p.keepers.ConcentratedKeeper.GetPools(ctx)
	if err != nil {
		return commondomain.BlockPools{}, err
	}

	// CFMM pools

	cfmmPools, err := p.keepers.GammKeeper.GetPools(ctx)
	if err != nil {
		return commondomain.BlockPools{}, err
	}

	// CosmWasm pools

	cosmWasmPools, err := p.keepers.CosmWasmPoolKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return commondomain.BlockPools{}, err
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

	return blockPools, nil
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
