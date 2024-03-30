package osmosis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cosmossdk.io/math"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/osmosis/v24/app"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer/osmosis"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

var (
	BtcAddr = "2Mt1ttL5yffdfCGxpfxmceNE4CRUcAsBbgQ"
)

type MockChain struct {
	H  uint64
	CR uint64
}

func (m *MockChain) SignalInboundTransfer(context.Context, observer.Transfer) error {
	return nil
}

func (m *MockChain) ListenOutboundTransfer() <-chan observer.Transfer {
	return make(<-chan observer.Transfer)
}

func (m *MockChain) Start(context.Context) error { return nil }

func (m *MockChain) Stop(context.Context) error { return nil }

func (m *MockChain) Height() (uint64, error) {
	return m.H, nil
}

func (m *MockChain) ConfirmationsRequired() (uint64, error) {
	return m.CR, nil
}

type OsmosisTestSuite struct {
	ts TestSuite
	hs *httptest.Server
	o  *osmosis.ChainClient
}

func NewOsmosisTestSuite(t *testing.T, ctx context.Context) OsmosisTestSuite {
	ts := NewTestSuite(t, ctx)

	s := httptest.NewServer(http.HandlerFunc(success(t)))
	cometRpc, err := rpchttp.New(s.URL, "/websocket")
	require.NoError(t, err)

	conn, err := grpc.DialContext(
		ctx,
		"test",
		grpc.WithContextDialer(ts.Dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err = keyring.NewAccount(
		osmosis.ModuleNameClient,
		Mnemonic1,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)
	client := osmosis.NewClientWithConnection(ChainId, conn, keyring)

	o := osmosis.NewChainClient(
		log.NewNopLogger(),
		&client,
		cometRpc,
	)
	require.NoError(t, err)

	return OsmosisTestSuite{ts, s, o}
}

func (ots *OsmosisTestSuite) Start(t *testing.T, ctx context.Context) {
	go ots.ts.Start(t)
	ots.o.Start(ctx)
}

func (ots *OsmosisTestSuite) Stop(t *testing.T, ctx context.Context) {
	ots.o.Stop(ctx)
	ots.hs.Close()
	ots.ts.Close(t)
}

var upgrader = websocket.Upgrader{}

func readNewBlockEvent(t *testing.T, path string) coretypes.ResultEvent {
	dataStr, err := os.ReadFile(path)
	require.NoError(t, err)
	result := coretypes.ResultEvent{}
	err = cmtjson.Unmarshal([]byte(dataStr), &result)
	require.NoError(t, err)
	return result
}

func readTxSearch(t *testing.T, path string) coretypes.ResultTxSearch {
	dataStr, err := os.ReadFile(path)
	require.NoError(t, err)
	result := coretypes.ResultTxSearch{}
	err = json.Unmarshal([]byte(dataStr), &result)
	require.NoError(t, err)
	return result
}

func success(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer c.Close()
			newBlock := readNewBlockEvent(t, "./test_events/new_block_event.json")
			newBlockResp := cmtrpctypes.NewRPCSuccessResponse(
				cmtrpctypes.JSONRPCIntID(1),
				newBlock,
			)
			newBlockRaw, err := json.Marshal(newBlockResp)
			require.NoError(t, err)
			err = c.WriteMessage(1, newBlockRaw)
			require.NoError(t, err)
		case http.MethodPost:
			blockResults := readTxSearch(t, "./test_events/tx_search.json")
			blockResultsResp := cmtrpctypes.NewRPCSuccessResponse(
				cmtrpctypes.JSONRPCIntID(0),
				blockResults,
			)
			blockResultsRaw, err := json.Marshal(blockResultsResp)
			require.NoError(t, err)
			_, err = w.Write(blockResultsRaw)
			require.NoError(t, err)
		default:
			t.Fatal("Unexpected request method", r.Method)
		}
	}
}

// TestSignalInboundTransfer verifies calling SignalInboundTransfer
// results in Tx being signed and sent to the chain
func TestSignalInboundTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ots := NewOsmosisTestSuite(t, ctx)

	expectedAcc := authtypes.NewBaseAccount(Addr1, nil, 1, 2)
	expReq1 := &authtypes.QueryAccountRequest{
		Address: Addr1.String(),
	}
	m, err := expectedAcc.Marshal()
	require.NoError(t, err)
	expResp1 := &authtypes.QueryAccountResponse{
		Account: &codectypes.Any{
			Value: m,
		},
	}
	ots.ts.AccServer.
		EXPECT().
		Account(gomock.Any(), expReq1).
		Times(1).
		Return(expResp1, nil)

	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err = keyring.NewAccount(
		osmosis.ModuleNameClient,
		Mnemonic1,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	in := observer.Transfer{
		SrcChain: observer.ChainIdBitcoin,
		DstChain: observer.ChainIdOsmosis,
		Id:       "deadbeef",
		Height:   42,
		Sender:   Addr1.String(),
		To:       BtcAddr,
		Asset:    "btc",
		Amount:   math.NewUint(10),
	}
	msg := bridgetypes.NewMsgInboundTransfer(
		in.Id,
		in.Sender,
		in.To,
		bridgetypes.AssetID{
			SourceChain: string(in.SrcChain),
			Denom:       in.Asset,
		},
		math.Int(in.Amount),
	)
	fees := sdktypes.NewCoins(sdktypes.NewCoin(osmosis.OsmoFeeDenom, osmosis.OsmoFeeAmount))
	expBytes := buildAndSignTx(
		t,
		keyring,
		expectedAcc.AccountNumber,
		expectedAcc.Sequence,
		msg,
		fees,
		osmosis.OsmoGasLimit,
	)
	expReq2 := &tx.BroadcastTxRequest{
		TxBytes: expBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
	}
	expResp2 := &tx.BroadcastTxResponse{
		TxResponse: &types.TxResponse{
			Height: 50,
			TxHash: "deadbeef",
		},
	}
	ots.ts.TxServer.
		EXPECT().
		BroadcastTx(gomock.Any(), gomock.Eq(expReq2)).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			req *tx.BroadcastTxRequest,
		) (*tx.BroadcastTxResponse, error) {
			fmt.Println("BroadcastTx")
			return expResp2, nil
		})
	ots.Start(t, ctx)

	err = ots.o.SignalInboundTransfer(ctx, in)
	require.NoError(t, err)

	ots.Stop(t, ctx)
}

// ListenOutboundTransfer verifies Osmosis properly collects transfers
// from the chain and sends it into the outbound channel
func TestListenOutboundTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ots := NewOsmosisTestSuite(t, ctx)

	height, err := ots.o.Height(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), height)
	ots.Start(t, ctx)

	// We expect to receive 3 Txs with `EventOutboundTransferType` events in this test
	// Only 2 of the Txs are successful, so we should receive only 2 event through the channel
	eventsOut := ots.o.ListenOutboundTransfer()
	transfers := [2]observer.Transfer{}
	for i := 0; i < len(transfers); i++ {
		require.Eventually(t, func() bool {
			transfers[i] = <-eventsOut
			return true
		}, time.Second, 100*time.Millisecond, "Timeout reading events from observer")
	}

	expTransfer0 := observer.Transfer{
		SrcChain: observer.ChainIdOsmosis,
		DstChain: observer.ChainIdBitcoin,
		Id:       "E765E65A3A513CCC3E2CE25BB6B47DBD7CA09AC6C7C380B84D96B88B3B0B8A70",
		Height:   5984109,
		Sender:   Addr1.String(),
		To:       BtcAddr,
		Asset:    "btc",
		Amount:   math.NewUint(10),
	}
	expTransfer1 := observer.Transfer{
		SrcChain: observer.ChainIdOsmosis,
		DstChain: observer.ChainIdBitcoin,
		Id:       "CE2D6798A8C8FD8685A29B543FDAEB31EED72A1EB5F570D889FF5E263AC7D19D",
		Height:   5984109,
		Sender:   Addr1.String(),
		To:       BtcAddr,
		Asset:    "btc",
		Amount:   math.NewUint(11),
	}
	require.Equal(t, expTransfer0, transfers[0])
	require.Equal(t, expTransfer1, transfers[1])
	require.Equal(t, 0, len(eventsOut))

	height, err = ots.o.Height(ctx)
	require.NoError(t, err)
	require.Equal(t, expTransfer0.Height, height)

	ots.Stop(t, ctx)
}

func TestChainClientConfirmationsRequired(t *testing.T) {
	tests := []struct {
		name        string
		assetId     bridgetypes.AssetID
		expectedErr error
		expectedRes uint64
	}{
		{
			"success",
			bridgetypes.AssetID{
				SourceChain: string(observer.ChainIdBitcoin),
				Denom:       string(observer.DenomBitcoin),
			},
			nil,
			5,
		},
		{
			"invalid source chain",
			bridgetypes.AssetID{
				SourceChain: "na",
				Denom:       string(observer.DenomBitcoin),
			},
			osmosis.ErrQuery,
			0,
		},
		{
			"invalid denom",
			bridgetypes.AssetID{
				SourceChain: string(observer.ChainIdBitcoin),
				Denom:       "na",
			},
			osmosis.ErrQuery,
			0,
		},
	}

	for _, tc := range tests {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			ots := NewOsmosisTestSuite(t, ctx)
			defer ots.Stop(t, ctx)

			ots.ts.ExpectTestConfirmationsRequired()
			ots.Start(t, ctx)

			cr, err := ots.o.ConfirmationsRequired(ctx, tc.assetId)
			if tc.expectedErr == nil {
				require.NoError(t, err, "test %s", tc.name)
				require.Equal(t, tc.expectedRes, cr, "test %s", tc.name)
			} else {
				require.ErrorIs(t, tc.expectedErr, err, "test %s", tc.name)
			}
		}()
	}
}
