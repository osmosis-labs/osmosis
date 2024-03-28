package domain

import "errors"

var (
	ErrNodeIsSynching = errors.New("node is synching, skipping block processing")
)
