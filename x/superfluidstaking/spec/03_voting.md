# Voting

## Governance proposal voting power

On `gov` module, to vote on the proposal, voting power is required.
Voting power of an address is calculated from OSMO stake amount and superfluid staking amount.

As in `staking` module, there will be a function to calculate superfluid staking amount that is registered on `app.go` and `gov` module will utilize it to calculate the total voting power and individual people's voting power.
