<!--
order: 8
-->

# Parameters

The incentives module contains the following parameters:

| Key                  | Type    | Example  |
| -------------------- | ------- | -------- |
| DistrEpochIdentifier | string  | "weekly" |
| MinAutostakingRate   | decimal | 0.5      |

Note:
- DistrEpochIdentifier is a epoch identifier, and module distribute rewards at the end of epochs.
As `epochs` module is handling multiple epochs, the identifier is required to check if distribution should be done at `AfterEpochEnd` hook
- MinAutostakingRate defines the rate of autostake amount per user after distribution of rewards for LPs.
For now, to make it simple, all the users' staking rates are same and it is defined by governance as a param.