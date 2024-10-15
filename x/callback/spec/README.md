# Callback

This module enables CosmWasm based smart contracts to receive callbacks at the end of a desired block. This is useful for scheduling actions to happen at an expected time by reserving execution in advance.

## Concepts

Callbacks are an intent submitted by a smart contract or a contract admin or a contract owner, which requests the protocol to execute an endpoint on the given contract for the desired height. The data structure of a callback can be found at [callback.proto](../../../proto/osmosis/callback/v1beta1/callback.proto#L12).

The authorized user can register a callback by providing the following:

1. Contract Address - The address of the contract which will receive the callback.
2. Job ID - User given number which can be used by the contract to handle different callbacks with custom logic.
3. Callback Height - The height at which the callback will be executed.
4. Fees - The total fees paid to successfully register a callback. [More](#fees)

### Fees

There are three types of fees that need to be paid to register a callback.

$fees = txFee + blockFee + futureFee$

where,

* fees is the total amount of fees to be paid to register a callback
* txFee is the transaction fee. [More](#1-transaction-fees)
* blockFee is the block reservation fee. [More](#2-block-reservation-fee)
* futureFee is the future reservation fee. [More](#3-future-reservation-fee)

#### 1. Transaction Fees

As the callbacks are executed by the protocol, the computation is subsidized by the validators. To ensure that the validators receive fair compensation, the transaction fees are paid upfront when registering a callback. As the gas consumption of the callback is not known at registration time, the user has to overpay for the callback. However, post completion of callback execution, any extra tx fee is refunded.

$txFee = callbackGasLimit_{params} \times estimateFees(1)$

where,

* txFee is the total transaction fees which need to be paid
* callbackGasLimit is a module param. [More](./01_state.md)

> **Note**

#### 2. Block Reservation Fee

This part of the fee is calculated based on how many callbacks are registered at the current block. The more filled a block's callback queue is, the more expensive it is to request further callbacks in that block.

$blockFee = count(callbacks_{currentHeight}) \times blockReservationFeeMultiplier_{params}$

where,

* blockFee is the block reservation fees which need to be paid
* count(callbacks) is the total number of callbacks already registered for the current block
* blockReservationFeeMultiplier is a module param. [More](./01_state.md)

#### 3. Future Reservation Fee

This part of the fee is calculated based on how far in the future does the user want to register their callback. The further in the future it is, the more expensive it is to request a callback.

$futureFee = (blockHeight_{callback} - blockHeight_{current}) \times futureReservationFeeMultiplier_{params}$

where,

* futureFee is the future reservation fees which need to be paid
* blockHeight is the respective height at the callback request and the current height
* futureReservationFeeMultiplier is a module param. [More](./01_state.md)

> **Note**
> Any extra fees paid, is kept as surplus fee and is refunded in case the callback is cancelled.

Post execution of a callback, all the fees are sent to the `fee_collector` account and are distributed to the validators and stakers

## How to use in CW contract

The callback is handled through the `sudo` entrypoint in the contract. The following Msg and endpoint needs to be implemented for the contract to be able to receive the callback.

```rust
// msg.rs
#[cw_serde]
pub enum SudoMsg {    
    Callback { job_id: u64 },
}
```

```rust
// contract.rs
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(_deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {        
        SudoMsg::Callback { job_id } => {
             !unimplemented()
        }
    }
}
```

A sample contract which shows how the feature can be used can be found [here](../../../contracts/callback-test/).

## Error Handling

As the contracts are executed during the protocol end blocker, it is not possible to return any execution errors to the user. However, the contract can use [x/cwerrors](../../cwerrors/spec/README.md) to get the errors when they happen.

## Contents

1. [State](./01_state.md)
2. [Messages](./02_messages.md)
3. [End Block](./03_end_block.md)
4. [Events](./04_events.md)
5. [Client](./05_client.md)
6. [Wasm bindings](./06_wasm_bindings.md)
7. [Module Errors](./07_errors.md)
