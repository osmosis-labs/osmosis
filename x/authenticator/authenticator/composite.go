package authenticator

import (
	"encoding/json"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type SubAuthenticatorInitData struct {
	AuthenticatorType string `json:"authenticator_type"`
	Data              []byte `json:"data"`
}

func subTrack(
	ctx sdk.Context,
	account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64,
	authenticatorId string,
	subAuthenticators []Authenticator,
) error {
	for id, auth := range subAuthenticators {
		subId := compositeId(authenticatorId, id)
		err := auth.Track(ctx, account, feePayer, msg, msgIndex, subId)
		if err != nil {
			return errorsmod.Wrapf(err, "sub-authenticator track failed (sub-authenticator id = %s)", subId)
		}
	}
	return nil
}

type PassingReq int

const (
	requireAllPass = iota
	requireAnyPass
)

func subHandleRequest(ctx sdk.Context, request AuthenticationRequest, subAuthenticators []Authenticator,
	passingReq PassingReq, signatureAssignment SignatureAssignment,
	f func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error,
) error {
	if passingReq != requireAllPass && passingReq != requireAnyPass {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid passing req")
	}

	var err error

	// Partitioned signatures are decoded and passed one by one as the signature of the sub-authenticator
	var signatures [][]byte
	if signatureAssignment == Partitioned {
		err = json.Unmarshal(request.Signature, &signatures)
		if err != nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to parse signatures")
		}
		if len(signatures) != len(subAuthenticators) {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid number of signatures")
		}
	}

	baseId := request.AuthenticatorId

	for i, auth := range subAuthenticators {
		// update the authenticator id to include the sub-authenticator id
		request.AuthenticatorId = compositeId(baseId, i)

		if signatureAssignment == Partitioned {
			request.Signature = signatures[i]
		}

		err = f(auth, ctx, request)

		if passingReq == requireAllPass && err != nil {
			return err
		}

		if passingReq == requireAnyPass && err == nil {
			return nil
		}
	}

	// require all pass return no error if it has not yet early returned
	if passingReq == requireAllPass {
		return nil
	}

	// require any pass return error it has not yet early returned
	if passingReq == requireAnyPass {
		return err
	}

	return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid passing req")
}

func onSubAuthenticatorsAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string, am *AuthenticatorManager) error {
	var initDatas []SubAuthenticatorInitData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return errorsmod.Wrapf(err, "failed to unmarshal sub-authenticator init data")
	}

	if len(initDatas) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	baseId := authenticatorId
	subAuthenticatorCount := 0
	for id, initData := range initDatas {
		for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				subId := compositeId(baseId, id)
				err := authenticatorCode.OnAuthenticatorAdded(ctx, account, initData.Data, subId)
				if err != nil {
					return errorsmod.Wrapf(err, "sub-authenticator `OnAuthenticatorAdded` failed (sub-authenticator id = %s)", subId)
				}

				subAuthenticatorCount++
				break
			}
		}
	}

	// If not all sub-authenticators are registered, return an error
	if subAuthenticatorCount != len(initDatas) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to initialize all sub-authenticators")
	}

	return nil
}

func onSubAuthenticatorsRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string, am *AuthenticatorManager) error {
	var initDatas []SubAuthenticatorInitData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return err
	}

	baseId := authenticatorId
	for id, initData := range initDatas {
		for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				subId := compositeId(baseId, id)
				err := authenticatorCode.OnAuthenticatorRemoved(ctx, account, initData.Data, subId)
				if err != nil {
					return errorsmod.Wrapf(err, "sub-authenticator `OnAuthenticatorRemoved` failed (sub-authenticator id = %s)", subId)
				}
				break
			}
		}
	}
	return nil
}

func compositeId(baseId string, subId int) string {
	return baseId + "." + strconv.Itoa(subId)
}
