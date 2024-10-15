# Messages

Section describes the processing of the module messages

## MsgUpdateParams

The module params can be updated via a governance proposal using the x/gov module. The proposal needs to include [MsgUpdateParams](../../../proto/osmosis/callback/v1beta1/tx.proto#L25) message. All the parameters need to be provided when creating the msg.

On success: 
* Module `Params` are updated to the new values

This message is expected to fail if:
* The msg is sent by someone who is not the x/gov module
* The param values are invalid

## MsgRequestCallback

A new callback can be registered by using the [MsgRequestCallback](../../../proto/osmosis/callback/v1beta1/tx.proto#L39) message.

On success:

* A callback is queued to be executed at the given height.
* The fee amount specified is transferred from the sender's account to the module account

This message is expected to fail if:

* Insufficient fees are sent
* The account has insufficient balance
* The contract with given address does not exist
* A callback with at given height for specified height with given job id already exists
* The callback request height is in the past or in the current block
* The sender is not authorized to request a callback. The callback can only be request by the following
  * The contract itself
  * The contract admin as set in the x/wasmd module
  * The contract owner as set in the x/rewards module

## MsgCancelCallback

An existing callback can be cancelled by using th [MsgCancelCallback](../../../proto/osmosis/callback/v1beta1/tx.proto#L58) message,

On success:

* The exisiting callback is removed from the execution queue.
* The txFee and surplusFee amount is refunded back to the sender.
* The rest of the fees are sent to fee_collector to be distributed to validators and stakers

This message is expected to fail if:

* Callback with specified block height, contract address and job id does not exist
* The sender is not authorized to cancel the callback. The callback can only be cancelled by the following:
  * The contract itself
  * The contract admin as set in the x/wasmd module
  * The contract owner as set in the x/rewards module
