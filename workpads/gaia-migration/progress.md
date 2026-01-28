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

**Commented out** (TODO for later):
- `TestAuthzMsg` - needs `apptesting` and `poolmanager/module` (Task 2.3)
- Imports: `apptesting`, `poolmanager/module`

**Test status**:
- Core code compiles ✅
- Some tests fail due to Osmosis-specific test data (uosmo, osmo addresses)
- Tests need updating to use Gaia denoms/addresses

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Created progress tracking file | AI Assistant |
| 2026-01-28 | Added osmomath and osmoutils entries | AI Assistant |
