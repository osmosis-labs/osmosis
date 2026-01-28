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

### Task 0.1a: Identify Required SDK Fork Features 📋 `pending` ⚠️ HIGH PRIORITY

**Description**: Analyze which Osmosis SDK fork features are used by the DEX modules and determine if they are available in upstream SDK 0.53. This is critical to assess early as it may fundamentally affect our migration approach.

**Why Important**: If the DEX modules depend on Osmosis-specific SDK fork features that don't exist in upstream SDK 0.53, we have several options:
1. Port those features to Gaia (adds complexity)
2. Refactor modules to not need those features (may be significant work)
3. Contribute missing features upstream (long-term, unlikely for this project timeline)

**Acceptance Criteria**:
- [ ] List all Osmosis SDK fork modifications (from `osmosis-labs/cosmos-sdk v0.50.14-v30-osmo`)
- [ ] For each fork modification, identify if DEX modules depend on it
- [ ] For each required fork feature, check if equivalent exists in SDK 0.53
- [ ] Document blockers or risks in `knowledge.md`
- [ ] Recommend approach for each missing feature

**Known Fork Areas to Investigate** (from knowledge.md):
- Bank module hooks / supply offsets
- Store fork (iavlFastNodeModuleWhitelist, async pruning)
- block-sdk fork from Skip protocol
- Any other custom SDK modifications

---

### Task 0.2: Analyze poolmanager Dependencies 📋 `pending`

**Description**: Map all internal and external dependencies of the `poolmanager` module.

**Acceptance Criteria**:
- [ ] List all Osmosis-internal imports
- [ ] List all cosmos-sdk imports
- [ ] List all third-party imports
- [ ] Identify which dependencies need to migrate first
- [ ] Update `knowledge.md` with module description and dependencies

---

### Task 0.3: Analyze concentrated-liquidity Dependencies 📋 `pending`

**Description**: Map all internal and external dependencies of the `concentrated-liquidity` module.

**Acceptance Criteria**:
- [ ] List all Osmosis-internal imports
- [ ] List all cosmos-sdk imports  
- [ ] List all third-party imports
- [ ] Identify which dependencies need to migrate first
- [ ] Update `knowledge.md` with module description and dependencies

---

### Task 0.4: Analyze gamm Dependencies 📋 `pending`

**Description**: Map all internal and external dependencies of the `gamm` module.

**Acceptance Criteria**:
- [ ] List all Osmosis-internal imports
- [ ] List all cosmos-sdk imports
- [ ] List all third-party imports
- [ ] Identify which dependencies need to migrate first
- [ ] Update `knowledge.md` with module description and dependencies

---

### Task 0.5: Analyze cosmwasmpool Dependencies 📋 `pending`

**Description**: Map all internal and external dependencies of the `cosmwasmpool` module.

**Acceptance Criteria**:
- [ ] List all Osmosis-internal imports
- [ ] List all cosmos-sdk imports
- [ ] List all third-party imports
- [ ] Identify which dependencies need to migrate first
- [ ] Update `knowledge.md` with module description and dependencies

---

### Task 0.6: Analyze protorev Dependencies 📋 `pending`

**Description**: Map all internal and external dependencies of the `protorev` module.

**Acceptance Criteria**:
- [ ] List all Osmosis-internal imports
- [ ] List all cosmos-sdk imports
- [ ] List all third-party imports
- [ ] Identify which dependencies need to migrate first
- [ ] Update `knowledge.md` with module description and dependencies

---

### Task 0.7: Build Dependency Graph 📋 `pending`

**Depends On**: Tasks 0.2-0.6

**Description**: Create a dependency DAG showing migration order from simplest to most complex.

**Acceptance Criteria**:
- [ ] Dependency graph documented in `knowledge.md`
- [ ] Migration order determined (leaf nodes first)
- [ ] Shared utilities (osmomath, osmoutils) positioned in graph

---

### Task 0.8: Define Testing Harness 📋 `pending`

**Description**: Design the three-level testing strategy and document setup requirements.

**Acceptance Criteria**:
- [ ] Unit test migration approach documented
- [ ] Integration test framework chosen
- [ ] Manual test setup documented (local node + mainnet data)
- [ ] Update `knowledge.md` with testing strategy

---

## Phase 1: Foundation Migration

_(Tasks will be added after Phase 0 completes and migration order is determined)_

### Task 1.0: Migrate First Leaf Dependency 📋 `pending`

**Depends On**: Task 0.7

**Description**: Migrate the first leaf node in the dependency graph (likely a utility package).

**Acceptance Criteria**:
- [ ] Package copied to Gaia
- [ ] Compiles in Gaia
- [ ] Unit tests pass
- [ ] Workflow documented and refined

---

## Phase 2: Core Module Migration

_(Tasks will be added as Phase 1 progresses)_

---

## Phase 3: Integration & Testing

_(Tasks will be added as Phase 2 progresses)_

---

## Notes

- Migration order will be determined by the dependency graph (Task 0.7)
- The workflow (copy → compile → adapt → test → integrate) will be refined with each module
- Focus on getting one module fully working before moving to the next

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial task structure created | AI Assistant |
| 2026-01-28 | Task 0.1 completed - SDK version differences documented | AI Assistant |
| 2026-01-28 | Added Task 0.1a - Identify Required SDK Fork Features (high priority) | AI Assistant |
