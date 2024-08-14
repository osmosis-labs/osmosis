## Abstract 

The Treasury module acts as the "central bank" of the Terra economy, measuring macroeconomic activity by [observing indicators](./01_concepts.md#Observed-Indicators) and adjusting [monetary policy levers](./01_concepts.md#Monetary-Policy-Levers) to modulate miner incentives toward stable, long-term growth.

> While the Treasury stabilizes miner demand through adjusting rewards, the [Market](../market/spec/README.md) is responsible for Terra price-stability through arbitrage and market maker.

## Contents
1. **[Concepts](01_concepts.md)**
    - [Voting Procedure](01_concepts.md#Observed-Indicators)
    - [Reward Band](01_concepts.md#Monetary-Policy-Levers)
    - [Slashing](01_concepts.md#Updating-Policies)
    - [Abstaining from Voting](01_concepts.md#Probation)
2. **[State](02_state.md)**
    - [TaxRate](02_state.md#TaxRate)
    - [RewardWeight](02_state.md#RewardWeight)
    - [TaxCap](02_state.md#TaxCap)
    - [TaxProceeds](02_state.md#TaxProceeds)
    - [EpochInitialIssuance](02_state.md#EpochInitialIssuance)
    - [Indicators](02_state.md#Indicators)
    - [CumulativeHeight](02_state.md#CumulativeHeight)
3. **[EndBlock](03_end_block.md)**
    - [EndBlocker](03_end_block.md#EndBlocker)
    - [Functions](03_end_block.md#Functions)
    - [PolicyConstraints](03_end_block.md#PolicyConstraints)
4. **[Porposals](04_proposals.md)**
    - [TaxRateUpdateProposal](04_proposals.md#TaxRateUpdateProposal)
    - [RewardWeightUpdateProposal](04_proposals.md#RewardWeightUpdateProposal)
5. **[Events](05_events.md)**
    - [EndBlocker](05_events.md#EndBlocker)
    - [Proposals](05_events.md#Proposals)
6. **[Parameters](06_params.md)**
