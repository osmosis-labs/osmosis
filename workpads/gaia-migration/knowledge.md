# Gaia Migration Knowledge Base

## Overview

**Goal**: Migrate core Osmosis DEX modules to Gaia, enabling all Osmosis DEX operations to run on Gaia with production-grade quality.

| Source | Target |
|--------|--------|
| Osmosis (`/Users/nicolas/devel/osmosis`) | Gaia (`/Users/nicolas/devel/gaia`) |

When complete, Gaia should be able to run all Osmosis DEX operations with full test coverage.

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

## Module Descriptions

### poolmanager

**Purpose**: _(to be documented)_

**Key Components**:
- _(to be documented)_

**External Dependencies**:
- _(to be documented)_

**Internal Dependencies**:
- _(to be documented)_

---

### concentrated-liquidity

**Purpose**: _(to be documented)_

**Key Components**:
- _(to be documented)_

**External Dependencies**:
- _(to be documented)_

**Internal Dependencies**:
- _(to be documented)_

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

```
(to be built after dependency analysis)
```

---

## Migration Workflow

### Per-Module Migration Steps

1. **Copy** - Copy module from Osmosis to Gaia
2. **Compile** - Attempt to compile in Gaia, document errors
3. **Adapt** - Update module to match Gaia SDK version and patterns
4. **Verify Compile** - Ensure clean compilation
5. **Unit Tests** - Run migrated unit tests, fix failures
6. **Integrate** - Wire module into Gaia app initialization
7. **Integration Tests** - Run or write integration tests
8. **Manual Tests** - Test on local node with realistic data

### Testing Strategy

| Level | Source | Purpose |
|-------|--------|---------|
| Unit Tests | Migrated from Osmosis | Verify module logic |
| Integration Tests | New (focused on user workflows) | Verify cross-module behavior |
| Manual Tests | Local node + mainnet data | Validate production readiness |

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
6. **NEW**: What Osmosis SDK fork features are required by the DEX modules, and are they available in upstream SDK 0.53?

---

## Lessons Learned

_(to be populated during migration)_

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial document creation | AI Assistant |
| 2026-01-28 | Documented SDK version differences (Task 0.1) - major gap: SDK 0.50→0.53, IBC v8→v10 | AI Assistant |
