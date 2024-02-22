package authenticator

import (
	"encoding/json"
	"strconv"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AllOfAuthenticator struct {
	SubAuthenticators []iface.Authenticator
	am                *AuthenticatorManager
}

var _ iface.Authenticator = &AllOfAuthenticator{}

func NewAllOfAuthenticator(am *AuthenticatorManager) AllOfAuthenticator {
	return AllOfAuthenticator{
		am:                am,
		SubAuthenticators: []iface.Authenticator{},
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

	if len(initDatas) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	for _, initData := range initDatas {
		for _, authenticatorCode := range aoa.am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				instance, err := authenticatorCode.Initialize(initData.Data)
				if err != nil {
					return nil, err
				}
				aoa.SubAuthenticators = append(aoa.SubAuthenticators, instance)
				continue
			}
		}
	}

	// If not all sub-authenticators are registered, return an error
	if len(aoa.SubAuthenticators) != len(initDatas) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to initialize all sub-authenticators")
	}

	return aoa, nil
}

func (aoa AllOfAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	if len(aoa.SubAuthenticators) == 0 {
		return iface.NotAuthenticated()
	}

	baseId := request.AuthenticatorId
	for id, auth := range aoa.SubAuthenticators {
		request.AuthenticatorId = baseId + "." + strconv.Itoa(id)
		result := auth.Authenticate(ctx, request)
		if !result.IsAuthenticated() {
			return result
		}
	}

	return iface.Authenticated()
}

func (aoa AllOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticatorId string) error {
	for id, auth := range aoa.SubAuthenticators {
		err := auth.Track(ctx, account, msg, authenticatorId+"."+strconv.Itoa(id))
		if err != nil {
			return err
		}
	}
	return nil
}

func (aoa AllOfAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
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

func (aoa AllOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	var initDatas []InitializationData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return err
	}
	if err := validateSubAuthenticatorData(initDatas, aoa.am); err != nil {
		return err
	}
	return nil
}

func (aoa AllOfAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return nil
}

func validateSubAuthenticatorData(initDatas []InitializationData, am *AuthenticatorManager) error {
	if len(initDatas) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	subAuthenticatorCount := 0
	for _, initData := range initDatas {
		for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				subAuthenticatorCount++
				break
			}
		}
	}
	// TODO: Should we recursively call OnAdded here?

	// If not all sub-authenticators are registered, return an error
	if subAuthenticatorCount != len(initDatas) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to initialize all sub-authenticators")
	}

	return nil
}
