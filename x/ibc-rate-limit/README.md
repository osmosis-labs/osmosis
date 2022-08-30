# # IBC Rate Limit

The ``IBC Rate Limit`` middleware implements an [IBC Middleware](https://github.com/cosmos/ibc-go/blob/f57170b1d4dd202a3c6c1c61dcf302b6a9546405/docs/ibc/middleware/develop.md) 
that wraps a [transfer](https://ibc.cosmos.network/main/apps/transfer/overview.html) app to regulate how much value can
flow in and out of the chain for a specific denom and channel.

## Contents

1. **[Concepts](#concepts)**
2. **[Parameters](#parameters)**
3. **[Contract](#contract)**
4. **[Integration](#integration)**

## Concepts

### Overview

The `x/ibc-rate-limit` module implements an IBC middleware and a transfer app wrapper. The middleware checks if the 
amount of value of a specific denom transferred through a channel has exceeded a quota defined by governance for 
that channel/denom. These checks are handled through a CosmWasm contract. The contract to be used for this is 
configured via a parameter.

### Middleware

To achieve this, the middleware  needs to implement  the `porttypes.Middleware` interface and the
`porttypes.ICS4Wrapper` interface. This allows the middleware to send and receive IBC messages by wrapping 
any IBC module, and be used as an ICS4 wrapper by a transfer module (for sending packets or writing acknowledgements).

Of those interfaces, just the following methods have custom logic:

 * `ICS4Wrapper.SendPacket` adds tracking of value sent via an ibc channel 
 * `Middleware.OnRecvPacket` adds tracking of value received via an ibc channel 
 * `Middleware.OnAcknowledgementPacket` undos the tracking of a sent packet if the acknowledgment is not a success
 * `OnTimeoutPacket` undos the tracking of a sent packet if the packet times out (is not relayed)

All other methods from those interfaces are passthroughs to the underlying implementations.

### Contract

The tracking contract uses the following concepts

1. **RateLimit** - tracks the value flow transferred and the quota for a path.
2. **Path** - is a (denom, channel) pair.
3. **Flow** - tracks the value that has moved through a path during the current time window.
4. **Quota** - is the percentage of the denom's total value that can be transferred through the path in a given period of time (duration)

## Parameters

The middleware uses the following parameters:

| Key             | Type   |
|-----------------|--------|
| ContractAddress | string |

1. **ContractAddress** -
   The contract address is the address of an instantiated version of the contract provided under `./contracts/`

## Contract

### Messages
The contract specifies the following messages:

#### Query
 * GetQuotas - Returns the quotas for a path

#### Exec
 * AddPath - Adds a list of quotas for a path
 * RemovePath - Removes a path
 * ResetPathQuota - If a rate limit has been reached, the contract's governance address can reset the quota so that transfers are allowed again

#### Sudo

Sudo messages can only be executed by the chain.

 * SendPacket - Increments the amount used out of the send quota and checks that the send is allowed. If it isn't, it will return a RateLimitExceeded error
 * RecvPacket - Increments the amount used out of the receive quota and checks that the receive is allowed. If it isn't, it will return a RateLimitExceeded error
 * UndoSend - If a send has failed, the undo message is used to remove its cost from the send quota

## Integration

The rate limit middleware wraps the `transferIBCModule` and is added as the entry route for IBC transfers.

The module is also provided to the underlying `transferIBCModule` as its `ICS4Wrapper`; previously, this would have 
pointed to a channel, which also implements the `ICS4Wrapper` interface.

This integration can be seen in [osmosis/app/keepers/keepers.go](https://github.com/osmosis-labs/osmosis/blob/main/app/keepers/keepers.go)

