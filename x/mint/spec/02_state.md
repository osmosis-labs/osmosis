<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current rewards information.

```go
type Minter struct {
	AnnualProvisions sdk.Dec   // current annual exptected provisions
}
```

## Params

Minting params are held in the global params store. 

```go
type Params struct {
	MintDenom               string        // type of coin to mint
	AnnualProvisions        sdk.Dec       // annual provisions
	MaxRewardPerEpoch       sdk.Dec       // maximum reward per epoch
	MinRewardPerEpoch       sdk.Dec       // minimum reward per epoch
	EpochDuration           time.Duration // duration of an epoch
	ReductionPeriodInEpochs int64         // number of epochs take to reduce rewards
	ReductionFactorForEvent sdk.Dec       // reduction multiplier to execute on each period
	EpochsPerYear           int64         // expected epochs per year
}
```