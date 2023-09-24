package authenticator

import (
	"encoding/json"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AnyOfAuthenticator struct {
	SubAuthenticators []iface.Authenticator
	executedAuths     []iface.Authenticator // track which sub-authenticators were executed

	am *AuthenticatorManager
}

type AnyOfAuthenticatorData struct {
	Data []iface.AuthenticatorData
}

var (
	_ iface.Authenticator     = &AnyOfAuthenticator{}
	_ iface.AuthenticatorData = &AnyOfAuthenticatorData{}
)

func NewAnyOfAuthenticator(am *AuthenticatorManager) AnyOfAuthenticator {
	return AnyOfAuthenticator{
		am:                am,
		SubAuthenticators: []iface.Authenticator{},
		executedAuths:     []iface.Authenticator{},
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

func (aoa AnyOfAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
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
) (iface.AuthenticatorData, error) {
	var authDataList []iface.AuthenticatorData
	for _, auth := range aoa.SubAuthenticators {
		data, err := auth.GetAuthenticationData(ctx, tx, messageIndex, simulate)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "a sub-authenticator failed to get authentication data")
		}
		authDataList = append(authDataList, data)
	}

	return AnyOfAuthenticatorData{Data: authDataList}, nil
}

func (aoa AnyOfAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	anyOfData, ok := authenticationData.(AnyOfAuthenticatorData)
	if !ok {
		return iface.Rejected("invalid authentication data for AnyOfAuthenticator", nil)
	}

	for idx, auth := range aoa.SubAuthenticators {
		result := auth.Authenticate(ctx, nil, msg, anyOfData.Data[idx])
		if result.IsAuthenticated() || result.IsRejected() {
			// TODO: Do we want to wrap the error in  case or rejection?
			return result
		}
	}
	return iface.NotAuthenticated()
}

func (aoa AnyOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	for _, auth := range aoa.SubAuthenticators {
		err := auth.Track(ctx, account, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (aoa AnyOfAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	// Call ConfirmExecution on executed sub-authenticators
	for _, auth := range aoa.executedAuths {
		if confirmation := auth.ConfirmExecution(ctx, nil, msg, authenticationData); confirmation.IsBlock() {
			return confirmation
		}
	}
	return iface.Confirm()
}

func (aoa AnyOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return err
	}
	return nil
}

func (aoa AnyOfAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
