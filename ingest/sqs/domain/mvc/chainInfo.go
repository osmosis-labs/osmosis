package mvc

import (
	"context"
)

// ChainInfoRepository represents the contract for a repository handling chain information
type ChainInfoRepository interface {
	// StoreLatestHeight stores the latest blockchain height
	StoreLatestHeight(ctx context.Context, tx Tx, height uint64) error

	// GetLatestHeight retrieves the latest blockchain height
	GetLatestHeight(ctx context.Context) (uint64, error)
}

type ChainInfoUsecase interface {
	// GetLatestHeight retrieves the latest blockchain height
	//
	// Despite being a getter, this method also validates that the height is updated within a reasonable time frame.
	//
	// Sometimes the node gets stuck and does not make progress.
	// However, it returns 200 OK for the status endpoint and claims to be not catching up.
	// This has caused the healthcheck to pass with false positives in production.
	// As a result, we need to keep track of the last seen height and time. Chain ingester pushes
	// the latest height into Redis. This method checks that the height is updated within a reasonable time frame.
	GetLatestHeight(ctx context.Context) (uint64, error)
}
