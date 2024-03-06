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
// Authenticator is an interface that encapsulates all authentication functionalities essential for
// verifying transactions, paying transaction fees, and managing gas consumption during verification.
type Authenticator interface {
    // Type returns the specific type of the authenticator, such as SignatureVerificationAuthenticator.
    // This type is used for registering and identifying the authenticator within the AuthenticatorManager.
    Type() string
    
    // StaticGas provides the fixed gas amount consumed for each invocation of this authenticator.
    // This is used for managing gas consumption during transaction verification.
    StaticGas() uint64
    
    // Initialize prepares the authenticator with necessary data from storage, specific to an account-authenticator pair.
    // This method is used for setting up the authenticator with data like a PublicKey for signature verification.
    Initialize(data []byte) (Authenticator, error)
    
    // Authenticate confirms the validity of a message using the provided authentication data.
    // NOTE: Any state changes made by this function will be discarded.
    // It's a core function within an ante handler to ensure message authenticity and enforce gas consumption.
    Authenticate(ctx sdk.Context, request AuthenticationRequest) error
    
    // Track allows the authenticator to record information, regardless of the transaction's authentication method.
    // NOTE: Any state changes made by this function will be written to the store as long as Authenticate succeeds and will not be reverted if the message execution fails.
    // This function is used for the authenticator to acknowledge the execution of specific messages by an account.
    Track(ctx sdk.Context, account sdk.AccAddress, feePayer sdk.AccAddress, msg sdk.Msg, msgIndex uint64, authenticatorId string) error
    
    // ConfirmExecution enforces transaction rules post-transaction, like spending and transaction limits.
    // It is used to verify execution-specific state and values, to allow authentication to be dependent on the effects of a transaction.
    ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error
    
    // OnAuthenticatorAdded handles the addition of an authenticator to an account.
    // It checks the data format and compatibility, to maintain account security and authenticator integrity.
    OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error
    
    // OnAuthenticatorRemoved manages the removal of an authenticator from an account.
    // This function is used for updating global data or preventing removal when necessary to maintain system stability.
    OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error
}
```

### Methods

#### `Type`

Returns the type of the authenticator (e.g., `SignatureVerificationAuthenticator`, `CosmWasmAuthenticator`). 
Each type must be registered within the `AuthenticatorManager`.

#### `StaticGas`

Provides the fixed gas amount that is consumed with each invocation of this authenticator

#### `Initialize`

Initializes the authenticator when retrieved from storage. It takes stored data (e.g., PublicKey for signature verification) as 
an argument to set up the authenticator.

#### `Authenticate`
Validates a message based on the provided transaction data (encoded as an `AuthenticationRequest`). 
Returns true if authenticated, false otherwise. 

Note that any state changes made by this function are intended to be discarded. The role of this function is 
transaction verification without affecting permanent state.

#### `Track`

Once authentication succeeds, the `Track` function is called on every authenticator that was involved in authenticating
the messages.


#### `ConfirmExecution`

Used in post-handler functions to enforce transaction rules like spending and transaction limits.

## Authenticator Manager

The authenticator manager the chain to register authenticators and retrieve them by type. Each authenticator type
represents the code to be executed.

To determine which authenticators will be used for each account, this module's keeper stores a mapping
between an account and a list of authenticators. A user can add or remove authenticators from their account using the
`MsgAddAuthenticator` and `MsgRemoveAuthenticator` messages.

Some authenticators may require additional data specific to the user being authenticated. To handle this, the user
can store data for each of their authenticators.

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

2. **Identify Fee Payer**: The first signer of the transaction is considered the fee payer.

3. **Authenticate Each Message**:

   - The associated account for every message is identified.
   - The system fetches the appropriate authenticators for that account.
   - The selected authenticator is retrieved from the transaction and used to determine which aithenticator to execute 
   - The authenticator tries to validate the message. Successful validation means the message is authenticated and execution can proceed.

4. **Gas Limit Reset**: Once the fee payer is authenticated, the gas limit is restored to its original value.

5. **Track Authenticated Messages**: After all messages are authenticated, the `Track` function notifies each executed authenticator . This allows authenticators to store any transaction-specific data they might need for the future.

6. **Execute Messages**: The transaction is executed.

7. **Confirm Execution**: After all messages are executed, the `ConfirmExecution` function is called for each of the authenticators that authenticated the tx. This allows authenticators to enforce rules that depend on the outcome of the message execution, like spending and transaction limits.

## Available Authenticator Types

### Signature Verification Authenticator

The signature verification authenticator is the default authenticator for all accounts. It verifies that the signer of a message is the same as the account associated with the message.

### AnyOf Authenticator

The anyOf authenticator allows you to specify a list of authenticators. If any of the authenticators in the list successfully authenticate a message, the message is authenticated.

### AllOf Authenticator

The allOf authenticator allows you to specify a list of authenticators. All authenticators in the list must successfully authenticate a message for the message to be authenticated.

## CosmWasm Authenticator

The CosmWasm Authenticator allows for the building of any custom authentication logic as a CosmWasm contract.
The contract needs to be instantiated before being added as an authenticator, since the contract address is required for `MsgAddAuthenticator`.

`MsgAddAuthenticator` arguments:

```text
sender: <bech32_address>
type: "CosmwasmAuthenticatorV1"
data: json_bytes({
    contract: "<contract_address>",
    params: [<byte_array>]
})
```

The `params` field allows users to configure any user-specific information that the authenticator contract may need so that shared contracts don't 
need to internally store data for each user. This allows for contract reuse even when user specific information is needed.
This field is standardized and must be json bytes. This way it can be easily parsed and displayed by clients, which makes them more human-friendly.

Contract storage should be used only when the authenticator needs to track any dynamic information required.

### Contract Interface

The contract needs to implement 3 authenticator hooks which are called by the authenticator through sudo entrypoint, it needs to handle the following messages:

```rs
#[cw_serde]
pub enum AuthenticatorSudoMsg {
    OnAuthenticatorAdded(OnAuthenticatorAddedRequest),
    OnAuthenticatorRemoved(OnAuthenticatorRemovedRequest),
    Authenticate(AuthenticationRequest),
    Track(TrackRequest),
    ConfirmExecution(ConfirmExecutionRequest),
}
```

The last three messages corresponds to steps 3, 5 and 7 of the [transaction authentication process](#transaction-authentication-overview) and the first
two messages are used to handle the addition and removal of the authenticator.

Request types are defined [here](https://docs.rs/osmosis-authenticators/latest/osmosis_authenticators).

## Queries

TODO: Add examples of queries and how to read them

-- 
# Design Decisions

## Authenticator Selection

### Different authenticators for each message

The classic cosmos sdk authentication authenticates the tx as a whole by verifying the signatures of each signer. This 
can be done because authentication logic is the same for all messages.  
This module allows for more fine-grained control over the authentication process, which means users can
configure how messages for their account are authenticated. Because of this, on a multi-message transaction, each
message will go through its own authentication process: 
 * fetch the authenticators for the account
 * select the authenticator to use for this message
 * authenticate the message with the selected authenticator

The authenticators used for each message may be different and require different signature types. 

Once all messages are authenticated, the track function is called on the used authenticators, messages are executed,
and if execution is successful, the confirm execution function is called on the used authenticators.

### Specifying authenticators

When a user submits a transaction, they must specify which authenticators to use for each message. This is done by
including the following TxExtension: 

```go
// TxExtension allows for additional authenticator-specific data in
// transactions.
message TxExtension {
  // selected_authenticators holds the authenticator_id for the chosen 
  // authenticator per message.
  repeated uint64 selected_authenticators = 1;
}
```
If the selected authenticators are not present, or there isn't one for each message, the transaction will be rejected
by this module.

At the moment, there is a backup authentication method that uses the classic Cosmos SDK authentication. If 
selected_authenticators are not specified, we default to the classic authentication method. This way existing
applications don't need to be aware of authenticators to get their txs processed. This is likely a temporary measure

## Introduced transaction restrictions

To simplify the design of the authenticator module, a few restrictions have been set on the type of transactions
that are accepted.

### Messages can only have one signer

On cosmos SDK versions before 0.50 it was possible to have multiple signers for a message. This will no longer be
the case after v0.50, and we have introduced this restriction here as it makes it more clear which account a message 
is associated with.

### The fee payer must be the first signer of the first message

The cosmos SDK allows for the fee payer to be any signer of the transaction but defaults to the first signer of the first
message. This module will enforce this restriction to simplify the gas management and authentication process.

## Fee Payer and Gas Consumption

Fees (that pay for gas consumption) must be paid by the fee payer. These fees are paid regardless of whether the message 
execution succeeds or fails, but cannot be charged (i.e.: the spend committed to state) until the authentication 
succeeds (otherwise an attacker could force fees to be deducted from an account they don't control). This is not a 
problem with classic authentication because validating each signature is a consistently cheap operation 
(signature verification).

In an authenticator model, the cost of authenticating a message can vary greatly depending on the authenticators used.
Moreover, if we enable the use of permissionless wasm contracts as authenticators, an attacker could write a contract
that consumes a lot of gas (i.e.: cpu cycles) during authentication and use that to spam the network without any cost. 

To prevent this type of spam, we need to limit the amount of gas that can be consumed during the authentication before 
the fee payer has been authenticated.

This limit is set to a fixed amount controlled by the param of this module as `maximum_unauthenticated_gas`

## Composite authenticators
- Composite id
- Composition logic and expected behaviour (between authenticate and confirm)
- State commit / reversion conditions for sub authenticators

# Authentication Lifecycle examples 
- Which hook runs and when
- State commit / reversion conditions

# Building composite authenticators

Examples


# Circuit Breaker

# Cosigner Flow

# Using authenticators from JS

TODO: Add examples of how to use authenticators from JS