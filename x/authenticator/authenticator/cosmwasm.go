package authenticator

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CosmwasmAuthenticator struct {
	contractKeeper *keeper.PermissionedKeeper
	cdc            codectypes.AnyUnpacker
	Handler        authsigning.SignModeHandler

	contractAddr sdk.AccAddress
}

var (
	_ iface.Authenticator = &CosmwasmAuthenticator{}
)

func NewCosmwasmAuthenticator(contractKeeper *keeper.PermissionedKeeper, cdc codectypes.AnyUnpacker) CosmwasmAuthenticator {
	return CosmwasmAuthenticator{
		contractKeeper: contractKeeper,
		cdc:            cdc,
	}
}

func (cwa CosmwasmAuthenticator) Type() string {
	return "CosmwasmAuthenticator"
}

func (cwa CosmwasmAuthenticator) StaticGas() uint64 {
	return 0
}

func (cwa CosmwasmAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	contractAddr, err := sdk.AccAddressFromBech32(string(data))
	if err != nil {
		return nil, err
	}
	cwa.contractAddr = contractAddr
	return cwa, nil
}

func (cwa CosmwasmAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int,
	simulate bool,
) (iface.AuthenticatorData, error) {
	signers, signatures, signingTx, err := GetCommonAuthenticationData(ctx, tx, messageIndex, simulate)
	if err != nil {
		return SignatureData{}, err
	}

	signBytes, err := cwa.Handler.GetSignBytes(data.SignMode, signerData, tx)
	if err != nil {
		return err
	}

	// Get the signature for the message at msgIndex
	return SignatureData{
		Signers:    signers,
		Signatures: signatures,
		Tx:         signingTx,
		Simulate:   simulate,
	}, nil
}

type AuthenticateMsg struct {
	TxData  []byte          `json:"tx_data"`
	Account sdk.AccAddress  `json:"account"`
	Msg     *codectypes.Any `json:"msg"`
}

func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	verificationData, ok := authenticationData.(SignatureData)
	if !ok {
		return iface.Rejected("invalid signature verification data", sdkerrors.ErrInvalidType)
	}

	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil
	}

	authMsg := AuthenticateMsg{
		TxData:  verificationData.Tx.GetSignBytes(),
		Account: account,
		Msg:     encodedMsg,
	}
	bz, err := json.Marshal(authMsg)
	if err != nil {
		return iface.Rejected("failed to marshall AuthenticateMsg", err)
	}

	result, err := cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return iface.Rejected("failed to sudo", err)
	}
	fmt.Println("result", result)
	// TODO: interpret the result
	if len(result) > 0 {
		return iface.Authenticated()
	}

	return iface.NotAuthenticated()
}

func (cwa CosmwasmAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (cwa CosmwasmAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
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
