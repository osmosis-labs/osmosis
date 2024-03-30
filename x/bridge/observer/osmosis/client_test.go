package osmosis_test

import (
	"context"
	"net"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/osmosis-labs/osmosis/v24/app"
	"github.com/osmosis-labs/osmosis/v24/tests/mocks"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer/osmosis"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

var (
	ChainId   = "test"
	Addr1     = types.MustAccAddressFromBech32("osmo1mfkl4p92lqvlf7fh5rpds3afvpase78fz5h36l")
	Mnemonic1 = "tonight april truth pelican manual door beyond inspire boil biology improve horse bean cotton festival display calm pitch ahead seed ice fee baby quality"
	Addr2     = types.MustAccAddressFromBech32("osmo1g9z8hxmr9qqfl0lxeahpf2m5hygveasx933lhp")
)

type TestSuite struct {
	Ctrl         *gomock.Controller
	Lis          *bufconn.Listener
	GrpcServer   *grpc.Server
	AccServer    *mocks.MockQueryServer
	TxServer     *mocks.MockServiceServer
	BridgeServer *mocks.MockBridgeQueryServer
}

func NewTestSuite(t *testing.T, ctx context.Context) TestSuite {
	ctrl := gomock.NewController(t)
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	accServer := mocks.NewMockQueryServer(ctrl)
	txServer := mocks.NewMockServiceServer(ctrl)
	bridgeServer := mocks.NewMockBridgeQueryServer(ctrl)
	authtypes.RegisterQueryServer(s, accServer)
	tx.RegisterServiceServer(s, txServer)
	bridgetypes.RegisterQueryServer(s, bridgeServer)
	return TestSuite{ctrl, lis, s, accServer, txServer, bridgeServer}
}

func (ts *TestSuite) Start(t *testing.T) {
	err := ts.GrpcServer.Serve(ts.Lis)
	require.NoError(t, err)
}

func (ts *TestSuite) Close(t *testing.T) {
	ts.GrpcServer.Stop()
	err := ts.Lis.Close()
	require.NoError(t, err)
	ts.Ctrl.Finish()
}

func (ts *TestSuite) Dialer() func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return ts.Lis.Dial()
	}
}

func (ts *TestSuite) ExpectTestConfirmationsRequired() {
	expReq := &bridgetypes.QueryParamsRequest{}
	expResp := &bridgetypes.QueryParamsResponse{
		Params: bridgetypes.Params{
			Assets: []bridgetypes.Asset{
				{
					Id: bridgetypes.AssetID{
						SourceChain: string(observer.ChainIdBitcoin),
						Denom:       string(observer.DenomBitcoin),
					},
					ExternalConfirmations: 5,
				},
			},
		},
	}
	ts.BridgeServer.
		EXPECT().
		Params(gomock.Any(), expReq).
		Times(1).
		Return(expResp, nil)
}

// TestAccountQuerySuccess verifies client properly sends account request
// and receives account response
func TestAccountQuerySuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ts := NewTestSuite(t, ctx)
	defer ts.Close(t)

	expectedReq := &authtypes.QueryAccountRequest{Address: Addr1.String()}
	expectedAcc := authtypes.NewBaseAccount(Addr1, nil, 1, 2)
	ts.AccServer.
		EXPECT().
		Account(gomock.Any(), expectedReq).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			req *authtypes.QueryAccountRequest,
		) (*authtypes.QueryAccountResponse, error) {
			res, err := expectedAcc.Marshal()
			require.NoError(t, err)
			ret := authtypes.QueryAccountResponse{
				Account: &codectypes.Any{
					Value: res,
				},
			}
			return &ret, nil
		})
	go ts.Start(t)

	conn, err := grpc.DialContext(
		ctx,
		"test",
		grpc.WithContextDialer(ts.Dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	client := osmosis.NewClientWithConnection(ChainId, conn, keyring)
	defer client.Close()

	acc, err := client.Account(ctx, Addr1)
	require.NoError(t, err)
	require.Equal(t, expectedAcc, &acc)
}

// TestSignTxSuccess verifies client properly signs transactions
func TestSignTxSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ts := NewTestSuite(t, ctx)
	defer ts.Close(t)

	expectedReq := &authtypes.QueryAccountRequest{Address: Addr1.String()}
	expectedAcc := authtypes.NewBaseAccount(Addr1, nil, 1, 2)
	ts.AccServer.
		EXPECT().
		Account(gomock.Any(), expectedReq).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			req *authtypes.QueryAccountRequest,
		) (*authtypes.QueryAccountResponse, error) {
			res, err := expectedAcc.Marshal()
			require.NoError(t, err)
			ret := authtypes.QueryAccountResponse{
				Account: &codectypes.Any{
					Value: res,
				},
			}
			return &ret, nil
		})
	go ts.Start(t)

	conn, err := grpc.DialContext(
		ctx,
		"test",
		grpc.WithContextDialer(ts.Dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	record, err := keyring.NewAccount(
		osmosis.ModuleNameClient,
		Mnemonic1,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)
	cpk, err := record.GetPubKey()
	require.NoError(t, err)
	client := osmosis.NewClientWithConnection(ChainId, conn, keyring)
	defer client.Close()

	coins := types.NewCoins(types.NewInt64Coin("uosmo", 100))
	msg := banktypes.NewMsgSend(Addr1, Addr2, coins)
	fees := types.NewCoins(types.NewInt64Coin("uosmo", 500))
	gasLimit := uint64(200000)
	bytes, err := client.SignTx(ctx, msg, fees, gasLimit)
	require.NoError(t, err)

	tx, err := app.GetEncodingConfig().TxConfig.TxDecoder()(bytes)
	require.NoError(t, err)

	fee, ok := tx.(types.FeeTx)
	require.True(t, ok, "Failed to cast Tx to FeeTx")
	require.Equal(t, gasLimit, fee.GetGas())
	require.Equal(t, fees, fee.GetFee())

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	require.True(t, ok, "Failed to cast Tx to SigVerifiableTx")
	pks, err := sigTx.GetPubKeys()
	require.NoError(t, err)
	require.Equal(t, 1, len(pks))
	require.Equal(t, cpk, pks[0])

	msgs := sigTx.GetMsgs()
	require.Equal(t, 1, len(msgs))
	require.Equal(t, msg, msgs[0])
	require.Equal(t, 1, len(msgs[0].GetSigners()))
	require.Equal(t, Addr1, msgs[0].GetSigners()[0])

	sigs, err := sigTx.GetSignaturesV2()
	require.NoError(t, err)
	require.Equal(t, 1, len(sigs))
	sig := sigs[0]
	require.Equal(t, expectedAcc.Sequence, sig.Sequence)

	signerData := authsigning.SignerData{
		ChainID:       ChainId,
		AccountNumber: expectedAcc.AccountNumber,
		Sequence:      expectedAcc.Sequence,
	}
	err = authsigning.VerifySignature(
		cpk,
		signerData,
		sig.Data,
		app.GetEncodingConfig().TxConfig.SignModeHandler(),
		tx,
	)
	require.NoError(t, err)
}

// Verifies client properly broadcasts transaction to chain
// and receives broadcast response
func TestBroadcastTxSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ts := NewTestSuite(t, ctx)
	defer ts.Close(t)

	expectedReq := &authtypes.QueryAccountRequest{Address: Addr1.String()}
	expectedAcc := authtypes.NewBaseAccount(Addr1, nil, 1, 2)
	ts.AccServer.
		EXPECT().
		Account(gomock.Any(), expectedReq).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			req *authtypes.QueryAccountRequest,
		) (*authtypes.QueryAccountResponse, error) {
			res, err := expectedAcc.Marshal()
			require.NoError(t, err)
			ret := authtypes.QueryAccountResponse{
				Account: &codectypes.Any{
					Value: res,
				},
			}
			return &ret, nil
		})

	keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
	_, err := keyring.NewAccount(
		osmosis.ModuleNameClient,
		Mnemonic1,
		"",
		types.FullFundraiserPath,
		hd.Secp256k1,
	)
	require.NoError(t, err)
	coins := types.NewCoins(types.NewInt64Coin("uosmo", 100))
	msg := banktypes.NewMsgSend(Addr1, Addr2, coins)
	fees := types.NewCoins(types.NewInt64Coin("uosmo", 500))
	gasLimit := uint64(200000)
	expectedTxBytes := buildAndSignTx(
		t,
		keyring,
		expectedAcc.AccountNumber,
		expectedAcc.Sequence,
		msg,
		fees,
		gasLimit,
	)
	expReq := tx.BroadcastTxRequest{
		TxBytes: expectedTxBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
	}
	expResp := types.TxResponse{
		Height: 1,
		TxHash: "deadbeef",
	}
	ts.TxServer.
		EXPECT().
		BroadcastTx(gomock.Any(), gomock.Eq(&expReq)).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			req *tx.BroadcastTxRequest,
		) (*tx.BroadcastTxResponse, error) {
			return &tx.BroadcastTxResponse{
				TxResponse: &expResp,
			}, nil
		})
	go ts.Start(t)

	conn, err := grpc.DialContext(
		ctx,
		"test",
		grpc.WithContextDialer(ts.Dialer()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := osmosis.NewClientWithConnection(ChainId, conn, keyring)
	defer client.Close()

	bytes, err := client.SignTx(ctx, msg, fees, gasLimit)
	require.NoError(t, err)

	resp, err := client.BroadcastTx(ctx, bytes)
	require.NoError(t, err)
	require.Equal(t, expResp, resp)
}

func TestConfirmationsRequiredSuccess(t *testing.T) {
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
			ts := NewTestSuite(t, ctx)
			defer ts.Close(t)

			ts.ExpectTestConfirmationsRequired()
			go ts.Start(t)

			conn, err := grpc.DialContext(
				ctx,
				"test",
				grpc.WithContextDialer(ts.Dialer()),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err)
			keyring := keyring.NewInMemory(app.GetEncodingConfig().Marshaler)
			client := osmosis.NewClientWithConnection(ChainId, conn, keyring)
			defer client.Close()

			cr, err := client.ConfirmationsRequired(ctx, tc.assetId)
			if tc.expectedErr == nil {
				require.NoError(t, err, "test %s", tc.name)
				require.Equal(t, tc.expectedRes, cr, "test %s", tc.name)
			} else {
				require.ErrorIs(t, tc.expectedErr, err, "test %s", tc.name)
			}
		}()
	}
}

func buildAndSignTx(
	t *testing.T,
	kr keyring.Keyring,
	accNum, accSeq uint64,
	msg types.Msg,
	fees types.Coins,
	gasLimit uint64,
) []byte {
	rec, err := kr.Key(osmosis.ModuleNameClient)
	require.NoError(t, err)
	cpk, err := rec.GetPubKey()
	require.NoError(t, err)
	addr, err := rec.GetAddress()
	require.NoError(t, err)

	tc := app.GetEncodingConfig().TxConfig
	builder := tc.NewTxBuilder()
	err = builder.SetMsgs(msg)
	require.NoError(t, err)
	sigData := signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
	}
	sig := signingtypes.SignatureV2{
		PubKey:   cpk,
		Data:     &sigData,
		Sequence: accSeq,
	}
	err = builder.SetSignatures(sig)
	require.NoError(t, err)
	builder.SetGasLimit(gasLimit)
	builder.SetFeeAmount(fees)
	mh := tc.SignModeHandler()
	signingData := signing.SignerData{
		ChainID:       ChainId,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	signBytes, err := mh.GetSignBytes(
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingData,
		builder.GetTx(),
	)
	sigData = signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
	}
	sigData.Signature, _, err = kr.SignByAddress(addr, signBytes)
	sig = signingtypes.SignatureV2{
		PubKey:   cpk,
		Data:     &sigData,
		Sequence: accSeq,
	}
	err = builder.SetSignatures(sig)
	require.NoError(t, err)
	bytes, err := tc.TxEncoder()(builder.GetTx())
	require.NoError(t, err)
	return bytes
}
