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

## Modules to Migrate

| Module | Description | Status |
|--------|-------------|--------|
| `poolmanager` | | 📋 pending |
| `concentrated-liquidity` | | 📋 pending |
| `gamm` | | 📋 pending |
| `cosmwasmpool` | | 📋 pending |
| `protorev` | | 📋 pending |

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
- `cosmossdk.io/store` - **⚠️ Uses Osmosis fork** for iavlFastNodeModuleWhitelist
- `cosmossdk.io/math`, `log` - Standard cosmossdk.io packages
- `github.com/cosmos/cosmos-sdk` - Uses fork via replace
- `github.com/cosmos/ibc-go/v8` - **Needs update to v10** for Gaia
- `github.com/CosmWasm/wasmvm/v2` - For CosmWasm helpers
- `github.com/cometbft/cometbft` - Uses Osmosis fork

**Internal Dependencies**:
- `osmomath` - Depends on osmomath (osmomath is the true leaf)

**Migration Notes**:
- Standalone Go module with own `go.mod`
- **Store fork is a potential blocker** - uses iavlFastNodeModuleWhitelist for sync performance
- Must remove all replace directives and use upstream dependencies
- IBC-go v8 → v10 update required
- SDK v0.50 → v0.53 update required

**Key Insight: Partial Migration**:
- We do NOT need to migrate all of osmoutils
- Only migrate the specific utilities used by our target DEX modules
- If the store fork features are only used by parts we don't need, we can skip them
- Need to identify exactly which osmoutils subpackages each DEX module imports

---

## Module Descriptions

### poolmanager

**Purpose**: Central router for all pool types. Manages pool creation, routing swaps across different pool types, taker fees, and provides a unified interface for interacting with any pool (GAMM, Concentrated Liquidity, CosmWasm pools).

**Key Components**:
- Pool routing and swap execution
- Taker fee management and distribution
- Multi-hop swap routing
- Pool creation delegation to specific pool modules
- Governance proposals for pool management

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
- `x/txfees/types` - Transaction fees types

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

**Migration Notes**:
- ✅ **No true circular dependency**: `poolmanager/types` defines interfaces only and does not import pool modules. Pool modules import `poolmanager/types` to implement interfaces. Keepers are wired via DI.
- Depends on osmomath and osmoutils which must be migrated first
- Can migrate `poolmanager/types` → `gamm` → `poolmanager/keeper` → other pools incrementally
- No direct SDK fork feature usage detected

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

**Purpose**: _(to be documented)_

**Key Components**:
- _(to be documented)_

**External Dependencies**:
- _(to be documented)_

**Internal Dependencies**:
- _(to be documented)_

---

### cosmwasmpool

**Purpose**: _(to be documented)_

**Key Components**:
- _(to be documented)_

**External Dependencies**:
- _(to be documented)_

**Internal Dependencies**:
- _(to be documented)_

---

### protorev

**Purpose**: _(to be documented)_

**Key Components**:
- _(to be documented)_

**External Dependencies**:
- _(to be documented)_

**Internal Dependencies**:
- _(to be documented)_

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

### Per-Module Migration Steps

1. **Copy** - Copy module from Osmosis to Gaia
2. **Compile** - Attempt to compile in Gaia, document all errors
3. **Adapt** - Update module to match Gaia SDK version and patterns
4. **Verify Compile** - Ensure clean compilation with no errors
5. **Unit Tests** - Run migrated unit tests, review and fix failures
6. **Integrate** - Wire module into Gaia app initialization
7. **Integration Tests** - Run existing integration tests; if none exist, write them
8. **Manual Tests** - Run a local node with realistic data, test with scripts

> **Note**: There may be an intermediate step before manual tests where we create realistic test data.

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

We will build a testing harness to iterate efficiently:
- Automated unit test execution per module
- Integration test framework for user workflow validation
- Local node setup with mainnet data import capability

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

---

## Open Questions

1. ~~What is the exact SDK version difference between Osmosis and Gaia?~~ ✅ Answered: SDK 0.50.14 (Osmosis fork) → 0.53.4 (Gaia)
2. Are there shared utility packages that need to migrate first (e.g., `osmomath`, `osmoutils`)?
3. What state/genesis migration is needed for each module?
4. How do we handle CosmWasm integration differences (wasmd v0.53 → v0.60)?
5. **NEW**: How do we handle the IBC v8 → v10 migration for modules that use IBC?
6. **CRITICAL**: What Osmosis SDK fork features are required by the DEX modules, and are they available in upstream SDK 0.53? (See Task 0.1a)

---

## Lessons Learned

_(to be populated during migration)_

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
