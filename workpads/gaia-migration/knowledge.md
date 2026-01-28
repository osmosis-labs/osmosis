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
| SDK Version | _(to be documented)_ | _(to be documented)_ |
| Key API Differences | _(to be documented)_ | _(to be documented)_ |

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

1. What is the exact SDK version difference between Osmosis and Gaia?
2. Are there shared utility packages that need to migrate first (e.g., `osmomath`, `osmoutils`)?
3. What state/genesis migration is needed for each module?
4. How do we handle CosmWasm integration differences?

---

## Lessons Learned

_(to be populated during migration)_

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial document creation | AI Assistant |
