package utils

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

// GetAccount retrieves the account associated with the first signer of a transaction message.
// It returns the account's address or an error if no signers are present.
func GetAccount(msg sdk.Msg) (sdk.AccAddress, error) {
	if len(msg.GetSigners()) != 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "messages must have exactly one signer")
	}
	return msg.GetSigners()[0], nil
}

// AuthenticatorStorage is an interface abstracting the only method from the keeper that we care about
type AuthenticatorStorage interface {
	GetAuthenticatorsForAccountOrDefault(ctx sdk.Context, account sdk.AccAddress) ([]int64, []types.Authenticator, error)
}

// ConfirmExecutionWithoutTx is a utility for msg executors that bypass the tx flow (i.e.: authz, ica)
// If the account's authenticators depend on the authenticator data, this will fail and execution will be blocked
func ConfirmExecutionWithoutTx(ctx sdk.Context, authStorage AuthenticatorStorage, msgs []sdk.Msg) error {
	for _, msg := range msgs {
		account, err := GetAccount(msg)
		if err != nil {
			return err
		}

		request := types.AuthenticationRequest{
			Account: account,
			// TODO: build this (we may want to split some of the generation of the request struct into helpers)
			//Msg:     msg,
		}

		_, allAuthenticators, err := authStorage.GetAuthenticatorsForAccountOrDefault(ctx, account)
		if err != nil {
			return err
		}
		for _, authenticator := range allAuthenticators {
			// Confirm Execution
			success := authenticator.ConfirmExecution(ctx, request)

			if success.IsBlock() {
				return errorsmod.Wrap(success.Error(), "authenticator failed to confirm execution without AuthenticationData")
			}
		}
	}
	return nil
}
