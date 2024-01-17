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
	GetLatestHeight(ctx context.Context) (uint64, error)
}
