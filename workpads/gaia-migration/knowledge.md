# Gaia Migration Knowledge Base

## Overview

**Goal**: Migrate core Osmosis DEX modules to Gaia, enabling all Osmosis DEX operations to run on Gaia with **production-grade quality**.

| Source | Target |
|--------|--------|
| Osmosis (`/Users/nicolas/devel/osmosis`) | Gaia (`/Users/nicolas/devel/gaia`) |

**Success Criteria**: When this project is complete, we should be able to run Gaia and execute **all Osmosis DEX operations** there. This is production-grade work, not a prototype.

**Approach**:
1. Both Gaia and Osmosis serve as references, but **Gaia is the main project** where we test and validate
2. First identify module dependencies to build a DAG (dependency graph)
3. Migrate from simplest (leaf nodes with no dependencies) to most complex
4. Each module must be fully working (all tests passing) before moving to the next

**Key Challenge**: Osmosis uses a different SDK version (v0.50.x fork) than Gaia (v0.53.4), so modules will not compile directly after copying. Each module requires adaptation.

---

## Executive Summary: Phase 0 Discovery Findings

### Key Conclusions

| Finding | Impact | Recommendation |
|---------|--------|----------------|
| **SDK Fork Features** | ✅ No blocker | DEX modules do NOT use bank hooks or supply offsets. Use upstream SDK 0.53. |
| **Store Fork** | ✅ No blocker | Performance optimization only. All code uses standard `store.KVStore` interface. |
| **osmoutils** | ✅ Manageable | Only 6 of 11 subpackages needed. All use standard APIs. |
| **x/epochs** | ✅ Use SDK version | SDK 0.53 x/epochs is wire-compatible. Minor hook interface adaptations needed. |
| **Circular Dependencies** | ✅ No issue | No true Go import cycles. `poolmanager/types` defines interfaces only. |
| **wasmd Version** | ⚠️ Requires work | v0.53 → v0.60 upgrade needed for cosmwasmpool. Gaia already has wasmd. |
| **IBC Version** | ⚠️ Requires work | v8 → v10 upgrade. osmoutils has IBC imports that need updating. |

### What Can Be Migrated As-Is (with version updates)

1. **osmomath** - True leaf, no Osmosis dependencies
2. **osmoutils** (minimal subset) - Standard store APIs, remove replace directives
3. **poolmanager/types** - Interfaces only, no keepers
4. **gamm** - Simpler pool type, no accumulator usage
5. **protorev** - MEV module, depends on pool modules

### What Requires More Adaptation

1. **concentrated-liquidity** - Most complex, heavy `osmoutils/accum` usage
2. **cosmwasmpool** - Requires wasmd v0.60 API compatibility check

### Modules Outside Scope (NOT migrating)

These Osmosis modules use SDK fork features and are NOT part of this migration:
- `x/tokenfactory` - Uses bank hooks
- `x/superfluid` - Uses supply offsets
- `x/mint` - Uses supply offsets
- `x/ibc-rate-limit` - Uses bank hooks

These modules are Osmosis-specific reward/incentive systems and are NOT part of this migration:
- `x/incentives` - Gauge-based reward distribution (Osmosis-specific)
- `x/pool-incentives` - Pool reward distribution (Osmosis-specific)
- `x/lockup` - LP token locking for rewards (Osmosis-specific)

These modules are excluded for complexity reasons (functionality will be reimplemented simpler):
- `x/txfees` - Fee distribution logic will be added directly to poolmanager via epoch hooks (see D3). The txfees module has unnecessary complexity: EIP-1559 mempool, fee token whitelist, fee decorators.

**Impact on DEX Modules**: 
- DEX modules originally had keeper interfaces for these (PoolIncentivesKeeper, IncentivesKeeper)
- These interfaces are **removed** during migration - core pool/swap functionality works without them
- CL migration features (CFMM → CL pool migration with superfluid) are **removed**

---

## Migration Plan

### Phase 1: Foundation (Leaf Dependencies)

| Step | Component | Description | Effort |
|------|-----------|-------------|--------|
| 1.1 | **osmomath** | Copy to Gaia, update `cosmossdk.io/math` to v1.5.3, remove replace directives | Low |
| 1.2 | **osmoutils (minimal)** | Copy 6 required subpackages, update IBC v8→v10, SDK v0.50→v0.53, remove replace directives | Medium |

### Phase 2: Core Pool Infrastructure

| Step | Component | Description | Effort |
|------|-----------|-------------|--------|
| 2.1 | **poolmanager/types** | Copy interfaces package, should compile standalone | Low |
| 2.2 | **gamm** | Copy module, adapt to SDK 0.53 patterns, move `MigrationPoolIDs` struct locally | Medium |
| 2.3 | **poolmanager/keeper** | Complete poolmanager, wire gamm as first pool type | Medium |

### Phase 3: Additional Pool Types

| Step | Component | Description | Effort |
|------|-----------|-------------|--------|
| 3.1 | **concentrated-liquidity** | Most complex module. Ensure `osmoutils/accum` works. Adapt legacy params. | High |
| 3.2 | **cosmwasmpool** | Verify wasmd v0.60 compatibility. Test with pre-compiled WASM contracts. | Medium |

### Phase 4: MEV & Integration

| Step | Component | Description | Effort |
|------|-----------|-------------|--------|
| 4.1 | **protorev** | Depends on all pool modules. Wire PostHandler. | Medium |
| 4.2 | **App Integration** | Wire all modules into Gaia app, genesis, upgrades | Medium |
| 4.3 | **Testing** | Unit tests, integration tests, manual testing with mainnet data | High |

### Per-Component Workflow

For each component above:

```
1. COPY      → Copy from Osmosis to Gaia
2. COMPILE   → Attempt build, document all errors
3. ADAPT     → Fix SDK version differences, update imports
4. VERIFY    → Clean compile with no errors
5. TEST      → Run unit tests, fix failures
6. INTEGRATE → Wire into Gaia app (if module)
7. VALIDATE  → Integration tests, manual verification
```

### Risk Mitigation

| Risk | Mitigation |
|------|------------|
| SDK API changes break modules | Document changes during adaptation, create compatibility shims if needed |
| wasmd v0.60 breaks cosmwasmpool | Test cosmwasmpool last, can defer if problematic |
| Accumulator issues in CL | Test `osmoutils/accum` thoroughly before CL migration |
| Genesis state incompatibility | Test export/import early, document format differences |

### Success Metrics

- [ ] All unit tests pass in Gaia
- [ ] Can create pools of all types (Balancer, Stableswap, CL, CosmWasm)
- [ ] Can execute swaps through poolmanager routing
- [ ] Protorev finds and executes arbitrage opportunities
- [ ] Genesis export/import works correctly

---

## Modules to Migrate

| Module | Description | Status |
|--------|-------------|--------|
| `poolmanager` | | 📋 pending |
| `concentrated-liquidity` | | 📋 pending |
| `gamm` | | 📋 pending |
| `cosmwasmpool` | | 📋 pending |
| `protorev` | | 📋 pending |

---

## Shared Module Dependencies

### x/epochs

**Purpose**: Provides time-based epoch hooks that trigger periodic operations across modules.

**Modules That Depend on It**:
- ~~`gamm` - EpochKeeper for epoch info~~ (removed - was via IncentivesKeeper which is excluded)
- `protorev` - EpochKeeper + epoch hooks for periodic route updates
- `poolmanager` - ⚠️ **NEW for Gaia** - epoch hooks for taker fee distribution (moved from txfees)

**SDK 0.53 Has x/epochs**: Yes, SDK 0.53 includes `x/epochs` module.

**Comparison: Osmosis vs SDK 0.53 x/epochs**:

| Aspect | Osmosis | SDK 0.53 |
|--------|---------|----------|
| EpochInfo fields | Identifier, StartTime, Duration, CurrentEpoch, CurrentEpochStartTime, EpochCountingStarted, CurrentEpochStartHeight | **Identical** |
| Proto field numbers | 1,2,3,4,5,6,8 | **Identical** |
| Hook context | `sdk.Context` | `context.Context` |
| GetModuleName() | Yes (for telemetry) | No |
| Panic handling | `osmoutils.ApplyFuncIfNoError` | Standard `errors.Join` |
| Depinject support | No | Yes (`EpochHooksWrapper`) |

**Key Differences in Hooks Interface**:

```go
// Osmosis
type EpochHooks interface {
    AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error
    BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error
    GetModuleName() string  // Extra method for telemetry
}

// SDK 0.53
type EpochHooks interface {
    AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error
    BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error
    // No GetModuleName - uses depinject for module identification
}
```

**Recommendation**: ✅ **Use SDK 0.53 x/epochs**

The SDK version is compatible and simpler. Required adaptations:
1. Change hook implementations to use `context.Context` instead of `sdk.Context`
2. Remove `GetModuleName()` from hook implementations (minor)
3. Use SDK's depinject patterns for hook registration
4. Accept slightly different panic handling (SDK doesn't catch panics, Osmosis does)

**Migration Notes**:
- EpochInfo type is wire-compatible (same proto fields) - genesis migration should work
- Hook interface change is minor - just context type and remove one method
- Osmosis uses custom panic recovery; SDK relies on standard error returns
- No need to port Osmosis x/epochs - use upstream SDK version

---

### Taker Fee Distribution Design

**Context**: In Osmosis, poolmanager collects taker fees at swap time and sends them to the `taker_fee_collector` module account. The actual distribution happens in `x/txfees` via epoch hooks (`AfterEpochEnd`).

**Decision**: ✅ **Add epoch hooks directly to poolmanager** instead of migrating `x/txfees`.

**Rationale**:
- `x/txfees` has significant complexity not needed for Gaia (EIP-1559 mempool, fee token whitelist, fee decorators)
- The core fee distribution logic is straightforward
- Keeps poolmanager self-contained for fee lifecycle

**What poolmanager needs to implement**:

```go
// AfterEpochEnd distributes accumulated taker fees
func (k Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
    if epochIdentifier != "day" {
        return nil
    }
    
    // 1. Get accumulated fees from taker_fee_collector
    // 2. Swap non-native fees to ATOM (base denom)
    // 3. Distribute according to TakerFeeParams:
    //    - CommunityPool % → FundCommunityPool()
    //    - Burn % → send to null address
    //    - StakingRewards % → send to fee_collector (auth module)
    // 4. Clear taker fee share accumulators (partner skims)
    
    return nil
}
```

**Key Design Points**:

| Aspect | Osmosis (txfees) | Gaia (poolmanager) |
|--------|------------------|-------------------|
| Base denom | OSMO | ATOM |
| Swap non-native fees | Yes, via protorev pools | Yes, via protorev pools |
| Smoothing buffer | Yes (gradual APR) | TBD - may simplify |
| 2-hop routing | Yes, for obscure tokens | TBD - may simplify |
| Epoch trigger | "day" | "day" |

**Dependencies for fee distribution**:
- `EpochsKeeper` - to register hooks (use SDK 0.53 x/epochs)
- `DistributionKeeper` - for `FundCommunityPool()`
- `BankKeeper` - for token transfers
- `ProtorevKeeper` - for finding swap routes (already a dependency)

**Open Items** (to investigate later):
1. Exact swap routing logic - use protorev's `GetPoolForDenomPairNoOrder`?
2. Whether to implement smoothing buffer or simplify to immediate distribution
3. Handling of failed swaps (leave in collector for next epoch?)
4. Taker fee share agreements (partner skims) - port the accumulator clearing logic

---

## Utility Packages (Leaf Dependencies)

### osmomath

**Purpose**: Extended math library providing `BigDec` (36-decimal precision) and aliases to `cosmossdk.io/math` types. Used by all DEX modules for precise mathematical operations.

**Key Components**:
- `BigDec` - High-precision decimal (36 places) for concentrated liquidity calculations
- Type aliases: `Dec`, `Int`, `Uint` → `cosmossdk.io/math.LegacyDec`, `Int`, `Uint`
- Helper functions: `exp2`, `sqrt`, binary search, rounding utilities

**External Dependencies**:
- `cosmossdk.io/math v1.4.0` - Core math types (need to update to v1.5.3 for Gaia)
- `github.com/cosmos/cosmos-sdk/types` - Only for `sdk.Coin` in some helpers
- Standard library: `math/big`, `encoding/json`, etc.

**Internal Dependencies**:
- ✅ **TRUE LEAF** - No Osmosis-internal dependencies

**Migration Notes**:
- Standalone Go module with own `go.mod`
- Currently uses Osmosis SDK fork via replace directive - must remove
- Update `cosmossdk.io/math` from v1.4.0 → v1.5.3
- Remove SDK replace directive, use upstream SDK types
- Should compile with minimal changes after version updates

---

### osmoutils

**Purpose**: General utility library providing accumulators, store helpers, CLI wrappers, CosmWasm helpers, and other shared functionality used across Osmosis modules.

**Key Components**:
- `accum/` - Accumulator for reward distribution
- `sumtree/` - Sum tree data structure
- `osmocli/` - CLI command wrappers and helpers
- `coinutil/` - Coin math utilities
- `cosmwasm/` - CosmWasm helpers
- Store helpers, encoding, parsing utilities

**External Dependencies**:
- `cosmossdk.io/store` - Uses standard KVStore interface (no fork APIs needed)
- `cosmossdk.io/math`, `log` - Standard cosmossdk.io packages
- `github.com/cosmos/cosmos-sdk` - Uses fork via replace (remove for Gaia)
- `github.com/cosmos/ibc-go/v8` - **Needs update to v10** for Gaia
- `github.com/CosmWasm/wasmvm/v2` - For CosmWasm helpers
- `github.com/cometbft/cometbft` - Uses Osmosis fork (use upstream for Gaia)

**Internal Dependencies**:
- `osmomath` - Depends on osmomath (osmomath is the true leaf)

**Migration Notes**:
- Standalone Go module with own `go.mod`
- ✅ **Store fork NOT required** - all subpackages use standard store.KVStore interface
- Must remove all replace directives and use upstream dependencies
- IBC-go v8 → v10 update required
- SDK v0.50 → v0.53 update required

### Minimal osmoutils Subset for DEX Modules

Analysis of which osmoutils subpackages each DEX module actually imports:

| DEX Module | osmoutils (root) | accum | osmocli | osmoassert | cosmwasm | observability |
|------------|------------------|-------|---------|------------|----------|---------------|
| **poolmanager** | ✅ | ❌ | ✅ | test only | ❌ | ❌ |
| **gamm** | ✅ | ❌ | ✅ | test only | ❌ | ❌ |
| **concentrated-liquidity** | ✅ | ✅ **critical** | ✅ | ✅ | ❌ | ✅ |
| **cosmwasmpool** | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ |
| **protorev** | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |

**Minimal Required Subpackages**:
1. `osmoutils` (root) - store helpers, used by ALL modules
2. `osmoutils/accum` - **CRITICAL for CL** - accumulator for spread rewards and incentives
3. `osmoutils/osmocli` - CLI helpers, used by ALL modules
4. `osmoutils/osmoassert` - assertions, used by CL and tests
5. `osmoutils/cosmwasm` - CosmWasm helpers, used by cosmwasmpool
6. `osmoutils/observability` - telemetry, used by CL

**NOT Required** (can be excluded):
- `osmoutils/sumtree` - not imported by DEX modules
- `osmoutils/coinutil` - not imported by DEX modules
- `osmoutils/partialord` - not imported by DEX modules
- `osmoutils/noapptest` - test utilities
- `osmoutils/wrapper` - database wrapper

**Key Finding**: ✅ All required subpackages use only standard `store.KVStore` interface methods (Get, Set, Delete, Has, Iterator). No store fork-specific APIs are called. The upstream SDK store will work.

---

## osmoutils Usage Summary (All Module Analyses Complete)

After analyzing all DEX modules, here is the complete picture of osmoutils usage:

| Module | Uses osmoutils/accum? | osmoutils subpackages used |
|--------|----------------------|---------------------------|
| poolmanager | No | root |
| concentrated-liquidity | **YES** | root, accum |
| gamm | No | root, osmocli |
| cosmwasmpool | No | root, cosmwasm |
| protorev | No | root |

**Key Insight**: Only `concentrated-liquidity` uses `osmoutils/accum`. All other DEX modules use simpler osmoutils patterns (root package helpers, CLI wrappers, or CosmWasm helpers).

**Implications for Migration**:
1. **osmoutils/accum is the critical path** - this is where store fork concerns may apply
2. **Other subpackages are simpler** - osmocli, cosmwasm, root helpers should work with upstream SDK
3. **If we can make accum work**, the rest of osmoutils migration is straightforward
4. **Fallback option**: If accum has blockers, we could potentially refactor CL module (significant work)

**Minimal osmoutils Subpackages Needed**:
- `osmoutils` (root) - store helpers, MustGet/MustSet, etc.
- `osmoutils/accum` - accumulator for CL spread rewards and incentives
- `osmoutils/osmocli` - CLI helpers for gamm
- `osmoutils/cosmwasm` - CosmWasm helpers for cosmwasmpool
- `osmoutils/osmoassert` - test assertions (tests only)

---

## Module Descriptions

### poolmanager

**Purpose**: Central router for all pool types. Manages pool creation, routing swaps across different pool types, taker fees, and provides a unified interface for interacting with any pool (GAMM, Concentrated Liquidity, CosmWasm pools).

**Key Components**:
- Pool routing and swap execution
- Taker fee collection (at swap time) and distribution (at epoch end)
- Multi-hop swap routing
- Pool creation delegation to specific pool modules
- Governance proposals for pool management
- ⚠️ **For Gaia**: Epoch hooks for fee distribution (moved from txfees)

**Cosmos SDK Dependencies**:
- `x/auth/types` - Account types
- `x/bank/types` - Bank types
- `x/distribution/types` - Community pool
- `x/gov/*` - Governance integration
- `x/params/types` - Params (legacy)
- `cosmossdk.io/core/appmodule`, `errors`, `math`, `store/types`

**Osmosis Internal Dependencies**:
- `osmomath` - Math utilities ⚠️ MUST MIGRATE FIRST
- `osmoutils` - General utilities ⚠️ MUST MIGRATE FIRST
- `x/pool-incentives/types` - Pool incentives types
- `x/txfees/types` - Transaction fees types ⚠️ NOT MIGRATING - define needed types/constants locally (module account names, etc.)

**Pool Module Relationships** (NOT circular - see note):
- `poolmanager/types` defines `PoolModuleI` interface (no imports from pool modules)
- `x/gamm`, `x/concentrated-liquidity`, `x/cosmwasmpool` import `poolmanager/types` to implement `PoolModuleI`
- `poolmanager/keeper` receives pool keepers via dependency injection at app wiring

**Required External Keepers**:
- `AccountI` - Account keeper
- `BankI` - Bank keeper (standard, no fork features needed)
- `CommunityPoolI` - Distribution keeper
- `StakingKeeper` - Staking keeper
- `PoolModuleI` - Generic pool interface (gamm, CL, cosmwasmpool implement this)
- `ConcentratedI` - CL-specific interface
- `PoolIncentivesKeeperI` - Pool incentives keeper
- `ProtorevKeeper` - Protorev keeper
- `WasmKeeper` - Wasm query keeper
- `EpochsKeeper` - ⚠️ NEW for Gaia - needed for fee distribution epoch hooks (use SDK 0.53 x/epochs)

**Migration Notes**:
- ✅ **No true circular dependency**: `poolmanager/types` defines interfaces only and does not import pool modules. Pool modules import `poolmanager/types` to implement interfaces. Keepers are wired via DI.
- Depends on osmomath and osmoutils which must be migrated first
- Can migrate `poolmanager/types` → `gamm` → `poolmanager/keeper` → other pools incrementally
- No direct SDK fork feature usage detected
- ⚠️ **Needs epoch hooks for fee distribution** - See "Taker Fee Distribution Design" section

---

### concentrated-liquidity

**Purpose**: Concentrated liquidity pools (Uniswap v3 style) that allow liquidity providers to specify price ranges for their liquidity, improving capital efficiency. This is the most complex pool type in Osmosis.

**Key Components**:
- Tick-based price ranges with configurable tick spacing
- Position management (create, add, withdraw liquidity)
- Swap execution with tick crossing logic
- Spread rewards (fees) accumulated per position
- Incentive distribution to liquidity providers
- Pool hooks for extensibility (CosmWasm contracts)
- Internal `math/` package for tick/price calculations
- Internal `model/` package for pool and position types
- Internal `swapstrategy/` for direction-specific swap logic

**Cosmos SDK Dependencies**:
- `cosmossdk.io/core/appmodule` - App module interface
- `cosmossdk.io/store` (prefix, types) - KV store access
- `cosmossdk.io/errors` - Error handling
- `cosmossdk.io/math` - Math types (Int, LegacyDec)
- `github.com/cosmos/cosmos-sdk/codec` - Encoding
- `github.com/cosmos/cosmos-sdk/types` - SDK types (sdk.Context, sdk.Coin, etc.)
- `github.com/cosmos/cosmos-sdk/types/query` - Pagination
- `github.com/cosmos/cosmos-sdk/x/params/types` - Legacy params
- `github.com/cosmos/cosmos-sdk/x/bank/types` - Bank types
- `github.com/cosmos/cosmos-sdk/x/gov/types` - Governance module address
- `github.com/cosmos/cosmos-sdk/telemetry` - Metrics

**Osmosis Internal Dependencies**:
- `osmomath` - BigDec and math utilities ⚠️ MUST MIGRATE FIRST
- `osmoutils` - General utilities ⚠️ MUST MIGRATE FIRST
  - Uses `osmoutils.MustGet`, `osmoutils.MustSet` for store helpers
  - Uses `osmoutils/accum` for accumulator (spread rewards, incentives)
- `x/poolmanager/types` - Pool interfaces (PoolI, CreatePoolMsg)
- `x/lockup/types` - For superfluid integration (lock types)
- `x/gamm` - Via GAMMKeeper interface (linked Balancer pools)
- `x/pool-incentives` - Via PoolIncentivesKeeper interface
- `x/incentives` - Via IncentivesKeeper interface

**Required External Keepers**:
- `AccountKeeper` - Module account management
- `BankKeeper` - Token transfers, minting, burning (standard, no fork features)
- `PoolManagerKeeper` - Pool creation, routing
- `GAMMKeeper` - Linked Balancer pool lookups
- `PoolIncentivesKeeper` - Pool gauge management
- `IncentivesKeeper` - Reward distribution
- `LockupKeeper` - Position locking (superfluid)
- `CommunityPoolKeeper` - Community pool funding
- `ContractKeeper` - CosmWasm contract sudo calls (pool hooks)

**Migration Notes**:
- Large, complex module with ~60 source files
- Uses `osmoutils/accum` heavily for reward distribution - critical path
- Has CosmWasm integration for pool hooks (requires wasmd)
- Implements `poolmanager/types.PoolModuleI` interface
- Uses legacy params (x/params) - may need migration to in-module params
- ✅ No direct SDK fork feature usage detected
- ⚠️ Heavy use of `osmoutils/accum` - need to ensure this subpackage works with upstream store

---

### gamm

**Purpose**: Generalized Automated Market Maker - the original pool type in Osmosis. Provides Balancer-style weighted pools and Stableswap pools (Curve-style).

**Key Components** (migrated to Gaia):
- `pool-models/balancer/` - Balancer-style weighted pools with configurable token weights
- `pool-models/stableswap/` - Curve-style stableswap pools optimized for similar-value assets
- `pool-models/internal/cfmm_common/` - Shared CFMM (constant function market maker) logic
- Pool lifecycle: create, join, exit, swap
- GAMM shares (LP tokens) minting/burning
- Governance proposals for pool parameters (scaling factor controller)

**Removed from Gaia Migration** (Osmosis-specific features):
- ~~Migration to concentrated-liquidity pools~~ - CL not in scope
- ~~PoolIncentivesKeeper~~ - Incentives system not in scope
- ~~IncentivesKeeper~~ - Incentives system not in scope
- ~~ConcentratedLiquidityKeeper~~ - CL not in scope
- ~~keeper/migrate.go~~ - CL migration functionality
- ~~simulation/~~ - Simulation framework

**Cosmos SDK Dependencies**:
- `cosmossdk.io/core/appmodule` - App module interface
- `cosmossdk.io/store/types` - KV store types
- `cosmossdk.io/math` - Math types (Int, LegacyDec)
- `github.com/cosmos/cosmos-sdk/codec` - Encoding
- `github.com/cosmos/cosmos-sdk/types` - SDK types
- `github.com/cosmos/cosmos-sdk/x/auth/types` - Module account permissions
- `github.com/cosmos/cosmos-sdk/x/params/types` - Legacy params
- `github.com/cosmos/cosmos-sdk/x/bank/types` - Bank types

**Osmosis Internal Dependencies** (migrated):
- `osmomath` - Math utilities ✅ migrated to `gaia/pkg/osmomath`
- `osmoutils` - General utilities ✅ migrated to `gaia/pkg/osmoutils`
  - Uses root `osmoutils` package (store helpers)
  - Uses `osmoutils/osmocli` (CLI helpers)
  - ✅ Does NOT use `osmoutils/accum` (simpler than CL!)
- `x/poolmanager/types` - Pool interfaces ✅ migrated to `gaia/x/poolmanager/types`
- `app/params` - App parameters ✅ using `gaia/app/params` with `BaseCoinUnit = "uatom"`

**Required External Keepers** (simplified for Gaia):
- `AccountKeeper` - Module account management
- `BankKeeper` - Token transfers, minting LP shares, burning
- `CommunityPoolKeeper` - Community pool funding
- `PoolManager` - Pool routing and creation delegation

**Migration Notes**:
- Well-established module, simpler than concentrated-liquidity
- Two pool types (Balancer, Stableswap) with internal CFMM logic
- Uses osmoutils but NOT the accumulator (simpler migration path)
- Uses legacy x/params (may need migration)
- ✅ No SDK fork features used directly
- ✅ Simpler osmoutils usage - no store fork concerns from accum
- ✅ **Migrated in Task 2.2** - core functionality preserved, incentives removed

---

### cosmwasmpool

**Purpose**: CosmWasm-based pools that allow custom pool logic implemented as smart contracts. Enables extensible pool types like Transmuter (1:1 swaps for similar assets) and orderbook pools.

**Key Components**:
- `model/` - Pool model that wraps a CosmWasm contract address
- `cosmwasm/msg/` - Message types for interacting with pool contracts (sudo, query)
- `bytecode/` - Pre-compiled WASM pool contracts (transmuter, orderbook)
- Pool lifecycle: create (instantiate contract), swap (sudo), query
- Governance: migrate pools to new contract versions, whitelist code IDs
- Transmuter pool: 1:1 swaps for similar-value assets (like stableswap but simpler)

**Cosmos SDK Dependencies**:
- `cosmossdk.io/core/appmodule` - App module interface
- `cosmossdk.io/store/types` - KV store types
- `github.com/cosmos/cosmos-sdk/codec` - Encoding
- `github.com/cosmos/cosmos-sdk/types` - SDK types
- `github.com/cosmos/cosmos-sdk/x/params/types` - Legacy params

**CosmWasm/wasmd Dependencies**:
- `github.com/CosmWasm/wasmd/x/wasm/types` - AccessConfig, ContractInfo
- `github.com/CosmWasm/wasmd/x/wasm/keeper` - Wasm keeper (tests)
- `github.com/CosmWasm/wasmd/x/wasm/ioutils` - File reading (CLI)
- ⚠️ Osmosis uses wasmd v0.53.3, Gaia uses v0.60.2 - API changes expected

**Osmosis Internal Dependencies**:
- `osmomath` - Math utilities ⚠️ MUST MIGRATE FIRST
- `osmoutils` - General utilities ⚠️ MUST MIGRATE FIRST
  - Uses root `osmoutils` package (store helpers)
  - Uses `osmoutils/cosmwasm` (CosmWasm helpers)
  - ✅ Does NOT use `osmoutils/accum` (simpler than CL!)
- `x/poolmanager/types` - Pool interfaces (PoolI, CreatePoolMsg)
- `x/poolmanager/events` - Pool events

**Required External Keepers**:
- `AccountKeeper` - Module address
- `BankKeeper` - Token transfers
- `PoolManagerKeeper` - Pool creation, routing
- `ContractKeeper` - CosmWasm contract execution (Instantiate, Sudo, Execute, Create, Migrate)
- `WasmKeeper` - CosmWasm queries and contract info

**Migration Notes**:
- Requires wasmd integration - Gaia already has this
- wasmd version upgrade v0.53 → v0.60 may have breaking changes
- Ships with pre-compiled WASM bytecode - need to verify compatibility
- Transmuter is a key pool type for similar-asset swaps
- Uses osmoutils but NOT the accumulator (simpler migration path)
- Uses legacy x/params (may need migration)
- ✅ No SDK fork features used directly
- ⚠️ Need to verify wasmd API compatibility between versions

---

### protorev

**Purpose**: MEV (Maximal Extractable Value) arbitrage module that finds and executes arbitrage opportunities across Osmosis pools. Captures value that would otherwise go to MEV searchers and directs it to the protocol.

**Key Components**:
- PostHandler: Executes arb opportunities after every swap transaction
- Route finding: Identifies profitable cyclic arb routes across pool types
- Statistics: Tracks profits, trades, and route performance
- Epoch hook: Periodic route updates and rebalancing
- Developer fees: Profit distribution mechanism
- Transient store: Temporary state during transaction execution

**Cosmos SDK Dependencies**:
- `cosmossdk.io/core/appmodule` - App module interface
- `cosmossdk.io/store/types` - KV store types (including TransientStoreKey)
- `cosmossdk.io/log` - Logging
- `github.com/cosmos/cosmos-sdk/codec` - Encoding
- `github.com/cosmos/cosmos-sdk/types` - SDK types
- `github.com/cosmos/cosmos-sdk/x/params/types` - Legacy params

**Osmosis Internal Dependencies**:
- `osmomath` - Math utilities ⚠️ MUST MIGRATE FIRST
- `osmoutils` - General utilities ⚠️ MUST MIGRATE FIRST
  - Uses root `osmoutils` package (store helpers)
  - ✅ Does NOT use `osmoutils/accum` (simpler than CL!)
- `x/poolmanager/types` - Pool interfaces for routing
- `x/gamm/types` - CFMMPoolI interface
- `x/txfees/types` - Referenced in protobuf (taker fee tracking)
- `x/epochs/types` - Epoch info for periodic updates
- `app/params` - Bond denom

**Required External Keepers**:
- `AccountKeeper` - Module address
- `BankKeeper` - Token transfers, minting profits, burning
- `GAMMKeeper` - Pool and poke operations
- `EpochKeeper` - Epoch info
- `PoolManagerKeeper` - Pool routing, estimates, taker fee tracking
- `ConcentratedLiquidityKeeper` - Max tick calculations for CL pools
- `DistributionKeeper` - Community pool funding

**Migration Notes**:
- Depends on ALL other DEX modules (poolmanager, gamm, concentrated-liquidity)
- Should be migrated LAST among the DEX modules
- PostHandler integration required at app level
- Uses transient store key (standard SDK feature)
- Uses osmoutils but NOT the accumulator (simpler migration path)
- Uses legacy x/params (may need migration)
- ✅ No SDK fork features used directly
- ⚠️ Has txfees dependency (may need to migrate x/txfees or mock it)

---

## Dependency Graph

The dependency graph determines migration order. Start with leaf nodes (no internal dependencies) and work up.

### Architecture Insight: No True Circular Dependencies

The poolmanager ↔ pool modules relationship is **NOT a Go import cycle**:

```
poolmanager/types  ←── defines interfaces (PoolI, PoolModuleI)
        ↑               NO imports from pool modules
        │
        ├── gamm ─────────────────── imports poolmanager/types, implements PoolModuleI
        ├── concentrated-liquidity ── imports poolmanager/types, implements PoolModuleI  
        └── cosmwasmpool ──────────── imports poolmanager/types, implements PoolModuleI

poolmanager/keeper ←── receives pool keepers via dependency injection at app wiring
```

**Key insight**: `poolmanager/types` only defines interfaces and can compile standalone. Pool modules import those types to implement the interfaces. The keeper receives pool module keepers via DI at runtime.

### Recommended Migration Order

```
1. osmomath              ← leaf dependency, no internal deps
2. osmoutils             ← leaf dependency, may use osmomath  
3. poolmanager/types     ← interfaces only, compiles standalone
4. gamm                  ← first pool type (simplest, most established)
5. poolmanager/keeper    ← can now route to gamm
6. concentrated-liquidity ← add next pool type
7. cosmwasmpool          ← add CosmWasm pools (requires wasmd)
8. protorev              ← uses poolmanager for arbitrage routing
```

Each step produces a **compilable, testable unit**. We can run gamm + poolmanager without CL or cosmwasmpool initially, then add pool types incrementally.

---

## Migration Workflow

### Two-Commit Rule (IMPORTANT)

**Every migration must use two commits for reviewability:**

1. **Copy Commit**: Raw copy with NO changes - exact files from Osmosis
2. **Adapt Commit**: ALL changes (imports, API fixes, SDK updates)

**Why**: This allows human reviewers to see EVERYTHING that changed:
```bash
# To review ALL changes (imports + adaptations):
git diff <copy-commit> <adapt-commit> -- path/to/component/
```

The diff shows exactly what we modified - nothing hidden.

**Tracking**: All migrations are documented in `workpads/gaia-migration/progress.md` with:
- Source and target paths
- Copy commit and adapt commit references
- List of adaptations made and why

### Per-Module Migration Steps

1. **Copy** - Copy module from Osmosis to Gaia (no changes)
2. **Commit (Copy)** - Commit raw copy as-is
3. **Update Imports** - Change import paths to Gaia module
4. **Compile** - Attempt to compile, document errors
5. **Adapt** - Fix SDK/IBC API changes
6. **Commit (Adapt)** - Commit all changes for review
7. **Unit Tests** - Run tests, fix failures
8. **Integrate** - Wire module into Gaia app
9. **Integration Tests** - Run or write integration tests
10. **Manual Tests** - Test with local node

> **Note**: Track all work in `progress.md` for reviewability.

### Workflow Evolution

This workflow will be **refined iteratively** as we migrate the first few modules. Once the first few modules are migrated, the workflow should be clear and explicit enough to be independently repeatable.

---

## Testing Strategy

Testing is done at **three levels**, each with a distinct purpose:

### Level 1: Unit Tests

| Aspect | Details |
|--------|---------|
| **Source** | Migrated from Osmosis |
| **Purpose** | Verify individual module logic works correctly |
| **Why Important** | Catches regressions in core module functionality; ensures the adaptation to the new SDK didn't break internal logic |
| **When to Run** | After module compiles, before integration |

### Level 2: Integration / E2E Tests

| Aspect | Details |
|--------|---------|
| **Source** | New tests, focused on user workflows |
| **Purpose** | Verify cross-module behavior and real user scenarios |
| **Why Important** | Unit tests can pass while integration fails; this catches issues in how modules interact with each other and with Gaia's existing modules |
| **When to Run** | After module is wired into Gaia app |

### Level 3: Manual Tests

| Aspect | Details |
|--------|---------|
| **Source** | Run a local node with mainnet data, interact with scripts |
| **Purpose** | Validate production readiness with realistic conditions |
| **Why Important** | Integration tests use synthetic data; manual tests with real mainnet data catch edge cases, state migration issues, and performance problems that only appear with production-scale data |
| **When to Run** | Final validation before considering a module "done" |

### Testing Harness

#### Level 1: Unit Test Migration

**Osmosis Pattern** (to migrate from):
- Uses `apptesting.KeeperTestHelper` from `app/apptesting/` package
- Full `OsmosisApp` instance with all keepers wired
- Suite-based tests using `testify/suite`
- Helper methods for pool creation, funding, etc.

**Gaia Pattern** (to migrate to):
- Uses `github.com/cosmos/cosmos-sdk/testutil/integration` package
- Lightweight `integration.App` with only necessary modules
- Fixture pattern with explicit keeper setup
- Follows SDK 0.53 conventions

**Migration Approach**:
1. Create a DEX test helper in Gaia: `tests/dex/test_common.go`
2. Build a fixture that wires DEX modules with minimal dependencies
3. Port Osmosis test helpers (pool creation, swap execution, etc.)
4. Migrate tests file by file, adapting import paths and context patterns

**Example Fixture Structure** (to be created):
```go
// tests/dex/test_common.go
type dexFixture struct {
    app        *integration.App
    sdkCtx     sdk.Context
    cdc        codec.Codec
    
    // Standard keepers
    accountKeeper authkeeper.AccountKeeper
    bankKeeper    bankkeeper.Keeper
    
    // DEX keepers (migrated)
    poolManagerKeeper poolmanagerkeeper.Keeper
    gammKeeper        gammkeeper.Keeper
    clKeeper          clkeeper.Keeper
    cosmwasmPoolKeeper cosmwasmpoolkeeper.Keeper
    protorevKeeper    protorevkeeper.Keeper
}
```

**Commands**:
```bash
# Run unit tests for a specific module
cd /Users/nicolas/devel/gaia
go test ./x/{module}/... -v

# Run all DEX tests
go test ./tests/dex/... -v
```

#### Level 2: Integration Tests

**Gaia Pattern** (existing):
- Located in `tests/integration/`
- Uses `integration.NewIntegrationApp()` from SDK
- Tests full message flows (MsgServer calls)
- Verifies keeper interactions work correctly

**DEX Integration Tests** (to create):
```
tests/integration/
├── test_common.go       # Existing fixture
├── dex_test.go          # NEW: Pool creation, swaps, routing
├── dex_arbitrage_test.go # NEW: Protorev MEV tests
└── ...
```

**Test Scenarios to Cover**:
1. **Pool Lifecycle**: Create Balancer → Create Stableswap → Create CL pool
2. **Swap Routing**: Multi-hop swaps through different pool types
3. **Liquidity Operations**: Add/remove liquidity, position management
4. **Protorev**: Verify arbitrage detection and execution
5. **Genesis**: Export → Import round-trip for all pool types

**Commands**:
```bash
# Run integration tests
cd /Users/nicolas/devel/gaia
go test ./tests/integration/... -v -run TestDex

# Run specific test
go test ./tests/integration/... -v -run TestDexSwapRouting
```

#### Level 3: E2E / Manual Tests

**E2E Tests** (Docker-based):
- Gaia has existing e2e framework in `tests/e2e/`
- Uses `dockertest` to spin up full chains
- Includes IBC relayer (Hermes) setup
- Add DEX-specific e2e tests: `tests/e2e/e2e_dex_test.go`

**Manual Testing** (LocalGaia with DEX):
- Similar to Osmosis's `localosmosis` setup
- Use Gaia's `contrib/single-node.sh` as base
- Create `tests/localgaia-dex/` with:
  - Docker compose setup
  - Pre-funded test accounts
  - Sample pool creation scripts
  - Swap test scripts

**Local Testing Setup**:
```bash
# Build Gaia binary with DEX modules
cd /Users/nicolas/devel/gaia
make build

# Start local node (simple, single validator)
./contrib/single-node.sh

# Or use docker-compose for more control
cd tests/localgaia-dex
docker-compose up
```

**Test Scripts to Create**:
```
tests/localgaia-dex/
├── docker-compose.yml
├── README.md
├── scripts/
│   ├── create_balancer_pool.sh
│   ├── create_cl_pool.sh
│   ├── create_stableswap_pool.sh
│   ├── execute_swap.sh
│   ├── test_multihop_swap.sh
│   └── test_protorev.sh
└── data/
    ├── pool_definitions/
    │   ├── balancer.json
    │   ├── stableswap.json
    │   └── cl.json
    └── test_accounts.txt
```

**Realistic Data Testing**:

Create `tests/localgaia-dex/fixtures/` with hand-crafted genesis state:
- 3-5 Balancer pools (different token pairs, weights)
- 2-3 Stableswap pools  
- 2-3 CL pools with positions at various price ranges
- Sample swap routes for multi-hop testing

**Validation approach**: Run operations on localgaia-dex and verify:
- Operations succeed/fail as expected
- Invariants hold (balances conserved, pool state consistent)
- Outputs match expectations (swap amounts, fees, slippage)

No need to compare with Osmosis - just validate correctness against defined expectations

#### Testing Priority Order

| Priority | Test Type | When to Run | Purpose |
|----------|-----------|-------------|---------|
| 1 | Unit tests | After each module compiles | Verify logic correctness |
| 2 | Integration tests | After keeper wiring | Verify cross-module interactions |
| 3 | E2E tests | After app integration | Verify full stack works |
| 4 | Manual tests | Final validation | Catch edge cases |

#### Test Infrastructure Files to Create

| File | Description | Priority |
|------|-------------|----------|
| `tests/dex/test_common.go` | DEX test fixture and helpers | P1 |
| `tests/dex/pool_helpers.go` | Pool creation/manipulation helpers | P1 |
| `tests/integration/dex_test.go` | Integration tests for DEX | P2 |
| `tests/e2e/e2e_dex_test.go` | E2E tests for DEX | P3 |
| `tests/localgaia-dex/` | Local test environment | P3 |

---

## SDK Version Differences

| Aspect | Osmosis | Gaia |
|--------|---------|------|
| **SDK Version** | v0.50.14 (fork: `osmosis-labs/cosmos-sdk v0.50.14-v30-osmo`) | v0.53.4 |
| **Go Version** | 1.23.4 | 1.24.0 |
| **IBC-go** | v8.7.0 | v10.5.0 |
| **CosmWasm/wasmd** | v0.53.3 | v0.60.2 |
| **CometBFT** | v0.38.21 | v0.38.21 (same) |
| **cosmossdk.io/core** | v0.11.0 (via replace) | v0.11.3 |
| **cosmossdk.io/store** | fork: `v1.1.1-v0.50.11-v28-osmo-2` | v1.1.2 |
| **cosmossdk.io/x/tx** | v0.13.7 | v0.14.0 |
| **cosmossdk.io/collections** | v0.4.0 | v1.2.1 |
| **cosmossdk.io/depinject** | v1.1.0 | v1.2.1 |

### Key Breaking Changes (SDK 0.50 → 0.53)

1. **collections v0.4 → v1.2**: Major version bump with likely API changes
2. **IBC-go v8 → v10**: Major IBC version upgrade; middleware and handler signatures may change
3. **CosmWasm v0.53 → v0.60**: Significant wasmd upgrade; contract ABI may differ
4. **Go 1.23 → 1.24**: Minor upgrade but may affect some dependencies

### Osmosis-Specific SDK Modifications

Osmosis uses forked versions with custom features:
- **cosmos-sdk**: Fork with custom changes (`osmosis-labs/cosmos-sdk v0.50.14-v30-osmo`)
- **store**: Fork with "iavlFastNodeModuleWhitelist" and async pruning features
- **block-sdk**: Fork from Skip protocol (`osmosis-labs/block-sdk/v2 v2.1.9-mempool`)

### Migration Impact Assessment

| Impact Level | Area |
|--------------|------|
| **HIGH** | SDK 0.50 → 0.53 requires updating module patterns (keeper, msg server, etc.) |
| **HIGH** | IBC v8 → v10 requires rewriting IBC integrations |
| **HIGH** | Osmosis SDK fork features may not exist in upstream SDK 0.53 |
| **MEDIUM** | CosmWasm integration changes |
| **LOW** | Go version upgrade (1.23 → 1.24) |

### Key API Differences Expected

1. **Module Registration**: SDK 0.53 uses enhanced depinject patterns
2. **Keeper Constructors**: May require additional context parameters
3. **Msg Server**: May have updated response patterns
4. **IBC Middleware**: v10 has different callback signatures
5. **Collections API**: v1.x has different iteration and storage patterns

---

## SDK Fork Features Analysis

### Summary

**The DEX modules do NOT require any Osmosis SDK fork features.** The fork features are used by other modules that are outside our migration scope.

### Osmosis SDK Fork (osmo-v53/0.53.4)

The Osmosis SDK fork has only **2 custom commits** on top of upstream SDK 0.53.4:

1. `76dc4a4d65 Add Osmosis bank hooks and supply offsets`
2. `523350f081 Add supply offset accessors to bank keeper interface`

These commits add:
- **Bank Hooks**: `TrackBeforeSend`, `BlockBeforeSend` hooks for tracking/blocking token transfers
- **Supply Offsets**: `GetSupplyOffset`, `AddSupplyOffset` for virtual supply tracking

### Modules That USE Fork Features

| Module | Fork Feature | Purpose |
|--------|-------------|---------|
| `x/tokenfactory` | Bank Hooks | Track/block transfers of factory tokens |
| `x/superfluid` | Supply Offsets | Virtual bonded token supply |
| `x/mint` | Supply Offsets | Epoch provisions offset |

### DEX Modules: Fork Feature Usage

| DEX Module | Uses Bank Hooks? | Uses Supply Offsets? |
|------------|------------------|---------------------|
| `poolmanager` | ❌ No | ❌ No |
| `concentrated-liquidity` | ❌ No | ❌ No |
| `gamm` | ❌ No | ❌ No |
| `cosmwasmpool` | ❌ No | ❌ No |
| `protorev` | ❌ No | ❌ No |

**Conclusion**: ✅ DEX modules can use upstream SDK 0.53 without any fork features.

### Store Fork Analysis

The osmoutils go.mod has a replace directive for the Osmosis store fork, which provides:
- `iavlFastNodeModuleWhitelist` - Performance optimization for syncing
- Async pruning - Performance optimization for snapshot nodes

**Key Finding**: osmoutils does NOT use any fork-specific APIs. It uses standard store operations:
- `store.Get()`, `store.Set()`, `store.Delete()`, `store.Has()`, `store.Iterator()`

These are identical in upstream SDK store. The fork only provides **performance optimizations at the node level**, not different functionality.

**Conclusion**: ✅ Can use upstream SDK store. Minor performance differences possible but functionally equivalent.

### References to Non-Migrated Modules

Some DEX module files reference modules we're NOT migrating (superfluid, tokenfactory, mint). These are **not blockers** - they relate to features that integrate with those modules:

| Reference | Location | Why Not a Blocker |
|-----------|----------|-------------------|
| `superfluidtypes.MigrationPoolIDs` | `gamm/keeper/migrate.go` | Superfluid migration feature - exclude or define struct locally |
| superfluid import | `concentrated-liquidity/pool_test.go` | Test for superfluid integration - exclude test |
| tokenfactory import | `cosmwasmpool/.../transmuter_test.go` | Test for tokenfactory integration - exclude test |
| mint import | `concentrated-liquidity/simulation/sim_msgs.go` | Simulation code - exclude or adapt |

**Bottom line**: These are integration points with modules outside our scope. They can be excluded, disabled, or stubbed without affecting core DEX functionality.

---

## Risks and Watch Items

| Risk | Description | Mitigation |
|------|-------------|------------|
| SDK version mismatch | Osmosis uses a different SDK version; modules won't compile directly | Document API differences, adapt per module |
| Hidden dependencies | Modules may depend on Osmosis-specific utilities | Map all dependencies before copying |
| State migration | Genesis/state format may differ | Test export/import paths |
| Test coverage gaps | Some Osmosis tests may not transfer cleanly | Write new integration tests |

---

## Decision Log

| D# | Decision | Rationale | Date |
|----|----------|-----------|------|
| D1 | Migrate to Gaia (not fork) | Align with ecosystem, reduce maintenance | 2026-01-28 |
| D2 | Start with simplest dependencies | Build confidence, establish workflow | 2026-01-28 |
| D3 | Add epoch hooks to poolmanager for fee distribution (not migrate txfees) | txfees has unnecessary complexity (EIP-1559, fee tokens); core distribution logic is simple; swap non-native to ATOM before distributing | 2026-01-28 |

---

## Open Questions

1. ~~What is the exact SDK version difference between Osmosis and Gaia?~~ ✅ Answered: SDK 0.50.14 (Osmosis fork) → 0.53.4 (Gaia)
2. Are there shared utility packages that need to migrate first (e.g., `osmomath`, `osmoutils`)?
3. What state/genesis migration is needed for each module?
4. How do we handle CosmWasm integration differences (wasmd v0.53 → v0.60)?
5. **NEW**: How do we handle the IBC v8 → v10 migration for modules that use IBC?
6. ~~**CRITICAL**: What Osmosis SDK fork features are required by the DEX modules?~~ ✅ Answered: **None!** DEX modules don't use bank hooks or supply offsets. See "SDK Fork Features Analysis" section.

---

## Lessons Learned

### L1: osmomath Migration (Task 1.1)

**What worked well**:
- Package was truly a leaf dependency - no internal Osmosis imports
- Gaia's `cosmossdk.io/math v1.5.3` is compatible with osmomath (which used v1.4.0)
- No SDK fork replace directive needed - Gaia's module already has correct deps
- All 23 source files copied without modification (except 2 test import paths)

**Key insight**: The SDK version upgrade (0.50 → 0.53) had no impact on osmomath because it only uses `cosmossdk.io/math` types (Int, LegacyDec), which are stable across versions.

**Location decision**: Placed in `pkg/osmomath/` following Gaia's existing pattern for shared packages (e.g., `pkg/address/`).

### L2: osmoutils Migration (Task 1.2)

**What worked well**:
- Bulk copy of subpackages then sed-based import replacement worked efficiently
- Most code compiled without changes after import updates
- Tests pass with minimal modifications

**IBC v10 API changes discovered**:
- `DenomTrace` type replaced with `Denom` type
- `denomTrace.Path` (field) → `denom.Path()` (method)
- `ParseDenomTrace` is deprecated, prefer `ExtractDenomFromPath`

**Unexpected additions**:
- `noapptest/` needed for test context creation
- `wrapper/` needed for accum tests (database wrapper for IAVL)

**Key insight**: The IBC v8→v10 migration has breaking changes in the transfer types. Always check method signatures when upgrading IBC versions.

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial document creation | AI Assistant |
| 2026-01-28 | Documented SDK version differences (Task 0.1) - major gap: SDK 0.50→0.53, IBC v8→v10 | AI Assistant |
| 2026-01-28 | Enhanced overview, testing strategy with detailed rationale for each level | AI Assistant |
| 2026-01-28 | Documented poolmanager dependencies; confirmed NO true circular dependency | AI Assistant |
| 2026-01-28 | Added recommended migration order based on dependency analysis | AI Assistant |
| 2026-01-28 | Completed osmomath analysis - confirmed TRUE LEAF dependency | AI Assistant |
| 2026-01-28 | Completed osmoutils analysis - depends on osmomath, uses store fork | AI Assistant |
| 2026-01-28 | Documented concentrated-liquidity dependencies - uses osmoutils/accum heavily | AI Assistant |
| 2026-01-28 | Documented gamm dependencies - simpler than CL, no accum usage | AI Assistant |
| 2026-01-28 | Documented cosmwasmpool dependencies - requires wasmd v0.53→v0.60 | AI Assistant |
| 2026-01-28 | Documented protorev dependencies - depends on all DEX modules, migrate last | AI Assistant |
| 2026-01-28 | Analyzed x/epochs - SDK 0.53 version can be used, minor hook adaptations needed | AI Assistant |
| 2026-01-28 | **SDK Fork Analysis Complete** - DEX modules do NOT require fork features (bank hooks/supply offsets used by tokenfactory/superfluid/mint only) | AI Assistant |
| 2026-01-28 | Minimal osmoutils subset identified - 6 subpackages needed, all use standard store.KVStore interface | AI Assistant |
| 2026-01-28 | Added Executive Summary and detailed Migration Plan based on Phase 0 findings | AI Assistant |
| 2026-01-28 | Testing Harness documented - 3-level strategy with fixture patterns, commands, and file structure | AI Assistant |
| 2026-01-28 | **D3: Taker Fee Distribution** - decided to add epoch hooks to poolmanager instead of migrating txfees; swap non-native fees to ATOM | AI Assistant |
| 2026-01-28 | **Scope Update**: Added incentives, pool-incentives, lockup to "Modules Outside Scope". DEX modules simplified by removing these dependencies. | AI Assistant |
| 2026-01-28 | **gamm Migration Complete** - core functionality preserved, incentives/CL migration removed. Simplified keeper to accountKeeper, bankKeeper, communityPoolKeeper only. | AI Assistant |
