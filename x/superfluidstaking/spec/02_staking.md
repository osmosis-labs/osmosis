# Staking

## Block voting power

On staking module, block voting power should consider superfluid staking amount.
To do that, staking module (native Cosmos SDK module) should be modified to call a function in superfluid staking for additional voting power.

Osmosis's forked Cosmos SDK could have modified staking module that calls registered function for additional voting power of a validator that is registered in `app.go`.