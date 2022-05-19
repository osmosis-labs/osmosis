```html
<!--
order: 3
-->
```

# Gov

`Pool Incentives` module uses the values set at genesis or values added
by chain governance to distribute part of the inflation minted by the
mint module to specified gauges.

```go
type DistrInfo struct {
 TotalWeight github_com_cosmos_cosmos_sdk_types.Int 
 Records     []DistrRecord                          
}

type DistrRecord struct {
 GaugeId  uint64                                 
 Weight github_com_cosmos_cosmos_sdk_types.Int 
}
```

`DistrInfo` internally manages the `DistrRecord` and total weight of all
`DistrRecord`. Governance can modify DistrInfo via
`UpdatePoolIncentivesProposal` proposal.

### UpdatePoolIncentivesProposal

```go
type UpdatePoolIncentivesProposal struct {
 Title       string       
 Description string      
 Records     []DistrRecord 
}
```

`UpdatePoolIncentivesProposal` can be used by governance to update
`DistrRecord`s.

```shell
osmosisd tx gov submit-proposal update-pool-incentives [gaugeIds] [weights]
```

Proposals can be proposed in using the CLI command format above.\
For example, to designate 100 weight to gauge id 2 and 200 weight to
gauge id 3, the following command can be used.

```shell
osmosisd tx gov submit-proposal update-pool-incentives 2,3 100,200
```
