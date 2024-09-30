package service

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NodeStatusChecker interface {
	// IsNodeSyncing checks if the node is syncing.
	// Returns true if the node is syncing, false otherwise.
	// Returns error if the node syncing status cannot be determined.
	IsNodeSyncing(ctx sdk.Context) (bool, error)
}

type nodeStatusChecker struct {
	// example format: tcp://localhost:26657
	address string
}

var _ NodeStatusChecker = (*nodeStatusChecker)(nil)

func NewNodeStatusChecker(address string) NodeStatusChecker {
	return &nodeStatusChecker{
		address: address,
	}
}

// IsNodeSyncing implements NodeStatusChecker.
func (n *nodeStatusChecker) IsNodeSyncing(ctx sdk.Context) (bool, error) {
	// Get client
	client, err := client.NewClientFromNode(n.address)
	if err != nil {
		return false, err
	}

	// Get status
	statusResult, err := client.Status(ctx)
	if err != nil {
		return false, err
	}

	// Return if node is catching up
	return statusResult.SyncInfo.CatchingUp, nil
}
