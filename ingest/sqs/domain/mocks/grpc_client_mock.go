package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v28/ingest/types"

	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
)

type GRPCClientMock struct {
	Error error
}

var _ domain.SQSGRPClient = &GRPCClientMock{}

// PushData implements domain.SQSGRPClient.
func (g *GRPCClientMock) PushData(ctx context.Context, height uint64, pools []sqsdomain.PoolI, takerFeesMap sqsdomain.TakerFeeMap) error {
	return g.Error
}
