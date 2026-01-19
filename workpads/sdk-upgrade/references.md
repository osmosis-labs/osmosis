# SDK Upgrade References

## §1 Official Documentation

### Cosmos SDK
- **SDK Docs**: <https://docs.cosmos.network/>
- **SDK Changelog**: <https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md>
- **SDK Upgrade Guide**: <https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md>
- **SDK v0.53 Release Notes**: <https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.0>

### Migration Guides
- **v0.50 → v0.53 Migration**: <https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md>
- **State Migrations**: <https://docs.cosmos.network/v0.53/build/migrations/>

#### v0.50.x → v0.53.x Migration Notes (UPGRADING.md v0.53.4)

| Note | Osmosis Impact | SDK Fork Impact |
|------|----------------|-----------------|
| **Coordinated chain upgrade required** | Requires standard on-chain upgrade process (gov proposal, upgrade handler, state migration testing). | Fork needs to be retired or rebased to upstream; existing fork-based upgrade workflows won’t apply post-migration. |
| **Unordered transactions (opt-in)** | Decide whether Osmosis wants unordered txs; if yes, update auth keeper wiring and ante `SigVerificationDecorator` options; otherwise no change. | Feature does not exist in fork; if adopting upstream, decide to enable/disable and ensure fork-only hooks don’t rely on sequence assumptions. |
| **Auth module `PreBlocker` required** | Update `ModuleManager.SetOrderPreBlockers` to include `authtypes.ModuleName`. | Fork does not enforce this; upstream wiring must include this change. |
| **New modules: `x/epoch` and `x/protocolpool` (optional)** | Osmosis already has `x/epochs` module; check for name conflicts and wiring differences if adopting SDK `x/epoch`. Protocol pool likely not used; if enabled, update distribution queries/messages to new module. | Fork does not include these modules; if upstreaming, decide whether to include and plan `StoreUpgrade`. |
| **Custom mint function (`MintFn`)** | If Osmosis uses custom inflation logic, move to `mintkeeper.WithMintFn`; ensure `InflationCalculationFn` is removed. | Fork code likely lacks `MintFn`; will need to port any customization into upstream-supported hooks. |
| **CheckTx handler support, ed25519 verification enabled** | Review any custom `CheckTx` logic; ensure any ed25519-based keys or tests still pass. | Fork may have diverged sig verification; upstream version should be accepted unless fork changes are re-applied. |
| **`testnet init-files` flag changes** | Update any local scripts that call `init-files` flags (validator count, staking denom, commit timeout, single-host). | Fork scripts may rely on old flags; these should be updated to upstream defaults. |

#### Version Availability Notes

- No stable `v0.51.x` or `v0.52.x` tags exist in upstream; the `v0.53.x` UPGRADING guide is the canonical path from `v0.50.x` → `v0.53.x`.

---

## §2 Reference Implementations

### Gaia (Cosmos Hub)
- **Repository**: <https://github.com/cosmos/gaia>
- **v25.3.0 Release**: <https://github.com/cosmos/gaia/releases/tag/v25.3.0>
- **SDK Upgrade PR**: (to be found - v25.0.0 upgrade)

### Other SDK v0.53 Chains
_(To be populated as references are found)_

### Previous Osmosis SDK Upgrades (Changelog)

| Upgrade | Notes | Osmosis Impact |
|---------|-------|----------------|
| **v26.0.0** (SDK v0.50 + Comet v0.38) | PR #8274 upgraded SDK and Comet; added tagged fork versions (v0.50.6-v26-osmo-1, v0.38.11-v26-osmo-1). | Indicates prior large SDK jump required state breaking upgrade; expect similar coordination for v0.53. |
| **v11 / v10.1** (SDK v0.45.0x-osmo) | PR #2245 and #2146 upgraded SDK fork with new governance deposit rules, concurrency query client, log changes, vesting CLI changes. | Past upgrades frequently required custom fork patches; expect equivalent reconciliation work for v0.50 → v0.53. |

---

## §3 Osmosis SDK Fork

### Fork Repository
- **Repository**: <https://github.com/osmosis-labs/cosmos-sdk>
- **Current Tag**: v0.50.14-v30-osmo
- **Comparison**: <https://github.com/cosmos/cosmos-sdk/compare/v0.50.14...osmosis-labs:cosmos-sdk:v0.50.14-v30-osmo>

### Key Commits/PRs
_(To be documented - list significant fork modifications)_

### Fork vs Upstream v0.50.14 (osmo-v30/0.50.14)

| Area | Fork Delta (from v0.50.14) | Osmosis Impact When Upgrading |
|------|----------------------------|-------------------------------|
| **Bank hooks + supply offsets** | New bank hooks (`TrackBeforeSend`, `BlockBeforeSend`) + supply offset keeper logic and migrations. | Must re-apply or replace if upstream lacks these; Osmosis modules using these hooks need equivalent APIs. |
| **Vesting (clawback + cliff)** | Added clawback vesting account, cliff vesting CLI, and proto changes. | Verify Osmosis uses clawback/cliff vesting; may need to migrate to upstream features or re-implement. |
| **Slashing + staking performance** | Slashing changes (no per-block sign info write), migration key adjustments, staking delegation index fixes. | Must ensure upstream slashing behavior matches; otherwise re-apply patches or adjust module usage. |
| **IAVL/store pruning changes** | Async pruning, pruning fixes, store/rootmulti adjustments, pruning error handling. | If upstream v0.53 lacks equivalent behavior, re-apply or accept behavior change; validate for mainnet performance. |
| **CheckTx/RecheckTx behavior** | Skip `ValidateBasic` on recheck; OTEL wiring for gRPC server. | Review any custom ante/recheck assumptions; keep or drop OTEL instrumentation. |
| **Proto + gRPC query deltas** | Regenerated protos and gRPC services (bank, vesting, slashing). | When moving to v0.53, re-run proto regen and reconcile client changes. |

---

## §4 Local Reference Repositories

Reference repositories are stored in `repos/` (gitignored).

### Setup Commands

```bash
cd workpads/sdk-upgrade/repos

# Clone upstream SDK for comparison
git clone --depth 1 --branch v0.53.4 https://github.com/cosmos/cosmos-sdk.git cosmos-sdk-v0.53.4

# Clone Osmosis SDK fork for comparison
git clone --depth 1 --branch v0.50.14-v30-osmo https://github.com/osmosis-labs/cosmos-sdk.git cosmos-sdk-osmo-fork

# Clone Gaia for reference
git clone --depth 1 --branch v25.3.0 https://github.com/cosmos/gaia.git gaia-v25.3.0

# Clone base SDK v0.50.14 for diff
git clone --depth 1 --branch v0.50.14 https://github.com/cosmos/cosmos-sdk.git cosmos-sdk-v0.50.14
```

### Repository Paths

| Repo | Local Path | Purpose |
|------|-----------|---------|
| SDK v0.53.4 | `repos/cosmos-sdk-v0.53.4/` | Target SDK version |
| SDK v0.50.14 | `repos/cosmos-sdk-v0.50.14/` | Base for fork comparison |
| Osmosis Fork | `repos/cosmos-sdk-osmo-fork/` | Current fork with patches |
| Gaia v25.3.0 | `repos/gaia-v25.3.0/` | Reference implementation |

### Local Checkout Notes

- **Local SDK checkout**: `/Users/nicolas/devel/cosmos-sdk`
- **Aligned branch**: `osmo-v30/0.50.14` (tracks `origin/osmo-v30/0.50.14`, tag `v0.50.14-v30-osmo`)

---

## §5 Useful Diff Commands

```bash
# Find Osmosis fork patches
cd repos
diff -rq cosmos-sdk-v0.50.14 cosmos-sdk-osmo-fork --exclude=".git" | head -50

# Compare specific modules
diff -u cosmos-sdk-v0.50.14/x/bank cosmos-sdk-osmo-fork/x/bank

# View SDK changelog between versions
cd cosmos-sdk-v0.53.4
git log --oneline v0.50.14..v0.53.4 -- CHANGELOG.md
```

---

## §6 IBC / WASM Compatibility

### IBC-Go
- **Repository**: <https://github.com/cosmos/ibc-go>
- **Compatibility Matrix**: <https://github.com/cosmos/ibc-go/blob/main/RELEASES.md>

### CosmWasm
- **Repository**: <https://github.com/CosmWasm/wasmd>
- **SDK Compatibility**: <https://github.com/CosmWasm/wasmd/blob/main/INTEGRATION.md>

---

## §7 Related Issues / PRs

_(Track relevant upstream issues and PRs here)_

| Link | Description | Status |
|------|-------------|--------|
| | | |

---

## Reference Quality Notes

### Highly Useful ⭐

| Reference | What Made It Useful | Used For |
|-----------|---------------------|----------|
| Cosmos SDK v0.53.4 UPGRADING.md | Canonical migration notes from v0.50 → v0.53 | Migration guide summary + repo-specific impact |
| Osmosis CHANGELOG.md | Historical SDK upgrade context and prior patterns | Prior upgrade notes |
| Local osmosis-labs/cosmos-sdk fork (`osmo-v30/0.50.14`) | Concrete diff vs upstream v0.50.14 | Fork delta summary |

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
| 2026-01-13 | Initial document creation | AI Assistant |
| 2026-01-19 | Add prior upgrade notes, migration guide summary, fork delta notes | AI Assistant |
