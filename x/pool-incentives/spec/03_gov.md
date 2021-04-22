<!--
order: 3
-->

# Gov

`Pool Incentives` module takes the uses the values set at genesis or values added by chain governance to distribute part of the inflation minted by the mint module to specified pots.

```go
type DistrInfo struct {
	TotalWeight github_com_cosmos_cosmos_sdk_types.Int 
	Records     []DistrRecord                          
}

type DistrRecord struct {
	PotId  uint64                                 
	Weight github_com_cosmos_cosmos_sdk_types.Int 
}
```
`DistrInfo` internally manages the `DistrRecord` and total weight of all `DistrRecord`. Governance can't modify DistrInfo. It can only add/remove/edit the `DistrRecord`.

### AddPoolIncentivesProposal
```go
type AddPoolIncentivesProposal struct {
	Title       string       
	Description string      
	Records     []DistrRecord 
}
```
`AddPoolIncentivesProposal` can be used by governance to add new `DistrRecord`.

```shell
osmosisd tx gov submit-proposal add-pool-incentives [potIds] [weights]
```
Proposals can be proposed in using the CLI command format above.  
For example, to designate 100 weight to pot id 2 and 200 weight to pot id 3, the following command can be used.

A pot id that's already registered can't be registered again. To change the weight, use the EditPoolIncentivesProposal as shown below, or RemovePoolIncentivesProposal to delete.

```shell
osmosisd tx gov submit-proposal add-pool-incentives 2,3 100,200
```

### EditPoolIncentivesProposal
```go
type EditPoolIncentivesProposal struct {
	Title       string      
	Description string       
	Records     []DistrRecord 
}
```
`EditPoolIncentivesProposal` is used by governance to modify the DistrRecord of a specific pot.

```shell
osmosisd tx gov submit-proposal edit-pool-incentives [potIds] [weights]
```
If no pot id that matches exists, the proposal can't be processed. In this case, use `AddPoolIncentivesProposal` first.

### RemovePoolIncentivesProposal
```go
type RemovePoolIncentivesProposal struct {
	Title       string   
	Description string   
	PotIds      []uint64 
}
```
`RemovePoolIncentivesProposal` is used by governance to delete the DistrRecord of a specific pot

```go
osmosisd tx gov submit-proposal remove-pool-incentives [potIds]
```
