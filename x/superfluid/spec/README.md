<!--
order: 0
title: "Superfluid Overview"
parent:
  title: "superfluid"
-->

# `superfluid`

## Abstract

Superfluid module provides a common interface for superfluid staking.

Module provides below functionalities
- Governance defined list of assets to allow get superfluid staked
- Superfluid staked assets implement an interface which has a function for "get risk-adjusted osmo value". We have to figure out how we do this with specifying denom.
- Every epoch, every assets value for "current superfluid amount staked" should be recorded in state. (We should only store the last UNBONDING_PERIOD_IN_EPOCHS number of records per past asset value)
- Later on, we should move the slashing module into here, and make it read from the superfluid staking module.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Keeper](05_keeper.md)**  
6. **[Hooks](06_hooks.md)**  
7. **[Queries](07_queries.md)**  
8. **[Params](08_params.md)**
9. **[Endblocker](09_endblocker.md)**