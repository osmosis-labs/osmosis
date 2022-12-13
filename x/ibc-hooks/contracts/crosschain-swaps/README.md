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
            "next_memo":null
        }
    }
}}
```

Channels are determined by the prefixes specified in the contract during
instantiation, so the user needs to provide a receiver with the supported
prefix. In future (once ack/timeout tracking is supported by the chain), we
could enable support for any channel by specifying it as a parameter.

The `slippage` can be set to a percentage of the twap price (as shown above), or as
the minimum amount of tokens expected to be received: `{"min_output_amount": "100"}`.


#### Optional keys

If `track_ibc_sends` is enabled during instantiation, the `failed_delivery` key
can be set to an address on Osmosis that will be allowed to recover the tokens
in case of a failure. This key is optional and if ommited will default to
`false`.

The `next_memo` key, if provided, will be added to the IBC transfer as the memo
for that transfer. This can be useful if the receiving chain also has IBC hooks
on transfers. In that case, this can be used to specify how the receiver should
deal with the received tokens (mostly useful when the receiver is a contract or
another ibc actor).


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


## Testing

The following is the procedure to test the contracts on a testnet.

### Requirements

You will need: 

* An osmosis testnet
* Another testnet with an open channel to the osmosis testnet
* Tokens on the non-osmosis testnet
* A pool on the osmosis testnet between the IBC'd native asset of the
  non-osmosis testnet and osmo
  
### Example

For the purpose of this test, we used the following testnets:

Osmosis:

``` toml
# The network chain ID
chain-id = "osmo-test-4"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "test"
# CLI output format (text|json)
output = "text"
# <host>:<port> to Tendermint RPC interface for this chain
node = "https://rpc-test.osmosis.zone:443"
# Transaction broadcasting mode (sync|async|block)
broadcast-mode = "block"
```

Gaia:

``` toml
# The network chain ID
chain-id = "theta-testnet-001"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "test"
# CLI output format (text|json)
output = "json"
# <host>:<port> to Tendermint RPC interface for this chain
node = "http://seed-02.theta-testnet.polypore.xyz:26657"
# Transaction broadcasting mode (sync|async|block)
broadcast-mode = "block"
```

### Deploying the contracts

Store the swaprouter code:

``` sh
> osmosisd-testnet tx wasm store ./bytecode/swaprouter.wasm --from owner  --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 -y
> export swaprouter_id=<the code id from above>
```

Instantiate the swaprouter using the code id received from the previous command:

``` sh
> export owner=<your account>
> osmosisd-testnet tx wasm instantiate $swaprouter_id '{"owner": "<owner bech32 addr>"}' --from owner --admin $owner --label swaprouter --yes
> export swaprouter_addr=<addr received from the above command>
```

Store the crosschain swaps code:

``` sh
> osmosisd-testnet tx wasm store ./bytecode/crosschain_swaps.wasm --from owner  --gas auto --gas-prices 0.1uosmo --gas-adjustment 1.3 -y
> export crosschain_swaps_id=<code id from the output of the prev command>
```

Instantiate the crosschain swaps contract using the code id received above. For
this you'll have to figure out the proper channel and prefix to use for the
testnet you're using and pass it in thge "channels" key:


``` sh
> osmosisd-testnet tx wasm instantiate $crosschain_swaps_id '{"swap_contract": "osmo1jd8fhpudhy8n77t57uqgq8jltc80khtrp2x0sflr0sa9useuyc7qcwc5ea", "track_ibc_sends": false, "channels": [["cosmos", "channel-314"]]}' --from owner --admin $owner --label=crosschain_swaps --yes
> export crosschain_swaps_addr=<contract addr from the response of the prec command>
```

### Transfer tokens

Make sure you have IBC tokens from the gaia testnet in the osmois one. If you don't, you can transfer with:

``` sh
> gaiad tx ibc-transfer transfer transfer channel-1238 osmo1pr8ktmhe095fc5stt5xrh4caw09xgtnasnwwf7 5850086uatom --from hub1 -y --gas auto --gas-prices 0.1uatom --gas-adjustment 1.3
```

You can also check the expected ibc denom on the osmosis side once you receive it

Note: For this the channels need to be active an a relayer needs to exist (or you have to manually relay the packets).

### Configuring the pools and swaprouter

Create a pool between the IBC'd atom and osmo:

``` sh
> osmosisd-testnet tx gamm create-pool --pool-file sample_pool.json --from owner
```

My pool file looks as follows:

``` json
{
        "weights": "1ibc/960B9755A955E1610CDE0F8AA35DDFCD1C31BDB35AE72E2702E1C2C2E778E603,1uosmo",
        "initial-deposit": "1000000ibc/960B9755A955E1610CDE0F8AA35DDFCD1C31BDB35AE72E2702E1C2C2E778E603,1000000uosmo",
        "swap-fee": "0.01",
        "exit-fee": "0.01",
        "future-governor": "168h"
}
```

Add the route to the swaprouter:

``` sh
 > osmosisd-testnet tx wasm execute $swaprouter_addr '{"set_route":{"input_denom":"ibc/960B9755A955E1610CDE0F8AA35DDFCD1C31BDB35AE72E2702E1C2C2E778E603","output_denom":"uosmo","pool_route":[{"pool_id":"720","token_out_denom":"uosmo"}]}}' --from owner -y
```

### Executing a crosschain swap

``` sh
gaiad tx ibc-transfer transfer transfer channel-1238 osmo1pr8ktmhe095fc5stt5xrh4caw09xgtnasnwwf7 1uatom --from hub1 -y --gas auto --gas-prices 0.1uatom --gas-adjustment 1.3 --memo '{"wasm": {"contract": "osmo1jx6d8x33yrysq0dyxpkwz48d2fn795y5p26ptsy4vchsjq52lteskp8n2q", "msg": {"osmosis_swap":{"input_coin":{"denom":"ibc/960B9755A955E1610CDE0F8AA35DDFCD1C31BDB35AE72E2702E1C2C2E778E603","amount":"100"},"output_denom":"uosmo","slippage":{"max_slippage_percentage":"20"},"receiver":"cosmos1qraj684l2deqwhe2x0dcz26xf28j5hrml8dlv5"}}}}'
```

