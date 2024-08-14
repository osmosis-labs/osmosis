<!--
order: 5
-->

# Events

The oracle module emits the following events:

## EndBlocker

| Type                 | Attribute Key | Attribute Value |
|----------------------|---------------|-----------------|
| policy_update        | tax_rate      | {taxRate}       |
| policy_update        | reward_weight | {rewardWeight}  |  
| policy_update        | tax_cap       | {taxCap}        |  

## Proposals

### TaxRateUpdateProposal

| Type            | Attribute Key | Attribute Value     |
|-----------------|---------------|---------------------|
| tax_rate_update | tax_rate      | {taxRate}           |

### RewardWeightUpdateProposal

| Type                 | Attribute Key | Attribute Value     |
|----------------------|---------------|---------------------|
| reward_weight_update | reward_weight | {rewardWeight}      |
