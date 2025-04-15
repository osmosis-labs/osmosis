package service

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v30/ingest/sqs/domain"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SQSStreamingService = sqsStreamingService

func (s *sqsStreamingService) ProcessBlockRecoverError(ctx sdk.Context) error {
	return s.processBlockRecoverError(ctx)
}

type TimeAfterFunc = timeAfterFunc

func (g *GRPCClient) Connect(ctx context.Context) {
	g.connect(ctx)
}

func (g *GRPCClient) SetConn(conn domain.ClientConn) {
	g.conn = conn
}

func (g *GRPCClient) SetTimeAfterFunc(timeAfterFunc func(time.Duration) <-chan time.Time) {
	g.timeAfterFunc = timeAfterFunc
}
