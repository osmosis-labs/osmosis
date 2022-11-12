# Counter

This contract is a modification of the standard cosmwasm `counter` contract.
Namely it tracks a counter, _by sender_.
This is done to let us be able to test wasmhooks in Osmosis better.