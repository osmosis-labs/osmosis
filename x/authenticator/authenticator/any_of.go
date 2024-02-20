package authenticator

import (
	"encoding/json"
	"strconv"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AnyOfAuthenticator struct {
	SubAuthenticators []iface.Authenticator
	am                *AuthenticatorManager
}

var (
	_ iface.Authenticator = &AnyOfAuthenticator{}
)

func NewAnyOfAuthenticator(am *AuthenticatorManager) AnyOfAuthenticator {
	return AnyOfAuthenticator{
		am:                am,
		SubAuthenticators: []iface.Authenticator{},
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
				break
			}
		}
	}

	// If not all sub-authenticators are registered, return an error
	if len(aoa.SubAuthenticators) != len(initDatas) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to initialize all sub-authenticators")
	}

	return aoa, nil
}

func (aoa AnyOfAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	baseId := request.AuthenticatorId
	for id, auth := range aoa.SubAuthenticators {
		request.AuthenticatorId = baseId + "." + strconv.Itoa(id)
		result := auth.Authenticate(ctx, request)
		if result.IsAuthenticated() || result.IsRejected() {
			return result
		}
	}
	return iface.NotAuthenticated()
}

func (aoa AnyOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticatorId string) error {
	for id, auth := range aoa.SubAuthenticators {
		err := auth.Track(ctx, account, msg, authenticatorId+"."+strconv.Itoa(id))
		if err != nil {
			return err
		}
	}
	return nil
}

func (aoa AnyOfAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	baseId := request.AuthenticatorId
	for id, auth := range aoa.SubAuthenticators {
		request.AuthenticatorId = baseId + "." + strconv.Itoa(id)
		result := auth.ConfirmExecution(ctx, request)
		if result.IsBlock() {
			return result
		}
	}
	return iface.Confirm()
}

func (aoa AnyOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return err
	}
	// TODO: Consume extra gas for each sub authenticator to avoid spam? (same on allOf)
	if err := validateSubAuthenticatorData(initDatas, aoa.am); err != nil {
		return err
	}
	return nil
}

func (aoa AnyOfAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
