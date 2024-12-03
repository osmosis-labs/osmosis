package mocks

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
)

type PoolsTransformerMock struct {
	PoolReturn     []sqsdomain.PoolI
	TakerFeeReturn sqsdomain.TakerFeeMap
	ErrReturn      error
}

var _ domain.PoolsTransformer = &PoolsTransformerMock{}

// Transform implements domain.PoolsTransformer.
func (p *PoolsTransformerMock) Transform(ctx types.Context, blockPools commondomain.BlockPools) ([]sqsdomain.PoolI, sqsdomain.TakerFeeMap, error) {
	return p.PoolReturn, p.TakerFeeReturn, p.ErrReturn
}
