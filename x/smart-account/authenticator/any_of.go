package authenticator

import (
	"encoding/json"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type SignatureAssignment string

const (
	Single      SignatureAssignment = "single"
	Partitioned SignatureAssignment = "partitioned"
)

type AnyOf struct {
	SubAuthenticators   []Authenticator
	am                  *AuthenticatorManager
	signatureAssignment SignatureAssignment
}

var _ Authenticator = &AnyOf{}

func NewAnyOf(am *AuthenticatorManager) AnyOf {
	return AnyOf{
		am:                  am,
		SubAuthenticators:   []Authenticator{},
		signatureAssignment: Single,
	}
}

func NewPartitionedAnyOf(am *AuthenticatorManager) AnyOf {
	return AnyOf{
		am:                  am,
		SubAuthenticators:   []Authenticator{},
		signatureAssignment: Partitioned,
	}
}

func (aoa AnyOf) Type() string {
	if aoa.signatureAssignment == Single {
		return "AnyOf"
	}
	return "PartitionedAnyOf"
}

func (aoa AnyOf) StaticGas() uint64 {
	var totalGas uint64
	for _, auth := range aoa.SubAuthenticators {
		totalGas += auth.StaticGas()
	}
	return totalGas
}

func (aoa AnyOf) Initialize(config []byte) (Authenticator, error) {
	// Decode the initialization data for each sub-authenticator
	var initDatas []SubAuthenticatorInitData
	if err := json.Unmarshal(config, &initDatas); err != nil {
		return nil, errorsmod.Wrap(err, "failed to parse sub-authenticators initialization data")
	}

	if len(initDatas) <= 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "anyOf must have at least 2 sub-authenticators")
	}

	// Call Initialize on each sub-authenticator with its appropriate data using AuthenticatorManager
	for _, initData := range initDatas {
		authenticatorCode := aoa.am.GetAuthenticatorByType(initData.Type)
		instance, err := authenticatorCode.Initialize(initData.Config)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to initialize sub-authenticator (type = %s)", initData.Type)
		}
		aoa.SubAuthenticators = append(aoa.SubAuthenticators, instance)
	}

	// If not all sub-authenticators are registered, return an error
	if len(aoa.SubAuthenticators) != len(initDatas) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to initialize all sub-authenticators")
	}

	return aoa, nil
}

func (aoa AnyOf) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	if len(aoa.SubAuthenticators) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	var subAuthErrors []string
	var err error

	// If the signature assignment is partitioned, we need to split the signatures and pass them to the sub-authenticators
	var signatures [][]byte
	if aoa.signatureAssignment == Partitioned {
		// Partitioned signatures are decoded and passed one by one as the signature of the sub-authenticator
		signatures, err = splitSignatures(request.Signature, len(aoa.SubAuthenticators))
		if err != nil {
			return err
		}
	}

	baseId := request.AuthenticatorId
	for i, auth := range aoa.SubAuthenticators {
		// update the authenticator id to include the sub-authenticator id
		request.AuthenticatorId = compositeId(baseId, i)
		// update the request to include the sub-authenticator signature
		if aoa.signatureAssignment == Partitioned {
			request.Signature = signatures[i]
		}
		err = auth.Authenticate(ctx, request)
		if err == nil { // Success!
			return nil
		}

		// If the sub-authenticator fails, we want to continue to the next one.
		// We accumulate any errors so that they can all be surfaced to the user
		subAuthErrors = append(subAuthErrors, fmt.Sprintf("[%s (id = %s)] %s; ", auth.Type(), request.AuthenticatorId, err))
	}

	if err != nil {
		// return all errors
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "all sub-authenticators failed to authenticate: %s", strings.Join(subAuthErrors, "; "))
	}

	return nil
}

func (aoa AnyOf) Track(ctx sdk.Context, request AuthenticationRequest) error {
	return subTrack(ctx, request, aoa.SubAuthenticators)
}

// ConfirmExecution is called on all sub-authenticators, but only the changes made by the authenticator that succeeds are written.
func (aoa AnyOf) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	var signatures [][]byte
	var err error

	// If the signature assignment is partitioned, we need to split the signatures and pass them to the sub-authenticators
	if aoa.signatureAssignment == Partitioned {
		// Partitioned signatures are decoded and passed one by one as the signature of the sub-authenticator
		signatures, err = splitSignatures(request.Signature, len(aoa.SubAuthenticators))
		if err != nil {
			return err
		}
	}
	var subAuthErrors []string

	baseId := request.AuthenticatorId
	for i, auth := range aoa.SubAuthenticators {
		// update the request to include the sub-authenticator id
		request.AuthenticatorId = compositeId(baseId, i)
		if aoa.signatureAssignment == Partitioned {
			// update the request to include the sub-authenticator signature
			request.Signature = signatures[i]
		}
		// We only want to write changes made by the authenticator that succeeds.
		// If the authenticator fails,its changes are discarded, and we want to continue to the next one.
		cacheCtx, write := ctx.CacheContext()
		err = auth.ConfirmExecution(cacheCtx, request)
		if err == nil {
			write()
			return nil
		}
		subAuthErrors = append(subAuthErrors, fmt.Sprintf("[%s (id = %s)] %s; ", auth.Type(), request.AuthenticatorId, err))
	}

	return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "all sub-authenticators failed to confirm execution: %s", strings.Join(subAuthErrors, "; "))
}

func (aoa AnyOf) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return onSubAuthenticatorsAdded(ctx, account, config, authenticatorId, aoa.am)
}

func (aoa AnyOf) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return onSubAuthenticatorsRemoved(ctx, account, config, authenticatorId, aoa.am)
}
