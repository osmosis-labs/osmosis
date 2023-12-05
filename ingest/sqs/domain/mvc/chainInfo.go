package mvc

import (
	"context"
	"time"
)

// ChainInfoRepository represents the contract for a repository handling chain information
type ChainInfoRepository interface {
	// StoreLatestHeight stores the latest blockchain height
	StoreLatestHeight(ctx context.Context, tx Tx, height uint64) error

	// GetLatestHeight retrieves the latest blockchain height
	GetLatestHeight(ctx context.Context) (uint64, error)

	// GetLatestHeightRetrievalTime retrieves the latest blockchain height retrieval time.
	GetLatestHeightRetrievalTime(ctx context.Context) (time.Time, error)

	// StoreLatestHeightRetrievalTime stores the latest blockchain height retrieval time.
	StoreLatestHeightRetrievalTime(ctx context.Context, time time.Time) error
}

type ChainInfoUsecase interface {
	GetLatestHeight(ctx context.Context) (uint64, error)
}
