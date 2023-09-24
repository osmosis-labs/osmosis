package authenticator

import (
	"encoding/json"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AllOfAuthenticator struct {
	SubAuthenticators []iface.Authenticator
	executedAuths     []iface.Authenticator
	am                *AuthenticatorManager
}

type AllOfAuthenticatorData struct {
	Data []iface.AuthenticatorData
}

var (
	_ iface.Authenticator     = &AllOfAuthenticator{}
	_ iface.AuthenticatorData = &AllOfAuthenticatorData{}
)

func NewAllOfAuthenticator(am *AuthenticatorManager) AllOfAuthenticator {
	return AllOfAuthenticator{
		am:                am,
		SubAuthenticators: []iface.Authenticator{},
		executedAuths:     []iface.Authenticator{},
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

func (aoa AllOfAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
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
) (iface.AuthenticatorData, error) {
	var authDataList []iface.AuthenticatorData
	for _, auth := range aoa.SubAuthenticators {
		data, err := auth.GetAuthenticationData(ctx, tx, messageIndex, simulate)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "a sub-authenticator failed to get authentication data")
		}
		authDataList = append(authDataList, data)
	}

	return AllOfAuthenticatorData{Data: authDataList}, nil
}

func (aoa AllOfAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	allOfData, ok := authenticationData.(AllOfAuthenticatorData)
	if !ok {
		return iface.Rejected("invalid authentication data for AllOfAuthenticator", nil)
	}

	for idx, auth := range aoa.SubAuthenticators {
		result := auth.Authenticate(ctx, account, msg, allOfData.Data[idx])
		if !result.IsAuthenticated() {
			return result
		}
	}
	return iface.Authenticated()
}

func (aoa AllOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	for _, auth := range aoa.SubAuthenticators {
		err := auth.Track(ctx, account, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (aoa AllOfAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	for _, auth := range aoa.executedAuths {
		if confirmation := auth.ConfirmExecution(ctx, nil, msg, authenticationData); confirmation.IsBlock() {
			return confirmation
		}
	}
	return iface.Confirm()
}

func (aoa AllOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return err
	}
	return nil
}
