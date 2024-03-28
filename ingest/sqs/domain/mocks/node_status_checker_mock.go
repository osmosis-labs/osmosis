package mocks

import (
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

// NodeStatusCheckerMock is a mock implementation of domain.NodeStatusChecker.
type NodeStatusCheckerMock struct {
	// IsSynching is the value to return when IsNodeSynching is called.
	IsSynching bool
	// IsNodeSynchingError is the error to return when IsNodeSynching is called.
	IsNodeSynchingError error
	// IsNodeSynchingCalled is a flag indicating if IsNodeSynching was called.
	IsNodeSynchingCalled bool
}

var _ domain.NodeStatusChecker = (*NodeStatusCheckerMock)(nil)

// IsNodeSynching implements domain.NodeStatusChecker.
func (n *NodeStatusCheckerMock) IsNodeSynching(ctx types.Context) (bool, error) {
	n.IsNodeSynchingCalled = true
	return n.IsSynching, n.IsNodeSynchingError
}
