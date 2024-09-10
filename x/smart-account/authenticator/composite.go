package authenticator

import (
	"encoding/json"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type SubAuthenticatorInitData struct {
	Type   string `json:"type"`
	Config []byte `json:"config"`
}

func subTrack(
	ctx sdk.Context,
	request AuthenticationRequest,
	subAuthenticators []Authenticator,
) error {
	baseId := request.AuthenticatorId
	for id, auth := range subAuthenticators {
		request.AuthenticatorId = compositeId(baseId, id)
		err := auth.Track(ctx, request)
		if err != nil {
			return errorsmod.Wrapf(err, "sub-authenticator track failed (sub-authenticator id = %s)", request.AuthenticatorId)
		}
	}
	return nil
}

func splitSignatures(signature []byte, total int) ([][]byte, error) {
	var signatures [][]byte
	err := json.Unmarshal(signature, &signatures)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to parse signatures")
	}
	if len(signatures) != total {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid number of signatures")
	}
	return signatures, nil
}

func onSubAuthenticatorsAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string, am *AuthenticatorManager) error {
	var initDatas []SubAuthenticatorInitData
	if err := json.Unmarshal(data, &initDatas); err != nil {
		return errorsmod.Wrapf(err, "failed to unmarshal sub-authenticator init data")
	}

	if len(initDatas) <= 1 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least 2 sub-authenticators must be provided, but got %d", len(initDatas))
	}

	baseId := authenticatorId
	subAuthenticatorCount := 0
	for id, initData := range initDatas {
		authenticatorCode := am.GetAuthenticatorByType(initData.Type)
		if authenticatorCode == nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "sub-authenticator failed to be added in function `OnAuthenticatorAdded` as type is not registered in manager")
		}
		subId := compositeId(baseId, id)
		err := authenticatorCode.OnAuthenticatorAdded(ctx, account, initData.Config, subId)
		if err != nil {
			return errorsmod.Wrapf(err, "sub-authenticator `OnAuthenticatorAdded` failed (sub-authenticator id = %s)", subId)
		}

		subAuthenticatorCount++
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
		authenticatorCode := am.GetAuthenticatorByType(initData.Type)
		if authenticatorCode == nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "sub-authenticator failed to be removed in function `OnAuthenticatorRemoved` as type is not registered in manager")
		}
		subId := compositeId(baseId, id)
		err := authenticatorCode.OnAuthenticatorRemoved(ctx, account, initData.Config, subId)
		if err != nil {
			return errorsmod.Wrapf(err, "sub-authenticator `OnAuthenticatorRemoved` failed (sub-authenticator id = %s)", subId)
		}
	}
	return nil
}

func compositeId(baseId string, subId int) string {
	return baseId + "." + strconv.Itoa(subId)
}
