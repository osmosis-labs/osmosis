package bitcoin_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer/bitcoin"
)

type Response struct {
	Result json.RawMessage   `json:"result"`
	Error  *btcjson.RPCError `json:"error"`
}

func readResponseFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	resp := Response{
		Result: data,
	}
	respRaw, err := json.Marshal(resp)
	require.NoError(t, err)
	return respRaw
}

func success(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			bytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, r.Body.Close())
			var cmd btcjson.Request
			err = json.Unmarshal(bytes, &cmd)
			require.NoError(t, err)

			var resp []byte

			switch cmd.Method {
			case "getblockhash":
				resp = readResponseFile(t, "./test_responses/block_hash.json")
			case "getblock":
				resp = readResponseFile(t, "./test_responses/block_verbose_tx.json")
			case "getrawtransaction":
				resp = readResponseFile(t, "./test_responses/vin_tx_verbose.json")
			}
			_, err = w.Write(resp)
			require.NoError(t, err)
		default:
			t.Fatal("Unexpected request method", r.Method)
		}
	}
}

// TestObserverSuccess verifies Observer properly processes observed transactions
func TestObserverSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(success(t)))
	defer s.Close()

	host, _ := strings.CutPrefix(s.URL, "http://")
	cfg := bitcoin.RpcConfig{
		Host:       host,
		DisableTls: true,
		User:       "test",
		Pass:       "test",
	}

	initialHeight := uint64(2582657)
	observer, err := bitcoin.NewObserver(
		log.NewNopLogger(),
		cfg,
		"2N4qEFwruq3zznQs78twskBrNTc6kpq87j1",
		initialHeight,
		time.Second,
	)
	require.NoError(t, err)

	observer.Start()

	// We expect Observer to observe 1 block with 2 Txs
	// Only 1 Tx is sent to our vault address,
	// so we should receive only 1 TxIn
	txs := observer.TxIns()
	var tx bitcoin.TxIn
	require.Eventually(t, func() bool {
		tx = <-txs
		return true
	}, time.Second, 100*time.Millisecond, "Timeout reading events from observer")

	expectedTx := bitcoin.TxIn{
		Id:          "f395b2cc8551aff25fe8d61fec159a6b93d29b9ff56a68c9d29df99a864fd74c",
		Height:      initialHeight,
		Sender:      "2Mt1ttL5yffdfCGxpfxmceNE4CRUcAsBbgQ",
		Destination: "2N4qEFwruq3zznQs78twskBrNTc6kpq87j1",
		Amount:      math.NewUint(10000),
		Memo:        "osmo13g23crzfp99xg28nh0j4em4nsqnaur02nek2wt",
	}
	require.Equal(t, expectedTx, tx)
	require.Equal(t, 0, len(txs))

	observer.Stop()
}

func TestInvalidRpcCfg(t *testing.T) {
	tests := []struct {
		name string
		cfg  bitcoin.RpcConfig
	}{
		{
			name: "Invalid Host URL",
			cfg: bitcoin.RpcConfig{
				Host:       "",
				DisableTls: true,
				User:       "test",
				Pass:       "test",
			},
		},
		{
			name: "Invalid User",
			cfg: bitcoin.RpcConfig{
				Host:       "127.0.0.1:1234",
				DisableTls: true,
				User:       "",
				Pass:       "test",
			},
		},
		{
			name: "Invalid Pass",
			cfg: bitcoin.RpcConfig{
				Host:       "127.0.0.1:1234",
				DisableTls: true,
				User:       "test",
				Pass:       "",
			},
		},
	}

	for _, tc := range tests {
		_, err := bitcoin.NewObserver(log.NewNopLogger(), tc.cfg, "", 0, time.Second)
		require.ErrorIs(t, err, bitcoin.ErrInvalidCfg)
	}
}

func TestInvalidVaultAddress(t *testing.T) {
	cfg := bitcoin.RpcConfig{
		Host:       "127.0.0.1:1234",
		DisableTls: true,
		User:       "test",
		Pass:       "test",
	}

	initialHeight := uint64(2582657)
	_, err := bitcoin.NewObserver(
		log.NewNopLogger(),
		cfg,
		"",
		initialHeight,
		time.Second,
	)
	require.ErrorIs(t, err, bitcoin.ErrInvalidCfg)
}
