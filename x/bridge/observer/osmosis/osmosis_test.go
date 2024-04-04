package osmosis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	comettypes "github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	TestCfg = observer.ChainConfig{
		Id:                        observer.ChainIdOsmosis,
		Mode:                      observer.ModeTestnet,
		MinOutboundTransferAmount: math.NewUint(1000),
	}
	BtcAddr              = "2Mt1ttL5yffdfCGxpfxmceNE4CRUcAsBbgQ"
	OsmosisValidatorAddr = "osmo1ajaeadkj8u4wgw3sfm8szu8hl992nngaex40fs"
)

var _ observer.Client = new(MockChain)

type MockChain struct {
	vHeight                uint64
	vConfirmationsRequired uint64
}

func (m *MockChain) SignalInboundTransfer(context.Context, observer.Transfer) error {
	return nil
}

func (m *MockChain) ListenOutboundTransfer() <-chan observer.Transfer {
	return make(<-chan observer.Transfer)
}

func (m *MockChain) Start(context.Context) error { return nil }

func (m *MockChain) Stop(context.Context) error { return nil }

func (m *MockChain) Height(context.Context) (uint64, error) {
	return m.vHeight, nil
}

func (m *MockChain) ConfirmationsRequired(context.Context, bridgetypes.AssetID) (uint64, error) {
	return m.vConfirmationsRequired, nil
}

type OsmosisTestSuite struct {
	ts TestSuite
	hs *httptest.Server
	o  *osmosis.ChainClient
}

func NewOsmosisTestSuite(t *testing.T, ctx context.Context) OsmosisTestSuite {
	ts := NewTestSuite(t)

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
	kr := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err = kr.NewAccount(
		osmosis.ModuleNameClient,
		Mnemonic1,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	client := osmosis.NewClient(ChainId, conn, kr, app.GetEncodingConfig().TxConfig)
	o := osmosis.NewChainClient(
		log.NewNopLogger(),
		TestCfg,
		client,
		cometRpc,
		app.GetEncodingConfig().TxConfig,
		OsmosisValidatorAddr,
	)
	require.NoError(t, err)

	return OsmosisTestSuite{ts, s, o}
}

func (ots *OsmosisTestSuite) Start(t *testing.T, ctx context.Context) {
	go ots.ts.Start(t)
	err := ots.o.Start(ctx)
	require.NoError(t, err)
}

func (ots *OsmosisTestSuite) Stop(t *testing.T, ctx context.Context) {
	err := ots.o.Stop(ctx)
	require.NoError(t, err)
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

func readTxCheckBytes(t *testing.T, id int, path string) []byte {
	dataStr, err := os.ReadFile(path)
	require.NoError(t, err)
	result := abci.ResponseCheckTx{}
	err = json.Unmarshal([]byte(dataStr), &result)
	require.NoError(t, err)
	checkResultsResp := cmtrpctypes.NewRPCSuccessResponse(
		cmtrpctypes.JSONRPCIntID(id),
		result,
	)
	checkResultsRaw, err := json.Marshal(checkResultsResp)
	require.NoError(t, err)
	return checkResultsRaw
}

func success(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer c.Close()
			newBlock := readNewBlockEvent(t, "./test_events/new_block.json")
			newBlockResp := cmtrpctypes.NewRPCSuccessResponse(
				cmtrpctypes.JSONRPCIntID(1),
				newBlock,
			)
			newBlockRaw, err := json.Marshal(newBlockResp)
			require.NoError(t, err)
			err = c.WriteMessage(1, newBlockRaw)
			require.NoError(t, err)
		case http.MethodPost:
			bytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, r.Body.Close())
			var req cmtrpctypes.RPCRequest
			err = json.Unmarshal(bytes, &req)
			require.NoError(t, err)
			jsonId, ok := req.ID.(cmtrpctypes.JSONRPCIntID)
			require.True(t, ok)
			id := int(jsonId)

			var resp []byte
			switch id {
			case 1:
				resp = readTxCheckBytes(t, id, "./test_events/tx_check_error.json")
			default:
				resp = readTxCheckBytes(t, id, "./test_events/tx_check_success.json")
			}

			_, err = w.Write(resp)
			require.NoError(t, err)
		default:
			t.Fatal("Unexpected request method", r.Method)
		}
	}
}

// TestSignalInboundTransfer verifies calling SignalInboundTransfer
// results in Tx being signed and sent to the chain
func TestSignalInboundTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
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

	kr := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err = kr.NewAccount(
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
		Asset:    bridgetypes.DefaultBitcoinDenomName,
		Amount:   math.NewUint(10),
	}
	msg := bridgetypes.NewMsgInboundTransfer(
		in.Id,
		OsmosisValidatorAddr, // NB! validator sends a message!
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
		kr,
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
	t.Skip("x/bridge needs to be wired to decode Txs")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ots := NewOsmosisTestSuite(t, ctx)

	height, err := ots.o.Height(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), height)
	ots.Start(t, ctx)
	defer ots.Stop(t, ctx)

	// We expect to observe 1 block with 3 Txs each with a `MsgOutboundTransfer` message:
	// - valid tx to BTC address
	// - failed tx
	// - tx with invalid destination address
	// - tx with amount of tokens below threshold
	// So, we should to receive only 1 Transfer
	transfers := ots.o.ListenOutboundTransfer()
	var transfer observer.Transfer
	require.Eventually(t, func() bool {
		transfer = <-transfers
		return true
	}, time.Second, 100*time.Millisecond, "Timeout waiting for transfer")

	expTransfer := observer.Transfer{
		SrcChain: observer.ChainIdOsmosis,
		DstChain: observer.ChainIdBitcoin,
		Id:       "8eb4b69be7144690f82a4e1485f4b85d23adc5267db5d3dab7affae57c8ce2a4",
		Height:   2801,
		Sender:   "osmo1pldlhnwegsj3lqkarz0e4flcsay3fuqgkd35ww",
		To:       "2Mt1ttL5yffdfCGxpfxmceNE4CRUcAsBbgQ",
		Asset:    bridgetypes.DefaultBitcoinDenomName,
		Amount:   math.NewUint(10),
	}
	require.Equal(t, expTransfer, transfer)
	require.Equal(t, 0, len(transfers))

	height, err = ots.o.Height(ctx)
	require.NoError(t, err)
	require.Equal(t, expTransfer.Height, height)
}

///////////////////////////

func TestParseBlock(t *testing.T) {
	t.SkipNow()

	event := readNewBlockEvent(t, "./test_events/blocks/block_370.json")

	newBlock, ok := event.Data.(comettypes.EventDataNewBlock)
	require.True(t, ok)

	js, err := json.MarshalIndent(newBlock, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(js))

	dec := app.GetEncodingConfig().TxConfig.TxDecoder()
	for _, tx := range newBlock.Block.Txs {
		decoded, err := dec(tx)
		require.NoError(t, err)
		for _, msg := range decoded.GetMsgs() {
			outbound, ok := msg.(*bridgetypes.MsgOutboundTransfer)
			require.True(t, ok)
			fmt.Println(outbound)
		}
	}
}

func TestBlocks(t *testing.T) {
	t.SkipNow()

	url := "http://127.0.0.1:26657"
	rpc, err := rpchttp.New(url, "/websocket")
	require.NoError(t, err)
	fmt.Println("Rpc")

	o := osmosis.NewChainClient(log.NewNopLogger(), TestCfg, nil, rpc, app.GetEncodingConfig().TxConfig, ValAddr)
	fmt.Println("Client")
	err = o.Start(context.Background())
	require.NoError(t, err)

	time.Sleep(time.Hour)

	o.Stop(context.Background())
}

var (
	ValMnemonic = "accident pipe try devote coin pet label brush fun myself carbon screen pen type impose grow marine famous live endless worth crew pact cute"
	ValAddr     = "osmo1pldlhnwegsj3lqkarz0e4flcsay3fuqgkd35ww"
)

func TestParams(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()

	grpcUrl := "127.0.0.1:9090"
	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err := keyring.NewAccount(
		osmosis.ModuleNameClient,
		ValMnemonic,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	o := osmosis.NewClient("my-test-chain", conn, keyring, app.GetEncodingConfig().TxConfig)
	require.NoError(t, err)

	msg := &bridgetypes.MsgUpdateParams{
		Sender: ValAddr,
		NewParams: bridgetypes.Params{
			Signers: []string{ValAddr},
			Assets: []bridgetypes.Asset{
				{
					Exponent:              10,
					ExternalConfirmations: 6,
					Id: bridgetypes.AssetID{
						SourceChain: "bitcoin",
						Denom:       "btc",
					},
					Status: bridgetypes.AssetStatus_ASSET_STATUS_OK,
				},
			},
			VotesNeeded: 1,
			Fee:         math.LegacyNewDec(0),
		},
	}
	fees := types.NewCoins(types.NewCoin("stake", math.NewInt(3000)))
	gasLimit := 1200000
	bytes, err := o.SignTx(ctx, msg, fees, uint64(gasLimit))
	require.NoError(t, err)

	resp, err := o.BroadcastTx(ctx, bytes)
	require.NoError(t, err)
	fmt.Println(resp)
	o.Close()
}

func TestInbound(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()

	grpcUrl := "127.0.0.1:9090"
	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err := keyring.NewAccount(
		osmosis.ModuleNameClient,
		ValMnemonic,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	o := osmosis.NewClient("my-test-chain", conn, keyring, app.GetEncodingConfig().TxConfig)
	require.NoError(t, err)

	msg := &bridgetypes.MsgInboundTransfer{
		ExternalId:     "deadbeef",
		ExternalHeight: 42,
		Sender:         ValAddr,
		DestAddr:       ValAddr,
		AssetId: bridgetypes.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		},
		Amount: math.Int(math.NewUint(10000)),
	}
	fees := types.NewCoins(types.NewCoin("stake", math.NewInt(1000)))
	gasLimit := 200000
	bytes, err := o.SignTx(ctx, msg, fees, uint64(gasLimit))
	require.NoError(t, err)

	resp, err := o.BroadcastTx(ctx, bytes)
	require.NoError(t, err)
	fmt.Println(resp)
	o.Close()
}

func TestOutbound(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()

	grpcUrl := "127.0.0.1:9090"
	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err := keyring.NewAccount(
		osmosis.ModuleNameClient,
		ValMnemonic,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	o := osmosis.NewClient("my-test-chain", conn, keyring, app.GetEncodingConfig().TxConfig)
	require.NoError(t, err)

	msg := &bridgetypes.MsgOutboundTransfer{
		Sender:   ValAddr,
		DestAddr: BtcAddr,
		AssetId: bridgetypes.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		},
		Amount: math.Int(math.NewUint(10)),
	}
	msg1 := msg
	msg2 := &bridgetypes.MsgOutboundTransfer{
		Sender:   ValAddr,
		DestAddr: Addr1.String(),
		AssetId: bridgetypes.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		},
		Amount: math.NewInt(10),
	}
	msg3 := &bridgetypes.MsgOutboundTransfer{
		Sender:   ValAddr,
		DestAddr: BtcAddr,
		AssetId: bridgetypes.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		},
		Amount: math.NewInt(1),
	}
	fees := types.NewCoins(types.NewCoin("stake", math.NewInt(1000)))
	gasLimit := 200000

	msgs := []sdk.Msg{msg3, msg2, msg1, msg}
	for _, m := range msgs {
		bytes, err := o.SignTx(ctx, m, fees, uint64(gasLimit))
		require.NoError(t, err)
		resp, err := o.BroadcastTx(ctx, bytes)
		require.NoError(t, err)
		fmt.Println(resp)
	}

	o.Close()
}
