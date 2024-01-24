package authenticator

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

type CosmwasmAuthenticator struct {
	contractKeeper *keeper.PermissionedKeeper
	ak             *authkeeper.AccountKeeper
	cdc            codectypes.AnyUnpacker
	sigModeHandler authsigning.SignModeHandler

	contractAddr        sdk.AccAddress
	authenticatorParams []byte
}

var (
	_ iface.Authenticator = &CosmwasmAuthenticator{}
)

func NewCosmwasmAuthenticator(contractKeeper *keeper.PermissionedKeeper, accountKeeper *authkeeper.AccountKeeper, sigModeHandler authsigning.SignModeHandler, cdc codectypes.AnyUnpacker) CosmwasmAuthenticator {
	return CosmwasmAuthenticator{
		contractKeeper: contractKeeper,
		sigModeHandler: sigModeHandler,
		ak:             accountKeeper,
		cdc:            cdc,
	}
}

func (cwa CosmwasmAuthenticator) Type() string {
	return "CosmwasmAuthenticatorV1"
}

func (cwa CosmwasmAuthenticator) StaticGas() uint64 {
	return 0
}

type CosmwasmAuthenticatorInitData struct {
	Contract string `json:"contract"`
	Params   []byte `json:"params"`
}

func (cwa CosmwasmAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	var initData CosmwasmAuthenticatorInitData
	err := json.Unmarshal(data, &initData)
	if err != nil {
		return nil, err
	}
	if len(initData.Contract) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing contract address")
	}
	contractAddr, err := sdk.AccAddressFromBech32(initData.Contract)
	if err != nil {
		return nil, err
	}
	cwa.contractAddr = contractAddr
	cwa.authenticatorParams = initData.Params
	return cwa, nil
}

type BaseData struct {
	Tx           sdk.Tx
	MessageIndex int
	Simulate     bool
}

var _ iface.AuthenticatorData = BaseData{}

func (cwa CosmwasmAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int,
	simulate bool,
) (iface.AuthenticatorData, error) {
	return BaseData{
		Tx:           tx,
		MessageIndex: messageIndex,
		Simulate:     simulate,
	}, nil
}

type SudoMsg struct {
	Authenticate     *AuthenticationRequest   `json:"authenticate,omitempty"`
	Track            *TrackRequest            `json:"track,omitempty"`
	ConfirmExecution *ConfirmExecutionRequest `json:"confirm_execution,omitempty"`
}

// TODO: decide when we want to reject and when to just not authenticate
func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	bd := authenticationData.(BaseData)
	authRequest, err := GenerateAuthenticationData(ctx, cwa.ak, cwa.sigModeHandler, account, msg, bd.Tx, bd.MessageIndex, bd.Simulate)
	if err != nil {
		return iface.Rejected("failed to generate authentication data", err)
	}
	// Add the authenticator params set for this authenticator in Initialize()
	authRequest.AuthenticatorParams = cwa.authenticatorParams

	bz, err := json.Marshal(SudoMsg{Authenticate: &authRequest})
	if err != nil {
		return iface.Rejected("failed to marshall AuthenticationRequest", err)
	}

	result, err := cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return iface.Rejected("failed to sudo", err)
	}

	authResult, err := UnmarshalAuthenticationResult(result)
	if err != nil {
		return iface.Rejected("failed to unmarshal authentication result", err)
	}
	return authResult
}

func GenerateAuthenticationData(ctx sdk.Context, ak *authkeeper.AccountKeeper, sigModeHandler authsigning.SignModeHandler, account sdk.AccAddress, msg sdk.Msg, tx sdk.Tx, messageIndex int, simulate bool) (AuthenticationRequest, error) {
	signers, txSignatures, _, err := GetCommonAuthenticationData(ctx, tx, -1, simulate)
	if err != nil {
		return AuthenticationRequest{}, sdkerrors.Wrap(err, "failed to get signes and signatures")
	}

	if len(msg.GetSigners()) != 1 {
		return AuthenticationRequest{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "only messages with a single signer are supported")
	}

	// Retrieve and build the signer data struct
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	baseAccount := ak.GetAccount(ctx, account)
	if !genesis {
		accNum = baseAccount.GetAccountNumber()
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      baseAccount.GetSequence(),
	}

	signBytes, err := sigModeHandler.GetSignBytes(txsigning.SignMode_SIGN_MODE_DIRECT, signerData, tx)
	if err != nil {
		return AuthenticationRequest{}, sdkerrors.Wrap(err, "failed to get signBytes")
	}

	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return AuthenticationRequest{}, sdkerrors.Wrap(err, "failed to encode msg")
	}

	timeoutTx, ok := tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return AuthenticationRequest{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithTimeoutHeight")
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return AuthenticationRequest{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithMemo")
	}

	msgs := make([]LocalAny, len(tx.GetMsgs()))
	for i, msg := range tx.GetMsgs() {
		encodedMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return AuthenticationRequest{}, sdkerrors.Wrap(err, "failed to encode msg")
		}
		msgs[i] = LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		}
	}

	txData := ExplicitTxData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      baseAccount.GetSequence(),
		TimeoutHeight: timeoutTx.GetTimeoutHeight(),
		Msgs:          msgs,
		Memo:          memoTx.GetMemo(),
	}

	signer := msg.GetSigners()[0]
	var signatures [][]byte
	var msgSignature []byte
	for i, signature := range txSignatures {
		// ToDo: deal with other signature types
		single, ok := signature.Data.(*txsigning.SingleSignatureData)
		if !ok {
			return AuthenticationRequest{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "failed to cast signature to SingleSignatureData")
		}
		signatures = append(signatures, single.Signature)
		if signers[i].Equals(signer) {
			msgSignature = single.Signature
		}
	}

	// should we pass ctx.IsReCheckTx() here? How about msgIndex?
	authRequest := AuthenticationRequest{
		Account: account,
		Msg: LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		},
		Signature: msgSignature, // currently only allowing one signer per message.
		TxData:    txData,
		SignModeTxData: SignModeData{ // TODO: Add other sign modes. Specifically textual when it becomes available
			Direct: signBytes,
		},
		SignatureData: SimplifiedSignatureData{
			Signers:    signers,
			Signatures: signatures,
		},
		Simulate:            simulate,
		AuthenticatorParams: nil,
	}
	return authRequest, nil
}

func (cwa CosmwasmAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "failed to encode msg")
	}
	trackRequest := TrackRequest{
		Account: account,
		Msg: LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		},
	}
	bz, err := json.Marshal(SudoMsg{Track: &trackRequest})
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "failed to marshall AuthenticationRequest")
	}

	_, err = cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return iface.Block(fmt.Errorf("failed to encode msg: %w", err))
	}

	// TODO: Do we want to pass the authentication data here? Should we wait until we have a usecase where we need it?
	confirmExecutionRequest := ConfirmExecutionRequest{
		Account: account,
		Msg: LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		},
	}
	bz, err := json.Marshal(SudoMsg{ConfirmExecution: &confirmExecutionRequest})
	if err != nil {
		return iface.Block(fmt.Errorf("failed to marshall AuthenticationRequest: %w", err))
	}

	result, err := cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return iface.Block(err)
	}
	confirmationResult, err := UnmarshalConfirmationResult(result)
	if err != nil {
		return iface.Block(fmt.Errorf("failed to unmarshal confirmation result: %w", err))
	}
	return confirmationResult
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	_, err := sdk.AccAddressFromBech32(string(data))
	if err != nil {
		return err
	}
	// TODO: check contract address length. Check contract exists?
	return nil
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func UnmarshalAuthenticationResult(data []byte) (iface.AuthenticationResult, error) {
	// Unmarshal type field
	var rawType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &rawType); err != nil {
		return nil, err
	}

	switch rawType.Type { // using snake case here because that's what cosmwasm defaults to
	case "authenticated":
		return iface.Authenticated(), nil
	case "not_authenticated":
		return iface.NotAuthenticated(), nil
	case "rejected":
		var content struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, err
		}
		return iface.Rejected(content.Msg, fmt.Errorf("cosmwasm contract error")), nil
	default:
		return nil, fmt.Errorf("invalid authentication result type: %s", rawType.Type)
	}
}

func UnmarshalConfirmationResult(data []byte) (iface.ConfirmationResult, error) {
	var rawType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &rawType); err != nil {
		return nil, err
	}

	switch rawType.Type { // using snake case here because that's what cosmwasm defaults to
	case "confirm":
		return iface.Confirm(), nil
	case "block":
		var content struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, err
		}
		return iface.Block(fmt.Errorf("cosmwasm contract error: %s", content.Msg)), nil
	default:
		return nil, fmt.Errorf("invalid confirmation result type: %s", rawType.Type)
	}
}
