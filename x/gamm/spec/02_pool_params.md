<!--
order: 2
-->

# Pool Parameters

The `x/gamm` module contains the following `Pool` parameters:

| Key                      | Type                      |
|--------------------------|---------------------------|
| SwapFee                  | sdk.Dec                   |
| ExitFee                  | sdk.Dec                   |
| SmoothWeightChangeParams | *SmoothWeightChangeParams |

- `SwapFee`: 
- `ExitFee`: 
- `SmoothWeightChangeParams`:

    The swap fee is the cut of all swaps that goes to the Liquidity Providers (LPs) for a pool. Suppose a pool has a swap fee `s`. Then if a user wants to swap T tokens in the pool, `sT` tokens go to the LP's, and then `(1 - s)T` tokens are swapped according to the AMM swap function.
1. ExitFee
    The exit fee is a fee that is applied to LP's that want to remove their liquidity from the pool. Suppose a pool has an exit fee `e`. If they currently have `S` LP shares, then when they remove their liquidity they get tokens worth `(1 - e)S` shares back. The remaining `eS` shares are then burned, and the tokens corresponding to these shares are kept as liquidity.
2. FutureGovernor
    Osmosis plans to allow every pool to act as a DAO, with its own governance in a future upgrade. To facilitate this transition, we allow pools to specify who the governor should be as a string. There are currently 3 options for the future governor.
    - Noone will govern it, this is done by leaving the future governor string as blank.
    - Allow a given address to govern it, this is done by setting the future governor as a bech32 address.
    - Lockups to a token. This is the full DAO scenario. The future governor specifies a token denomination `Denom`, and a lockup duration `Duration`. This says that "all tokens of denomination `Denom` that are locked up for `Duration` or longer, have equal say in governance of this pool".
3. Weights
    This defines the weights of the pool. https://balancer.fi/whitepaper.pdf
    TODO Add better description of how the weights affect things here.
4. SmoothWeightChangeParams
    SmoothWeightChangeParams allows pool governance to smoothly change the weights of the assets it holds in the pool. So it can slowly move from a 2:1 ratio, to a 1:1 ratio. Currently, smooth weight changes are implemented as a linear change in weight ratios over a given duration of time. So weights changed from 4:1 to 2:2 over 2 days, then at day 1 of the change, the weights would be 3:1.5, and at day 2 its 2:2, and will remain at these weight ratios.
