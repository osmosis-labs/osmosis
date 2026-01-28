# DEX Module Dependency Graph

This document provides a visual representation of the dependency relationships between the DEX modules being migrated from Osmosis to Gaia.

---

## Overview

The migration follows a **bottom-up** approach: start with leaf dependencies (no internal imports) and work up to modules that depend on everything.

---

## Visual Dependency Graph

```
                              ┌──────────────────────────────────────────────────────────────────────┐
                              │                           PROTOREV                                   │
                              │              MEV arbitrage across all pool types                     │
                              │   Uses: poolmanager, gamm, concentrated-liquidity, cosmwasmpool     │
                              └───────────────────────────────────┬──────────────────────────────────┘
                                                                  │
                                                                  │ depends on
                                                                  ▼
                ┌─────────────────────────────────────────────────────────────────────────────────────┐
                │                                                                                     │
    ┌───────────┼────────────────────────────────────────┬────────────────────────────────┐           │
    │           │                                        │                                │           │
    ▼           ▼                                        ▼                                ▼           │
┌────────┐ ┌─────────────────────────────────┐  ┌─────────────────────────┐  ┌───────────────────────┐│
│  GAMM  │ │    CONCENTRATED-LIQUIDITY       │  │      COSMWASMPOOL       │  │    (SDK Modules)      ││
│        │ │                                 │  │                         │  │                       ││
│Balancer│ │  CL pools (Uniswap v3 style)    │  │  CosmWasm-based pools   │  │ bank, auth, staking,  ││
│Stable- │ │  Uses: osmoutils/accum heavily  │  │  (Transmuter, orderbook)│  │ distribution, epochs  ││
│  swap  │ │                                 │  │  Requires wasmd v0.60   │  │                       ││
└────┬───┘ └───────────────┬─────────────────┘  └────────────┬────────────┘  └───────────────────────┘│
     │                     │                                  │                                       │
     │                     │ implements                       │                                       │
     │                     │ PoolModuleI                      │                                       │
     └──────────┬──────────┴──────────────────────────────────┘                                       │
                │                                                                                     │
                ▼                                                                                     │
┌───────────────────────────────────────────────┐                                                     │
│           POOLMANAGER/KEEPER                  │◄────────────────────────────────────────────────────┘
│                                               │
│  Central router - routes swaps to pool types  │
│  Collects taker fees, manages pool lifecycle  │
└───────────────────────┬───────────────────────┘
                        │
                        │ uses interfaces from
                        ▼
┌───────────────────────────────────────────────┐
│           POOLMANAGER/TYPES                   │
│                                               │
│  Defines PoolI, PoolModuleI interfaces        │
│  NO dependencies on pool implementations      │
└───────────────────────┬───────────────────────┘
                        │
                        │ depends on
                        ▼
┌───────────────────────────────────────────────┐
│               OSMOUTILS                       │
│                                               │
│  Subpackages needed:                          │
│  ├─ osmoutils (root) - store helpers          │
│  ├─ osmoutils/accum - reward accumulator      │
│  ├─ osmoutils/osmocli - CLI helpers           │
│  ├─ osmoutils/osmoassert - test assertions    │
│  ├─ osmoutils/cosmwasm - CosmWasm helpers     │
│  └─ osmoutils/observability - telemetry       │
└───────────────────────┬───────────────────────┘
                        │
                        │ depends on
                        ▼
┌───────────────────────────────────────────────┐
│                OSMOMATH                       │
│                                               │
│  TRUE LEAF - No Osmosis dependencies          │
│  BigDec (36-decimal precision)                │
│  Type aliases to cosmossdk.io/math            │
└───────────────────────────────────────────────┘
```

---

## Module-to-Module Dependencies

### Core DEX Modules

| Module | Depends On (Internal) | Depends On (External) |
|--------|----------------------|----------------------|
| **osmomath** | _(none - true leaf)_ | `cosmossdk.io/math` |
| **osmoutils** | osmomath | IBC-go, cosmos-sdk, wasmvm |
| **poolmanager/types** | osmomath, osmoutils | cosmos-sdk types |
| **gamm** | poolmanager/types, osmomath, osmoutils | bank, auth, staking, epochs |
| **poolmanager/keeper** | poolmanager/types, osmomath, osmoutils | bank, auth, distribution |
| **concentrated-liquidity** | poolmanager/types, osmomath, osmoutils/accum | bank, auth, wasmd |
| **cosmwasmpool** | poolmanager/types, osmomath, osmoutils | bank, wasmd |
| **protorev** | poolmanager, gamm, CL | bank, distribution, epochs |

### osmoutils Subpackage Usage by Module

| Module | root | accum | osmocli | osmoassert | cosmwasm | observability |
|--------|------|-------|---------|------------|----------|---------------|
| poolmanager | ✅ | ❌ | ✅ | test | ❌ | ❌ |
| gamm | ✅ | ❌ | ✅ | test | ❌ | ❌ |
| concentrated-liquidity | ✅ | **✅** | ✅ | ✅ | ❌ | ✅ |
| cosmwasmpool | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ |
| protorev | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |

> **Note**: Only `concentrated-liquidity` uses `osmoutils/accum` (for spread rewards and incentives).

---

## Migration Order (Topological Sort)

The recommended migration order follows the dependency graph from leaves to root:

```
Step 1: osmomath
   │
   │  ✅ Completed
   ▼
Step 2: osmoutils (minimal subset)
   │
   │  ✅ Completed
   ▼
Step 3: poolmanager/types
   │
   │  ✅ Completed
   ▼
Step 4: gamm
   │
   │  🚧 In Progress
   ▼
Step 5: poolmanager/keeper
   │
   │  📋 Pending
   ▼
Step 6: concentrated-liquidity     Step 7: cosmwasmpool
   │                                  │
   │  📋 Pending                      │  📋 Pending
   │                                  │
   └────────────────┬─────────────────┘
                    │
                    ▼
             Step 8: protorev
                    │
                    │  📋 Pending
                    ▼
               COMPLETE
```

### Rationale for Order

1. **osmomath** - True leaf, no Osmosis dependencies
2. **osmoutils** - Depends only on osmomath; provides utilities for all modules
3. **poolmanager/types** - Interfaces only; needed before pool implementations
4. **gamm** - Simplest pool type; establishes pool module pattern
5. **poolmanager/keeper** - Can now route to gamm; validates router logic
6. **concentrated-liquidity** - More complex; heavy accum usage
7. **cosmwasmpool** - Requires wasmd; can be done in parallel with CL
8. **protorev** - Depends on all pool types; must be last

---

## External Dependencies (SDK & Third-Party)

### SDK Modules Used

```
┌─────────────────────────────────────────────────────┐
│                 COSMOS SDK v0.53                    │
├─────────────────────────────────────────────────────┤
│                                                     │
│  x/bank ────────────────► token transfers           │
│  x/auth ────────────────► account management        │
│  x/staking ─────────────► staking info              │
│  x/distribution ────────► community pool            │
│  x/epochs ──────────────► periodic hooks            │
│  x/params (legacy) ─────► module params             │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Third-Party Dependencies

```
┌─────────────────────────────────────────────────────┐
│              THIRD-PARTY DEPENDENCIES               │
├─────────────────────────────────────────────────────┤
│                                                     │
│  IBC-go v10 ────────────► IBC denom utilities       │
│  wasmd v0.60 ───────────► CosmWasm integration      │
│  CometBFT v0.38 ────────► consensus (unchanged)     │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## Circular Dependency Clarification

There are **NO true circular dependencies** in the module graph.

The perceived circular relationship between poolmanager and pool modules is actually:

```
poolmanager/types  ◄── defines interfaces (PoolI, PoolModuleI)
        ▲              NO imports from pool modules
        │
        ├── gamm ─────────────────── imports poolmanager/types
        ├── concentrated-liquidity ── imports poolmanager/types  
        └── cosmwasmpool ──────────── imports poolmanager/types

poolmanager/keeper ◄── receives pool keepers via DEPENDENCY INJECTION
                       at app wiring time, not import time
```

**Key insight**: `poolmanager/types` only defines interfaces and can compile standalone. Pool modules import those types to implement the interfaces. The keeper receives pool module keepers via DI at runtime.

---

## Modules NOT Being Migrated

These Osmosis modules are **out of scope** because they use SDK fork features or are not needed:

| Module | Reason | Fork Feature Used |
|--------|--------|-------------------|
| x/tokenfactory | Uses bank hooks | `TrackBeforeSend`, `BlockBeforeSend` |
| x/superfluid | Uses supply offsets | `GetSupplyOffset`, `AddSupplyOffset` |
| x/mint | Uses supply offsets | Epoch provisions offset |
| x/ibc-rate-limit | Uses bank hooks | Transfer tracking |
| x/txfees | Unnecessary complexity | _(none, just complex)_ |

> **Note**: The DEX modules do NOT depend on these fork features. They use standard SDK APIs.

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial dependency graph created | AI Assistant |
