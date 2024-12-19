package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v28/ingest/types"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
)

type PoolsTransformerMock struct {
	PoolReturn     []types.PoolI
	TakerFeeReturn types.TakerFeeMap
	ErrReturn      error
}

var _ domain.PoolsTransformer = &PoolsTransformerMock{}

// Transform implements domain.PoolsTransformer.
func (p *PoolsTransformerMock) Transform(ctx sdk.Context, blockPools commondomain.BlockPools) ([]types.PoolI, types.TakerFeeMap, error) {
	return p.PoolReturn, p.TakerFeeReturn, p.ErrReturn
}
