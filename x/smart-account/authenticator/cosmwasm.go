package authenticator

import (
	"encoding/json"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type CosmwasmAuthenticator struct {
	contractKeeper types.ContractKeeper
	ak             authante.AccountKeeper
	cdc            codectypes.AnyUnpacker

	contractAddr        sdk.AccAddress
	authenticatorParams []byte
}

var _ Authenticator = &CosmwasmAuthenticator{}

func NewCosmwasmAuthenticator(contractKeeper *keeper.PermissionedKeeper, accountKeeper authante.AccountKeeper, cdc codectypes.AnyUnpacker) CosmwasmAuthenticator {
	return CosmwasmAuthenticator{
		contractKeeper: contractKeeper,
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

func (cwa CosmwasmAuthenticator) Initialize(config []byte) (Authenticator, error) {
	contractAddr, params, err := parseInitData(config)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to parse initialization data")
	}
	cwa.contractAddr = contractAddr
	cwa.authenticatorParams = params
	return cwa, nil
}

type OnAuthenticatorAddedRequest struct {
	Account             sdk.AccAddress `json:"account"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
	AuthenticatorId     string         `json:"authenticator_id"`
}

type OnAuthenticatorRemovedRequest struct {
	Account             sdk.AccAddress `json:"account"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
	AuthenticatorId     string         `json:"authenticator_id"`
}

type SudoMsg struct {
	OnAuthenticatorAdded   *OnAuthenticatorAddedRequest   `json:"on_authenticator_added,omitempty"`
	OnAuthenticatorRemoved *OnAuthenticatorRemovedRequest `json:"on_authenticator_removed,omitempty"`
	Authenticate           *AuthenticationRequest         `json:"authenticate,omitempty"`
	Track                  *TrackRequest                  `json:"track,omitempty"`
	ConfirmExecution       *ConfirmExecutionRequest       `json:"confirm_execution,omitempty"`
}

func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	// Add the authenticator params set for this authenticator in Initialize()
	request.AuthenticatorParams = cwa.authenticatorParams

	bz, err := json.Marshal(SudoMsg{Authenticate: &request})
	if err != nil {
		return errorsmod.Wrapf(err, "failed to marshall AuthenticationRequest")
	}

	_, err = cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) Track(ctx sdk.Context, request AuthenticationRequest) error {
	trackRequest := TrackRequest{
		AuthenticatorId:     request.AuthenticatorId,
		Account:             request.Account,
		FeePayer:            request.FeePayer,
		FeeGranter:          request.FeeGranter,
		Fee:                 request.Fee,
		Msg:                 request.Msg,
		MsgIndex:            request.MsgIndex,
		AuthenticatorParams: cwa.authenticatorParams,
	}
	bz, err := json.Marshal(SudoMsg{Track: &trackRequest})
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to marshall TrackRequest")
	}

	_, err = cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	confirmExecutionRequest := ConfirmExecutionRequest{
		AuthenticatorId:     request.AuthenticatorId,
		Account:             request.Account,
		FeePayer:            request.FeePayer,
		FeeGranter:          request.FeeGranter,
		Fee:                 request.Fee,
		Msg:                 request.Msg,
		MsgIndex:            request.MsgIndex,
		AuthenticatorParams: cwa.authenticatorParams,
	}
	bz, err := json.Marshal(SudoMsg{ConfirmExecution: &confirmExecutionRequest})
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshall ConfirmExecutionRequest")
	}

	_, err = cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	contractAddr, params, err := parseInitData(config)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(SudoMsg{OnAuthenticatorAdded: &OnAuthenticatorAddedRequest{
		Account:             account,
		AuthenticatorParams: params,
		AuthenticatorId:     authenticatorId,
	}})
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to marshall OnAuthenticatorAddedRequest")
	}

	_, err = cwa.contractKeeper.Sudo(ctx, contractAddr, bz)
	if err != nil {
		return err
	}

	return nil
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	contractAddr, params, err := parseInitData(config)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(SudoMsg{OnAuthenticatorRemoved: &OnAuthenticatorRemovedRequest{
		Account:             account,
		AuthenticatorParams: params,
		AuthenticatorId:     authenticatorId,
	}})
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to marshall OnAuthenticatorRemovedRequest")
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
		return nil, nil, errorsmod.Wrap(err, "invalid contract address")
	}

	// params are optional, early return if they are not present
	if initData.Params == nil || len(initData.Params) == 0 {
		return contractAddr, nil, nil
	}

	// check if initData.Params is valid json bytes
	var jsonTest map[string]interface{}
	err = json.Unmarshal(initData.Params, &jsonTest)
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "params must be valid json bytes")
	}

	return contractAddr, initData.Params, nil
}
