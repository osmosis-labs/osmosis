<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current rewards information.

```go
type Minter struct {
    EpochProvisions sdk.Dec   // Rewards for the current epoch
}
```

## Params

Minting params are held in the global params store.

```go
type Params struct {
    MintDenom               string        // type of coin to mint
    GenesisEpochProvisions  sdk.Dec       // initial epoch provisions at genesis
    EpochIdentifier         string        // identifier of epoch
    ReductionPeriodInEpochs int64         // number of epochs between reward reductions
    ReductionFactor sdk.Dec               // reduction multiplier to execute on each period
}
```
