package authenticator

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AnyOfAuthenticator struct {
	SubAuthenticators []Authenticator
	executedAuths     []Authenticator // track which sub-authenticators were executed

	am *AuthenticatorManager
}

type AnyOfAuthenticatorData struct {
	Data []AuthenticatorData
}

var _ Authenticator = &AnyOfAuthenticator{}
var _ AuthenticatorData = &AnyOfAuthenticatorData{}

func NewAnyOfAuthenticator(am *AuthenticatorManager) AnyOfAuthenticator {
	return AnyOfAuthenticator{
		am:                am,
		SubAuthenticators: []Authenticator{},
		executedAuths:     []Authenticator{},
	}
}

type InitializationData struct {
	AuthenticatorType string `json:"authenticator_type"`
	Data              []byte `json:"data"`
}

func (aoa AnyOfAuthenticator) Type() string {
	return "AnyOfAuthenticator"
}

func (aoa AnyOfAuthenticator) StaticGas() uint64 {
	var totalGas uint64
	for _, auth := range aoa.SubAuthenticators {
		totalGas += auth.StaticGas()
	}
	return totalGas
}

func (aoa AnyOfAuthenticator) Initialize(data []byte) (Authenticator, error) {
	// Decode the initialization data for each sub-authenticator
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return nil, err
	}

	// Call Initialize on each sub-authenticator with its appropriate data using AuthenticatorManager
	for _, initData := range initDatas {
		for _, authenticatorCode := range aoa.am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				instance, err := authenticatorCode.Initialize(initData.Data)
				if err != nil {
					return nil, err // Handling the error by returning it
				}
				aoa.SubAuthenticators = append(aoa.SubAuthenticators, instance)
			}
		}
	}

	return aoa, nil
}

func (aoa AnyOfAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int8,
	simulate bool,
) (AuthenticatorData, error) {
	var authDataList []AuthenticatorData
	for _, auth := range aoa.SubAuthenticators {
		data, err := auth.GetAuthenticationData(ctx, tx, messageIndex, simulate)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "a sub-authenticator failed to get authentication data")
		}
		authDataList = append(authDataList, data)
	}

	return AnyOfAuthenticatorData{Data: authDataList}, nil
}

func (aoa AnyOfAuthenticator) Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticationData AuthenticatorData,
) (bool, error) {
	anyOfData, ok := authenticationData.(AnyOfAuthenticatorData)
	if !ok {
		return false, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "invalid authentication data for AnyOfAuthenticator")
	}

	aoa.executedAuths = []Authenticator{} // Reset the executed authenticators
	for idx, auth := range aoa.SubAuthenticators {
		success, err := auth.Authenticate(ctx, msg, anyOfData.Data[idx])
		// TODO: fix static check here;
		// SA4005: ineffective assignment to field AnyOfAuthenticator.executedAuth
		aoa.executedAuths = append(aoa.executedAuths, auth) // nolint:staticcheck
		if success {
			return true, nil
		}
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func (aoa AnyOfAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData AuthenticatorData, msg sdk.Msg) {
	// Call AuthenticationFailed on executed sub-authenticators
	for _, auth := range aoa.executedAuths {
		auth.AuthenticationFailed(ctx, authenticatorData, msg)
	}
}

func (aoa AnyOfAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticationData AuthenticatorData) bool {
	// Call ConfirmExecution on executed sub-authenticators
	for _, auth := range aoa.executedAuths {
		if !auth.ConfirmExecution(ctx, msg, authenticationData) {
			return false
		}
	}
	return true
}
