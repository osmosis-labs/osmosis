package observer_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

type MockChain struct {
	In               chan observer.Transfer
	Out              chan observer.Transfer
	HeightRes        atomic.Uint64
	ConfirmationsRes atomic.Uint64
}

func NewMockChain(h uint64, cr uint64) *MockChain {
	mc := MockChain{
		In:               make(chan observer.Transfer),
		Out:              make(chan observer.Transfer),
		HeightRes:        atomic.Uint64{},
		ConfirmationsRes: atomic.Uint64{},
	}
	mc.HeightRes.Store(h)
	mc.ConfirmationsRes.Store(cr)
	return &mc
}

func (m *MockChain) SignalInboundTransfer(ctx context.Context, in observer.Transfer) error {
	m.In <- in
	return nil
}

func (m *MockChain) ListenOutboundTransfer() <-chan observer.Transfer {
	return m.Out
}

func (m *MockChain) Start(context.Context) error {
	return nil
}

func (m *MockChain) Stop(context.Context) error {
	return nil
}

func (m *MockChain) Height(context.Context) (uint64, error) {
	return m.HeightRes.Load(), nil
}

func (m *MockChain) ConfirmationsRequired(context.Context, bridgetypes.AssetID) (uint64, error) {
	return m.ConfirmationsRes.Load(), nil
}

// TestObserver verifies Observer properly receives transfers from src chains
// and sends them to dst chain
func TestObserver(t *testing.T) {
	osmo := NewMockChain(15, 3)
	btc := NewMockChain(15, 3)
	chains := make(map[observer.ChainId]observer.Client)
	chains[observer.ChainIdOsmosis] = osmo
	chains[observer.ChainIdBitcoin] = btc
	o := observer.NewObserver(log.NewNopLogger(), chains, 100*time.Millisecond)

	ctx := context.Background()
	err := o.Start(ctx)
	require.NoError(t, err)

	btcOut := observer.Transfer{
		SrcChain: observer.ChainIdBitcoin,
		DstChain: observer.ChainIdOsmosis,
		Id:       "from-btc",
		Height:   10,
		Sender:   "btc-sender",
		To:       "osmo-recipient",
		Asset:    bridgetypes.DefaultBitcoinDenomName,
		Amount:   math.NewUint(10),
	}
	btc.Out <- btcOut
	osmoIn := observer.Transfer{}
	require.Eventually(t, func() bool {
		osmoIn = <-osmo.In
		return true
	}, time.Second, 100*time.Millisecond, "Timeout receiving transfer")
	require.Equal(t, btcOut, osmoIn)

	err = o.Stop(ctx)
	require.NoError(t, err)
}

// TestObserverLowConfirmation verifies Observer sends transfers to
// dst chains only when there is enough confirmations
func TestObserverLowConfirmation(t *testing.T) {
	osmo := NewMockChain(15, 3)
	btc := NewMockChain(15, 3)
	chains := make(map[observer.ChainId]observer.Client)
	chains[observer.ChainIdOsmosis] = osmo
	chains[observer.ChainIdBitcoin] = btc
	o := observer.NewObserver(log.NewNopLogger(), chains, 100*time.Millisecond)

	ctx := context.Background()
	err := o.Start(ctx)
	require.NoError(t, err)

	btcOut := observer.Transfer{
		SrcChain: observer.ChainIdBitcoin,
		DstChain: observer.ChainIdOsmosis,
		Id:       "from-btc",
		Height:   15,
		Sender:   "btc-sender",
		To:       "osmo-recipient",
		Asset:    bridgetypes.DefaultBitcoinDenomName,
		Amount:   math.NewUint(10),
	}
	btc.Out <- btcOut
	osmoIn := observer.Transfer{}
	received := false
	select {
	case osmoIn = <-osmo.In:
		received = true
	case <-time.After(time.Millisecond * 500):
	}
	require.False(t, received)
	btc.HeightRes.Store(20)
	require.Eventually(t, func() bool {
		osmoIn = <-osmo.In
		return true
	}, time.Second, 100*time.Millisecond, "Timeout receiving transfer")
	require.Equal(t, btcOut, osmoIn)

	err = o.Stop(ctx)
	require.NoError(t, err)
}
