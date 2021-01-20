<!--
order: 1
-->

# Concept

**Disclaimer: This is work in progress. Mechanisms are susceptible to change.**

The `x/farm` module is an inter-module accessible, generalized yield farming reward distribution module based onf the [F1 fee distribution module](https://github.com/cosmos/cosmos-sdk/tree/master/docs/spec/fee_distribution).

It should be noted that the `x/farm` module doesn't support features such as slashing and commission.

**Notice:** The `x/farm` module doesn't include msgs, and only manages the state. The user's shares or assets that was deposited through the `x/farm` module is not custodied by the `x/farm` module. Deposited assets as well as the yield farming reward must be transferred **through**(and not **to**) the `x/farm` module to the other module that is using the `x/farm` module.