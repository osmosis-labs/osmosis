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

#### Module Wiring Deltas (Osmosis v31 vs Gaia v25.3.0)

| Area | Observation | Osmosis Impact |
|------|-------------|----------------|
| **PreBlocker ordering** | Gaia sets `SetOrderPreBlockers(upgradetypes.ModuleName, authtypes.ModuleName)`; Osmosis only sets upgrade module. | Add `authtypes.ModuleName` to pre-blockers when moving to SDK v0.53. |
| **Epoch module naming** | Osmosis uses `x/epochs` (custom module). SDK v0.53 introduces optional `x/epoch`. | Avoid wiring SDK `x/epoch` unless explicitly desired; ensure no name conflict and plan store upgrade if added. |
| **Protocol pool** | Gaia v25.3.0 does not wire `x/protocolpool`; SDK v0.53 describes it as optional. | Likely skip `x/protocolpool` unless Osmosis wants the new community pool semantics; if added, requires store upgrade and distribution wiring changes. |
| **Capability module (IBC v10)** | Osmosis still wires `capability` module and uses scoped keepers. IBC-Go v10 removes capability module. | Plan removal of capability module and related store keys when upgrading IBC to v10; adjust begin/init ordering accordingly. |

#### Version Availability Notes

- No stable `v0.51.x` or `v0.52.x` tags exist in upstream; the `v0.53.x` UPGRADING guide is the canonical path from `v0.50.x` → `v0.53.x`.

#### State/Store Upgrade Plan (Draft)

| Area | Store/State Change | Osmosis Impact |
|------|--------------------|----------------|
| **IBC capability module (v10)** | Remove `capability` KV store key and mem store key; delete scoped keepers and module wiring. | Requires a store upgrade that removes the capability store keys and updates app wiring. |
| **Optional SDK modules** | `x/epoch` and `x/protocolpool` require new store keys if adopted. | Decide whether to add; if yes, include store keys and upgrade handler additions. |
| **Wasm module (v0.60.x)** | Potential x/wasm state migrations or params changes per wasmd changelog. | Audit `x/wasm` migrations between v0.53.x + wasmd v0.60.x; add migration handlers as needed. |
| **IBC / ICA / PFM / rate-limit** | Version bumps may include module-specific migrations (params or store layout). | Review module migrations in v10 series and add to upgrade handler if required. |

**Upgrade handler outline**:
1. Define `storetypes.StoreUpgrades` with added/removed store keys (capability removal; optional adds).
2. Run module migrations in order (SDK/IBC/wasm) using `UpgradeKeeper`.
3. Validate no orphaned store keys remain (remove capability keys if present).

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

#### IBC-Go v10 Migration Notes (v8.x → v10)

Sources:
- Migration guide: <https://ibc.cosmos.network/v10/migrations/v8_1-to-v10/>
- Release notes: <https://docs.cosmos.network/ibc/v10.1.x/changelog/release-notes>
- GitHub releases: <https://github.com/cosmos/ibc-go/releases>

| Change | Osmosis Impact |
|--------|----------------|
| **Module versioning in lockstep with v10** | Update import paths to `/v10` for ibc-go core and `08-wasm`; ensure all IBC-related modules use v10 series. |
| **Capability module removed** | Remove `CapabilityKeeper`, scoped keepers, and capability store keys; adjust app wiring and any IBC stack code relying on capabilities. |
| **ICS29 fee middleware removed** | If Osmosis uses IBC fee middleware, remove from stack and module manager; verify no fee-related store keys or params remain. |
| **Channel upgradeability removed** | Remove any channel upgrade feature usage or wiring. |
| **IBC v2 support added (optional)** | Decide whether to wire IBC v2 transfer stack (`transferv2` + callbacks v2). If not, keep v1 stack. |
| **Legacy proposal route removal** | Remove legacy 02-client proposal route wiring if present. |
| **API removals (LookupModuleByChannel/Port, ChannelCapabilityPath, ICA Authenticate/ClaimCapability)** | Update any Osmosis modules or middleware referencing these APIs. |

### CosmWasm
- **Repository**: <https://github.com/CosmWasm/wasmd>
- **SDK Compatibility**: <https://github.com/CosmWasm/wasmd/blob/main/INTEGRATION.md>

#### Wasmd v0.60.x / CosmWasm Compatibility Notes

Sources:
- Wasmd releases: <https://github.com/CosmWasm/wasmd/releases>
- Wasmd changelog: <https://github.com/CosmWasm/wasmd/blob/main/CHANGELOG.md>
- INTEGRATION.md: <https://github.com/CosmWasm/wasmd/blob/main/INTEGRATION.md>
- Advisory (IBC channel open error handling): <https://advisories.gitlab.com/pkg/golang/github.com/cosmwasm/wasmd/GHSA-79xg-q4qm-7v9w/>

| Change | Osmosis Impact |
|--------|----------------|
| **No clear v0.60.2 migration guide published** | Plan to audit the tag/changelog directly; treat as potentially consensus-affecting until confirmed. |
| **IBC channel open error handling fix (CWA-2025-006)** | v0.60.1+ fixes erroneous channel open when contract errors; ensure behavior change is acceptable for Osmosis IBC hooks and tests. |
| **WasmVM minor version bumps are consensus-breaking** | Align wasmvm to wasmd v0.60.x (`v2.2.4` in Gaia); requires coordinated chain upgrade. |
| **Go toolchain requirement bump (v0.60.0+)** | Ensure build toolchain matches wasmd requirements during upgrade CI. |
| **Capability flags / feature gates** | Review `INTEGRATION.md` for capability flags and ensure Osmosis enables required features when moving to v0.60.x. |

### Dependency Alignment Matrix (SDK v0.53.4 baseline)

Source: `gaia` v25.3.0 `go.mod` (SDK v0.53.4).

| Package | Target Version | Notes |
|---------|----------------|-------|
| **IBC-Go** | `github.com/cosmos/ibc-go/v10 v10.5.0` | Gaia uses IBC-Go v10 for SDK v0.53.4. |
| **IBC-Go 08-wasm** | `github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10 v10.3.0` | Matches IBC-Go v10 series. |
| **IBC Apps PFM** | `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10 v10.1.0` | Gaia baseline. |
| **IBC Apps Rate Limiting** | `github.com/cosmos/ibc-apps/modules/rate-limiting/v10 v10.1.0` | Gaia baseline. |
| **Wasmd** | `github.com/CosmWasm/wasmd v0.60.2` | Gaia baseline for SDK v0.53.4. |
| **WasmVM** | `github.com/CosmWasm/wasmvm/v2 v2.2.4` | Matches Gaia. |
| **CometBFT** | `github.com/cometbft/cometbft v0.38.20` | Gaia baseline for SDK v0.53.4. |
| **CometBFT DB** | `github.com/cometbft/cometbft-db v1.0.4` | Gaia baseline. |
| **cosmossdk.io/api** | `v0.9.2` | Gaia baseline. |
| **cosmossdk.io/client/v2** | `v2.0.0-beta.9` | Gaia baseline. |
| **cosmossdk.io/core** | `v0.11.3` | Gaia baseline. |
| **cosmossdk.io/errors** | `v1.0.2` | Gaia baseline. |
| **cosmossdk.io/log** | `v1.6.1` | Gaia baseline. |
| **cosmossdk.io/math** | `v1.5.3` | Gaia baseline. |
| **cosmossdk.io/store** | `v1.1.2` | Gaia baseline. |
| **cosmossdk.io/x/tx** | `v0.14.0` | Gaia baseline. |
| **cosmossdk.io/x/upgrade** | `v0.2.0` | Gaia baseline. |
| **cosmossdk.io/x/evidence** | `v0.2.0` | Gaia baseline. |
| **cosmossdk.io/x/feegrant** | `v0.2.0` | Gaia baseline. |

#### Conflicts vs current Osmosis `go.mod`

- **IBC-Go**: Osmosis uses `v8.7.0` + `08-wasm` `v0.4.2-*`; target is v10.x.
- **IBC apps**: Osmosis uses `packet-forward-middleware/v8` and `async-icq/v8`; Gaia baseline is v10 series (no async-icq in Gaia).
- **Wasmd**: Osmosis uses `v0.53.3`; Gaia baseline is `v0.60.2`.
- **CometBFT**: Osmosis uses `v0.38.17`; Gaia baseline is `v0.38.20`.
- **cosmossdk.io/client/v2**: Osmosis uses `v2.0.0-beta.6`; Gaia baseline is `beta.9`.
- **cosmossdk.io/core**: Osmosis uses `v0.12.1-*` but replaces to `v0.11.0`; Gaia baseline is `v0.11.3`.
- **cosmossdk.io/errors/log/store/x/*:** Osmosis pins older versions (`errors v1.0.1`, `log v1.6.0`, `store v1.1.1`, `x/tx v0.13.7`, `x/upgrade v0.1.4`, `x/evidence v0.1.1`).
- **cosmossdk.io/store replace**: Osmosis uses a forked `cosmossdk.io/store` replace; needs reconciliation for v0.53.4.

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
| Gaia v25.3.0 `go.mod` | Authoritative dependency versions for SDK v0.53.4 | Dependency alignment matrix |
| IBC-Go v10 migration guide | Concrete IBC v10 breaking changes and wiring changes | IBC v10 migration notes |
| IBC-Go v10 release notes | API removals and module changes | IBC v10 migration notes |
| Wasmd changelog/releases | Primary source for v0.60.x changes | Wasmd compatibility notes |
| Wasmd integration docs | Capability flags and runtime compatibility | Wasmd compatibility notes |

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
| 2026-01-19 | Add dependency alignment matrix from Gaia v25.3.0 | AI Assistant |
| 2026-01-19 | Add IBC-Go v10 migration notes | AI Assistant |
| 2026-01-19 | Add Wasmd v0.60.x compatibility notes | AI Assistant |
| 2026-01-19 | Add module wiring deltas vs Gaia v25.3.0 | AI Assistant |
| 2026-01-19 | Add draft state/store upgrade plan | AI Assistant |
