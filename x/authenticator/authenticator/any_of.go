package authenticator

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AnyOfAuthenticator struct {
	SubAuthenticators []Authenticator
	am                *AuthenticatorManager
}

var (
	_ Authenticator = &AnyOfAuthenticator{}
)

func NewAnyOfAuthenticator(am *AuthenticatorManager) AnyOfAuthenticator {
	return AnyOfAuthenticator{
		am:                am,
		SubAuthenticators: []Authenticator{},
	}
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
	var initDatas []SubAuthenticatorInitData
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

func (aoa AnyOfAuthenticator) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	if len(aoa.SubAuthenticators) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	err := subHandleRequest(
		ctx, request, aoa.SubAuthenticators, requireAnyPass,
		func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error {
			err := auth.Authenticate(ctx, request)

			if err != nil {
				ctx.Logger().Error("sub-authenticator failed to authenticate", "id", request.AuthenticatorId, "authenticator", auth.Type(), "error", err.Error())
			}

			return err
		},
	)

	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "all sub-authenticators failed to authenticate")
	}

	return nil
}

func (aoa AnyOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64, authenticatorId string) error {
	return subTrack(ctx, account, feePayer, msg, msgIndex, authenticatorId, aoa.SubAuthenticators)
}

func (aoa AnyOfAuthenticator) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	return subHandleRequest(
		ctx, request, aoa.SubAuthenticators, requireAnyPass,
		func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error {
			return auth.ConfirmExecution(ctx, request)
		},
	)
}

func (aoa AnyOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return onSubAuthenticatorsAdded(ctx, account, data, authenticatorId, aoa.am)
}

func (aoa AnyOfAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return onSubAuthenticatorsRemoved(ctx, account, data, authenticatorId, aoa.am)
}
