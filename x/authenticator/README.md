# x/authenticators Module

## General Explanation

The `x/authenticators` module provides a robust and extensible framework for authenticating transactions. 

Unlike traditional authentication methods, this module allows you to use multiple types of authenticators,
each with their own set of rules and conditions for transaction approval.

## Authenticator Interface

The `Authenticator` interface is the cornerstone of this module, encapsulating the essential functionalities 
for transaction authentication.

Here is a look at the Go code defining the interface:

```go
type Authenticator interface {
    Type() string
    StaticGas() uint64
    Initialize(data []byte) (Authenticator, error)
    GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (AuthenticatorData, error)
    Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error
    Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) AuthenticationResult
    ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) ConfirmationResult
}
```
### Methods

#### `Type`
Returns the type of the authenticator (e.g., `SignatureVerificationAuthenticator`, `CosmWasmAuthenticator`). Each type must be registered within the `AuthenticatorManager`.

#### `StaticGas`
Defines the static gas consumption for each call to the authenticator.

#### `Initialize`
Initializes the authenticator when retrieved from storage. It takes stored data (e.g., PublicKey for signature verification) as an argument to set up the authenticator.

#### `GetAuthenticationData`
Retrieves required authentication data from a transaction. Used in ante handlers to ensure the user has the correct permissions to execute a message.

#### `Track`
Tracks any information the authenticator may need, regardless of how the transaction is authenticated (e.g., via authz, ICA).

#### `Authenticate`
Validates a message based on the signer and data. Returns true if authenticated, false otherwise. Consumes gas within this function.

#### `ConfirmExecution`
Used in post-handler functions to enforce transaction rules like spending and transaction limits.

## Authenticator Manager

The authenticator manager the chain to register authenticators and retrieve them by type. Each authenticator type
represents the code to be executed. 

To determine which authenticators will be used for each account, the authenticator manager also stores a mapping 
between an account and a list of authenticators. A user can add or remove authenticators from their account using the
`MsgAddAuthenticator` and `MsgRemoveAuthenticator` messages.

Some authenticators may require additional data specific to the user being authenticated. To handle this, the user
can store data for each authenticator type. 

### Messages

#### `MsgAddAuthenticator`
Adds an authenticator to the user's account. The authenticator must be registered with the authenticator manager.

Example:

```go
AddAuthenticator(account, "SignatureVerificationAuthenticator", pubKeyBytes) 
```

#### `MsgRemoveAuthenticator`
Removes an authenticator from the user's account. The authenticator must be registered with the authenticator manager.

example:

```go
RemoveAuthenticator(account, authenticatorGlobalId)
``` 

## Transaction Authentication Overview

1. **Initial Gas Limit**: A temporary gas limit is set for fee payer authentication. This is a spam prevention measure to safeguard computational resources.

2. **Identify Fee Payer**: By default, the first signer of the transaction is considered the fee payer.

3. **Authenticate Each Message**:
    - The associated account for every message is identified.
    - The system fetches the appropriate authenticators for that account.
    - Each authenticator tries to validate the message. Successful validation means the message is authenticated.

4. **Gas Limit Reset**: Once the fee payer is authenticated, the gas limit is restored to its original value.

5. **Track Authenticated Messages**: After all messages are authenticated, the `Track` function notifies each authenticator. This allows authenticators to store any transaction-specific data they might need for future reference, irrespective of the authentication method used for the transaction.

6. **Execute Messages**: The transaction is executed.

7. **Confirm Execution**: After all messages are authenticated, the `ConfirmExecution` function is called for each authenticator. This allows authenticators to enforce transaction rules like spending and transaction limits.

## Available Authenticator Types

### Signature Verification Authenticator

The signature verification authenticator is the default authenticator for all accounts. It verifies that the signer of a message is the same as the account associated with the message.

### PassKey Authenticator

The passkey authenticator is the authenticator for Passkeys. It verifies that the signer of a message is in the authentication store and also a secp256r1 key.

### Spend Limit Authenticator

The spend limit authenticator enforces a spend limit for each account. 

### AnyOf Authenticator

The anyOf authenticator allows you to specify a list of authenticators. If any of the authenticators in the list successfully authenticate a message, the message is authenticated.

### AllOf Authenticator

The allOf authenticator allows you to specify a list of authenticators. All authenticators in the list must successfully authenticate a message for the message to be authenticated.

