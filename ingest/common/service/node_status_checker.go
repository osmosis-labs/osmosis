package service

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v26/ingest/sqs/domain"
)

type nodeStatusChecker struct {
	// example format: tcp://localhost:26657
	address string
}

var _ domain.NodeStatusChecker = (*nodeStatusChecker)(nil)

func NewNodeStatusChecker(address string) domain.NodeStatusChecker {
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
