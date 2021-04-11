<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current rewards information.

```go
type Minter struct {
	EpochProvisions sdk.Dec   // current epoch's provisions
}
```

## Params

Minting params are held in the global params store. 

```go
type Params struct {
	MintDenom               string        // type of coin to mint
	GenesisEpochProvisions  sdk.Dec       // initial epoch provisions at genesis
	EpochDuration           time.Duration // duration of an epoch
	ReductionPeriodInEpochs int64         // number of epochs take to reduce rewards
	ReductionFactorForEvent sdk.Dec       // reduction multiplier to execute on each period
}
```