package bitcoin_test

import (
	"context"
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
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer/bitcoin"
)

var (
	BtcVault = "2N4qEFwruq3zznQs78twskBrNTc6kpq87j1"
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

func TestListenOutboundTransfer(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(success(t)))
	defer s.Close()

	host, _ := strings.CutPrefix(s.URL, "http://")
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         host,
		DisableTLS:   true,
		HTTPPostMode: true,
		User:         "test",
		Pass:         "test",
		Params:       chaincfg.TestNet3Params.Name,
	}, nil)
	require.NoError(t, err)

	initialHeight := uint64(2582657)
	b, err := bitcoin.NewBitcoin(
		log.NewNopLogger(),
		client,
		BtcVault,
		time.Second,
		initialHeight,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	b.Start(ctx)

	// We expect Observer to observe 1 block with 2 Txs
	// Only 1 Tx is sent to our vault address,
	// so we should receive only 1 TxIn
	txs := b.ListenOutboundTransfer()
	var out observer.OutboundTransfer
	require.Eventually(t, func() bool {
		out = <-txs
		return true
	}, time.Second, 100*time.Millisecond, "Timeout reading transfer")

	expOut := observer.OutboundTransfer{
		DstChain: observer.ChainId_OSMO,
		Id:       "ef4cd511c64834bde624000b94110c9f184388566a97d68d355339294a72dadf",
		Height:   initialHeight,
		Sender:   "2Mt1ttL5yffdfCGxpfxmceNE4CRUcAsBbgQ",
		To:       "osmo13g23crzfp99xg28nh0j4em4nsqnaur02nek2wt",
		Asset:    string(observer.Denom_BITCOIN),
		Amount:   math.NewUint(10000),
	}
	require.Equal(t, expOut, out)
	require.Equal(t, 0, len(txs))

	b.Stop(ctx)
}

func TestInvalidVaultAddress(t *testing.T) {
	_, err := bitcoin.NewBitcoin(
		log.NewNopLogger(),
		nil,
		"",
		time.Second,
		0,
	)
	require.ErrorIs(t, err, bitcoin.ErrInvalidCfg)
}
