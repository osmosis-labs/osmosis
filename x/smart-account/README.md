# x/smart-account Module

## General Explanation

The `x/smart-account` module provides a robust and extensible framework for authenticating transactions.

Unlike traditional authentication methods, this module allows you to use multiple types of authenticators,
each with their own set of rules and conditions for transaction approval.

## Architecture Overview

### Circuit Breaker

The module is designed to be used as a replacement for the default Cosmos SDK authentication mechanism. This is 
configured as an ante handler (executed before the transaction messages are processed). For safety, we have included
a circuit breaker that allows the module to be disabled if necessary. Once the module is enabled, the user needs to 
opt-in into using authenticators by selecting the authenticators it wants to use for each message. This is specified
in the `selected_authenticators` field of the transaction extension. If selected_authenticators are not provided, the
transaction defaults to the classic Cosmos SDK authentication method.

The flow is as follows:

![Circuit Breaker](/x/smart-account/images/circuit_breaker.jpg)

### Authenticator Flow

After passing the circuit breaker, if the transaction uses authenticators, the flow becomes as follows:

The authenticator ante handler iterates over each message in the transaction. For each message, the following steps occur:

 * The message signer is selected as the "account" for this message. The authenticator for that account is selected based on the selected_authenticator provided in the tx
    * Validation occurs to ensure that the selected authenticator is valid and that the account has the authenticator.
 * The selected authenticator attempts to authenticate the message by calling its Authenticate() function. 
   * If authentication fails, the process stops and the transaction is rejected. No changes to state are made.
   * If authentication succeeds, closure is generated for the message.
     * This closure remembers which authenticator was used and will be called later if the whole tx (all messages) are authenticated. 
 * Fees for the transaction are collected

 After all messages are authenticated successfully:

 * The Call Track() on all messages step is executed, notifying the authenticators involved.
 * If all track calls finish successfully, the changes are written to the data store.

The process then executes all the authenticated messages. If the transaction fails at this point, the execution 
changes are discarded. Please note that the authenticator changes (committed in track) are not reverted!

If the execution is successful, we continue in the post handler:

 * For each message, an account and authenticator are selected again.
 * The ConfirmExecution() function is called on the selected authenticator, allowing it to enforce post-execution rules.
   * If ConfirmExecution() succeeds for all authenticators, the changes are written to the data store.
   * If ConfirmExecution() fails for any authenticator, or if the "Execute All Messages" step fails, the changes are discarded.

![Authenticator Flow](/x/smart-account/images/authentication_flow.jpg)

### Authenticator Implementations

The implementation of each authenticator type is done by a Go struct that implements the `Authenticator` interface. 
This interface defines the functions that need to be implemented and will be described in detail in the next section. 

For authenticators to be available, they need to be registered with the `AuthenticatorManager`. This manager is 
responsible for retrieving authenticators by their unique type.

![Authenticator Implementations](/x/smart-account/images/authenticator_manager.jpg)

Since implementations are custom code, they can encode complex authentication logic like calling each other, or
calling cosmwasm contracts to authenticate the messages.

### Authenticator configuration for accounts

Accounts have the flexibility to be linked with multiple authenticators, a setup maintained in the system's storage 
and managed by the module's Keeper. The keeper is responsible for adding and removing 
authenticators, as well as storing any user data that the authenticators may need. 

This is where the association of specific authenticators with accounts is stored. 

![Account Authenticator Configuration](/x/smart-account/images/keeper.jpg)

One way of seeing this data is as the instantiation information necessary to use the authenticator for a specific 
account. For example, a `SignatureVerification` contains the code necessary to verify a signature, but
it needs to know which public key to use when verifying it. An account can configure the 
`SignatureVerification` to be one of their authenticators and would need to provide the public key it wants 
to use for verification in the configuration data.


To make an authenticator work for a specific account, you just need to feed it the right information. For example, 
the `SignatureVerification` needs to know which public key to check when verifying a signature. 
So, if you're setting this up for your account, you have to configure it with the public key you want as part of the 
account-authenticator link.

This is done by using the `MsgAddAuthenticator` message, which is covered in detail in a later section. When 
authenticators are added to accounts, they should validate that the necessary data is available and correct in their
`OnAuthenticatorAdded` function.


## Authenticator Interface

The `Authenticator` interface is the cornerstone of this module, encapsulating the essential functionalities
for transaction authentication.

Here is a look at the Go code defining the interface:

```go
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
```

### Methods

#### `Type`

Returns the type of the authenticator (e.g., `SignatureVerification`, `CosmWasmAuthenticator`).
Each type must be registered within the `AuthenticatorManager`.

#### `StaticGas`

Provides the fixed gas amount that is consumed with each invocation of this authenticator

#### `Initialize`

Initializes the authenticator when retrieved from storage. It takes the stored config (e.g., PublicKey for signature 
verification) as an argument to set up the authenticator.

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

The authenticator manager allows the chain to register authenticators and retrieve them by type. Each authenticator 
type represents the code to be executed.

To determine which authenticators will be used for each account, this module's keeper stores a mapping
between an account and a list of authenticators. A user can add or remove authenticators from their account using the
`MsgAddAuthenticator` and `MsgRemoveAuthenticator` messages.

Some authenticators may require additional config specific to the user being authenticated. To handle this, the user
can store config data for each of their authenticators.

### Messages

#### `MsgAddAuthenticator`

Adds an authenticator to the user's account. The authenticator must be registered with the authenticator manager.

Example:

```go
AddAuthenticator(account, "SignatureVerification", pubKeyBytes)
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

### SignatureVerification Authenticator

The signature verification authenticator is the default authenticator for all accounts. It verifies that the signer of a message is the same as the account associated with the message.

### AnyOf Authenticator

The anyOf authenticator allows you to specify a list of authenticators. If any of the authenticators in the list successfully authenticate a message, the message is authenticated.

### AllOf Authenticator

The allOf authenticator allows you to specify a list of authenticators. All authenticators in the list must successfully authenticate a message for the message to be authenticated.

### MessageFilter Authenticator

The message filter authenticator allows you to match the incoming message against a message pattern specified in the
authenticator configuration for the user. If the message matches the pattern, the message is authenticated. Otherwise, 
the message is rejected.

#### Patterns

The message filter patterns are specified as a json object with a `@type` field for the type of the message to match and
a serialization of the remaining fields of the message. For example, to match a `MsgSend` message with a specific denom
and recipient, the pattern would look like this:

```json
{
  "@type": "/cosmos.bank.v1beta1.MsgSend",
  "amount": [
    {
      "denom": "uatom"
    }
  ]
}
```

Note that there are other fields in MsgSend that are not specified in the pattern. This is because the pattern only
matches the fields that are specified; all other fields are ignored.

Similarly, to match a `MsgSwapExactIn` message with a specific sender and token, the pattern would look like this:

```json
{
   "@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
   "sender":"osmo1...", 
   "token_in":{
      "denom":"inputDenom"
   }
}
```

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

The contract needs to implement 5 authenticator hooks which are called by the authenticator through sudo entrypoint, it needs to handle the following messages:

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

- fetch the authenticators for the account
- select the authenticator to use for this message
- authenticate the message with the selected authenticator

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

## Authenticator Ids

Each authenticator associated with an account has a unique id that is used to identify it. This id is an incrementing
number that is globally unique. For example, if user1 and user2 add authenticators to their accounts, these will have
ids 1 and 2 respectively. If user1 adds a second authenticator, it will have id 3.

## Composite authenticators

Composite authenticators are authenticators that can use other "sub-authenticators" to authenticate a message.

There are two composite authenticators implemented in this module: `AnyOf` and `AllOf` authenticators. These take a
list of authenticators and authenticate a message if any or all of the authenticators in the list authenticate the
message. A similar logic applies to track and confirm.

### Composite Ids

When a composite authenticator calls a sub authenticator, it is its responsibility to update the authenticator id
passed down to the callee so that it can identify itself. This is done by combining the id of the top level
authenticator with the position of the sub-authenticator in the list. For example, if a user has
`AllOf(AnyOf(sig1, sig2), CosmwasmAuthenticator(contract1, params))` as it's authenticator, and it has a
global id of 86, the cosmwasm authenticator will receive `86.1` as its authenticator id. If that instead was another
composite authenticator, its first sub-authenticator would receive `86.1.0` as its authenticator id.

In the general case, if we call a composite authenticator with id `a` and it has `n` sub-authenticators, the sub-authenticator
at position `i` will receive the id `a.i`. If the sub-authenticator is itself a composite authenticator with `m` sub-authenticators,
the sub-authenticator at position `j` will receive the id `a.i.j`.

### Confirm Execution call order

The call logic on confirm execution behaves the same way as the call logic on authenticate and is stateless, i.e.: it 
does not have any information about which authenticators were called during the authenticate call. 

This may lead to a situation where two authenticators inside an AnyOf (or a more complex composition logic) get called
in a way that makes it so that after the transaction only one of the methods have been called on each.

For example, if the user's authenticators are `AnyOf(A,B)`, you could have a case where Authenticate fails on A and 
succeeds on B, whereas ConfirmExecution succeeds on A. There will then never be a call to ConfirmExecution on B.

This is the expected behaviour and authenticator authors should be aware that *there is no guarantee that both methods
will be called on their authenticator*. There is a proposal to improve this in the future tracked in 
[#8373](https://github.com/osmosis-labs/osmosis/issues/8373)

### Composition Logic

#### AllOf

When using an AllOf authenticator `AllOf(a,b,...)`, the msg will be authenticated iff `authenticate(a) && authenticate(b) && ...` is true.

The track function will be called for all the sub-authenticators. Any state changes made in track will be committed if
authentication and track succeeds, regardless of the outcome of the message execution.

If the message execution succeeds, the confirm function will be called for all the sub-authenticators.

The message will succeed and its state changes be committed iff `confirm(a) && confirm(b) && ...` is true. Otherwise,
the all state changes will be reverted except the ones made in track, which are always committed.

#### AnyOf

When using an AnyOf authenticator `AnyOf(a,b,...)`, the msg will be authenticated iff `authenticate(a) || authenticate(b) || ...` is true.

In this case, because we are not tracking which sub-authenticators authenticated the message, the track function must be
called on _all_ the sub-authenticators. Any state changes made in track will be committed if authentication and track
succeeds, regardless of the outcome of the message execution.

Similarly, when calling confirm on an AnyOf authenticator, we call confirm on _all_ the sub-authenticators.

The message will succeed and its state changes be committed iff `confirm(a) || confirm(b) || ...` is true. Otherwise,
the all state changes will be reverted except the ones made in track, which are always committed.

### Selecting sub-authenticators

At the moment, we do not support selecting sub-authenticators when submitting a tx. This would be useful when dealing
with an `AnyOf` so that a user can specify which sub-authenticator to select. This would also allow us to avoid calling
all sub-authenticators when during the track and confirm steps. This is a feature that could be added in the future.

### Composite signatures

There are two ways in which we may want to provide signatures to a composite authenticator:

- Simple: the same signature is passed to all sub-authenticators
- Partitioned: the signatures are encoded as a json array, and each sub-authenticator is given its own part of the signature

As an example, consider a simple multi-sig implemented through composite authenticators. In this case, a user would assign
an authenticator that looks like `PartitionedAllOf(pubkey1, pubkey2)` to their account. If this authenticator were a
simple `AnyOf` there would be no way of providing a signature that satisfies both pubkeys. Here, however, we can
provide each signature encoded in a json array.

The user provides the bytes `"[sig1, sig2]"` as their signature. The `PartitionedAllOf` authenticator will then decode
this json array and pass `sig1` to the first sub-authenticator and `sig2` to the second sub-authenticator.

# Constructing composite authenticators

As a user, you may want to combine composite authenticators to create more complex authentication logic. Here are some
examples of how to do that:

## One Click trading

A hot key can be configured so that it is only allowed to execute swap messages, and fail if the transaction leaves the
user with a lower balance than a certain threshold. This can be done by using the following authenticator:

`AllOf(SignatureVerification(usersPubKey), AnyOf(MessageFilter(SwapMsg1), MessageFilter(SwapMsg2)), CosmwasmAuthenticator(spendLimitContract, params))`

## Multisig

A simple multisig design can be done by using a `PartitionedAllOf` authenticator. This authenticator will take a list of
pubkeys and authenticate the message if signatures for the pubkeys are present in the signature.

`PartitionedAllOf(pubkey1, pubkey2, pubkey3)`

## Cosigner

An off chain cosigner can be configured to be required for all messages on an account. This way, an api (cosigner service)
can analyze and simulate the transactions for security and provide their signature iff the transaction is safe.
This could be achieved by using the following authenticator:

`AllOf(SignatureVerification(cosignerPubKey), AnyOf(...)` where the rest of the user's authenticators are
under the `AnyOf`.

For a more complex cosigner, the user could use a `CosmwasmAuthenticator` that calls a contract that implements the
cosigner logic. This would allow some manager to rotate the cosigner key or implement some recovery logic in case the
cosigner is unavailable or misbehaving.

## Succession/Inheritance protocol

A user can configure an inheritance authenticator so that a beneficiary can take over their account if the account has
been inactive for a certain period of time. This can be done by using the following authenticator:

`AnyOf(CosmwasmAuthenticator(inheritanceContract, params), ...other_authenticators)`

# Authentication Lifecycle examples

- Which hook runs and when
- State commit / reversion conditions

# Circuit Breaker

For the initial release, this module will be provided behind a "circuit breaker" or feature switch. This means that
the feature will be controlled by the `is_smart_account_active` parameter. If that parameter is set to false,
the authenticator module will not be used and the classic cosmos sdk authentication will be used instead.

# Using authenticators from JS

TODO: Add examples of how to use authenticators from JS
