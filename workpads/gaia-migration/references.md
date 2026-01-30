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

## §6 CosmWasm Pool Contracts

### Overview

The `cosmwasmpool` module uses pre-compiled WASM contracts for pool logic. These contracts are compiled against Osmosis-specific dependencies and have hardcoded assumptions about the chain they run on.

### Contract Source Repositories (Cloned)

| Contract | Repository | Local Path |
|----------|------------|------------|
| **Transmuter** | <https://github.com/osmosis-labs/transmuter> | `workpads/gaia-migration/repos/transmuter/` |
| **Orderbook** | <https://github.com/osmosis-labs/orderbook> | `workpads/gaia-migration/repos/orderbook/` |
| **Osmosis Rust** | <https://github.com/osmosis-labs/osmosis-rust> | `workpads/gaia-migration/repos/osmosis-rust/` |

### Key Discovery: Proto Compatibility

**Gaia's tokenfactory** (`github.com/cosmos/tokenfactory v0.53.5`) uses **the same proto type URLs** as Osmosis's tokenfactory. This is because it's derived from Osmosis's proto definitions.

| Proto Type URL | Gaia Support |
|---------------|--------------|
| `/osmosis.tokenfactory.v1beta1.MsgCreateDenom` | ✅ Same proto |
| `/osmosis.tokenfactory.v1beta1.MsgMint` | ✅ Same proto |
| `/osmosis.tokenfactory.v1beta1.MsgBurn` | ✅ Same proto |
| `/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata` | ✅ Same proto |

**Implication**: Existing WASM contracts may work on Gaia without recompilation!

### Bytecode Files in Codebase

Located in `x/cosmwasmpool/bytecode/`:

| File | Source | Purpose |
|------|--------|---------|
| `transmuter.wasm` | transmuter repo (v1) | Basic transmuter pool |
| `transmuter_v3.wasm` | transmuter repo (v3) | Transmuter with alloyed assets |
| `transmuterv3.wasm` | transmuter repo (v3) | Duplicate/alias of v3 |
| `transmuter_migrate.wasm` | transmuter repo | Migration entrypoint contract |
| `sumtree_orderbook.wasm` | orderbook repo | Orderbook pool contract |

### Key Dependencies

**Transmuter** (contracts/transmuter/Cargo.toml):
```toml
osmosis-std = "0.26.0"       # Osmosis-specific proto bindings
osmosis-test-tube = "26.0.1" # Integration testing
cosmwasm-std = "2.2.2"       # Core CosmWasm runtime
```

**Orderbook** (contracts/sumtree-orderbook/Cargo.toml):
```toml
osmosis-std = "0.25.0"       # Osmosis-specific proto bindings
osmosis-test-tube = "25.0.0" # Integration testing
cosmwasm-std = "1.5.4"       # Core CosmWasm runtime
```

### The Bech32 Prefix Problem

**Issue**: Contracts compiled with `osmosis-std` have Osmosis chain assumptions baked in:
- Address validation expects `osmo` prefix
- Tokenfactory denom generation uses `osmo` prefix
- Proto types are Osmosis-specific

**Symptoms in Gaia tests**:
1. `uploadAndInstantiateContract` hardcodes `"osmo"` bech32 prefix (pool_module_test.go:563)
2. Transmuter pool creation may fail address validation
3. Alloyed asset denoms generated with wrong prefix

**Root Cause**: The `osmosis-std` package at `github.com/osmosis-labs/osmosis-rust/packages/osmosis-std` contains:
- Proto-generated types from Osmosis chain
- Stargate message types specific to Osmosis
- Tests that assume Osmosis chain configuration

### Usage: Testing vs Production

| Context | Bytecode Source | Purpose |
|---------|-----------------|---------|
| **Unit Tests** | `x/cosmwasmpool/bytecode/` | Verify Go module logic works with real contracts |
| **Integration Tests** | `x/cosmwasmpool/bytecode/` | Test full pool lifecycle |
| **Osmosis Mainnet** | Uploaded via governance | Production pools |
| **Gaia (future)** | TBD - needs recompilation | Production pools on Gaia |

### Compilation Process

**Standard Osmosis compilation**:
```bash
# In transmuter repo
cd contracts/transmuter
cargo build --release --target wasm32-unknown-unknown

# Or use optimizer (recommended for production)
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/rust-optimizer:0.14.0
```

**For Gaia** (requires investigation):
1. Fork osmosis-std → create gaia-std (or cosmos-std)
2. Update contract dependencies to use new bindings
3. Recompile with Gaia chain parameters
4. Upload new bytecode

---

## §7 CosmWasm Test Issues (Current State)

### Test Failures in `x/cosmwasmpool`

| Test | Issue | Root Cause |
|------|-------|------------|
| `TestUploadCodeIdAndWhitelist/error:_cw_pool_module_account_does_not_have_upload_access` | Expected error but got nil | Wasmd permission model changed |
| `TestUploadCodeIdAndWhitelist/happy_path` | Code ID mismatch (expected 1, got 2) | Pre-existing code uploads in test setup |
| `TestInitializePool/invalid_pool_type` | CL pool creation fails | Permissionless CL creation disabled |
| `TestSudoGasLimit/contract_consumes_less_than_limit` | Bech32 address error | Hardcoded "osmo" prefix in test helper |

### Test Helpers with Hardcoded Prefixes

**pool_module_test.go:563**:
```go
bech32CWAddr, err = sdk.Bech32ifyAddressBytes("osmo", rawCWAddr)
```
Should be:
```go
bech32CWAddr, err = sdk.Bech32ifyAddressBytes("cosmos", rawCWAddr)
```

### Permissionless CL Pool Creation

Tests that create CL pools fail because permissionless creation is disabled by default. The apptesting infrastructure should enable this in genesis params.

---

## §8 Reference Quality Notes

### Highly Useful ⭐

| Reference | What Made It Useful | Used For |
|-----------|---------------------|----------|
| <https://github.com/osmosis-labs/transmuter> | Complete contract source with build config | Understanding transmuter compilation |
| <https://github.com/osmosis-labs/orderbook> | Complete orderbook contract source | Understanding orderbook compilation |
| transmuter README | Detailed API documentation | Understanding instantiate/execute messages |

### Less Useful Than Expected

| Reference | Why It Didn't Help | Alternative |
|-----------|-------------------|-------------|
| osmosis-std lib.rs | Just a module re-export, no bech32 config visible | Check types/ submodules |

### New References to Consider

| Source | Why It Might Help |
|--------|-------------------|
| <https://github.com/osmosis-labs/osmosis-rust> | Full osmosis-std source, needed to understand proto bindings |
| <https://github.com/CosmWasm/wasmd> v0.60 release notes | API changes from v0.53 to v0.60 |
| cosmwasm-std documentation | Understanding chain-agnostic vs chain-specific patterns |

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-28 | Initial document creation | AI Assistant |
| 2026-01-28 | Added §6 CosmWasm Pool Contracts with transmuter/orderbook sources, dependencies, and bech32 prefix problem | AI Assistant |
| 2026-01-28 | Added §7 CosmWasm Test Issues documenting current test failures and causes | AI Assistant |
| 2026-01-28 | Cloned transmuter, orderbook, osmosis-rust repos. Updated §6 with proto compatibility discovery - Gaia tokenfactory uses same protos! | AI Assistant |