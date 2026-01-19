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

### Dependency Alignment (SDK v0.53.4 baseline)

Baseline derived from Gaia v25.3.0.

| Dependency | Target Version |
|------------|----------------|
| IBC-Go | `github.com/cosmos/ibc-go/v10 v10.5.0` |
| IBC-Go 08-wasm | `github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10 v10.3.0` |
| Wasmd | `github.com/CosmWasm/wasmd v0.60.2` |
| CometBFT | `github.com/cometbft/cometbft v0.38.20` |
| cosmossdk.io/* | Align to Gaia v25.3.0 (`api v0.9.2`, `client/v2 v2.0.0-beta.9`, `core v0.11.3`, `errors v1.0.2`, `log v1.6.1`, `store v1.1.2`, `x/tx v0.14.0`, `x/upgrade v0.2.0`, `x/evidence v0.2.0`, `x/feegrant v0.2.0`) |

#### Noted Conflicts (current Osmosis vs baseline)
- IBC-Go v8 → v10 upgrade required; `08-wasm` module also jumps to v10 series.
- Wasmd v0.53.3 → v0.60.2 upgrade required.
- CometBFT v0.38.17 → v0.38.20 upgrade required.
- `cosmossdk.io/*` packages lag Gaia; `cosmossdk.io/store` is fork-replaced and must be reconciled.

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
| Bank hooks + supply offsets | Adds `TrackBeforeSend`/`BlockBeforeSend` hooks and supply offset support. | 📋 pending | Identify Osmosis modules using hooks; re-apply or replace with upstream features. |
| Clawback + cliff vesting | Adds clawback vesting account + cliff vesting CLI/protos. | 📋 pending | Determine if Osmosis relies on clawback/cliff; re-apply or migrate to upstream alternative. |
| Slashing perf + migration tweak | Stops per-block sign info writes; includes slashing migration key change. | 📋 pending | Check upstream v0.53 behavior; re-apply if perf regression or migration mismatch. |
| IAVL pruning + fast nodes | Async pruning, pruning fixes, per-module fast nodes. | 📋 pending | Validate if upstream v0.53 includes equivalents; re-apply if needed for mainnet performance. |
| ReCheckTx ValidateBasic | Skip ValidateBasic on recheck. | 📋 pending | Confirm upstream behavior; keep if Osmosis depends on recheck behavior. |
| OTEL gRPC interceptor | Adds OTEL span attributes in gRPC server. | 📋 pending | Decide whether to keep instrumentation or drop with upstream logging. |
| Governance + query fixes | Query all proposals, pagination checks, whitelist settings parse. | 📋 pending | Verify upstream v0.53; re-apply if still missing. |
| Misc backports | Denom regex removal, types speedups, epoch account access changes, supply offset helpers. | 📋 pending | Audit against upstream v0.53 and remove/re-apply as needed. |

### Fork Diff Overview (v0.50.14 → osmo-v30/0.50.14)

High-level categories from `git diff --name-only`:
- **Bank module**: hooks + supply offsets (`x/bank/*`, related proto/gen files).
- **Vesting module**: clawback/cliff and new CLI/proto (`x/auth/vesting/*`, `proto/cosmos/vesting/*`).
- **Slashing + staking**: perf and migration tweaks (`x/slashing/*`, `x/staking/*`).
- **Store/IAVL**: pruning + rootmulti changes (`store/*`).
- **Baseapp/server**: recheck behavior + OTEL (`baseapp/*`, `server/*`).
- **Gov queries**: pagination fixes (`x/gov/*`).
- **Types**: coin/dec coin tweaks (`types/*`).
- **Tests + CI**: integration/vesting tests and workflow adjustments.

Full file list is available via:
`git diff --name-only v0.50.14..osmo-v30/0.50.14`

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
| 2026-01-19 | Document SDK v0.53.4 dependency baseline and conflicts | AI Assistant |
| 2026-01-19 | Draft fork patch reconciliation map | AI Assistant |
| 2026-01-19 | Document fork diff overview and categories | AI Assistant |
