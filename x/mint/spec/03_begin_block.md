<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## NextAnnualProvisions

Rewards in the codebase are thought of in annual terms.
(TODO: Why?)

The target annual provision is recalculated on each reduction period (default 3 years).
At the time of reduction, annual provision is multiplied by reduction factor (default `2/3`),
consequently the rewards of the next period will be lowered by `1 - reduction factor`.

```go
func (m Minter) NextAnnualProvisions(params Params) sdk.Dec {
    return m.AnnualProvisions.Mul(params.ReductionFactorForEvent)
}
```

## EpochProvision

Calculate the provisions generated for each epoch based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount`. These rewards are transferred to a `FeeCollector`, which handles distributing the rewards per the chains needs. (See TODO.md for details) This fee collector is specified as the `auth` module's `FeeCollector` `ModuleAccount`.

```go
func (m Minter) EpochProvision(params Params) sdk.Coin {
    provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.EpochsPerYear)))
    return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
```
