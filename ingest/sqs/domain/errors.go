package domain

import "errors"

var (
	ErrNodeIsSyncing = errors.New("node is syncing, skipping block processing")
)
