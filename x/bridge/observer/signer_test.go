package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// osmosis.bridge.v1beta1.EventOutboundTransfer
// osmosis.bridge.v1beta1.EventInboundTransfer

func TestObserver(t *testing.T) {
	t.Skip("Requires connection to a node")

	rpcUrl := "http://localhost:26657" // Local net
	signer, err := NewSigner(rpcUrl)
	require.NoError(t, err)

	signer.Start()

	time.Sleep(time.Second * 10)

	require.NoError(t, signer.Stop())
}
