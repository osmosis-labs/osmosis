package utils

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
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
	GetAuthenticatorsForAccountOrDefault(ctx sdk.Context, account sdk.AccAddress) ([]types.InitializedAuthenticator, error)
}
