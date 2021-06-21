<!--
order: 1
-->

# Concepts

Osmosis is giving airdrops to the users to get more users trying the network.
At initial, module stores all airdrop users with amounts from genesis inside KVStore.

20% of airdrop is given to the user at initial and rest of them is distributed equally for the action made by the user.

To incentivize user to claim in time, after `DurationOfDecay` pass, the amount start to reduce linearly and ends after `DurationUntilDecay`.

After end, all remaining tokens are sent to the community pool.
