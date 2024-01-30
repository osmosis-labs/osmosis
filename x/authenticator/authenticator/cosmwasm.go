package authenticator

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
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
	contractAddr, params, err := parseInitData(data)
	if err != nil {
		return nil, err
	}
	cwa.contractAddr = contractAddr
	cwa.authenticatorParams = params
	return cwa, nil
}

func (cwa CosmwasmAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int,
	simulate bool,
) (iface.AuthenticatorData, error) {
	// We ignore message index here as we want all signers and signatures to be passed to the contract
	signers, signatures, signingTx, err := GetCommonAuthenticationData(ctx, tx, -1, simulate)
	if err != nil {
		return SignatureData{}, err
	}

	return SignatureData{
		Signers:    signers,
		Signatures: signatures,
		Tx:         signingTx,
		Simulate:   simulate,
	}, nil
}

type SignModeData struct {
	Direct  []byte `json:"sign_mode_direct"`
	Textual string `json:"sign_mode_textual"`
}

type LocalAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

type ExplicitTxData struct {
	ChainID       string     `json:"chain_id"`
	AccountNumber uint64     `json:"account_number"`
	Sequence      uint64     `json:"sequence"`
	TimeoutHeight uint64     `json:"timeout_height"`
	Msgs          []LocalAny `json:"msgs"`
	Memo          string     `json:"memo"`
}

type simplifiedSignatureData struct {
	Signers    []sdk.AccAddress `json:"signers"`
	Signatures [][]byte         `json:"signatures"`
}

type AuthenticationRequest struct {
	Account             sdk.AccAddress          `json:"account"`
	Msg                 LocalAny                `json:"msg"`
	Signature           []byte                  `json:"signature"` // Only allowing messages with a single signer
	SignModeTxData      SignModeData            `json:"sign_mode_tx_data"`
	TxData              ExplicitTxData          `json:"tx_data"`
	SignatureData       simplifiedSignatureData `json:"signature_data"`
	Simulate            bool                    `json:"simulate"`
	AuthenticatorParams []byte                  `json:"authenticator_params,omitempty"`
}

type TrackRequest struct {
	Account sdk.AccAddress `json:"account"`
	Msg     LocalAny       `json:"msg"`
}

type ConfirmExecutionRequest struct {
	Account sdk.AccAddress `json:"account"`
	Msg     LocalAny       `json:"msg"`
}

type OnAuthenticatorAddedRequest struct {
	Account             sdk.AccAddress `json:"account"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type OnAuthenticatorRemovedRequest struct {
	Account             sdk.AccAddress `json:"account"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type SudoMsg struct {
	Authenticate           *AuthenticationRequest         `json:"authenticate,omitempty"`
	Track                  *TrackRequest                  `json:"track,omitempty"`
	ConfirmExecution       *ConfirmExecutionRequest       `json:"confirm_execution,omitempty"`
	OnAuthenticatorAdded   *OnAuthenticatorAddedRequest   `json:"on_authenticator_added,omitempty"`
	OnAuthenticatorRemoved *OnAuthenticatorRemovedRequest `json:"on_authenticator_removed,omitempty"`
}

// TODO: decide when we want to reject and when to just not authenticate
func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	if len(msg.GetSigners()) != 1 {
		return iface.Rejected("only messages with a single signer are supported", sdkerrors.ErrInvalidType)
	}

	signatureData, ok := authenticationData.(SignatureData)
	if !ok {
		return iface.Rejected("invalid signature verification data", sdkerrors.ErrInvalidType)
	}

	// Retrieve and build the signer data struct

	// TODO: ctx.isGenesis() replacement?
	// old: genesis := ctx.isGenesis() || ctx.BlockHeight() == 0
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	baseAccount := cwa.ak.GetAccount(ctx, account)
	if !genesis {
		accNum = baseAccount.GetAccountNumber()
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      baseAccount.GetSequence(),
	}

	signBytes, err := cwa.sigModeHandler.GetSignBytes(txsigning.SignMode_SIGN_MODE_DIRECT, signerData, signatureData.Tx)
	if err != nil {
		return iface.Rejected("failed to get signBytes", err)
	}

	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return iface.Rejected("failed to encode msg", err)
	}

	timeoutTx, ok := signatureData.Tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return iface.Rejected("failed to cast tx to TxWithTimeoutHeight", sdkerrors.ErrInvalidType)
	}
	memoTx, ok := signatureData.Tx.(sdk.TxWithMemo)
	if !ok {
		return iface.Rejected("failed to cast tx to TxWithMemo", sdkerrors.ErrInvalidType)
	}

	msgs := make([]LocalAny, len(signatureData.Tx.GetMsgs()))
	for i, msg := range signatureData.Tx.GetMsgs() {
		encodedMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return iface.Rejected("failed to encode msg", err)
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
	for i, signature := range signatureData.Signatures {
		single, ok := signature.Data.(*txsigning.SingleSignatureData)
		if !ok {
			return iface.Rejected("failed to cast signature to SingleSignatureData", sdkerrors.ErrInvalidType)
		}
		signatures = append(signatures, single.Signature)
		if signatureData.Signers[i].Equals(signer) {
			msgSignature = single.Signature
		}
	}
	// should we pass ctx.IsReCheckTx() here?
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
		SignatureData: simplifiedSignatureData{
			Signers:    signatureData.Signers,
			Signatures: signatures,
		},
		Simulate:            signatureData.Simulate,
		AuthenticatorParams: cwa.authenticatorParams,
	}
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
	contractAddr, params, err := parseInitData(data)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(SudoMsg{OnAuthenticatorAdded: &OnAuthenticatorAddedRequest{
		Account:             account,
		AuthenticatorParams: params,
	}})
	if err != nil {
		return err
	}

	_, err = cwa.contractKeeper.Sudo(ctx, contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	contractAddr, params, err := parseInitData(data)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(SudoMsg{OnAuthenticatorRemoved: &OnAuthenticatorRemovedRequest{
		Account:             account,
		AuthenticatorParams: params,
	}})
	if err != nil {
		return err
	}

	_, err = cwa.contractKeeper.Sudo(ctx, contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) ContractAddress() sdk.AccAddress {
	return cwa.contractAddr
}

func (cwa CosmwasmAuthenticator) Params() []byte {
	return cwa.authenticatorParams
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

func parseInitData(data []byte) (sdk.AccAddress, []byte, error) {
	var initData CosmwasmAuthenticatorInitData
	err := json.Unmarshal(data, &initData)
	if err != nil {
		return nil, nil, err
	}

	// check if contract address is empty
	if len(initData.Contract) == 0 {
		return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "missing contract address")
	}

	// check if contract address is valid
	contractAddr, err := sdk.AccAddressFromBech32(initData.Contract)
	if err != nil {
		return nil, nil, err
	}

	// params are optional, early return if they are not present
	if initData.Params == nil || len(initData.Params) == 0 {
		return contractAddr, nil, nil
	}

	// check if initData.Params is valid json bytes
	var jsonTest map[string]interface{}
	err = json.Unmarshal(initData.Params, &jsonTest)
	if err != nil {
		return nil, nil, err
	}

	return contractAddr, initData.Params, nil
}
