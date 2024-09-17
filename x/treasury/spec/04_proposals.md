## Proposals

The Treasury module defines special proposals which allow the [Tax Rate](./02_state.md#TaxRate) and [Reward Weight](./02_state.md#RewardWeight) values in the `KVStore` to be voted on and changed accordingly, subject to the [policy constraints](./03_end_block.md#PolicyConstraints) imposed by `pc.Clamp()`.

The treasury module will define two proposals to add or remove tax exemption list. Transaction among addresses in tax exemption list will not be taxed.

### TaxRateUpdateProposal

```go
type TaxRateUpdateProposal struct {
	Title       string  // Title of the Proposal
	Description string  // Description of the Proposal
	TaxRate     osmomath.Dec // target TaxRate
}
```

::: details JSON Example

```json
{
  "type": "treasury/TaxRateUpdateProposal",
  "value": {
    "title": "proposal title",
    "description": "proposal description",
    "tax_rate": "0.001000000000000000"
  }
}
```

### RewardWeightUpdateProposal

```go
type RewardWeightUpdateProposal struct {
	Title        string  // Title of the Proposal
	Description  string  // Description of the Proposal
	RewardWeight osmomath.Dec // target RewardWeight
}
```

::: details JSON Example

```json
{
  "type": "treasury/RewardWeightUpdateProposal",
  "value": {
    "title": "proposal title",
    "description": "proposal description",
    "reward_weight": "0.001000000000000000"
  }
}
```

### AddBurnTaxExemptionAddressProposal

```go
type AddBurnTaxExemptionAddressProposal struct {
	Title            string     // Title of the Proposal
	Description      string     // Description of the Proposal
	ExemptionAddress []string   // List of addresses to be added to tax exemption
}
```

::: details JSON Example

```json
{
  "type": "treasury/AddBurnTaxExemptionAddressProposal",
  "value": {
    "title": "proposal title",
    "description": "proposal description",
    "exemption_address": ["terra1dczz24r33fwlj0q5ra7rcdryjpk9hxm8rwy39t","terra1qt8mrv72gtvmnca9z6ftzd7slqhaf8m60aa7ye"]
  }
}
```

### RemoveBurnTaxExemptionAddressProposal

```go
type RemoveBurnTaxExemptionAddressProposal struct {
	Title            string     // Title of the Proposal
	Description      string     // Description of the Proposal
	ExemptionAddress []string   // List of addresses to be removed from tax exemption
}
```

::: details JSON Example

```json
{
  "type": "treasury/RemoveBurnTaxExemptionAddressProposal",
  "value": {
    "title": "proposal title",
    "description": "proposal description",
    "exemption_address": ["terra1dczz24r33fwlj0q5ra7rcdryjpk9hxm8rwy39t","terra1qt8mrv72gtvmnca9z6ftzd7slqhaf8m60aa7ye"]
  }
}
```