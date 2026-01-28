# Gaia Migration References

## §1 Local Repositories

| Repo | Path | Purpose |
|------|------|---------|
| Osmosis | `/Users/nicolas/devel/osmosis` | Source of modules to migrate |
| Gaia | `/Users/nicolas/devel/gaia` | Target for migration, primary test environment |
| Cosmos SDK | `/Users/nicolas/devel/cosmos-sdk` | SDK reference for API changes |

---

## §2 Key Osmosis Paths

### Modules to Migrate

| Module | Path |
|--------|------|
| poolmanager | `x/poolmanager/` |
| concentrated-liquidity | `x/concentrated-liquidity/` |
| gamm | `x/gamm/` |
| cosmwasmpool | `x/cosmwasmpool/` |
| protorev | `x/protorev/` |

### Potential Dependencies

| Package | Path | Notes |
|---------|------|-------|
| osmomath | `osmomath/` | Math utilities |
| osmoutils | `osmoutils/` | General utilities |

---

## §3 Key Gaia Paths

| Item | Path | Notes |
|------|------|-------|
| App setup | `app/` | Module wiring |
| Existing modules | `x/` | Reference for patterns |
| go.mod | `go.mod` | SDK version and dependencies |

---

## §4 Official Documentation

### Cosmos SDK

- **SDK Docs**: <https://docs.cosmos.network/>
- **Module Development**: <https://docs.cosmos.network/main/building-modules/intro>

### Osmosis

- **Osmosis Docs**: <https://docs.osmosis.zone/>
- **GitHub**: <https://github.com/osmosis-labs/osmosis>

### Gaia

- **GitHub**: <https://github.com/cosmos/gaia>

---

## §5 Useful Commands

### Dependency Analysis

```bash
# Find imports in a module
cd /Users/nicolas/devel/osmosis
grep -r "github.com/osmosis-labs" x/poolmanager/ --include="*.go" | grep import

# Check Gaia SDK version
cd /Users/nicolas/devel/gaia
grep "github.com/cosmos/cosmos-sdk" go.mod

# Check Osmosis SDK version
cd /Users/nicolas/devel/osmosis
grep "github.com/cosmos/cosmos-sdk" go.mod
```

### Compilation Testing

```bash
# Test compile in Gaia after adding module
cd /Users/nicolas/devel/gaia
go build ./...

# Run specific module tests
go test ./x/{module}/...
```

---

## §6 Reference Quality Notes

### Highly Useful ⭐

| Reference | What Made It Useful | Used For |
|-----------|---------------------|----------|
| | | |

### Less Useful Than Expected

| Reference | Why It Didn't Help | Alternative |
|-----------|-------------------|-------------|
| | | |

### New References to Consider

| Source | Why It Might Help |
|--------|-------------------|
| | |

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial document creation | AI Assistant |
