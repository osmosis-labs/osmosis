## Abstract

The Market module contains the logic for atomic swaps between Terra currencies (e.g. UST<>KRT), as well as between Terra and Luna (e.g. SDT<>Luna).

The ability to guarantee an available, liquid market with fair exchange rates between different Terra denominations and between Terra and Luna is critical for user-adoption and price-stability.

The price stability of TerraSDR's peg to the SDR is achieved through Terra<>Luna arbitrage activity against the protocol's algorithmic market-maker which expands and contracts Terra supply to maintain the peg.

## Contents

1. **[Concepts](01_concepts.md)**
    - [Swap Fees](01_concepts.md#Swap-Fees)
    - [Market Making Algorithm](01_concepts.md#Market-Making-Algorithm)
    - [Virtual Liquidity Pools](01_concepts.md#Virtual-Liquidity-Pools)
    - [Swap Procedure](01_concepts.md#Swap-Procedure)
    - [Seigniorage](01_concepts.md#Seigniorage)
2. **[State](02_state.md)**
    - [TerraPoolDelta](02_state.md#TerraPoolDelta)
3. **[EndBlock](03_end_block.md)**
    - [Replenish Pool](03_end_block.md#Replenish-Pool)
4. **[Messages](04_messages.md)**
    - [MsgSwap](04_messages.md#MsgSwap)
    - [MsgSwapSend](04_messages.md#MsgSwapSend)
    - [Functions](04_messages.md#Functions)
5. **[Events](05_events.md)**
    - [Handlers](05_events.md#Handlers)
5. **[Parameters](06_params.md)**