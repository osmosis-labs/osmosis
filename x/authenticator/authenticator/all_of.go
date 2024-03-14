package authenticator

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type AllOfAuthenticator struct {
	SubAuthenticators   []Authenticator
	am                  *AuthenticatorManager
	signatureAssignment SignatureAssignment
}

var _ Authenticator = &AllOfAuthenticator{}

func NewAllOfAuthenticator(am *AuthenticatorManager) AllOfAuthenticator {
	return AllOfAuthenticator{
		am:                  am,
		SubAuthenticators:   []Authenticator{},
		signatureAssignment: Single,
	}
}

func NewPartitionedAllOfAuthenticator(am *AuthenticatorManager) AllOfAuthenticator {
	return AllOfAuthenticator{
		am:                  am,
		SubAuthenticators:   []Authenticator{},
		signatureAssignment: Partitioned,
	}
}

func (aoa AllOfAuthenticator) Type() string {
	if aoa.signatureAssignment == Single {
		return "AllOfAuthenticator"
	}
	return "PartitionedAllOfAuthenticator"
}

func (aoa AllOfAuthenticator) StaticGas() uint64 {
	var totalGas uint64
	for _, auth := range aoa.SubAuthenticators {
		totalGas += auth.StaticGas()
	}
	return totalGas
}

func (aoa AllOfAuthenticator) Initialize(data []byte) (Authenticator, error) {
	var initDatas []SubAuthenticatorInitData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return nil, errorsmod.Wrap(err, "failed to parse sub-authenticators initialization data")
	}

	if len(initDatas) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	for _, initData := range initDatas {
		for _, authenticatorCode := range aoa.am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				instance, err := authenticatorCode.Initialize(initData.Data)
				if err != nil {
					return nil, errorsmod.Wrapf(err, "failed to initialize sub-authenticator (type = %s)", initData.AuthenticatorType)
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

func (aoa AllOfAuthenticator) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	if len(aoa.SubAuthenticators) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	return subHandleRequest(ctx, request, aoa.SubAuthenticators, requireAllPass, aoa.signatureAssignment, func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error {
		return auth.Authenticate(ctx, request)
	})
}

func (aoa AllOfAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64, authenticatorId string) error {
	return subTrack(ctx, account, feePayer, msg, msgIndex, authenticatorId, aoa.SubAuthenticators)
}

func (aoa AllOfAuthenticator) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	return subHandleRequest(ctx, request, aoa.SubAuthenticators, requireAllPass, aoa.signatureAssignment, func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error {
		return auth.ConfirmExecution(ctx, request)
	})
}

func (aoa AllOfAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return onSubAuthenticatorsAdded(ctx, account, data, authenticatorId, aoa.am)
}

func (aoa AllOfAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return onSubAuthenticatorsRemoved(ctx, account, data, authenticatorId, aoa.am)
}
