# SDK Upgrade Knowledge Base

## Overview

**Goal**: Upgrade Osmosis from the Cosmos SDK v0.50.x fork to the default Cosmos SDK v0.53.4 (matching Gaia).

| Current State | Target State |
|---------------|--------------|
| osmo-labs/cosmos-sdk v0.50.14-v30-osmo | cosmos/cosmos-sdk v0.53.4 |
| Osmosis v31.0.0 | Osmosis v32+ (TBD) |

---

## Context

### Osmosis Current SDK Fork

- **Repository**: `github.com/osmosis-labs/cosmos-sdk`
- **Tag**: `v0.50.14-v30-osmo`
- **Base**: Cosmos SDK v0.50.14 with Osmosis-specific patches
- **Key modifications**: (to be documented)

### Gaia Reference

- **Version**: Gaia v25.3.0
- **SDK**: Cosmos SDK v0.53.4
- **Migration path**: Gaia v25.0.0 moved from v0.50.x → v0.53.0

---

## Technical Specifications

### SDK Version Jump Analysis

| SDK Version | Major Changes |
|-------------|---------------|
| v0.50.x → v0.51.x | (to be documented) |
| v0.51.x → v0.52.x | (to be documented) |
| v0.52.x → v0.53.x | (to be documented) |

### Breaking Changes Checklist

- [ ] Module API changes
- [ ] Keeper interface changes
- [ ] Message handler changes
- [ ] Genesis state changes
- [ ] Client/CLI changes
- [ ] Proto changes
- [ ] Ante handler changes
- [ ] Store key changes

---

## Decision Log

| D# | Decision | Rationale | Date |
|----|----------|-----------|------|
| D1 | Target SDK v0.53.4 | Match Gaia v25.3.0 for ecosystem compatibility | 2026-01-13 |
| D2 | Move to upstream SDK | Reduce maintenance burden of fork | 2026-01-13 |

---

## Osmosis Fork Modifications

### Patches to Evaluate

Document each Osmosis fork modification and determine:
1. Is it still needed?
2. Can it be upstreamed?
3. Is there an SDK alternative?

| Patch Area | Description | Status | Action Needed |
|------------|-------------|--------|---------------|
| (TBD) | | 📋 pending | |

---

## Migration Strategy

### Phase 1: Analysis
- [ ] Document all fork modifications
- [ ] Identify breaking changes v0.50 → v0.53
- [ ] Map Osmosis patches to SDK features

### Phase 2: Implementation
- [ ] Update go.mod
- [ ] Fix compilation errors
- [ ] Update module implementations
- [ ] Update tests

### Phase 3: Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Simulation tests pass
- [ ] E2E tests pass

### Phase 4: Validation
- [ ] State migration testing
- [ ] Mainnet fork testing
- [ ] Performance benchmarks

---

## Open Questions

1. **Fork patches**: Which Osmosis SDK patches are critical vs nice-to-have?
2. **IBC compatibility**: Any IBC changes between SDK versions?
3. **WASM compatibility**: CosmWasm compatibility with SDK v0.53?
4. **State migration**: Are there state migrations between versions?

---

## Lessons Learned

_(To be updated during implementation)_

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-13 | Initial document creation | AI Assistant |
