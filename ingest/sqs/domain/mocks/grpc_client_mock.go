package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/types"
)

type GRPCClientMock struct {
	Error error
}

var _ domain.SQSGRPClient = &GRPCClientMock{}

// PushData implements domain.SQSGRPClient.
func (g *GRPCClientMock) PushData(ctx context.Context, height uint64, pools []types.PoolI, takerFeesMap types.TakerFeeMap) error {
	return g.Error
}
