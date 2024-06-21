package mocks

import (
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

var _ domain.Ingester = &SQSIngesterMock{}

// SQSIngesterMock is a mock implementation of domain.Ingester.
type SQSIngesterMock struct {
	// AllBlockDataError is the error to return when ProcessAllBlockData is called.
	AllBlockDataError error
	// ChangedBlockDataError is the error to return when ProcessChangedBlockData is called.
	ChangedBlockDataError error
	// IsProcessAllBlockDataCalled is a flag indicating if ProcessAllBlockData was called.
	IsProcessAllBlockDataCalled bool
	// IsProcessAllChangedDataCalled is a flag indicating if ProcessChangedBlockData was called.
	IsProcessAllChangedDataCalled bool
	// LastChangedPoolsObserved is the last changed pools observed by the mock when
	// ProcessChangedBlockData is called.
	LastChangedPoolsObserved domain.BlockPools
	// If this is non-empty, ProcessAllBlockData(...) will panic with this message.
	ProcessAllBlockDataPanicMsg string
}

// ProcessAllBlockData implements domain.Ingester.
func (s *SQSIngesterMock) ProcessAllBlockData(ctx types.Context) ([]poolmanagertypes.PoolI, error) {
	if s.ProcessAllBlockDataPanicMsg != "" {
		panic(s.ProcessAllBlockDataPanicMsg)
	}

	s.IsProcessAllBlockDataCalled = true
	return s.LastChangedPoolsObserved.CosmWasmPools, s.AllBlockDataError
}

// ProcessChangedBlockData implements domain.Ingester.
func (s *SQSIngesterMock) ProcessChangedBlockData(ctx types.Context, changedPools domain.BlockPools) error {
	s.IsProcessAllChangedDataCalled = true
	s.LastChangedPoolsObserved = changedPools
	return s.ChangedBlockDataError
}
