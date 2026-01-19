# SDK Upgrade Tasks

## Status Legend

```
📋 pending      - Not yet started
🚧 in_progress  - Currently working on  
✅ completed    - Finished and verified
🚫 blocked      - Cannot proceed
```

---

## Phase 0: Research & Alignment

### Task 0.1: Collect Prior SDK Upgrade Notes ✅ `completed`

**Description**: Review Osmosis changelog for prior SDK upgrade patterns and capture key notes.

**Acceptance Criteria**:
- [x] Prior SDK upgrade notes summarized in `references.md`

---

### Task 0.2: Review v0.50 → v0.53 Migration Guide ✅ `completed`

**Description**: Summarize Cosmos SDK v0.53 migration guidance with repo-specific impact notes.

**Acceptance Criteria**:
- [x] Migration guide notes added to `references.md`
- [x] Osmosis + SDK fork impact captured for each note

---

### Task 0.3: Align Local SDK Checkout ✅ `completed`

**Description**: Ensure `/Users/nicolas/devel/cosmos-sdk` is on the same branch/tag used by Osmosis v31.0.0.

**Acceptance Criteria**:
- [x] Branch set to `osmo-v30/0.50.14`
- [x] Alignment noted in `references.md`

---

### Task 0.4: Compare Fork vs Upstream ✅ `completed`

**Description**: Summarize differences between the Osmosis SDK fork and upstream v0.50.14.

**Acceptance Criteria**:
- [x] High-level diff summary captured in `references.md`

---

### Task 0.5: Potential Upgrade Planning Tasks ✅ `completed`

**Description**: Create a task breakdown for the upgrade execution plan.

**Acceptance Criteria**:
- [x] Tasks added for dependency alignment, module wiring, store upgrade planning, and test matrix

---

### Task 0.6: Dependency Alignment Matrix ✅ `completed`

**Description**: Define target versions for IBC-Go, Wasmd, CometBFT, and `cosmossdk.io/*` packages compatible with SDK v0.53.4.

**Acceptance Criteria**:
- [x] Version matrix documented
- [x] Conflicts with Osmosis modules identified

---

### Task 0.6a: IBC-Go v10 Migration Research ✅ `completed`

**Description**: Review IBC-Go v10 release notes and migration docs to identify breaking changes, module wiring changes, and required upgrades for Osmosis.

**Acceptance Criteria**:
- [x] IBC-Go v10 migration notes summarized in `references.md`
- [x] API or module changes impacting Osmosis identified
- [x] Known conflicts or upgrade risks documented

---

### Task 0.6b: Wasmd / CosmWasm SDK Compatibility Research ✅ `completed`

**Description**: Review Wasmd v0.60.x and CosmWasm SDK compatibility notes to identify breaking changes and required upgrades for Osmosis.

**Acceptance Criteria**:
- [x] Wasmd v0.60.x migration/compat notes summarized in `references.md`
- [x] Contract/runtime compatibility risks documented
- [x] Osmosis-specific conflicts or upgrade steps identified

---

### Task 0.7: Module Wiring & Conflicts ✅ `completed`

**Description**: Plan wiring changes for v0.53, including `PreBlocker`, `x/epoch` vs `x/epochs`, and `x/protocolpool`.

**Acceptance Criteria**:
- [x] Wiring deltas documented
- [x] Store upgrade needs identified
- [x] Module name conflicts resolved

---

### Task 0.8: State/Store Upgrade Plan ✅ `completed`

**Description**: Identify new store keys, migrations, and state transformations needed for v0.53.

**Acceptance Criteria**:
- [x] Store keys list prepared
- [x] Migration handlers mapped
- [x] Upgrade handler outline drafted

---

### Task 0.9: Fork Patch Reconciliation ✅ `completed`

**Description**: Map each fork patch to an upstream equivalent or re-application plan.

**Acceptance Criteria**:
- [x] Fork patch list mapped to upstream
- [x] Re-apply plan drafted for critical patches

---

### Task 0.10: Upgrade Test Matrix ✅ `completed`

**Description**: Define tests for upgrade, migration, and rollback safety.

**Acceptance Criteria**:
- [x] State export/import test plan
- [x] Mainnet fork test plan
- [x] E2E and simulation test plan

---

## Phase 1: Analysis & Design

### Task 1.1: Baseline Inputs & Reference Checkouts ✅ `completed`

**Description**: Use existing local checkouts as baselines; only clone extra references into `workpads/sdk-upgrade/repos` if needed.

**Acceptance Criteria**:
- [x] `/Users/nicolas/devel/osmosis` baseline noted
- [x] `/Users/nicolas/devel/cosmos-sdk` on `osmo-v30/0.50.14`
- [x] Any extra reference checkouts (if needed) live under `workpads/sdk-upgrade/repos/`

---

### Task 1.2: Document Fork Patches (from local checkout) 📋 `pending`

**Description**: Identify and document all Osmosis SDK fork modifications using the local fork checkout.

**Acceptance Criteria**:
- [ ] List all modified files vs upstream v0.50.14
- [ ] Categorize patches (bug fix, feature, optimization)
- [ ] Assess each patch (still needed, can upstream, SDK alternative exists)
- [ ] Update knowledge.md with findings

---

### Task 1.3: SDK v0.50 → v0.53 Change Review 📋 `pending`

**Description**: Review breaking changes/new features from the SDK changelog and UPGRADING guide (v0.53.4).

**Acceptance Criteria**:
- [ ] Breaking changes documented
- [ ] New features documented
- [ ] Deprecations documented
- [ ] Update knowledge.md with findings

---

### Task 1.4: Dependency Compatibility Check 📋 `pending`

**Description**: Validate compatibility of IBC-Go, Wasmd, CometBFT, and `cosmossdk.io/*` with SDK v0.53.4.

**Acceptance Criteria**:
- [ ] Version matrix documented
- [ ] Conflicts with Osmosis modules identified
- [ ] Required bumps tracked

---

### Task 1.5: Wiring + Store Upgrade Design 📋 `pending`

**Description**: Plan v0.53 wiring changes, module conflicts, and store upgrades.

**Acceptance Criteria**:
- [ ] `PreBlocker` ordering plan documented
- [ ] `x/epoch` vs `x/epochs` decision documented
- [ ] `x/protocolpool` inclusion decision documented
- [ ] Store upgrade needs identified

---

### Task 1.6: Gaia v25 Migration Reference 📋 `pending`

**Description**: Extract applicable patterns from Gaia v25 migration.

**Acceptance Criteria**:
- [ ] Relevant Gaia PRs/changes identified
- [ ] Migration patterns documented
- [ ] References added to `references.md`

---

### Task 1.7: Fork Patch Reconciliation Plan 📋 `pending`

**Description**: Map each fork patch to upstream equivalents or re-apply plan.

**Depends On**: Task 1.2, Task 1.3

**Acceptance Criteria**:
- [ ] Fork patch list mapped to upstream
- [ ] Re-apply plan drafted for critical patches

---

### Task 1.8: Upgrade Test Matrix 📋 `pending`

**Description**: Define tests for upgrade, migration, and rollback safety.

**Acceptance Criteria**:
- [ ] State export/import test plan
- [ ] Mainnet fork test plan
- [ ] E2E and simulation test plan

---

## Phase 2: Implementation

### Task 2.1: Create SDK Upgrade Branch 📋 `pending`

**Depends On**: Phase 1 complete

**Acceptance Criteria**:
- [ ] Branch created from main
- [ ] go.mod updated with target SDK version
- [ ] Initial compilation attempted

---

### Task 2.2: Fix Compilation Errors 📋 `pending`

**Depends On**: Task 2.1

**Acceptance Criteria**:
- [ ] All packages compile
- [ ] No type errors
- [ ] No import errors

---

### Task 2.3: Update Module Implementations 📋 `pending`

**Depends On**: Task 2.2

**Acceptance Criteria**:
- [ ] All Osmosis modules updated for new APIs
- [ ] Keeper interfaces updated
- [ ] Message handlers updated

---

### Task 2.4: Reapply Critical Fork Patches 📋 `pending`

**Depends On**: Task 2.3, Task 1.2

**Acceptance Criteria**:
- [ ] Critical patches identified in Task 1.2 reapplied
- [ ] Patches tested
- [ ] Document any patches that couldn't be reapplied

---

## Phase 3: Testing

### Task 3.1: Unit Tests 📋 `pending`

**Depends On**: Task 2.4

**Acceptance Criteria**:
- [ ] All unit tests pass
- [ ] Test coverage maintained

**Command**: `go test ./...`

---

### Task 3.2: Integration Tests 📋 `pending`

**Depends On**: Task 3.1

**Acceptance Criteria**:
- [ ] Integration tests pass
- [ ] IBC tests pass

---

### Task 3.3: Simulation Tests 📋 `pending`

**Depends On**: Task 3.2

**Acceptance Criteria**:
- [ ] Simulation tests pass
- [ ] No panics or state corruption

---

### Task 3.4: E2E Tests 📋 `pending`

**Depends On**: Task 3.3

**Acceptance Criteria**:
- [ ] E2E tests pass
- [ ] All critical flows tested

---

## Phase 4: Validation

### Task 4.1: State Migration Testing 📋 `pending`

**Depends On**: Phase 3 complete

**Acceptance Criteria**:
- [ ] Export/import works
- [ ] State migration successful
- [ ] No data loss

---

### Task 4.2: Mainnet Fork Testing 📋 `pending`

**Depends On**: Task 4.1

**Acceptance Criteria**:
- [ ] Node syncs from mainnet state
- [ ] Transactions execute correctly
- [ ] No panics or errors

---

### Task 4.3: Performance Benchmarks 📋 `pending`

**Depends On**: Task 4.2

**Acceptance Criteria**:
- [ ] Benchmarks run
- [ ] No significant performance regression
- [ ] Document any changes

---

## Notes

_(Space for task-related notes during execution)_
