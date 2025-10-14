package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	indexerdomain "github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

var _ indexerdomain.PairPublisher = &MockPairPublisher{}

// MockPairPublisher is a mock implementation of the PairPublisherI interface.
type MockPairPublisher struct {
	PublishPoolPairsError       error
	PublishPoolPairsCalled      bool
	CalledWithPools             []poolmanagertypes.PoolI
	CalledWithCreatedPoolIDs    map[uint64]commondomain.PoolCreation
	NumPoolPairPublished        int
	NumPoolPairWithCreationData int
}

func (m *MockPairPublisher) PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI, createdPoolIDs map[uint64]commondomain.PoolCreation) error {
	m.PublishPoolPairsCalled = true
	m.CalledWithPools = pools
	m.CalledWithCreatedPoolIDs = createdPoolIDs
	m.NumPoolPairPublished += len(pools)
	for _, pool := range pools {
		if _, ok := createdPoolIDs[pool.GetId()]; ok {
			m.NumPoolPairWithCreationData++
		}
	}
	return m.PublishPoolPairsError
}
