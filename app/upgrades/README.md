# Osmosis Upgrades

This folder contains sub-folders for every osmosis upgrade. (Both state migrations, and hard forks)
It also defines upgrade & hard fork structs, that each upgrade implements. 
These then get included in the application app.go to run the upgrade.

## Version History

* v1 - Initial version
* v3 - short blurb on prop19 and the fork
* v4 - Berylium State Migration
* v5 - Boron State migration
* v6 - hard fork for IBC bug fix
* v7 - Carbon State migration
* v8 - Nitrogen State migration

## Upgrade types 

There are two upgrade types exposed, `Upgrade` and `Fork`. 
An `Upgrade` defines an upgrade that is to be acted upon by state migrations from the SDK `x/upgrade` module.
A `Fork` defines a hard fork that changes some logic at a block height. 
If the goal is to have a new binary be compatible with the old binary prior to the upgrade height,
as is the case for all osmosis `Fork`s, then all logic changes must be height-gated or in the `BeginForkLogic` code.

```go
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string
	// Function that creates an upgrade handler
	CreateUpgradeHandler func(mm *module.Manager, configurator module.Configurator, keepers *keepers.AppKeepers) upgradetypes.UpgradeHandler
	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}

type Fork struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string
	// height the upgrade occurs at
	UpgradeHeight int64

	// Function that runs some custom state transition code at the beginning of a fork.
	BeginForkLogic func(ctx sdk.Context, keepers *keepers.AppKeepers)
}
```