package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ingesttypes "github.com/osmosis-labs/osmosis/v31/ingest/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v31/ingest/sqs/domain"
)

type PoolsTransformerMock struct {
	PoolReturn     []ingesttypes.PoolI
	TakerFeeReturn ingesttypes.TakerFeeMap
	ErrReturn      error
}

var _ domain.PoolsTransformer = &PoolsTransformerMock{}

// Transform implements domain.PoolsTransformer.
func (p *PoolsTransformerMock) Transform(ctx sdk.Context, blockPools commondomain.BlockPools) ([]ingesttypes.PoolI, ingesttypes.TakerFeeMap, error) {
	return p.PoolReturn, p.TakerFeeReturn, p.ErrReturn
}
