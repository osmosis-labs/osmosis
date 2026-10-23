# Go security dependency follow-up

Snapshot: 2026-10-23, open Dependabot alerts in
[osmosis-labs/osmosis](https://github.com/osmosis-labs/osmosis/security/dependabot).
Python, Rust and JavaScript alerts are outside this change's scope.
This follow-up branches directly from `main`; the earlier small-bump and
Redoc-removal changes are not included.

GitHub reports each manifest separately. A patched workspace/root selection
**does not** repair a nested module for standalone consumers. Do not dismiss
nested-module alerts just because the workspace uses a newer version.

## Deferred: small bumps from the separate follow-up

- [ ] Root `go.mod`: `github.com/klauspost/compress` 1.18.5 → **1.18.7**
  ([#692](https://github.com/osmosis-labs/osmosis/security/dependabot/692)).
- [ ] Root: `github.com/moby/sys/user` 0.3.0 → **0.4.1**
  ([#705](https://github.com/osmosis-labs/osmosis/security/dependabot/705)).
- [ ] Root: `github.com/aws/aws-sdk-go-v2/service/s3` 1.97.1 → **1.97.3**,
  including the required AWS companion-module patch bumps
  ([#704](https://github.com/osmosis-labs/osmosis/security/dependabot/704)).
- [ ] `osmoutils`, `x/epochs`, `x/ibc-hooks`: `filippo.io/edwards25519`
  1.1.0 → **1.1.1** (alerts #489–491).
- [ ] Same three modules: `github.com/dvsekhvalnov/jose2go` 1.6.0 → **1.7.0**
  (alerts #453–455).
- [x] `x/ibc-hooks`: `github.com/ulikunitz/xz` 0.5.11 → **0.5.15**
  ([#440](https://github.com/osmosis-labs/osmosis/security/dependabot/440)).

Nine small-bump alert instances remain deferred on this branch. The `xz`
instance is covered by the go-getter companion upgrades below.

## Completed: nested-module toolchain and dependency alignment

All four nested modules now declare Go **1.25.8**, matching root and workspace.
The obsolete **go1.24.1** toolchain directives are removed. Updates were targeted
with `GOWORK=off`, not `go work sync` or a repository-wide `go get -u`.

Standalone dependency alignment now mirrors root's `cosmossdk.io/core` **0.11.0**
replacement in all four modules and its existing store fork in `x/epochs` and
`x/ibc-hooks`. These resolve the SDK's `comet.BlockInfo` and IAVL fast-node
whitelist API mismatches without changing SDK or CometBFT fork versions.

- [x] **Critical: `golang.org/x/crypto`** in `osmomath`, `osmoutils`, `x/epochs`,
  and `x/ibc-hooks`: aligned with root **0.55.0** (minimum fix **0.52.0**).
  Covers 60 snapshot alert instances, including SSH
  authentication/constraint bypasses, deadlocks, panics and resource exhaustion.
  Representative alert: [#622](https://github.com/osmosis-labs/osmosis/security/dependabot/622).
- [x] **Critical/high: `google.golang.org/grpc`** in all four nested modules:
  aligned with root **1.83.2** (minimum fix **1.83.1**).
  Covers 20 snapshot instances across five advisories: authorization bypass,
  xDS RBAC/header validation and HTTP/2 denial of service.
  Representative alert: [#506](https://github.com/osmosis-labs/osmosis/security/dependabot/506).
  Generated code and Cosmos SDK compatibility passed existing standalone,
  workspace and IBC-hooks integration tests; no code regeneration was needed.
- [x] **Medium: `golang.org/x/net`** in all four nested modules: aligned
  with root **0.58.0** (minimum fix **0.55.0**).
  Four HTML-parser DoS alerts (#565–568;
  [GHSA-5cv4-jp36-h3mw](https://github.com/advisories/GHSA-5cv4-jp36-h3mw)).
- [x] **High: `github.com/klauspost/compress`** in all four nested modules:
  upgraded to **1.18.7**. Four alerts (#688–691;
  [GHSA-259r-337f-4rfw](https://github.com/advisories/GHSA-259r-337f-4rfw)).
- [x] **High: `github.com/hashicorp/go-getter`** in `osmoutils` and
  `x/ibc-hooks`: upgraded to **1.8.6** (Go **1.25.8+**), covering arbitrary file
  reads (#537–538;
  [GHSA-92mm-2pjq-r785](https://github.com/advisories/GHSA-92mm-2pjq-r785)).
  Required AWS v2/cloud/telemetry companion modules were updated by Go's module
  resolver. S3 was explicitly raised to **1.97.3** rather than introducing the
  known-vulnerable **1.97.1** minimum. Standalone and workspace tests passed.

## Deferred: SDK, consensus and container-runtime changes

- [ ] **High: `github.com/cometbft/cometbft` in `osmomath`**: 0.38.12 →
  **0.38.21+** (root already requires 0.38.23). Five alerts: #311, #354,
  #359, #443 and #479, covering vote-extension/BitArray panics, blocksync
  stalls, block-part validation and commit-signature/block-time consistency.
  Review the effective workspace replacements before updating: `osmoutils`,
  `x/epochs` and `x/ibc-hooks` still replace CometBFT with
  `github.com/osmosis-labs/cometbft` **0.38.17-v28-osmo-1**. Verify the fork
  contains later security fixes; do not simply remove it. Run consensus and
  state-compatibility tests before changing the effective consensus engine.
- [ ] **High: `github.com/CosmWasm/wasmd` in `x/ibc-hooks`**: 0.53.2 →
  **0.53.3+** (#419;
  [GHSA-79xg-q4qm-7v9w](https://github.com/advisories/GHSA-79xg-q4qm-7v9w)).
  Although a patch release, it changes IBC channel-opening error behavior.
  Validate against the SDK/IBC/wasmvm versions with channel-handshake and
  rollback tests rather than treating it as a utility-library bump.
- [ ] **High/medium: Docker/Moby and Docker CLI in root**: migrate the
  `github.com/ory/dockertest/v3` dependency chain together, not independent
  `+incompatible` major-version bumps. CLI needs **29.2.0+** for #500
  ([GHSA-p436-gjf2-799p](https://github.com/advisories/GHSA-p436-gjf2-799p)).
  Docker alerts #433, #514 and #549–551 cover firewall isolation, plugin
  validation and archive/copy host-filesystem attacks. Several have **no
  patched version recorded by GitHub**, so bumping Docker to 28 alone cannot
  close this group. Check for maintained replacement module paths and run
  Docker-backed tests on Linux and relevant Windows paths.
- [ ] **High/medium: `github.com/opencontainers/runc` in `osmoutils`**:
  1.1.14 → **1.3.6+** (#447, #449, #451, #563). Earlier escape fixes require
  1.2.8, but the latest `/dev` symlink advisory requires 1.3.6. Coordinate
  the runtime/API and toolchain changes with the container test dependencies;
  root already selects 1.3.6.
- [ ] **High: `github.com/shamaton/msgpack/v2`** in root, `osmoutils`,
  `x/epochs` and `x/ibc-hooks`: GitHub lists **no patched release**
  (#501–504;
  [GHSA-h9q6-hc68-35rp](https://github.com/advisories/GHSA-h9q6-hc68-35rp)).
  Trace the importing dependency, check whether untrusted MessagePack reaches
  it, and plan a parent upgrade, replacement or reviewed backport. Do not
  substitute an unverified version or dismiss the alert.

## Already patched in root before this change

These 30 open alert instances are stale relative to this checkout's root
manifest; let GitHub rescan after merge. This is not a claim that the nested
modules or effective fork replacements are patched.

| Root dependency | Existing version | Alert instances |
| --- | --- | --- |
| `golang.org/x/crypto` | 0.55.0 | 15 |
| `google.golang.org/grpc` | 1.83.2 | 5 |
| `golang.org/x/net` | 0.58.0 | 1 |
| `github.com/opencontainers/runc` | 1.3.6 | 4 |
| `github.com/hashicorp/go-getter` | 1.8.6 | 1 |
| `github.com/CosmWasm/wasmd` | 0.54.10 | 1 |
| `filippo.io/edwards25519` | 1.1.1 | 1 |
| `github.com/dvsekhvalnov/jose2go` | 1.7.0 | 1 |
| `github.com/ulikunitz/xz` | 0.5.15 | 1 |

Reconciliation against the snapshot: **150 Go alerts = 91 bumped here
(90 nested-alignment instances + 1 xz) + 30 already patched at root
+ 29 deferred manifest-level instances**. GitHub closes alerts only after
rescanning its tracked branch; no alerts were manually dismissed.

## Validation and follow-up acceptance checks

Passed locally with **GOTOOLCHAIN=go1.25.8**:

```sh
# In each of osmomath, osmoutils, x/epochs and x/ibc-hooks:
GOWORK=off go mod tidy
GOWORK=off go mod verify
GOWORK=off go test -mod=readonly -p 4 ./...
GOWORK=off go list -mod=readonly -m -json all

# From the repository root, with the workspace enabled:
go test -mod=readonly -p 4 ./osmomath/... ./osmoutils/... ./x/epochs/... ./x/ibc-hooks/...
go test -mod=readonly -p 4 ./tests/ibc-hooks/...
go list -mod=readonly -m -json all
```

The pre-existing standalone SDK API mismatches are resolved by the core/store
alignment above. `x/ibc-hooks` itself has no test files; the root's IBC-hooks
integration suite passed. No full-chain upgrade, consensus/state-compatibility,
or Docker-backed runtime validation is claimed by this change.

For each follow-up, test both workspace and standalone module builds, inspect
`go list -m -json all` for effective replacements, run relevant integration and
state-compatibility tests, and re-fetch Dependabot alerts after merge. Do not
assume a clean root build or a newer upstream tag proves a fork is patched.
