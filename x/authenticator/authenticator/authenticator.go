package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	SignatureVerificationAuthenticatorType = "SignatureVerificationAuthenticator"
)

type AuthenticatorData interface{}

type Authenticator interface {
	Gas() uint64

	Type() string

	Initialize(data []byte) (Authenticator, error)

	GetAuthenticationData(
		ctx sdk.Context,
		tx sdk.Tx,
		messageIndex uint8,
		simulate bool,
	) (AuthenticatorData, error)

	Authenticate(
		ctx sdk.Context,
		msg sdk.Msg,
		authenticationData AuthenticatorData,
	) (bool, error)

	ConfirmExecution(
		ctx sdk.Context,
		msg sdk.Msg,
		authenticated bool,
		authenticationData AuthenticatorData,
	) bool

	// Optional Hooks. ToDo: Revisit this when adding the authenticator storage and messages
	// OnAuthenticatorAdded(...) bool
	// OnAuthenticatorRemoved(...) bool
}
