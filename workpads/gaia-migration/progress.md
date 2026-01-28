# Gaia Migration Progress

Tracks what changed for each migrated component. Each entry shows:
- What was copied
- What adaptations were made (if any)
- Commit references for review

---

## Review Guide

For each component, there are **two commits**:
1. **Copy commit**: Raw copy with only import path changes
2. **Adapt commit**: Any API/logic changes needed for SDK 0.53 / IBC v10

To review adaptations, diff between the two commits:
```bash
git diff <copy-commit> <adapt-commit> -- path/to/component/
```

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
| **Target** | TBD |
| **Copy Commit** | TBD |
| **Adapt Commit** | TBD |
| **Adaptations** | TBD |

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Created progress tracking file | AI Assistant |
| 2026-01-28 | Added osmomath and osmoutils entries | AI Assistant |
