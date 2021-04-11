<!--
order: 4
-->

# Parameters

The minting module contains the following parameters:

| Key                     | Type             | Example                |
|-------------------------|------------------|------------------------|
| MintDenom               | string           | "uatom"                |
| AnnualProvisions        | string (dec)     | "800000000"            |
| MaxRewardPerEpoch       | string (dec)     | "0.200000000000000000" |
| MinRewardPerEpoch       | string (dec)     | "0.070000000000000000" |
| EpochDuration           | string (time ns) | "172800000000000"      |
| ReductionPeriodInEpochs | string (int64)   | "156"                  |
| ReductionFactorForEvent | string (dec)     | "0.5"                  |
| EpochsPerYear           | string (int64)   | "6311520"              |
