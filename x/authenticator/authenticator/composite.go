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
