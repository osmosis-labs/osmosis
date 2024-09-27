package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	indexerdomain "github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v26/x/poolmanager/types"
)

var _ indexerdomain.PairPublisher = &MockPairPublisher{}

// MockPairPublisher is a mock implementation of the PairPublisherI interface.
type MockPairPublisher struct {
	PublishPoolPairsError    error
	PublishPoolPairsCalled   bool
	NumPoolsPublished        int
	NumPoolsWithCreationData int
}

func (m *MockPairPublisher) PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI, createdPoolIDs map[uint64]commondomain.PoolCreation) error {
	m.PublishPoolPairsCalled = true
	m.NumPoolsPublished += len(pools)
	for _, pool := range pools {
		if _, ok := createdPoolIDs[pool.GetId()]; ok {
			m.NumPoolsWithCreationData++
		}
	}
	return m.PublishPoolPairsError
}
