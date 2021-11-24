# Capability

::: warning NOTE
Osmosis's capability module inherits from Cosmos SDK's [`capability`](https://docs.cosmos.network/master/modules/capability/) module. This document is a stub, and covers mainly important Osmosis-specific notes about how it is used.
:::

The capability module allows you to provision, track, and authenticate multi-owner capabilities at runtime.

The keeper maintains two states: persistent and ephemeral in-memory. The persistent state maintains a globally unique autoincrementing index and a map from the capability index to a set of capability owners that are defined as a module and a capability name tuple. The ephemeral in-memory state tracks the actual capabilities, represented as addresses in local memory, with both forward and reverse indexes. The forward index maps the module name and capability tuples to the capability name. The reverse index maps the module and capability name to the capability itself.

The keeper allows the creation of scoped subkeepers, which are tied to a particular module by name. Scoped subkeepers must be created and passed to modules when you initialize the application. Then, the modules can use them to claim capabilities they receive and retrieve capabilities that they own by name. Additionally, they can create new capabilities and authenticate capabilities passed by other modules. A scoped subkeeper cannot escape its scope, so a module cannot interfere with or inspect capabilities owned by other modules.

The keeper provides no other core functionality that can be found in other modules, such as queriers, REST and CLI handlers, and the genesis state.

## Initialization

When you initialize the application, the keeper must be instantiated with a persistent store key and an ephemeral in-memory store key.

```
type App struct {
  // ...

  capabilityKeeper *capability.Keeper
}

func NewApp(...) *App {
  // ...

  app.capabilityKeeper = capability.NewKeeper(codec, persistentStoreKey, memStoreKey)
}
```

After the keeper is created, it can create scoped subkeepers, which are passed to other modules that can create, authenticate, and claim capabilities. After all the necessary scoped subkeepers are created and the state is loaded, you must initialize and seal the main capability keeper to populate the ephemeral in-memory state and to prevent more scoped subkeepers from being created.

```
func NewApp(...) *App {
  // ...

  // Initialize and seal the capability keeper so all persistent capabilities
  // are loaded in-memory and prevent any further modules from creating scoped
  // sub-keepers.
  ctx := app.BaseApp.NewContext(true, tmproto.Header{})
  app.capabilityKeeper.InitializeAndSeal(ctx)

  return app
}
```

## Concepts

Capabilities are multi-owner. A scoped subkeeper can create a capability via `NewCapability,` which creates a unique, unforgeable, object-capability reference. The newly created capability is automatically persisted. The calling module does not need to call `ClaimCapability`. Calling `NewCapability` creates the capability with the calling module and name as a tuple to be the capabilities first owner.

Capabilities can be claimed by other modules, which add them as owners. `ClaimCapability` allows a module to claim a capability key that it has received from another module so that `GetCapability` calls made later will succeed. If a module that receives a capability wants to access it by name later, `ClaimCapability` must be called. Because capabilities are multi-owner, if multiple modules have one capability reference, all the modules own it. If a module receives a capability from another module but does not call `ClaimCapability`, it may use it in the executing transaction but will not be able to access it afterward.

Any module can call `AuthenticateCapability` to check whether a capability corresponds to a particular name, including un-trusted user input, with which the calling module previously associated it.

`GetCapability` allows a module to fetch a capability that it has previously claimed by name. The module is not allowed to retrieve capabilities that it does not own.

### Stores

- MemStore

## States

- Index
- CapabilityOwners
- Capability
