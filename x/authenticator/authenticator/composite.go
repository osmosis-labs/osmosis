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
	account sdk.AccAddress, msg sdk.Msg, msgIndex uint64,
	authenticatorId string,
	subAuthenticators []Authenticator,
) error {
	for id, auth := range subAuthenticators {
		err := auth.Track(ctx, account, msg, msgIndex, compositeId(authenticatorId, id))
		if err != nil {
			return err
		}
	}
	return nil
}

type PassingReq int

const (
	requireAllPass = iota
	requireAnyPass
)

func subHandleRequest(
	ctx sdk.Context,
	request AuthenticationRequest,
	subAuthenticators []Authenticator,
	passingReq PassingReq,
	f func(auth Authenticator, ctx sdk.Context, request AuthenticationRequest) error,
) error {
	if passingReq != requireAllPass && passingReq != requireAnyPass {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid passing req")
	}

	var err error

	for id, auth := range subAuthenticators {
		// update the authenticator id to include the sub-authenticator id
		request.AuthenticatorId = compositeId(request.AuthenticatorId, id)

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
		return err
	}

	if len(initDatas) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no sub-authenticators provided")
	}

	baseId := authenticatorId
	subAuthenticatorCount := 0
	for id, initData := range initDatas {
		for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
			if authenticatorCode.Type() == initData.AuthenticatorType {
				err := authenticatorCode.OnAuthenticatorAdded(ctx, account, initData.Data, compositeId(baseId, id))

				if err != nil {
					return err
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
				err := authenticatorCode.OnAuthenticatorRemoved(ctx, account, initData.Data, compositeId(baseId, id))
				if err != nil {
					return err
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
