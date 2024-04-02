package osmosis

import (
	"context"
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"

	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

var ModuleNameClient = "osmo-client"

type Client struct {
	chainId      string
	keyring      keyring.Keyring
	grpcConn     *grpc.ClientConn
	txConfig     client.TxConfig
	txClient     tx.ServiceClient
	accClient    authtypes.QueryClient
	bridgeClient bridgetypes.QueryClient
}

// NewClient returns new instance of `Client` with
// Tx service client and Auth query client created
func NewClient(
	chainId string,
	grpcConn *grpc.ClientConn,
	keyring keyring.Keyring,
	txConfig client.TxConfig,
) *Client {
	return &Client{
		chainId:   chainId,
		keyring:   keyring,
		grpcConn:  grpcConn,
		txClient:  tx.NewServiceClient(grpcConn),
		accClient: authtypes.NewQueryClient(grpcConn),
		txConfig:  txConfig,
	}
}

// Close closes client's GRPC connections
func (c *Client) Close() {
	_ = c.grpcConn.Close()
}

// SignTx signs provided message with internal keyring
func (c *Client) SignTx(
	ctx context.Context,
	msg sdk.Msg,
	fees sdk.Coins,
	gasLimit uint64,
) ([]byte, error) {
	key, err := c.keyring.Key(ModuleNameClient)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}
	cpk, err := key.GetPubKey()
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}
	addr, err := key.GetAddress()
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}

	acc, err := c.Account(ctx, addr)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}

	txBuilder, err := c.buildUnsigned(
		cpk,
		acc.Sequence,
		msg,
		fees,
		gasLimit,
	)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}
	txBytes, err := c.sign(txBuilder, cpk, addr, acc.AccountNumber, acc.Sequence)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrSignTx, err.Error())
	}

	return txBytes, nil
}

// BroadcastTx broadcasts given message to chain
func (c *Client) BroadcastTx(ctx context.Context, txBytes []byte) (sdk.TxResponse, error) {
	req := tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
	}
	res, err := c.txClient.BroadcastTx(ctx, &req)
	if err != nil {
		return sdk.TxResponse{}, errorsmod.Wrapf(ErrBroadcastTx, err.Error())
	}
	return *res.TxResponse, nil
}

// Account queries account information by given account address
func (c *Client) Account(ctx context.Context, addr sdk.AccAddress) (authtypes.BaseAccount, error) {
	req := authtypes.QueryAccountRequest{
		Address: addr.String(),
	}
	acc, err := c.accClient.Account(ctx, &req)
	if err != nil {
		return authtypes.BaseAccount{}, errorsmod.Wrapf(ErrQuery, "Account %s", err.Error())
	}

	ba := authtypes.BaseAccount{}
	err = ba.Unmarshal(acc.GetAccount().Value)
	if err != nil {
		return authtypes.BaseAccount{}, errorsmod.Wrapf(ErrQuery, "Account %s", err.Error())
	}

	return ba, nil
}

// ConfirmationsRequired returns the amount of confirmations required for the specified asset
func (c *Client) ConfirmationsRequired(
	ctx context.Context,
	assetId bridgetypes.AssetID,
) (uint64, error) {
	params, err := c.bridgeClient.Params(ctx, new(bridgetypes.QueryParamsRequest))
	if err != nil {
		return 0, errorsmod.Wrapf(ErrQuery, "bridge/params: %s", err.Error())
	}
	idx := slices.IndexFunc(params.GetParams().Assets, func(a bridgetypes.Asset) bool {
		return a.Id == assetId
	})
	const idxNotFound = -1
	if idx == idxNotFound {
		return 0, errorsmod.Wrapf(
			ErrQuery,
			"bridge/params: asset with id %s not found",
			assetId.String(),
		)
	}
	return params.GetParams().Assets[idx].ExternalConfirmations, nil
}

// buildUnsigned creates unassigned transaction with provided message, fees and gas limit.
// Initializes transaction signatures
func (c *Client) buildUnsigned(
	cpk types.PubKey,
	accSeq uint64,
	msg sdk.Msg,
	fees sdk.Coins,
	gasLimit uint64,
) (client.TxBuilder, error) {
	txBuilder := c.txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to set tx messages")
	}

	sigData := signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
	}

	sig := signingtypes.SignatureV2{
		PubKey:   cpk,
		Data:     &sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to init tx signatures")
	}
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fees)

	return txBuilder, nil
}

// sign signs transaction using client's keyring
func (c *Client) sign(
	txBuilder client.TxBuilder,
	cpk types.PubKey,
	addr sdk.AccAddress,
	accNum uint64,
	accSeq uint64,
) ([]byte, error) {
	modeHandler := c.txConfig.SignModeHandler()
	signingData := signing.SignerData{
		ChainID:       c.chainId,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	signBytes, err := modeHandler.GetSignBytes(
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to get sign bytes")
	}
	sigData := signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
	}
	sigData.Signature, _, err = c.keyring.SignByAddress(addr, signBytes)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to sign with keyring")
	}

	if !cpk.VerifySignature(signBytes, sigData.Signature) {
		return nil, fmt.Errorf("Failed to verify signature")
	}

	sig := signingtypes.SignatureV2{
		PubKey:   cpk,
		Data:     &sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to set tx signatures")
	}

	txBytes, err := c.txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to encode tx")
	}
	return txBytes, nil
}
