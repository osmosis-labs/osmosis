# State

Section describes all stored by the module objects and their storage keys.

Refer to the [callback.proto](../../../proto/osmosis/callback/v1beta1/callback.proto) for objects fields description.

## Params

[Params](../../../proto/osmosis/callback/v1beta1/callback.proto#L38) object is used to store the module params.

The params value can only be updated by x/gov module via a governance upgrade proposal.

Storage keys:

* Params: `ParamsKey -> ProtocolBuffer(Params)`

## Callback

[Callback](../../../proto/osmosis/callback/v1beta1/callback.proto#L12) object is used to store the callbacks which are registered.

The callbacks are pruned after they are executed.

Storage keys:

* Callback: `CallbacksKey | BlockHeight | ContractAddress | JobID -> ProtocolBuffer(Callback)`
