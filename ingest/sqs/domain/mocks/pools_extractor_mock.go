package mocks

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
)

type PoolsExtractorMock struct {
	// AllBlockDataError is the error to return when ProcessAllBlockData is called.
	AllBlockDataError error
	// ChangedBlockDataError is the error to return when ProcessChangedBlockData is called.
	ChangedBlockDataError error
	// IsProcessAllBlockDataCalled is a flag indicating if ProcessAllBlockData was called.
	IsProcessAllBlockDataCalled bool
	// IsProcessAllChangedDataCalled is a flag indicating if ProcessChangedBlockData was called.
	IsProcessAllChangedDataCalled bool
	// IsProcessAllCreatedDataCalled is a flag indicating if ProcessAllCreatedDataCalled was called.
	IsProcessAllCreatedDataCalled bool
	// IsPoolTrackerReset is a flag indicating if ResetPoolTracker was called.
	IsPoolTrackerReset bool
	// If this is non-empty, ProcessAllBlockData(...) will panic with this message.
	ProcessAllBlockDataPanicMsg string
	// Block pools to return
	BlockPools commondomain.BlockPools
	// CreatedPoolIDs is the map of pool IDs that were created in the block.
	CreatedPoolIDs map[uint64]commondomain.PoolCreation
}

var _ commondomain.PoolExtractor = &PoolsExtractorMock{}

// ExtractAll implements commondomain.PoolExtractor.
func (p *PoolsExtractorMock) ExtractAll(ctx types.Context) (commondomain.BlockPools, map[uint64]commondomain.PoolCreation, error) {
	if p.ProcessAllBlockDataPanicMsg != "" {
		panic(p.ProcessAllBlockDataPanicMsg)
	}

	p.IsProcessAllBlockDataCalled = true
	return p.BlockPools, p.CreatedPoolIDs, p.AllBlockDataError
}

// ExtractChanged implements commondomain.PoolExtractor.
func (p *PoolsExtractorMock) ExtractChanged(ctx types.Context) (commondomain.BlockPools, error) {
	p.IsProcessAllChangedDataCalled = true
	return p.BlockPools, p.ChangedBlockDataError
}

// ExtractCreated implements commondomain.PoolExtractor.
func (p *PoolsExtractorMock) ExtractCreated(ctx types.Context) (commondomain.BlockPools, map[uint64]commondomain.PoolCreation, error) {
	p.IsProcessAllCreatedDataCalled = true
	return p.BlockPools, p.CreatedPoolIDs, nil
}

// ResetPoolTracker implements commondomain.PoolExtractor.
func (p *PoolsExtractorMock) ResetPoolTracker(ctx types.Context) {
	p.IsPoolTrackerReset = true
}
