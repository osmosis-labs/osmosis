# Crosschain Swaps

The following contract is a *swap and forward* contract that takes the received
tokens, swaps them via a swaprouter contract, and sends them to an IBC receiver.

The goal is to use this contract to provide *crosschain swaps*: sending an ICS20
transfer on chain A, receiving it on osmosis, swapping for a different token,
and forwarding to a different chain.

## Instantiation

To instantiate the contract, you need to specify the following parameters:

 * swap_contract: the swaprouter contract to be used
 * track_ibc_sends: true|false. Specifies if the contract should track the sent ibc packets for recovery. This should be false in v1
 * channels: a list of (bech32 prefix, channel_id) that the contract will allow. 

### Example instantiation message

``` json
{"swap_contract": "osmo1thiscontract", "track_ibc_sends": false, "channels": [["cosmos", "channel-0"], ["juno", "channel-42"]]}
```

## Usage

### Via IBC

Assuming the current implementation of the wasm middleware on Osmosis, the memo
of an IBC transfer to do crosschain swaps would look as follows:

``` json
{"wasm": {
    "contract": "osmo1crosschainswapscontract", 
    "msg": {
        "osmosis_swap": {
            "input_coin": {"denom":"token0","amount":"1000"}, 
            "output_denom":"token1",
            "slippage":{"max_slippage_percentage":"5"},
            "receiver":"juno1receiver",
            "failed_delivery":null
        }
    }
}}
```

Channels are hard-coded in the contract, so the user should specify one of the
supported chains: "axelar", "juno", "cosmoshub". In future (once ack/timeout
tracking is supported by the chain), any channel should be allowed. 

If `track_ibc_sends` is enabled during instantiation, the `failed_delivery` key
can be set to an address on Osmosis that will be allowed to recover the tokens
in case of a failure. This key is optional and if ommited will default to
`false`.

The `slippage` can be set to a percentage of the twap price (as shown above), or as
the minimum amount of tokens expected to be received: `{"min_output_amount": "100"}`.


## Requirements

To use this contract for crosschain swaps, the following are needed:

 * The chain needs a wasm execute middleware that executes a contract when
   receiving a wasm directive in the memo.
 * The swaprouter contract should be instantiated
 
Optional:

This contract can be configured to track the acks or timeouts of the sent
packets. For that we needed:

 * The chain to provide a way to call a contract when the ack or timeout for a
   packet is received. 
 * Instantiate the contract with the `track_ibc_sends` set to true. Defaults
   to false.
