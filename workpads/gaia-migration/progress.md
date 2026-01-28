# Gaia Migration Progress

Tracks what changed for each migrated component. Each entry shows:
- What was copied
- What adaptations were made (if any)
- Commit references for review

---

## Review Guide

For each component, there are **two commits**:
1. **Copy commit**: Raw copy with NO changes (exact files from Osmosis)
2. **Adapt commit**: ALL changes (imports + API fixes)

To review everything that changed:
```bash
git diff <copy-commit> <adapt-commit> -- path/to/component/
```

This shows ALL modifications - imports, SDK updates, IBC fixes - nothing hidden.

---

## Phase 1: Foundation

### osmomath

| Aspect | Details |
|--------|---------|
| **Source** | `osmosis/osmomath/` |
| **Target** | `gaia/pkg/osmomath/` |
| **Files** | 23 .go files |
| **Gaia Commit** | `57a107393` |
| **Adaptations** | None - pure copy with import updates |

**Import changes**:
- `github.com/osmosis-labs/osmosis/osmomath` → `github.com/cosmos/gaia/v26/pkg/osmomath` (in test files only)

**Review**: No logic changes. Import-only migration.

---

### osmoutils

| Aspect | Details |
|--------|---------|
| **Source** | `osmosis/osmoutils/` |
| **Target** | `gaia/pkg/osmoutils/` |
| **Files** | 51 .go files (8 subpackages) |
| **Gaia Commit** | `74c359d58` |
| **Adaptations** | IBC v10 API change in `ibc.go` |

**Subpackages migrated**:
- `osmoutils/` (root)
- `osmoutils/accum/`
- `osmoutils/osmocli/`
- `osmoutils/osmoassert/`
- `osmoutils/cosmwasm/`
- `osmoutils/observability/`
- `osmoutils/noapptest/`
- `osmoutils/wrapper/`

**Import changes**:
- `github.com/osmosis-labs/osmosis/osmomath` → `github.com/cosmos/gaia/v26/pkg/osmomath`
- `github.com/osmosis-labs/osmosis/osmoutils` → `github.com/cosmos/gaia/v26/pkg/osmoutils`
- `github.com/cosmos/ibc-go/v8` → `github.com/cosmos/ibc-go/v10`

**API adaptations** (requires review):

```diff
// pkg/osmoutils/ibc.go line 69
- if denomTrace.Path != "" {
+ if denomTrace.Path() != "" {
```

**Reason**: IBC v10 changed `DenomTrace` type to `Denom`, and `Path` field became `Path()` method.

---

## Phase 2: Core Pool Infrastructure

### poolmanager/types

| Aspect | Details |
|--------|---------|
| **Source** | `osmosis/x/poolmanager/types/` |
| **Target** | `gaia/x/poolmanager/types/` |
| **Files** | 24 .go files |
| **Copy Commit** | `6db70b42f` |
| **Adapt Commit** | `dc4acb8d0` |

**To review all changes**:
```bash
git diff 6db70b42f dc4acb8d0 -- x/poolmanager/types/
```

**Import changes**:
- `github.com/osmosis-labs/osmosis/osmomath` → `github.com/cosmos/gaia/v26/pkg/osmomath`
- `github.com/osmosis-labs/osmosis/osmoutils` → `github.com/cosmos/gaia/v26/pkg/osmoutils`
- `github.com/osmosis-labs/osmosis/v31/x/poolmanager/types` → `github.com/cosmos/gaia/v26/x/poolmanager/types`
- `github.com/osmosis-labs/osmosis/v31/app/params` → `github.com/cosmos/gaia/v26/app/params`

**Added to Gaia app/params**:
- `BaseCoinUnit = "uatom"` (DEX modules need this constant)
- `SetAddressPrefixes()` (test helper for bech32 address validation)

**Commented out** (TODO for Task 2.3):
- `TestAuthzMsg` - needs `poolmanager/module.AppModuleBasic{}`

**Test status**: ✅ All unit tests pass (after Task 0.9 fixes)

---

## Test Infrastructure (Task 0.9)

| Aspect | Details |
|--------|---------|
| **Commit** | `0c758f641` |

**Files created**:
- `tests/dex/test_helpers.go`

**Provides**:
- `TestMessageAuthzSerialization(t, msg, module)` - Gaia equivalent of Osmosis apptesting
- `GenerateTestAddrs()` - generates valid/invalid test addresses
- `TestDenom = "uatom"`, `SecondaryTestDenom = "uion"` - test constants

**Test fixes applied to poolmanager/types**:
- Added `init()` to set bech32 prefixes before address creation
- Changed `invalidAddr` to malformed bech32 string ("cosmos1invalid")
- Updated test data to use different denoms (uatom/uion instead of uatom/uatom)
- Fixed expected keys (uosmo → uatom)

---

## gamm (Task 2.2)

| Aspect | Details |
|--------|---------|
| **Source** | `osmosis/x/gamm/` |
| **Target** | `gaia/x/gamm/` |
| **Files** | 95 .go files (copy), 80 modified (adapt) |
| **Copy Commit** | `28e055001` |
| **Adapt Commit** | `83cd5bfbc` |

**To review all changes**:
```bash
git diff 28e055001 83cd5bfbc -- x/gamm/
```

**Import changes**:
- `osmomath`, `osmoutils` → `gaia/pkg/`
- `x/gamm/*`, `x/poolmanager/*` → `gaia/x/`
- `app/params` → `gaia/app/params`

**Removed dependencies** (Osmosis-specific, not needed for core DEX):
- `x/incentives` - reward gauges
- `x/pool-incentives` - pool reward distribution  
- `x/superfluid` - superfluid staking
- `x/concentrated-liquidity` - CL migration
- `simulation/simtypes` - simulation framework

**Removed files**:
- `keeper/migrate.go`, `migrate_test.go` - CL migration
- `simulation/` - simulation code
- `pool-models/internal/test_helpers/` - test helpers (needs recreation)

**Added files**:
- `x/poolmanager/events/` - swap event emission

**Stubbed (not supported)**:
- CL migration queries return "unimplemented"
- CL migration gov proposals return "not supported"

**Test status**:
- `cfmm_common` tests pass ✅
- Other tests need apptesting adaptation (use Osmosis test framework)

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Created progress tracking file | AI Assistant |
| 2026-01-28 | Added osmomath and osmoutils entries | AI Assistant |
| 2026-01-28 | Added poolmanager/types entry | AI Assistant |
| 2026-01-28 | Added test infrastructure (Task 0.9) | AI Assistant |
| 2026-01-28 | Added gamm entry (Task 2.2) | AI Assistant |
