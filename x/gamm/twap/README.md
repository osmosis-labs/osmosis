# TWAP

We maintain TWAP entries for every gamm pool.

## Module API

## File layout

**api.go** is the main file you should look at for what you should depend upon.

**logic.go** is the main file you should look at for how the TWAP implementation works.

- types/* - Implement TwapRecord, GenesisState. Define AMM interface, and methods to format keys.
- api.go - Public API, that other users / modules can/should depend on
- hook_listener.go - Defines hooks & calls to logic.go, for triggering actions on 
- keeper.go - generic SDK boilerplate (defining a wrapper for store keys + params)
- logic.go - Implements all TWAP module 'logic'. (Arithmetic, defining what to get/set where, etc.)
- module.go - SDK AppModule interface implementation.
- store.go - Managing logic for getting and setting things to underlying stores

## Basic architecture notes

