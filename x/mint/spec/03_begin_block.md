<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## NextAnnualProvisions

The target annual provision is recalculated on each reduction period (default 3 years).
At the time of reduction, annual provision is multiplied by reduction factor (default `0.5`).

```
func (m Minter) NextAnnualProvisions(params Params) sdk.Dec {
	return m.AnnualProvisions.Mul(params.ReductionFactorForEvent)
}
```

## EpochProvision

Calculate the provisions generated for each epoch based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount` and then transferred to the `auth`'s `FeeCollector` `ModuleAccount`.

```
func (m Minter) EpochProvision(params Params) sdk.Coin {
	provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.EpochsPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
```
