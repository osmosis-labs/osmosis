# Gaia Migration Tasks

## Status Legend

```
📋 pending      - Not yet started
🚧 in_progress  - Currently working on  
✅ completed    - Finished and verified
🚫 blocked      - Cannot proceed
```

---

## Phase 0: Discovery & Planning

### Task 0.1: Document SDK Version Differences ✅ `completed`

**Description**: Compare Osmosis and Gaia SDK versions and document key API differences that will affect migration.

**Acceptance Criteria**:
- [x] Osmosis SDK version documented (v0.50.14 fork)
- [x] Gaia SDK version documented (v0.53.4)
- [x] Key breaking changes between versions identified (SDK 0.50→0.53, IBC v8→v10, CosmWasm v0.53→v0.60)
- [x] Update `knowledge.md` with findings

---

### Task 0.1a: Identify Required SDK Fork Features ✅ `completed`

**Depends On**: Tasks 0.2-0.6 (need dependency analysis first to know what SDK features modules use)

**Description**: Analyze which Osmosis SDK fork features are used by the DEX modules and determine if they are available in upstream SDK 0.53. This is critical to assess early as it may fundamentally affect our migration approach.

**Why Important**: If the DEX modules depend on Osmosis-specific SDK fork features that don't exist in upstream SDK 0.53, we have several options:
1. Port those features to Gaia (adds complexity)
2. Refactor modules to not need those features (may be significant work)
3. Contribute missing features upstream (long-term, unlikely for this project timeline)

**Acceptance Criteria**:
- [x] List all Osmosis SDK fork modifications (from `osmosis-labs/cosmos-sdk v0.50.14-v30-osmo`)
- [x] For each fork modification, identify if DEX modules depend on it
- [x] For each required fork feature, check if equivalent exists in SDK 0.53
- [x] Document blockers or risks in `knowledge.md`
- [x] Recommend approach for each missing feature

**Key Findings**:
- ✅ **NO BLOCKERS** - DEX modules do NOT use any SDK fork features
- SDK fork has only 2 commits: bank hooks + supply offsets
- These features are used by tokenfactory, superfluid, mint (NOT our DEX modules)
- Store fork is for performance only; osmoutils uses standard store APIs
- Minor: `gamm/migrate.go` uses `superfluidtypes.MigrationPoolIDs` - trivial struct, can be moved
- **Recommendation**: Use upstream SDK 0.53 without modifications

---

### Task 0.1b: Analyze osmomath Dependencies ✅ `completed`

**Description**: Map all dependencies of the `osmomath` package. This is a leaf dependency that must be migrated first.

**Why Important**: osmomath is used by all DEX modules for mathematical operations. It must compile in Gaia before any module can be migrated.

**Acceptance Criteria**:
- [x] List all external imports (cosmos-sdk, cosmossdk.io, third-party)
- [x] Confirm no Osmosis-internal dependencies (should be a leaf)
- [x] Identify any SDK version-specific APIs that may need adaptation
- [x] Update `knowledge.md` with findings

**Key Findings**:
- ✅ TRUE LEAF - no Osmosis-internal dependencies
- Standalone Go module with own go.mod
- Uses `cosmossdk.io/math` v1.4.0 (need v1.5.3 for Gaia)
- Has SDK fork replace directive that must be removed
- Should compile with minimal version updates

---

### Task 0.1c: Analyze osmoutils Dependencies ✅ `completed`

**Description**: Map all dependencies of the `osmoutils` package. This is a leaf dependency that must be migrated first.

**Why Important**: osmoutils provides general utilities used across DEX modules. It must compile in Gaia before any module can be migrated.

**Acceptance Criteria**:
- [x] List all external imports (cosmos-sdk, cosmossdk.io, third-party)
- [x] Check if it depends on osmomath (would make osmomath the true leaf)
- [x] Identify any SDK version-specific APIs that may need adaptation
- [x] Update `knowledge.md` with findings

**Key Findings**:
- Depends on osmomath → confirms osmomath is true leaf
- ⚠️ Uses store fork (iavlFastNodeModuleWhitelist) - potential blocker
- Uses IBC-go v8 (needs v10 for Gaia)
- Has multiple replace directives for SDK, CometBFT, store
- **Important**: We only need to migrate the osmoutils subpackages actually used by DEX modules, not the entire package

---

### Task 0.1d: Investigate Store Fork Requirement ✅ `completed`

**Description**: The osmoutils package uses an Osmosis store fork with `iavlFastNodeModuleWhitelist` feature. Determine if this is required for the DEX modules or if upstream SDK v0.53 store can be used.

**Why Important**: If the store fork is required, we may need to:
1. Port the feature to Gaia's store
2. Find an equivalent in SDK v0.53
3. Accept potential performance impact without it

**Acceptance Criteria**:
- [x] Understand what iavlFastNodeModuleWhitelist does
- [x] Identify which modules/features depend on it
- [x] Check if SDK v0.53 has equivalent functionality
- [x] Recommend approach: use upstream store or port feature
- [x] Update `knowledge.md` with findings

**Key Findings** (from Task 0.1a analysis):
- Store fork provides **performance optimizations only** (fast node whitelist, async pruning)
- osmoutils uses only standard store APIs (Get, Set, Delete, Has, Iterator)
- No store fork-specific APIs are called by DEX modules
- **Recommendation**: Use upstream SDK store - functionality is identical, minor performance difference acceptable

---

### Task 0.1e: Compare Tokenfactory Implementations ✅ `completed`

**Description**: Gaia has its own tokenfactory module. If any Osmosis DEX modules depend on tokenfactory, we need to compare the Osmosis and Gaia implementations to understand compatibility and determine if we can use Gaia's native implementation.

**Why Important**: Tokenfactory in Osmosis uses bank hooks and supply offsets (the SDK fork features). If DEX modules depend on tokenfactory, we need to:
1. Understand if Gaia's tokenfactory has equivalent functionality
2. Identify any API differences that would require adaptation
3. Determine if Gaia's implementation can serve as a drop-in replacement

**Acceptance Criteria**:
- [x] Identify which DEX modules (if any) import or depend on tokenfactory
- [ ] ~~Document Osmosis tokenfactory's key APIs and features~~ (not needed)
- [ ] ~~Document Gaia tokenfactory's key APIs and features~~ (not needed)
- [ ] ~~Compare implementations and identify differences~~ (not needed)
- [ ] ~~Determine if Gaia's tokenfactory can be used as-is~~ (not needed)
- [x] Update `knowledge.md` with findings and recommendations

**Key Findings** (from Task 0.1a analysis):
- ✅ **DEX modules do NOT depend on tokenfactory**
- Only reference: `cosmwasmpool/.../transmuter_test.go` (test file only, not production)
- **No comparison needed** - tokenfactory is outside our migration scope
- Gaia's native tokenfactory is unaffected by this migration

---

### Task 0.1f: Analyze x/epochs Dependency ✅ `completed`

**Description**: The `x/epochs` module is used by gamm and protorev (and other Osmosis modules). SDK 0.53 has its own `x/epochs` module. Determine if we can use the SDK version or need to port Osmosis's version.

**Why Important**: Epochs provides time-based hooks that trigger periodic operations. Multiple DEX modules depend on it:
- `gamm` - uses EpochKeeper for epoch info
- `protorev` - uses EpochKeeper + epoch hooks for periodic updates

**Acceptance Criteria**:
- [x] Analyze Osmosis x/epochs API (EpochInfo type, hooks interface)
- [x] Analyze SDK 0.53 x/epochs API
- [x] Compare the two implementations
- [x] Determine if SDK epochs can be used as drop-in replacement
- [x] Document any API differences that need adaptation
- [x] Update `knowledge.md` with findings and recommendations

**Key Findings**:
- ✅ EpochInfo type is wire-compatible (identical proto fields)
- ✅ SDK 0.53 x/epochs can be used as replacement
- Minor hook interface changes needed:
  - `sdk.Context` → `context.Context`
  - Remove `GetModuleName()` method
- Osmosis uses osmoutils panic recovery; SDK uses standard errors
- **Recommendation**: Use SDK 0.53 x/epochs, adapt hook implementations

---

### Task 0.2: Analyze poolmanager Dependencies ✅ `completed`

**Description**: Map all internal and external dependencies of the `poolmanager` module.

**Acceptance Criteria**:
- [x] List all Osmosis-internal imports
- [x] List all cosmos-sdk imports
- [x] List all third-party imports
- [x] Identify which dependencies need to migrate first
- [x] Update `knowledge.md` with module description and dependencies

**Key Findings**:
- Must migrate `osmomath` and `osmoutils` first
- Has circular dependencies with gamm, CL, and cosmwasmpool
- No SDK fork features used directly

---

### Task 0.3: Analyze concentrated-liquidity Dependencies ✅ `completed`

**Description**: Map all internal and external dependencies of the `concentrated-liquidity` module.

**Acceptance Criteria**:
- [x] List all Osmosis-internal imports
- [x] List all cosmos-sdk imports  
- [x] List all third-party imports
- [x] Identify which dependencies need to migrate first
- [x] Update `knowledge.md` with module description and dependencies

**Key Findings**:
- Largest/most complex DEX module (~60 source files)
- Depends on: `osmomath`, `osmoutils` (including `accum`), `poolmanager/types`, `lockup/types`
- Uses keepers: GAMMKeeper, PoolIncentivesKeeper, IncentivesKeeper, LockupKeeper, ContractKeeper
- Heavy use of `osmoutils/accum` for reward distribution (critical path)
- Has CosmWasm pool hooks integration
- ✅ No SDK fork features used directly
- ⚠️ Uses legacy x/params (may need migration)

---

### Task 0.4: Analyze gamm Dependencies ✅ `completed`

**Description**: Map all internal and external dependencies of the `gamm` module.

**Acceptance Criteria**:
- [x] List all Osmosis-internal imports
- [x] List all cosmos-sdk imports
- [x] List all third-party imports
- [x] Identify which dependencies need to migrate first
- [x] Update `knowledge.md` with module description and dependencies

**Key Findings**:
- Two pool models: Balancer (weighted) and Stableswap (curve-style)
- Depends on: `osmomath`, `osmoutils` (root, osmocli), `poolmanager/types`, `concentrated-liquidity/types`
- Also: `incentives/types`, `pool-incentives/types`, `epochs/types`
- ✅ Does NOT use `osmoutils/accum` - simpler than CL
- Has migration feature to CL pools (bidirectional dependency)
- ✅ No SDK fork features used directly

---

### Task 0.5: Analyze cosmwasmpool Dependencies ✅ `completed`

**Description**: Map all internal and external dependencies of the `cosmwasmpool` module.

**Acceptance Criteria**:
- [x] List all Osmosis-internal imports
- [x] List all cosmos-sdk imports
- [x] List all third-party imports
- [x] Identify which dependencies need to migrate first
- [x] Update `knowledge.md` with module description and dependencies

**Key Findings**:
- CosmWasm-based pools (Transmuter, orderbook) via smart contracts
- Depends on: `osmomath`, `osmoutils` (root, cosmwasm), `poolmanager/types`
- ⚠️ Requires wasmd integration (v0.53 → v0.60 upgrade)
- Ships with pre-compiled WASM bytecode
- ✅ Does NOT use `osmoutils/accum` - simpler than CL
- ✅ No SDK fork features used directly
- Gaia already has wasmd - check API compatibility

---

### Task 0.6: Analyze protorev Dependencies ✅ `completed`

**Description**: Map all internal and external dependencies of the `protorev` module.

**Acceptance Criteria**:
- [x] List all Osmosis-internal imports
- [x] List all cosmos-sdk imports
- [x] List all third-party imports
- [x] Identify which dependencies need to migrate first
- [x] Update `knowledge.md` with module description and dependencies

**Key Findings**:
- MEV arbitrage module - finds and executes arb opportunities
- Depends on ALL DEX modules: `poolmanager`, `gamm`, `concentrated-liquidity`
- Also: `osmomath`, `osmoutils`, `epochs`, `txfees` (proto reference)
- Uses PostHandler for transaction-level arb execution
- ✅ Does NOT use `osmoutils/accum` - simpler than CL
- ✅ No SDK fork features used directly
- Should be migrated LAST (depends on all other DEX modules)

---

### Task 0.7: Build Dependency Graph ✅ `completed`

**Depends On**: Tasks 0.2-0.6

**Description**: Create a dependency DAG showing migration order from simplest to most complex.

**Acceptance Criteria**:
- [x] Dependency graph documented in `knowledge.md`
- [x] Migration order determined (leaf nodes first)
- [x] Shared utilities (osmomath, osmoutils) positioned in graph

**Key Findings**:
- Graph and 8-step migration order documented in knowledge.md § Dependency Graph
- No true circular dependencies - poolmanager/types defines interfaces only
- Migration order: osmomath → osmoutils → poolmanager/types → gamm → poolmanager/keeper → CL → cosmwasmpool → protorev

---

### Task 0.7a: Determine Minimal osmoutils Subset ✅ `completed`

**Depends On**: Tasks 0.2-0.6, 0.7

**Description**: After completing all module dependency analyses, review which osmoutils subpackages are actually imported by our target DEX modules. Determine the minimal subset needed and identify if store fork features can be avoided.

**Why Important**: osmoutils has store fork dependencies that may be blockers. If we can avoid importing those subpackages, migration becomes much simpler.

**Acceptance Criteria**:
- [x] List all osmoutils subpackages imported by each DEX module
- [x] Identify which subpackages use store fork features
- [x] Determine if we can avoid store fork by using minimal subset
- [x] Update migration plan based on findings
- [x] Update `knowledge.md` with minimal osmoutils requirements

**Key Findings**:
- 6 subpackages needed: root, accum, osmocli, osmoassert, cosmwasm, observability
- 5 subpackages can be excluded: sumtree, coinutil, partialord, noapptest, wrapper
- ✅ **ALL required subpackages use standard store.KVStore interface only**
- ✅ **Store fork NOT required** - upstream SDK store will work
- `osmoutils/accum` is critical for concentrated-liquidity (spread rewards, incentives)

---

### Task 0.8: Define Testing Harness ✅ `completed`

**Description**: Design the three-level testing strategy and document setup requirements.

**Acceptance Criteria**:
- [x] Unit test migration approach documented
- [x] Integration test framework chosen
- [x] Manual test setup documented (local node + mainnet data)
- [x] Update `knowledge.md` with testing strategy

**Key Findings**:
- **Unit Tests**: Migrate from Osmosis `apptesting.KeeperTestHelper` to SDK's `integration.App` fixture pattern
- **Integration Tests**: Use Gaia's existing `tests/integration/` pattern with `cosmos-sdk/testutil/integration`
- **E2E Tests**: Extend Gaia's Docker-based `tests/e2e/` framework for DEX testing
- **Manual Tests**: Create `tests/localgaia-dex/` similar to Osmosis's localosmosis
- Test infrastructure files to create: `tests/dex/test_common.go`, `tests/integration/dex_test.go`, etc.

---

## Phase 1: Foundation Migration

### Task 1.1: Migrate osmomath 📋 `pending`

**Description**: Migrate the `osmomath` package to Gaia. This is the true leaf dependency with no Osmosis-internal imports.

**Workflow**: Copy → Compile → Adapt → Test

**Acceptance Criteria**:
- [ ] Copy `osmomath/` to Gaia (location TBD: likely `pkg/osmomath/` or `x/dex/osmomath/`)
- [ ] Update `cosmossdk.io/math` from v1.4.0 → v1.5.3
- [ ] Remove SDK fork replace directive
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Document any API adaptations needed

---

### Task 1.2: Migrate osmoutils (minimal subset) 📋 `pending`

**Depends On**: Task 1.1

**Description**: Migrate the minimal osmoutils subset needed by DEX modules. Only 6 subpackages required.

**Subpackages to Migrate**:
- `osmoutils/` (root) - store helpers
- `osmoutils/accum/` - accumulator (critical for CL)
- `osmoutils/osmocli/` - CLI helpers
- `osmoutils/osmoassert/` - assertions
- `osmoutils/cosmwasm/` - CosmWasm helpers
- `osmoutils/observability/` - telemetry

**Subpackages to EXCLUDE**:
- `osmoutils/sumtree/`, `coinutil/`, `partialord/`, `noapptest/`, `wrapper/`

**Acceptance Criteria**:
- [ ] Copy required subpackages to Gaia
- [ ] Update IBC-go v8 → v10 imports
- [ ] Update SDK v0.50 → v0.53 imports
- [ ] Remove all replace directives (SDK, CometBFT, store)
- [ ] Update osmomath import path to Gaia location
- [ ] Clean compile with no errors
- [ ] All unit tests pass for migrated subpackages

---

## Phase 2: Core Pool Infrastructure

### Task 2.1: Migrate poolmanager/types 📋 `pending`

**Depends On**: Task 1.2

**Description**: Migrate `poolmanager/types` package. This defines interfaces only (PoolI, PoolModuleI) and should compile standalone.

**Acceptance Criteria**:
- [ ] Copy `x/poolmanager/types/` to Gaia
- [ ] Update imports (osmomath, osmoutils, SDK)
- [ ] Clean compile with no errors
- [ ] Document interface definitions for pool modules to implement

---

### Task 2.2: Migrate gamm 📋 `pending`

**Depends On**: Task 2.1

**Description**: Migrate the `gamm` module (Balancer and Stableswap pools). This is the simplest pool type and most established.

**Key Adaptations**:
- Move `superfluidtypes.MigrationPoolIDs` struct to gamm/types (trivial 2-field struct)
- Exclude superfluid migration features or stub them
- Update SDK patterns for v0.53

**Acceptance Criteria**:
- [ ] Copy `x/gamm/` to Gaia
- [ ] Update all imports (osmomath, osmoutils, poolmanager/types, SDK)
- [ ] Adapt legacy x/params if needed
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Wire module into Gaia app (basic registration)

---

### Task 2.3: Complete poolmanager 📋 `pending`

**Depends On**: Task 2.2

**Description**: Migrate `poolmanager` keeper and complete the module. Wire gamm as the first pool type.

**Acceptance Criteria**:
- [ ] Copy remaining `x/poolmanager/` (keeper, module, etc.)
- [ ] Update all imports
- [ ] Wire gamm as pool module via dependency injection
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Integration test: create Balancer pool, execute swap

---

## Phase 3: Additional Pool Types

### Task 3.1: Migrate concentrated-liquidity 📋 `pending`

**Depends On**: Task 2.3

**Description**: Migrate the concentrated-liquidity module. This is the most complex pool type with heavy `osmoutils/accum` usage.

**Key Challenges**:
- Heavy use of `osmoutils/accum` for spread rewards and incentives
- CosmWasm pool hooks integration
- Legacy x/params migration

**Acceptance Criteria**:
- [ ] Copy `x/concentrated-liquidity/` to Gaia
- [ ] Verify `osmoutils/accum` works correctly
- [ ] Update all imports
- [ ] Adapt legacy x/params if needed
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Wire as pool module in poolmanager
- [ ] Integration test: create CL pool, add liquidity, execute swap

---

### Task 3.2: Migrate cosmwasmpool 📋 `pending`

**Depends On**: Task 2.3

**Description**: Migrate the cosmwasmpool module for CosmWasm-based pools (Transmuter, orderbook).

**Key Challenges**:
- wasmd v0.53 → v0.60 API compatibility
- Pre-compiled WASM bytecode compatibility
- Gaia already has wasmd - verify integration

**Acceptance Criteria**:
- [ ] Copy `x/cosmwasmpool/` to Gaia
- [ ] Verify wasmd v0.60 API compatibility
- [ ] Update all imports
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Wire as pool module in poolmanager
- [ ] Integration test: instantiate Transmuter contract, execute swap

---

## Phase 4: MEV & Integration

### Task 4.1: Migrate protorev 📋 `pending`

**Depends On**: Tasks 3.1, 3.2

**Description**: Migrate the protorev MEV arbitrage module. Depends on all pool modules.

**Key Components**:
- PostHandler for transaction-level arbitrage
- Route finding across pool types
- Epoch hooks for periodic updates

**Acceptance Criteria**:
- [ ] Copy `x/protorev/` to Gaia
- [ ] Update all imports
- [ ] Wire PostHandler into Gaia app
- [ ] Clean compile with no errors
- [ ] All unit tests pass
- [ ] Integration test: verify arb detection across pool types

---

### Task 4.2: App Integration 📋 `pending`

**Depends On**: Task 4.1

**Description**: Complete Gaia app integration for all DEX modules.

**Acceptance Criteria**:
- [ ] All modules registered in app.go
- [ ] Genesis import/export working
- [ ] Upgrade handler if needed
- [ ] CLI commands available
- [ ] gRPC/REST endpoints working
- [ ] Clean build of full Gaia binary

---

### Task 4.3: Testing & Validation 📋 `pending`

**Depends On**: Task 4.2

**Description**: Comprehensive testing to validate production readiness.

**Test Levels**:
1. **Unit Tests**: All migrated tests passing
2. **Integration Tests**: User workflow scenarios
3. **Manual Tests**: Local node with realistic data

**Acceptance Criteria**:
- [ ] All unit tests pass
- [ ] Create pools of all types (Balancer, Stableswap, CL, CosmWasm)
- [ ] Execute swaps through poolmanager routing
- [ ] Multi-hop swaps work correctly
- [ ] Protorev finds and executes arbitrage
- [ ] Genesis export/import round-trip works
- [ ] Performance acceptable for production use

---

## Notes

- Each task follows workflow: `COPY → COMPILE → ADAPT → VERIFY → TEST → INTEGRATE → VALIDATE`
- Focus on getting one component fully working before moving to the next
- Document all adaptations and lessons learned in `knowledge.md`
- Commit progress after each task completion

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial task structure created | AI Assistant |
| 2026-01-28 | Task 0.1 completed - SDK version differences documented | AI Assistant |
| 2026-01-28 | Added Task 0.1a - Identify Required SDK Fork Features (high priority) | AI Assistant |
| 2026-01-28 | Added Task 0.1b - Compare Tokenfactory Implementations | AI Assistant |
| 2026-01-28 | Task 0.3 completed - concentrated-liquidity dependencies documented | AI Assistant |
| 2026-01-28 | Task 0.4 completed - gamm dependencies documented (simpler than CL, no accum) | AI Assistant |
| 2026-01-28 | Task 0.5 completed - cosmwasmpool dependencies documented (requires wasmd) | AI Assistant |
| 2026-01-28 | Task 0.6 completed - protorev dependencies documented (depends on all DEX) | AI Assistant |
| 2026-01-28 | Task 0.1f added and completed - x/epochs comparison (use SDK version) | AI Assistant |
| 2026-01-28 | Task 0.1a completed - SDK fork analysis shows NO blockers for DEX migration | AI Assistant |
| 2026-01-28 | Tasks 0.7, 0.1d, 0.1e completed - dependency graph already documented; store fork and tokenfactory questions resolved by 0.1a | AI Assistant |
| 2026-01-28 | Task 0.7a completed - minimal osmoutils subset identified; all use standard store APIs | AI Assistant |
| 2026-01-28 | Added concrete Phase 1-4 tasks matching migration plan in knowledge.md | AI Assistant |
| 2026-01-28 | Task 0.8 completed - Testing Harness defined with 3-level strategy (unit/integration/e2e) | AI Assistant |
