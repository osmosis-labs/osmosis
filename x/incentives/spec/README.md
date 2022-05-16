# Incentives

## Abstract

Incentives module provides general interface to give yield to stakers.

The yield to be given to stakers are stored in `gauge` and it is
distributed on epoch basis to the stakers who meet specific conditions.

Anyone can create gauge and add rewards to the gauge, there is no way to
take it out other than distribution.

There are two kinds of `gauges`, perpetual and non-perpetual ones.

- Non perpetual ones get removed from active queue after the the
    distribution period finish but perpetual ones persist.
- For non perpetual ones, they distribute the tokens equally per epoch
    during the `gauge` is in the active period.
- For perpetual ones, it distribute all the tokens at a single time
    and somewhere else put the tokens regularly to distribute the
    tokens, it's mainly used to distribute minted OSMO tokens to LP
    token stakers.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Hooks](05_hooks.md)**\
6. **[Queries](06_queries.md)**\
7. **[Params](07_params.md)**
