package observer_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer"
)

type MockChain struct {
	In  chan observer.InboundTransfer
	Out chan observer.OutboundTransfer
	H   uint64
	CR  uint64
}

func NewMockChain(h uint64, cr uint64) *MockChain {
	return &MockChain{
		In:  make(chan observer.InboundTransfer),
		Out: make(chan observer.OutboundTransfer),
		H:   h,
		CR:  cr,
	}
}

func (m *MockChain) SignalInboundTransfer(ctx context.Context, in observer.InboundTransfer) error {
	m.In <- in
	return nil
}

func (m *MockChain) ListenOutboundTransfer() <-chan observer.OutboundTransfer {
	return m.Out
}

func (m *MockChain) Start(context.Context) error {
	return nil
}

func (m *MockChain) Stop() error {
	return nil
}

func (m *MockChain) Height() (uint64, error) {
	return m.H, nil
}

func (m *MockChain) ConfirmationsRequired() (uint64, error) {
	return m.CR, nil
}

// TestObserver verifies Observer properly receives transfers from src chains
// and sends them to dst chain
func TestObserver(t *testing.T) {
	osmo := NewMockChain(15, 3)
	btc := NewMockChain(15, 3)
	chains := make(map[observer.ChainId]observer.Chain)
	chains[observer.ChainId_OSMO] = osmo
	chains[observer.ChainId_BITCOIN] = btc
	o := observer.NewObserver(log.NewNopLogger(), chains, 100*time.Millisecond)

	ctx := context.Background()
	err := o.Start(ctx)
	require.NoError(t, err)

	btcOut := observer.OutboundTransfer{
		DstChain: observer.ChainId_OSMO,
		Id:       "from-btc",
		Height:   10,
		Sender:   "btc-sender",
		To:       "osmo-recipient",
		Asset:    "btc",
		Amount:   math.NewUint(10),
	}
	expOsmoIn := observer.InboundTransfer{
		SrcChain: observer.ChainId_BITCOIN,
		Id:       btcOut.Id,
		Height:   btcOut.Height,
		Sender:   btcOut.Sender,
		To:       btcOut.To,
		Asset:    btcOut.Asset,
		Amount:   btcOut.Amount,
	}

	btc.Out <- btcOut
	osmoIn := observer.InboundTransfer{}
	require.Eventually(t, func() bool {
		osmoIn = <-osmo.In
		return true
	}, time.Second, 100*time.Millisecond, "Timeout receiving transfer")
	require.Equal(t, expOsmoIn, osmoIn)

	err = o.Stop()
	require.NoError(t, err)
}

// TestObserverLowConfirmation verifies Observer sends transfers to
// dst chains only when there is enough confirmations
func TestObserverLowConfirmation(t *testing.T) {
	osmo := NewMockChain(15, 3)
	btc := NewMockChain(15, 3)
	chains := make(map[observer.ChainId]observer.Chain)
	chains[observer.ChainId_OSMO] = osmo
	chains[observer.ChainId_BITCOIN] = btc
	o := observer.NewObserver(log.NewNopLogger(), chains, 100*time.Millisecond)

	ctx := context.Background()
	err := o.Start(ctx)
	require.NoError(t, err)

	btcOut := observer.OutboundTransfer{
		DstChain: observer.ChainId_OSMO,
		Id:       "from-btc",
		Height:   15,
		Sender:   "btc-sender",
		To:       "osmo-recipient",
		Asset:    "btc",
		Amount:   math.NewUint(10),
	}
	expOsmoIn := observer.InboundTransfer{
		SrcChain: observer.ChainId_BITCOIN,
		Id:       btcOut.Id,
		Height:   btcOut.Height,
		Sender:   btcOut.Sender,
		To:       btcOut.To,
		Asset:    btcOut.Asset,
		Amount:   btcOut.Amount,
	}

	btc.Out <- btcOut
	osmoIn := observer.InboundTransfer{}
	received := false
	select {
	case osmoIn = <-osmo.In:
		received = true
	case <-time.After(time.Millisecond * 500):
	}
	require.False(t, received)
	btc.H = 20
	require.Eventually(t, func() bool {
		osmoIn = <-osmo.In
		return true
	}, time.Second, 100*time.Millisecond, "Timeout receiving transfer")
	require.Equal(t, expOsmoIn, osmoIn)

	err = o.Stop()
	require.NoError(t, err)
}
