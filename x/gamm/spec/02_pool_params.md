# Pool Parameters

The `x/gamm` module contains the following `Pool` parameters:

  Key                        Type
  --------------------------; ----------------------------;
  SwapFee                    sdk.Dec
  ExitFee                    sdk.Dec
  SmoothWeightChangeParams   \*SmoothWeightChangeParams

- `SwapFee`: The swap fee is the cut of all swaps that goes to the
    Liquidity Providers (LPs) for a pool. Suppose a pool has a swap fee
    `sf`. Then if a user wants to swap `T` tokens in the pool, `sf * T`
    tokens go to the LP's, and then `(1 - sf) * T` tokens are swapped
    according to the AMM swap function.
- `ExitFee`: The exit fee is a fee that is applied to LP's that want
    to remove their liquidity from the pool. Suppose a pool has an exit
    fee `ef`. If they currently have `S` LP shares, then when they
    remove their liquidity they get tokens worth `(1 - ef) * S` shares
    back. The remaining `ef * S` shares are then burned, and the tokens
    corresponding to these shares are kept as liquidity.
- `SmoothWeightChangeParams`: These params allows pool governance to
    smoothly change the weights of the assets it holds in the pool. E.g.
    it can slowly move from a 2:1 ratio, to a 1:1 ratio. The params
    consist of `StartTime`, `Duration`, `InitialPoolWeights` and
    `TargetPoolWeights`, where the latter two params are a list of
    `PoolAsset` that define the `Token` and `Weight`. Currently, smooth
    weight changes are implemented as a linear change in weight ratios
    over a given duration of time. So weights changed from 4:1 to 2:2
    over 2 days, then at day 1 of the change, the weights would be
    3:1.5, and at day 2 its 2:2, and will remain at these weight ratios.
