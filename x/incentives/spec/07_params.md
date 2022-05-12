# Parameters

The incentives module contains the following parameters:

  Key                    Type     Example
  ---------------------- -------- ----------
  DistrEpochIdentifier   string   "weekly"

Note: DistrEpochIdentifier is a epoch identifier, and module distribute
rewards at the end of epochs. As `epochs` module is handling multiple
epochs, the identifier is required to check if distribution should be
done at `AfterEpochEnd` hook
