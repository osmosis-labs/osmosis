package mocks

import (
	"sync"
	"context"

	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/domain"
	ingesttypes "github.com/osmosis-labs/osmosis/v29/ingest/types"

	"google.golang.org/grpc/connectivity"
)

type ClientConn struct {
	mu              sync.Mutex
	State           connectivity.State
	StateChanges    []connectivity.State
	waitShouldBlock bool
}

func (m *ClientConn) GetState() connectivity.State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.State
}

func (m *ClientConn) Connect() {}

var _ domain.ClientConn = &ClientConn{}

func (m *ClientConn) WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool {
	if m.waitShouldBlock {
		<-ctx.Done()
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.StateChanges) > 0 {
		m.State = m.StateChanges[0]
		m.StateChanges = m.StateChanges[1:]
	}

	return true
}

type GRPCClientMock struct {
	Error error
}

var _ domain.SQSGRPClient = &GRPCClientMock{}

// PushData implements domain.SQSGRPClient.
func (g *GRPCClientMock) PushData(ctx context.Context, height uint64, pools []ingesttypes.PoolI, takerFeesMap ingesttypes.TakerFeeMap) error {
	return g.Error
}

// IsConnected implements domain.SQSGRPClient.
func (g *GRPCClientMock) IsConnected() error {
	return g.Error
}
