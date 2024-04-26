package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitializedAuthenticator denotes an authenticator fetched from the store and prepared for use.
type InitializedAuthenticator struct {
	Id            uint64
	Authenticator Authenticator
}

// Authenticator is an interface that encapsulates all authentication functionalities essential for
// verifying transactions, paying transaction fees, and managing gas consumption during verification.
type Authenticator interface {
	// Type returns the specific type of the authenticator, such as SignatureVerification.
	// This type is used for registering and identifying the authenticator within the AuthenticatorManager.
	Type() string

	// StaticGas provides the fixed gas amount consumed for each invocation of this authenticator.
	// This is used for managing gas consumption during transaction verification.
	StaticGas() uint64

	// Initialize prepares the authenticator with necessary data from storage, specific to an account-authenticator pair.
	// This method is used for setting up the authenticator with data like a PublicKey for signature verification.
	Initialize(config []byte) (Authenticator, error)

	// Authenticate confirms the validity of a message using the provided authentication data.
	// NOTE: Any state changes made by this function will be discarded.
	// It's a core function within an ante handler to ensure message authenticity and enforce gas consumption.
	Authenticate(ctx sdk.Context, request AuthenticationRequest) error

	// Track allows the authenticator to record information, regardless of the transaction's authentication method.
	// NOTE: Any state changes made by this function will be written to the store as long as Authenticate succeeds and will not be reverted if the message execution fails.
	// This function is used for the authenticator to acknowledge the execution of specific messages by an account.
	Track(ctx sdk.Context, request AuthenticationRequest) error

	// ConfirmExecution enforces transaction rules post-transaction, like spending and transaction limits.
	// It is used to verify execution-specific state and values, to allow authentication to be dependent on the effects of a transaction.
	ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error

	// OnAuthenticatorAdded handles the addition of an authenticator to an account.
	// It checks the data format and compatibility, to maintain account security and authenticator integrity.
	OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error

	// OnAuthenticatorRemoved manages the removal of an authenticator from an account.
	// This function is used for updating global data or preventing removal when necessary to maintain system stability.
	OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error
}
