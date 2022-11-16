# Counter

This contract is a modification of the standard cosmwasm `counter` contract.
Namely it tracks a counter, _by sender_.
This is done to let us be able to test wasmhooks in Osmosis better.

This contract tracks any funds sent to it by adding it to the state under the `sender` key.

This way we can verify that, independently of the sender, the funds will end up under the 
`WasmHooksModuleAccount` address when the contract is executed via an IBC send that goes 
through the wasmhooks module.
