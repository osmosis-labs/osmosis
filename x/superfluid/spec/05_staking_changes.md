<!--
order: 5
-->

# Staking power updates

We need to be concerned with how/when validators enter and leave the active set.

We expect the guarantee that there is an Intermediary account for every (active validator, superfluid denom) pair, and every (unbonding validator, superfluid denom) pair. (TODO: Where/why)

We also want to avoid resource exhaustion attacks. We relegate concerns around upper-bounding the number of active + unbonding validators to the staking module. This module is liable to potentially cause a 100-1000x amplification factor on this workload.

## How we handle it now

- Intermediary accounts are not created on SetSuperfluidAsset
- They are created at-time-of-need on MsgSuperfluidDelegate

## Future optimizations

- Adjust the guarantee to expect there to be an intermediary account for every (active, SFD) + (ubd, SFD) pairs with at least one osmo delegated to it. (Optimistic optimization, requires intermediate delegation account creation on MsgSuperfluidDelegate)
  - EDIT: Seems like this is actually what we do.
