package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// SignatureVerificationAuthenticatorType represents a type of authenticator specifically designed for
	// secp256k1 signature verification.
	SignatureVerificationAuthenticatorType = "SignatureVerificationAuthenticator"
)

// AuthenticatorData represents the data required for verifying a signer's address and message signature.
type AuthenticatorData interface{}

// Authenticator is an interface employed to encapsulate all authentication functionalities essential for
// verifying transactions, paying transaction fees, and managing gas consumption during verification.
type Authenticator interface {
	// Type() defines the various types of authenticators, such as SignatureVerificationAuthenticator
	// or CosmWasmAuthenticator. Each authenticator type must be registered within the AuthenticatorManager,
	// and these types are used to link the data structure with the authenticator's logic.
	Type() string

	// StaticGas() specifies the gas consumption per signature verification by the authenticator.
	StaticGas() uint64

	// Initialize is used when an authenticator associated with an account is retrieved
	// from storage. The data stored for each (account, authenticator) pair is provided
	// to this method. For instance, the SignatureVerificationAuthenticator requires a PublicKey
	// for signature verification, but the auth module is unaware of the public key. By storing
	// the public key alongside the authenticator, we can Initialize() the code with that data,
	// allowing the same authenticator code to verify signatures for different public keys.
	Initialize(data []byte) (Authenticator, error)

	// GetAuthenticationData retrieves any required authentication data from a transaction.
	// It returns an interface defined as a concrete type by the implementer of the interface.
	// This is used within an ante handler in conjunction with Authenticate to ensure that
	// the user possesses the correct permissions to execute a message.
	GetAuthenticationData(
		ctx sdk.Context, // The SDK Context is utilized to access data associated with authentication data.
		tx sdk.Tx, // The transaction is passed to the getter function to parse signatures and signers.
		messageIndex int8, // The message index is used to extract specific signers and signatures from the authentication data.
		simulate bool, // Simulate is used to perform transaction simulation.
	) (AuthenticatorData, error)

	// Authenticate validates a message based on the signer and data parsed from the GetAuthenticationData function.
	// It returns true if authenticated, or false if not authenticated. This function is used within an ante handler.
	// Note: Gas consumption occurs within this function.
	Authenticate(
		ctx sdk.Context, // The SDK Context is used to access data for authentication and to consume gas.
		account sdk.AccAddress, // The account being authenticated (typically msg.GetSigners()[0]).
		msg sdk.Msg, // A message is passed into the authenticate function, allowing authenticators to utilize its information.
		authenticationData AuthenticatorData, // The authentication data is used to authenticate a message.
	) AuthenticationResult

	// ConfirmExecution is employed in the post-handler function to enforce transaction rules,
	// such as spending and transaction limits. It accesses the account's owned state to store
	// and verify these values.
	ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) ConfirmationResult

	// Optional Hooks (TODO: Revisit this section)
	// OnAuthenticatorAdded(...) bool
	// OnAuthenticatorRemoved(...) bool
}
