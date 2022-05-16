# Parameters

The `x/gamm` module contains the following parameters:

  Key               Type        Example
  -----------------; -----------; --------------------------------------------;
  PoolCreationFee   sdk.Coins   \[{"denom":"uosmo","amount":"100000000"}\]

## PoolCreationFee

This parameter defines the amount of coins paid to community pool at the
time of pool creation which is introduced to prevent spam pool creation.
