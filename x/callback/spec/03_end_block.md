# End Block

Section describes the module state changes on the ABCI end block call

## Callback Execution

Every end block we iterate over all the callbacks registered at that height. For each of the registered callback we,

1. Create a CallbackMsg

   It is a json encoded msg which includes the job id and is sent to the contract

2. Execute the callback

   A new sdk context is used with a limited gas meter. The gas limit is set to the value of the module param [CallbackGasLimit](../../../proto/osmosis/callback/v1beta1/callback.proto). Execute using the Sudo entrypoint and track the amount of gasUsed and errors, if any.

3. Handle error

   If there was any error during the execution of the callback, whether from the contract returning an error, or an out of gas error, set the error with the [x/cwerrors](../../cwerrors/spec/README.md) module with the appropriate error code.

   If the callback was successfull, throw a success event.

4. Calculate tx fees

   Based on the gas used, calculate the transaction fees for the executed callback. If the calculated fee is less than what was paid, refund the surplus to the address which registered the callback.

5. Distribute fees

   The consumed tx fees and all the other fees are sent to the fee collector to be distributed to the validators and stakers.

6. Cleanup

   Remove the callback entry from the state
