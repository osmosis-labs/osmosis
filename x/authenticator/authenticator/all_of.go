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
				instance, err := authenticatorCode.Initialize([]byte(initData.Data))
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

func (aoa AllOfAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) AuthenticationResult {
	allOfData, ok := authenticationData.(AllOfAuthenticatorData)
	if !ok {
		return Rejected("invalid authentication data for AllOfAuthenticator", nil)
	}

	aoa.executedAuths = []Authenticator{}
	for idx, auth := range aoa.SubAuthenticators {
		success := auth.Authenticate(ctx, nil, msg, allOfData.Data[idx])
		aoa.executedAuths = append(aoa.executedAuths, auth)
		if !success.IsAuthenticated() {
			return success
		}
	}
	return Authenticated()
}

func (aoa AllOfAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) ConfirmationResult {
	for _, auth := range aoa.executedAuths {
		if confirmation := auth.ConfirmExecution(ctx, nil, msg, authenticationData); confirmation.IsBlock() {
			return confirmation
		}
	}
	return Confirm()
}
