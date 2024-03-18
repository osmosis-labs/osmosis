package bitcoin_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer/bitcoin"
)

func TestObserver(t *testing.T) {
	cfg := bitcoin.RpcConfig{
		Host:     "127.0.0.1:18334",
		Endpoint: "ws",
		User:     "test",
		Pass:     "test",
	}

	observer, err := bitcoin.NewObserver(log.NewNopLogger(), cfg, "")
	require.NoError(t, err)

	err = observer.Start()
	require.NoError(t, err)

	time.Sleep(time.Second * 10)

	err = observer.Stop()
	require.NoError(t, err)
}
