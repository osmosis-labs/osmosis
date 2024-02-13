package authenticator

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

type SudoMsg struct {
	Authenticate     *iface.AuthenticationRequest `json:"authenticate,omitempty"`
	Track            *TrackRequest                `json:"track,omitempty"`
	ConfirmExecution *ConfirmExecutionRequest     `json:"confirm_execution,omitempty"`
}

func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	// Add the authenticator params set for this authenticator in Initialize()
	request.AuthenticatorParams = cwa.authenticatorParams

	bz, err := json.Marshal(SudoMsg{Authenticate: &request})
	if err != nil {
		// REVIEW Q: Should this be reject or just not authenticated?
		return iface.Rejected("failed to marshall AuthenticationRequest", err)
	}

	result, err := cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		// REVIEW Q: Should this be reject or just not authenticated?
		return iface.Rejected("failed to sudo", err)
	}

	authResult, err := UnmarshalAuthenticationResult(result)
	if err != nil {
		// REVIEW Q: Should this be reject or just not authenticated?
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
		Msg: iface.LocalAny{
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

func (cwa CosmwasmAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	// TODO: Do we want to pass the authentication data here? Should we wait until we have a usecase where we need it?
	confirmExecutionRequest := ConfirmExecutionRequest{
		Account: request.Account,
		Msg:     request.Msg,
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
