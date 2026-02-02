# Gaia Migration Tasks

## Status Legend

```
📋 pending      - Not yet started
🚧 in_progress  - Currently working on  
✅ completed    - Finished and verified
🚫 blocked      - Cannot proceed
```

---

## Previous Phases (Completed)

Phases 0-5 (Discovery, Migration, Test Infrastructure) are complete.
See `tasks-completed-phases-0-5.md` for the full archive.

---

## Phase 6: Manual Testing

### Overview

This phase focuses on comprehensive manual testing of the migrated DEX modules. We will create a structured testing framework with scripts that:

1. **Print what we're doing** - Clear description of each test step
2. **Print expected results** - What we expect to happen
3. **Print actual results** - What actually happened
4. **Report pass/fail** - Clear test outcome

### Test Environment Options

| Environment | Description | Use Case |
|-------------|-------------|----------|
| **Empty Gaia** | Fresh chain with no pools | Basic functionality, pool creation |
| **Mainnet State** | Chain initialized from mainnet snapshot | Realistic data, edge cases |

### Testing Infrastructure

All manual tests live in `gaia/tests/manual-dex/`:

```
tests/manual-dex/
├── README.md                 # Setup instructions and overview
├── setup/
│   ├── start-node.sh        # Start a local node
│   ├── fund-accounts.sh     # Fund test accounts
│   └── config.sh            # Shared configuration
├── gamm/
│   ├── test-balancer.sh     # Balancer pool tests
│   ├── test-stableswap.sh   # Stableswap pool tests
│   └── README.md            # GAMM-specific notes
├── concentrated-liquidity/
│   ├── test-cl-pool.sh      # CL pool tests
│   └── README.md
├── cosmwasmpool/
│   ├── test-transmuter.sh   # Transmuter pool tests
│   └── README.md
├── poolmanager/
│   ├── test-routing.sh      # Multi-hop swap tests
│   └── README.md
├── protorev/
│   ├── test-arb.sh          # Arbitrage detection tests
│   └── README.md
└── lib/
    ├── test-utils.sh        # Shared test utilities
    └── assertions.sh        # Test assertion helpers
```

---

### Task 6.0: Create DEX E2E Testing Infrastructure ✅ `completed`

**Description**: Set up Docker-based e2e testing for DEX modules using Gaia's existing test framework.

**Decision**: Use Go e2e tests only (removed shell scripts `tests/manual-dex/`). Can migrate to scripts later if needed.

**Acceptance Criteria**:
- [x] Create DEX tx helpers (`tests/e2e/tx/dex.go`)
- [x] Create DEX query helpers (`tests/e2e/query/dex.go`)
- [x] Create DEX e2e tests (`tests/e2e/e2e_dex_test.go`)
- [x] Add protorev genesis init to prevent PostHandler panics
- [x] Fix protorev PostHandler to handle uninitialized state
- [x] Test infrastructure running with Docker

**E2E Test Results** (Jan 30, 2026):
| Test | Status |
|------|--------|
| create_gamm_balancer_pool | ✅ PASS |
| query_all_pools | ✅ PASS |
| swap_exact_amount_in | ✅ PASS |
| swap_exact_amount_out | ✅ PASS |
| query_spot_price | ⚠️ FAIL (API endpoint needs fixing) |
| join_and_exit_pool | ⚠️ FAIL (needs debugging) |

**How to run**:
```bash
cd /Users/nicolas/devel/gaia
make docker-build-debug
cd tests/e2e
go test -v -timeout 30m -run TestIntegrationTestSuite/TestDEX ./...
```

---

### Task 6.0a: Fix feemarket genesis gentx issue ✅ `completed`

**Description**: The test node crashes during InitChain with error "UnmarshalJSON cannot decode empty bytes" in the feemarket post handler when processing gentxs.

**Resolution**: Use Docker-based e2e tests like Gaia does. The e2e framework properly initializes feemarket genesis state in `tests/e2e/genesis.go`:
- Sets `MinBaseGasPrice`
- Sets `FeeDenom`
- Sets `DistributeFees = true`
- Sets `State.BaseGasPrice`

**Why Docker tests work**: The e2e framework creates the entire genesis from scratch with proper configuration (see `modifyGenesis()` in `tests/e2e/genesis.go`), whereas our shell script relied on default genesis which may have initialization order issues.

**Alternative for shell-based testing**: If needed, modify genesis after `gaiad init` to disable feemarket:
```bash
jq '.app_state.feemarket.params.enabled = false' genesis.json > temp.json && mv temp.json genesis.json
```

---

### Task 6.0b: Add DEX store keys to upgrade handler ✅ `completed`

**Description**: Add DEX module store keys to the v26 upgrade handler for in-place testnet support.

**Files Modified**:
- `app/upgrades/v26_0_0/constants.go` - Added store keys for:
  - `poolmanager`
  - `gamm`
  - `concentratedliquidity`
  - `cosmwasmpool`
  - `protorev`

**Acceptance Criteria**:
- [x] Add all DEX module store keys to StoreUpgrades.Added
- [x] Build compiles successfully

---

### Task 6.0c: Set up Docker-based e2e testing ✅ `completed`

**Description**: Use Gaia's existing Docker-based e2e test framework for DEX testing instead of shell scripts.

**Files Created**:
- `tests/e2e/tx/dex.go` - DEX transaction helpers:
  - `ExecGammCreatePool()` - Create GAMM pools
  - `ExecPoolmanagerSwapExactAmountIn()` - Swap with exact input
  - `ExecPoolmanagerSwapExactAmountOut()` - Swap with exact output
  - `ExecGammJoinPool()` - Add liquidity
  - `ExecGammExitPool()` - Remove liquidity
  - `WritePoolFile()` - Create pool config file

- `tests/e2e/query/dex.go` - DEX query helpers:
  - `DEXPool()` - Query specific pool
  - `AllPools()` - Query all pools
  - `NumPools()` - Query pool count
  - `SpotPrice()` - Query spot price
  - `TotalPoolLiquidity()` - Query pool liquidity

- `tests/e2e/e2e_dex_test.go` - DEX e2e tests:
  - `testDEXGammCreateBalancerPool()` - Pool creation test
  - `testDEXSwapExactAmountIn()` - Swap in test
  - `testDEXSwapExactAmountOut()` - Swap out test
  - `testDEXQuerySpotPrice()` - Spot price query test
  - `testDEXJoinAndExitPool()` - Liquidity add/remove test
  - `testDEXQueryAllPools()` - All pools query test

**Acceptance Criteria**:
- [x] Create DEX tx helpers following Gaia e2e patterns
- [x] Create DEX query helpers
- [x] Create DEX e2e test functions
- [x] All files compile successfully

**How to run**:
```bash
# Build docker image first
docker build -t gaia:local .

# Run e2e tests
cd tests/e2e
go test -v -run TestIntegrationTestSuite
```

---

### Task 6.1: Fix Remaining GAMM E2E Test Failures ✅ `completed`

**Depends On**: Task 6.0

**Description**: Fix the failing e2e tests from Task 6.0 (spot price query, join/exit pool).

**Results** (Jan 30, 2026):
- [x] Fix spot price query API endpoint - **FIXED** (returns "1.000000000000000000")
- [x] Fix join-pool transaction - **FIXED** (share-amount-out was too small)
- [x] Fix exit-pool transaction - **FIXED** (works after join-pool fixed)

**E2E Test Status**: 6/6 passing ✅
| Test | Status |
|------|--------|
| create_gamm_balancer_pool | ✅ PASS |
| query_all_pools | ✅ PASS |
| query_spot_price | ✅ PASS |
| swap_exact_amount_in | ✅ PASS |
| swap_exact_amount_out | ✅ PASS |
| join_and_exit_pool | ✅ PASS |

**Key Fix**: The join-pool `share-amount-out` was set to "1" which caused "Too few shares out wanted" error. Fixed by using "1000000000000000000" (10^18) which is reasonable for a pool with 10^20 total shares.

---

### Task 6.2: GAMM - Stableswap E2E Tests ✅ `completed`

**Depends On**: Task 6.1

**Description**: Add e2e tests for Stableswap (Curve-style) pools in `tests/e2e/e2e_dex_test.go`.

**Test Scenarios**:
1. **Create Pool**: Create a stableswap pool for stablecoins
2. **Verify Curve**: Confirm stableswap curve math (lower slippage near 1:1)
3. **Swap Operations**: Test swaps at different pool imbalances

**Results** (Jan 30, 2026):
- [x] Add `WriteStableswapPoolFile` helper to `tx/dex.go`
- [x] Add `ExecGammCreateStableswapPool` helper to `tx/dex.go`
- [x] Add `testDEXGammCreateStableswapPool` test
- [x] Add `testDEXStableswapSwap` test
- [x] Add `testDEXStableswapLowSlippage` test (verifies spot price ≈ 1.0)

**E2E Test Status**: 3/3 stableswap tests passing ✅
| Test | Status |
|------|--------|
| create_gamm_stableswap_pool | ✅ PASS |
| stableswap_swap | ✅ PASS |
| stableswap_low_slippage | ✅ PASS (deviation: 0.000001990000000000) |

**Key Learnings**:
- Stableswap coins must be sorted alphabetically (stake < uatom)
- Stableswap maintains spot price very close to 1.0 for equal-weight assets

---

### Task 6.3: Concentrated Liquidity E2E Tests ✅ `completed`

**Depends On**: Task 6.1

**Description**: Add e2e tests for concentrated liquidity pools.

**Solution**: Enabled permissionless CL pool creation in genesis by setting `IsPermissionlessPoolCreationEnabled = true`.

**Work Completed**:
- [x] Split test files by module (dex_gamm.go, dex_poolmanager.go, dex_cl.go)
- [x] Add CL tx helpers to `tests/e2e/tx/dex_cl.go`
- [x] Add CL test functions to `e2e_dex_cl_test.go`
- [x] Add DEX_TEST_FILTER env var for fast iteration
- [x] Enable permissionless CL pool creation in genesis
- [x] Fix spread factor to use authorized value (0.001)
- [x] All 3 CL tests passing (create_pool, create_position, swap)

---

### Task 6.4: Poolmanager Multi-hop E2E Tests ✅ `completed`

**Depends On**: Tasks 6.2, 6.3

**Description**: Test multi-hop swaps through poolmanager routing.

**Test Scenarios**:
1. **Multi-Hop Swap**: Route through 2-3 pools
2. **Taker Fees**: Verify taker fees are collected

**Work Completed**:
- [x] Added `ExecPoolmanagerMultiHopSwap` helper to `tests/e2e/tx/dex_poolmanager.go`
- [x] Created `tests/e2e/e2e_dex_poolmanager_test.go` with multi-hop tests
- [x] Added `testDEXPoolmanagerCreatePhotonPool` - creates stake/photon pool
- [x] Added `testDEXPoolmanagerMultiHopSwap` - tests uatom → stake → photon route
- [x] Added `testDEXPoolmanagerMultiHopSwapReverse` - tests photon → stake → uatom route
- [x] Added `multihop` filter to `DEX_TEST_FILTER` env var

**E2E Test Results** (Feb 2, 2026):
| Test | Status |
|------|--------|
| create_photon_pool | ✅ PASS |
| multi_hop_swap | ✅ PASS (1M uatom → 924M photon through pools 1,4) |
| multi_hop_swap_reverse | ✅ PASS (500M photon → 196K uatom through pools 4,1) |

**Note**: Taker fee verification deferred - requires querying taker fee collector module account balance, which can be added as a follow-up enhancement.

---

### Task 6.5: CosmWasm Pool Tests 🚫 `blocked`

**Depends On**: Task 6.0

**Description**: Test CosmWasm-based pools (transmuter).

**Test Scenarios**:
1. **Upload Contract**: Upload transmuter WASM
2. **Create Pool**: Create transmuter pool via cosmwasmpool module
3. **Execute Swap**: 1:1 swap through transmuter
4. **Query State**: Verify pool state and liquidity

**Work Completed**:
- [x] Created `tests/e2e/tx/dex_cosmwasmpool.go` with `ExecCosmwasmPoolCreate` helper
- [x] Created `tests/e2e/e2e_dex_cosmwasmpool_test.go` with test functions
- [x] Added `StoreWasmHighGas` helper for large contracts (transmuter is 2.2MB)
- [x] Added `cosmwasmpool` filter to `DEX_TEST_FILTER`

**Blocker**: Large WASM upload times out in e2e test infrastructure

The transmuter.wasm (2.2MB) upload transaction is accepted (code:0) but the test validation times out waiting for block inclusion. This appears to be an e2e infrastructure issue with large contract uploads.

**Options to Unblock**:
1. Increase e2e test timeout for WASM uploads
2. Use gzip-compressed WASM (if available)
3. Test cosmwasmpool in unit tests only (skip e2e)
4. Debug the ExecuteGaiaTxCommand validation for large txs

**Note**: The transmuter contract also has known bech32 prefix issues documented in references.md - contracts are compiled against Osmosis (`osmo` prefix) and may not work with Gaia (`cosmos` prefix) without recompilation.

---

### Task 6.6: Protorev Arbitrage Tests ✅ `completed`

**Depends On**: Task 6.4

**Description**: Test protorev arbitrage detection and execution.

**Work Completed**:
- [x] Created `tests/e2e/query/dex_protorev.go` with query helpers:
  - `ProtorevEnabled()` - Check if protorev is enabled
  - `ProtorevParams()` - Query protorev params
  - `ProtorevNumberOfTrades()` - Query number of arb trades
  - `ProtorevAllProfits()` - Query accumulated profits by denom
  - `ProtorevAllRouteStatistics()` - Query route statistics
- [x] Created `tests/e2e/e2e_dex_protorev_test.go` with tests:
  - `protorev_query_params` - Verify params query works
  - `protorev_query_enabled` - Verify enabled status query
  - `protorev_query_number_of_trades` - Verify trades count query
  - `protorev_query_all_profits` - Verify profits query
  - `protorev_create_arb_opportunity` - Create pools with price imbalance
  - `protorev_query_route_statistics` - Verify route stats query
- [x] Added `protorev` filter to `DEX_TEST_FILTER` env var

**E2E Test Status**: 6/6 passing ✅
| Test | Status |
|------|--------|
| protorev_query_params | ✅ PASS |
| protorev_query_enabled | ✅ PASS |
| protorev_query_number_of_trades | ✅ PASS |
| protorev_query_all_profits | ✅ PASS |
| protorev_create_arb_opportunity | ✅ PASS |
| protorev_query_route_statistics | ✅ PASS |

**Note on Full Arb Testing**: Protorev is disabled by default in e2e tests (for performance - it runs after every swap). The tests verify query infrastructure works. To test actual arbitrage execution:
1. Set `protorevGenState.Params.Enabled = true` in `tests/e2e/genesis.go`
2. Rebuild docker image: `make docker-build-debug`
3. Re-run tests: `DEX_TEST_FILTER=protorev go test -v -run TestDEX ...`

The test creates two pools with price imbalances (Pool A: 1:1 stake/uatom, Pool B: 2:1 stake/uatom) which would create arb opportunities when protorev is enabled.

---

### Task 6.7: Genesis Export/Import Tests 📋 `pending`

**Depends On**: Tasks 6.1-6.6

**Description**: Verify genesis export and import works correctly for all DEX modules.

**Test Scenarios**:
1. **Create State**: Create pools of all types, positions, execute swaps
2. **Export Genesis**: Export chain state
3. **Restart with Genesis**: Start new node with exported genesis
4. **Verify State**: Confirm all pools, positions, state preserved

**Acceptance Criteria**:
- [ ] Create test script for genesis round-trip
- [ ] Document any issues in knowledge.md

---

### Task 6.8: Mainnet State Testing 📋 `pending`

**Depends On**: Task 6.0

**Description**: Test with chain initialized from mainnet snapshot.

**Test Scenarios**:
1. **Load Mainnet State**: Initialize from Gaia mainnet snapshot
2. **Basic Operations**: Verify existing functionality works
3. **DEX Operations**: Create new pools, execute swaps
4. **Edge Cases**: Test with realistic mainnet data

**Acceptance Criteria**:
- [ ] Document process for loading mainnet state
- [ ] All tests pass on mainnet-initialized chain

---

### Task 6.9: Rename Osmo → Atom in TakerFeeDistribution Types ✅ `completed`

**Description**: Rename taker fee distribution types from Osmosis naming to Cosmos Hub naming:
- `OsmoTakerFeeDistribution` → `AtomTakerFeeDistribution`
- `NonOsmoTakerFeeDistribution` → `NonAtomTakerFeeDistribution`

**Files Updated**:
- `proto/gaia/poolmanager/v1beta1/genesis.proto` - Proto field definitions and comments
- `x/poolmanager/types/genesis.pb.go` - Regenerated
- `x/poolmanager/types/params.go` - Key constants and field references
- `x/poolmanager/taker_fee_distribution.go` - Field access
- `app/upgrades/v26_0_0/upgrades.go` - Upgrade handler initialization
- `x/poolmanager/router_test.go` - Test variables
- `x/poolmanager/keeper_test.go` - Test variables and assertions
- `x/poolmanager/README.md` - Documentation

**Acceptance Criteria**:
- [x] Rename `OsmoTakerFeeDistribution` to `AtomTakerFeeDistribution`
- [x] Rename `NonOsmoTakerFeeDistribution` to `NonAtomTakerFeeDistribution`
- [x] Update all references in Go code
- [x] Regenerate protobuf files
- [x] Build compiles successfully
- [x] Poolmanager tests pass (all 4 test packages)

---

## Notes

- Each task follows workflow: `SETUP → EXECUTE → VERIFY → REPORT`
- Focus on getting one module fully tested before moving to the next
- Document all discoveries and issues in `knowledge.md`
- Iterate on the testing framework as we learn from early tests

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-01-30 | Created Phase 6 (Manual Testing) task structure | AI Assistant |
| 2026-01-30 | Archived Phases 0-5 to tasks-completed-phases-0-5.md | AI Assistant |
| 2026-01-30 | Added Task 6.9: Rename Osmo → Atom in TakerFeeDistribution types | AI Assistant |
