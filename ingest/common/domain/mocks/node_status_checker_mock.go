package mocks

import (
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
)

// NodeStatusCheckerMock is a mock implementation of domain.NodeStatusChecker.
type NodeStatusCheckerMock struct {
	// IsSyncing is the value to return when IsNodeSyncing is called.
	IsSyncing bool
	// IsNodeSyncingError is the error to return when IsNodeSyncing is called.
	IsNodeSyncingError error
	// IsNodeSyncingCalled is a flag indicating if IsNodeSyncing was called.
	IsNodeSyncingCalled bool
}

var _ domain.NodeStatusChecker = (*NodeStatusCheckerMock)(nil)

// IsNodeSyncing implements domain.NodeStatusChecker.
func (n *NodeStatusCheckerMock) IsNodeSyncing(ctx types.Context) (bool, error) {
	n.IsNodeSyncingCalled = true
	return n.IsSyncing, n.IsNodeSyncingError
}
