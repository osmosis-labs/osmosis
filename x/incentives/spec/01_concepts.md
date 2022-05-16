# Concepts

The purpose of `incentives` module is to provide incentives to the users
who lock specific token for specific period of time.

Locked tokens can be of any denom, including LP tokens, IBC tokens, and
native tokens. The incentive amount is entered from the provider
directly via a specific message type. Rewards for a given pool of locked
up tokens are pooled into a gauge until the disbursement time. At the
disbursement time, they are distributed pro-rata to members of the pool.
