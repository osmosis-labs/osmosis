package observer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// osmosis.bridge.v1beta1.EventOutboundTransfer
// osmosis.bridge.v1beta1.EventInboundTransfer

func TestObserver(t *testing.T) {
	// rpcUrl := "https://rpc.testnet.osmosis.zone:443" // Osmosis testnet
	rpcUrl := "http://localhost:26657" // Local net
	observer, err := NewObesrver(rpcUrl)
	require.NoError(t, err)

	// query := "tm.event = 'osmosis.bridge.v1beta1.EventOutboundTransfer'"
	query := "tm.event = 'NewBlock'"
	observer.Start(query)

	time.Sleep(time.Second * 10)

	require.NoError(t, observer.Stop())
}
