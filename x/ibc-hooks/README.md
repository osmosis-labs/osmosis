# IBC-hooks

## Wasm Hooks

The wasm hook is an IBC middleware which is used to allow ICS-20 token transfers to initiate contract calls.
This allows cross-chain contract calls, that involve token movement. 
This is useful for a variety of usecases.
One of primary importance is cross-chain swaps, which is an extremely powerful primitive.

The mechanism enabling this is a `memo` field on every ICS20 transfer packet as of [IBC v3.4.0](https://medium.com/the-interchain-foundation/moving-beyond-simple-token-transfers-d42b2b1dc29b).
Wasm hooks is an IBC middleware that parses an ICS20 transfer, and if the `memo` field is of a particular form, executes a wasm contract call. We now detail the `memo` format for `wasm` contract calls, and the execution guarantees provided.

### Cosmwasm Contract Execution Format

Before we dive into the IBC metadata format, we show the cosmwasm execute message format, so the reader has a sense of what are the fields we need to be setting in.
The cosmwasm `MsgExecuteContract` is defined [here](https://github.com/CosmWasm/wasmd/blob/4fe2fbc8f322efdaf187e2e5c99ce32fd1df06f0/x/wasm/types/tx.pb.go#L340-L349
) as the following type:

```go
type MsgExecuteContract struct {
	// Sender is the actor that committed the message in the sender chain
	Sender string
	// Contract is the address of the smart contract
	Contract string
	// Msg json encoded message to be passed to the contract
	Msg RawContractMessage
	// Funds coins that are transferred to the contract on execution
	Funds sdk.Coins
}
```

So we detail where we want to get each of these fields from:

* Sender: We cannot trust the sender of an IBC packet, the counterparty chain has full ability to lie about it. 
We cannot risk this sender being confused for a particular user or module address on Osmosis.
So we replace the sender with an account to represent the sender prefixed by the channel and a wasm module prefix.
This is done by setting the sender to `Bech32(Hash("ibc-wasm-hook-intermediary" || channelID || sender))`, where the channelId is the channel id on the local chain. 
* Contract: This field should be directly obtained from the ICS-20 packet metadata
* Msg: This field should be directly obtained from the ICS-20 packet metadata.
* Funds: This field is set to the amount of funds being sent over in the ICS 20 packet. One detail is that the denom in the packet is the counterparty chains representation of the denom, so we have to translate it to Osmosis' representation.

> **_WARNING:_**  Due to a [bug](https://twitter.com/SCVSecurity/status/1682329758020022272) in the packet forward middleware, we cannot trust the sender from chains that use PFM. Until that is fixed, we recommend chains to not trust the sender on contracts executed via IBC hooks. 

So our constructed cosmwasm message that we execute will look like:

```go
msg := MsgExecuteContract{
	// Sender is the that actor that signed the messages
	Sender: "osmo1-hash-of-channel-and-sender",
	// Contract is the address of the smart contract
	Contract: packet.data.memo["wasm"]["ContractAddress"],
	// Msg json encoded message to be passed to the contract
	Msg: packet.data.memo["wasm"]["Msg"],
	// Funds coins that are transferred to the contract on execution
	Funds: sdk.NewCoin{Denom: ibc.ConvertSenderDenomToLocalDenom(packet.data.Denom), Amount: packet.data.Amount}
```

### ICS20 packet structure

So given the details above, we propogate the implied ICS20 packet data structure.
ICS20 is JSON native, so we use JSON for the memo format.

```json 
{
    //... other ibc fields that we don't care about
    "data":{
    	"denom": "denom on counterparty chain (e.g. uatom)",  // will be transformed to the local denom (ibc/...)
        "amount": "1000",
        "sender": "addr on counterparty chain", // will be transformed
        "receiver": "contract addr or blank",
    	"memo": {
           "wasm": {
              "contract": "osmo1contractAddr",
              "msg": {
                "raw_message_fields": "raw_message_data",
              }
            }
        }
    }
}
```

An ICS20 packet is formatted correctly for wasmhooks iff the following all hold:

* `memo` is not blank
* `memo` is valid JSON
* `memo` has at least one key, with value `"wasm"`
* `memo["wasm"]` has exactly two entries, `"contract"` and `"msg"`
* `memo["wasm"]["msg"]` is a valid JSON object
* `receiver == "" || receiver == memo["wasm"]["contract"]`

We consider an ICS20 packet as directed towards wasmhooks iff all of the following hold:

* `memo` is not blank
* `memo` is valid JSON
* `memo` has at least one key, with name `"wasm"`

If an ICS20 packet is not directed towards wasmhooks, wasmhooks doesn't do anything.
If an ICS20 packet is directed towards wasmhooks, and is formated incorrectly, then wasmhooks returns an error.

### Execution flow

Pre wasm hooks:

* Ensure the incoming IBC packet is cryptogaphically valid
* Ensure the incoming IBC packet is not timed out.

In Wasm hooks, pre packet execution:

* Ensure the packet is correctly formatted (as defined above)
* Edit the receiver to be the hardcoded IBC module account

In wasm hooks, post packet execution:

* Construct wasm message as defined before
* Execute wasm message
* if wasm message has error, return ErrAck
* otherwise continue through middleware

## Ack callbacks

A contract that sends an IBC transfer, may need to listen for the ACK from that packet. To allow
contracts to listen on the ack of specific packets, we provide Ack callbacks. 

### Design

The sender of an IBC transfer packet may specify a callback for when the ack of that packet is received in the memo 
field of the transfer packet. 

Crucially, _only_ the IBC packet sender can set the callback.

### Use case

The crosschain swaps implementation sends an IBC transfer. If the transfer were to fail, we want to allow the sender
to be able to retrieve their funds (which would otherwise be stuck in the contract). To do this, we allow users to 
retrieve the funds after the timeout has passed, but without the ack information, we cannot guarantee that the send 
hasn't failed (i.e.: returned an error ack notifying that the receiving change didn't accept it)

### Implementation

#### Callback information in memo

For the callback to be processed, the transfer packet's memo should contain the following in its JSON:

`{"ibc_callback": "osmo1contractAddr"}`

The wasm hooks will keep the mapping from the packet's channel and sequence to the contract in storage. When an ack is
received, it will notify the specified contract via a sudo message.

#### Interface for receiving the Acks and Timeouts

The contract that awaits the callback should implement the following interface for a sudo message:

```rust
#[cw_serde]
pub enum IBCLifecycleComplete {
    #[serde(rename = "ibc_ack")]
    IBCAck {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
        /// String encoded version of the ack as seen by OnAcknowledgementPacket(..)
        ack: String,
        /// Weather an ack is a success of failure according to the transfer spec
        success: bool,
    },
    #[serde(rename = "ibc_timeout")]
    IBCTimeout {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
    },
}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    #[serde(rename = "ibc_lifecycle_complete")]
    IBCLifecycleComplete(IBCLifecycleComplete),
}
```

### Async Acks

IBC supports the ability to send an ack back to the sender of the packet asynchronously. This is useful for
cases where the packet is received, but the ack is not immediately known. For example, if the packet is being
forwarded to another chain, the ack may not be known until the packet is received on the other chain.

Note this ACK does not imply full revertability. It is possible that unrevertable actions have occurred 
even if there is an Ack Error. (This is distinct from the behavior of ICS-20 transfers). If you want to ensure 
revertability, your contract should be implemented in a way that actions are not finalized until a success ack
is received.

#### Use case

Async acks are useful in cases where the contract needs to wait for a response from another chain before
returning a result to the caller. 

For example, if you want to send tokens to another chain after the contract is executed you need to
add a new ibc packet and wait for its ack. 

In the synchronous acks case, the caller will receive an ack from the contract before the second packet 
has been processed. This means that the caller will have to wait (and potentially track) if the second 
packet has been processed successfully or not. 

With async acks, you contract can take this responsibility and only send an ack to the caller once the 
second packet has been processed

#### Making contract Acks async

To support this, we allow contracts to return an `IBCAsync` response from the function being executed when the
packet is received. That response specifies that the ack should be handled asynchronously. 

Concretely the contract should return:

```rust
#[cw_serde]
pub struct OnRecvPacketAsyncResponse {
    pub is_async_ack: bool,
}
```

if `is_async_ack` is set to true, `OnRecvPacket` will return `nil` and the ack will not be written. Instead, the
contract wil be stored as the "ack actor" for the packet so that only that contract is allowed to send an ack 
for it.

It is up to the contract developers to decide which conditions will trigger the ack to be sent. 

#### Sending an async ack

To send the async ack, the contract needs to send the MsgEmitIBCAck message to the chain. This message will 
then make a sudo call to the contract requesting the ack and write the ack to state. 

That message can be specified in the contract as: 

```rust
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/osmosis.ibchooks.MsgEmitIBCAck")]
pub struct MsgEmitIBCAck {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub packet_sequence: u64,
    #[prost(string, tag = "3")]
    pub channel: ::prost::alloc::string::String,
}
```

The contract is expected to implement the following sudo message handler:

```rust
#[cw_serde]
pub enum IBCAsyncOptions {
    #[serde(rename = "request_ack")]
    RequestAck {
        /// The source channel (osmosis side) of the IBC packet
        source_channel: String,
        /// The sequence number that the packet was sent with
        packet_sequence: u64,
    },
}

#[cw_serde]
pub enum SudoMsg {
    #[serde(rename = "ibc_async")]
    IBCAsync(IBCAsyncOptions),
}
```

and that sudo call should return an `IBCAckResponse`:

```rust
#[cw_serde]
#[serde(tag = "type", content = "content")]
pub enum IBCAck {
    AckResponse{
        packet: Packet,
        contract_ack: ContractAck,
    },
    AckError {
        packet: Packet,
        error_description: String,
        error_response: String,
    }
}
```

Note: the sudo call is required to potentially allow anyone to send the MsgEmitIBCAck message. For now, however,
this is artificially limited so that the message can only be send by the same contract. This could be expanded in
the future if needed. 

# Testing strategy

See go tests.`