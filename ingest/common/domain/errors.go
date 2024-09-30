package commondomain

import (
	"errors"
	"fmt"
)

var (
	ErrNodeIsSyncing = errors.New("node is syncing, skipping block processing")
)

type NodeSyncCheckError struct {
	Err error
}

func (e *NodeSyncCheckError) Error() string {
	return fmt.Sprintf("failed to check if node is syncing: %v", e.Err)
}
