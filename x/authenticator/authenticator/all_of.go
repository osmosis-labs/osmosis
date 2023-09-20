package authenticator

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AllOfAuthenticator struct {
	SubAuthenticators []Authenticator
	executedAuths     []Authenticator
	am                *AuthenticatorManager
}

type AllOfAuthenticatorData struct {
	Data []AuthenticatorData
}

var _ Authenticator = &AllOfAuthenticator{}
var _ AuthenticatorData = &AllOfAuthenticatorData{}

func NewAllOfAuthenticator(am *AuthenticatorManager) AllOfAuthenticator {
	return AllOfAuthenticator{
		am:                am,
		SubAuthenticators: []Authenticator{},
		executedAuths:     []Authenticator{},
	}
}

func (aoa AllOfAuthenticator) Type() string {
	return "AllOfAuthenticator"
}

func (aoa AllOfAuthenticator) StaticGas() uint64 {
	var totalGas uint64
	for _, auth := range aoa.SubAuthenticators {
		totalGas += auth.StaticGas()
	}
	return totalGas
}

func (aoa AllOfAuthenticator) Initialize(data []byte) (Authenticator, error) {
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return nil, err
	}

	for _, initData := range initDatas {
		for _, authenticatorCode := range aoa.am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				instance, err := authenticatorCode.Initialize(initData.Data)
				if err != nil {
					return nil, err
				}
				aoa.SubAuthenticators = append(aoa.SubAuthenticators, instance)
			}
		}
	}

	return aoa, nil
}

func (aoa AllOfAuthenticator) GetAuthenticationData(
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

	return AllOfAuthenticatorData{Data: authDataList}, nil
}

func (aoa AllOfAuthenticator) Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticationData AuthenticatorData,
) (bool, error) {
	allOfData, ok := authenticationData.(AllOfAuthenticatorData)
	if !ok {
		return false, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "invalid authentication data for AllOfAuthenticator")
	}

	aoa.executedAuths = []Authenticator{}
	for idx, auth := range aoa.SubAuthenticators {
		success, err := auth.Authenticate(ctx, msg, allOfData.Data[idx])
		// TODO: fix static check here;
		// SA4005: ineffective assignment to field AllOfAuthenticator.executedAuth
		aoa.executedAuths = append(aoa.executedAuths, auth) // nolint:staticcheck
		if !success {
			return false, err
		}
	}
	return true, nil
}

func (aoa AllOfAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData AuthenticatorData, msg sdk.Msg) {
	for _, auth := range aoa.executedAuths {
		auth.AuthenticationFailed(ctx, authenticatorData, msg)
	}
}

func (aoa AllOfAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticationData AuthenticatorData) bool {
	for _, auth := range aoa.executedAuths {
		if !auth.ConfirmExecution(ctx, msg, authenticationData) {
			return false
		}
	}
	return true
}
