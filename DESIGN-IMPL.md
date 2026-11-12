# Recovery lock for Osmosis: implementation reference

Base: Osmosis v32.0.0 (commit e37a6f4752b2cace9ab921004d03116ffd4a6be3). All lock code is in `app/recoverylock`. Outside it, the Osmosis files change only to wire the lock in, plus one reward receiver hook in `x/incentives` (nil unless set).

## What it is

A consensus rule for a fixed list of compromised wallets. While it is active, a listed wallet can act only in transactions whose fee is paid by one of a fixed set of fee payer accounts, and only with the allowed messages (below). These pay out only to the wallet or the recovery address (undelegate, withdraw rewards and commission, begin unlocking, CL withdraw and collect), move funds to the one recovery address (send, CL position transfer, cw20 and cw721 transfer), point payouts at the wallet or the recovery address (withdraw address, lock reward receiver) or remove cw20 and cw721 rights granted earlier. Bank coins leave a listed wallet only to the recovery address. cw20 allowances and cw721 approvals granted before activation stay usable until they are removed (see Known limits). The upgrade itself moves nobody's funds. The lock has no store of its own and needs no migration; its only state write is the staking hooks' reset of an unsafe withdraw address (see Staking hooks). In block execution before the activation height, and from the expiry time on, every part of it returns at once without reading state, using gas or emitting events.

## Layout

| Path | Content |
| --- | --- |
| `app/recoverylock/core` | Chain independent shared code: configuration, active window, listed set, message rule registry, strict contract exit parser. No Cosmos SDK dependency. |
| `app/recoverylock/sdk50` | Cosmos SDK v0.50 adapter: send restriction, staking guard, router wrapper, ante decorator, staking hooks, standard message rules. |
| `app/recoverylock/config.go` | Loads and checks the embedded `params.json` and `listed.txt`. |
| `app/recoverylock/plugin.go` | Osmosis message rules and the lock incentives receiver rule. |
| `app/recoverylock/testbuild_on.go`, `testbuild_off.go` | The `recoverylock_test` build tag switch. |
| `app/keepers/keepers.go`, `app/ante.go`, `app/app.go` | Wiring. |
| `x/incentives/keeper/keeper.go`, `distribute.go` | Reward receiver hook (nil unless set). |

## Configuration

Both files are embedded in the binary with `go:embed`. `MustConfig` (config.go) panics at startup on any problem, so a misconfigured binary never runs.

`listed.txt`: one bech32 account address per line; empty lines and lines starting with `#` are skipped.

`params.json` is parsed strictly by `parseParams`: unknown keys, keys in the wrong case, duplicate keys at any depth, wrong value types and trailing data are refused. A JSON `null` leaves a field at its zero value, which fails every later check that needs a value; for `test_values` it means false.

| Field | Meaning |
| --- | --- |
| `recovery_address` | The only destination for coins, CL positions and cw20 or cw721 tokens that leave a listed wallet through the lock's rules. |
| `fee_payers` | Accounts allowed to pay fees for transactions that involve a listed address (1 to 16). |
| `activation_height` | First block height at which the lock applies. |
| `expiry_time` | Block time (unix seconds) from which the lock stops applying. |
| `activation_time_estimate` | Expected block time of the activation height (release builds). |
| `unbonding_seconds` | The chain's staking unbonding time (release builds). |
| `estimate_ref_height`, `estimate_ref_time` | A real mainnet block height and its block time, taken when the config is built (release builds). |
| `list_sha256` | sha256 of `listed.txt`, byte for byte. |
| `list_count` | Number of addresses in `listed.txt`. |
| `test_values` | `true` marks test values. |

Checks in every build (`build` in config.go and `core.NewConfig`): the list is not empty; 1 to 16 fee payers; the list hash and count match; every address decodes with the Osmosis bech32 codec and has 20 bytes; no address is a module account (looked up by bech32 in the app's blocked address map; its sanctioned entries are 0x hex keys and are not matched); no duplicate listed address or fee payer (compared as bytes, so another spelling of the same address is caught); no fee payer is listed or is the recovery address; the recovery address is not listed; activation height and expiry time are positive.

Extra checks in release builds (any build without `-tags recoverylock_test`):

- `test_values` must be false, and no address may be a known test address (`knownTestAddresses`: the addresses derived from the seeds in `testSeeds` and `testHashAddrs`, which cover the committed test values and the seeded test keys of the app tests and of a local test network).
- At least 8 fee payers (`MinReleaseFeePayers`).
- `unbonding_seconds` at least 1209600 (14 days, the Osmosis mainnet value, `MinReleaseUnbondingSeconds`) and at most one year.
- `expiry_time >= activation_time_estimate + 2 x unbonding_seconds + 7 days` (`ExpiryMarginSeconds`). Two periods because an attacker can fill all 7 unbonding entries of a delegation before activation: the recovery operator's undelegation can then only start when the first of those entries completes, up to one period after activation, and completes one period later. The 7 days absorb block time drift and a slow restart.
- `checkEstimate`: `activation_height > estimate_ref_height`, and the implied block time `(activation_time_estimate - estimate_ref_time) / (activation_height - estimate_ref_height)` must be between 0.5 s and 3 s (`MinMillisPerBlock`, `MaxMillisPerBlock`). Osmosis mainnet block header times give 0.83 s per block over heights 73836404 to 74836404 and 1.27 s over 71836404 to 73836404. The band refuses a height or time off by a factor of 10, a time in milliseconds, or a height of another chain.

Estimate method (also in the comment on the `params` struct): take the latest block (height h0, time t0) and the block 10000 heights earlier (time t1); b = (t0 - t1) / 10000; estimate = t0 + (H - h0) x b rounded up to the hour, where H is the activation height; use the later of this and the planned halt time.

### Test values and release values

The committed `params.json` and `listed.txt` are test values for a local chain: `test_values: true`, activation height 1000000000, expiry 4102444800 (2100-01-01 UTC), 3 listed addresses whose private keys are sha256 of `recoverylock-test-listed-<i>`, 8 test fee payers. Only a build with `-tags recoverylock_test` starts with them; every other build panics at startup. For a release, both files are replaced with the real list, recovery address, fee payers, activation height, time estimate, reference block, unbonding time and expiry.

At startup the node logs one line `recovery lock config` with `mode` (test or release), `activation_height`, `expiry_time`, `expiry_utc`, `recovery_address`, `fee_payers` (count), `list_count` and `list_sha256`.

## Activation and expiry

`core.Config.Active(height, time)` is `height >= activation_height && time < expiry_time`. `sdk50.Lock.active` is the single entry check of every component:

- In block execution it uses the block height and block time.
- In CheckTx it checks the next block (context height + 1). Right after a restart the context height is 0, so it uses the last committed height (`bApp.LastBlockHeight`) + 1. If the block time is unknown it checks the height only. Both only make the mempool stricter, never consensus.

The lock is switched on by a coordinated binary upgrade at a halt height equal to `activation_height`: the old binary stops before it applies that block, and the lock binary executes it with the lock active. It switches itself off at `expiry_time`: from then on, block execution is as upstream.

## Rules and where they are enforced

| Rule | Where |
| --- | --- |
| Coin rule: coins leaving a listed address may only go to the recovery address. Covers fees, sends, multi-send, IBC escrow, module deposits and contract funds, because they all go through `SendCoins` or `InputOutputCoins`. | `sdk50/components.go`: `Lock.SendRestriction`, added with `BankKeeper.AppendSendRestriction` in keepers.go |
| Staking guard: a listed delegator cannot delegate (`DelegateCoins` skips the send restriction). | `sdk50/components.go`: `StakingBank.DelegateCoinsFromAccountToModule`, the bank keeper given to `stakingkeeper.NewKeeper` |
| A transaction that involves a listed address (as signer of any message at any depth, as fee payer or as fee granter) must have a fee payer from the configured set, no fee granter and a non-zero fee, and every message with a listed signer must pass its rule. | `sdk50/components.go`: `Decorator.AnteHandle`; `sdk50/lock.go`: `checkMsg` |
| Only registered message types are allowed for a listed signer; each has its own rule. | `core/rules.go`: `Registry.Check` |
| Dispatched messages (authz `MsgExec`, CosmWasm, ICA host) with a listed signer at any depth are refused while active, including messages the rules would allow. | `sdk50/components.go`: `Router.wrap` |
| Nesting deeper than 6 levels of `MsgExec` is refused while active, for every transaction, because a level that is not unpacked could hide a listed signer. | `sdk50/lock.go`: `MaxDepth`, `checkMsg` |
| Staking rewards and validator commission paid outside transactions never reach an unsafe withdraw address. | `sdk50/hooks.go`: `StakingHooks` |
| Lock incentives of a listed owner go to the owner unless the receiver is the owner or the recovery address. | `plugin.go`: `LockRewardReceiver`; `x/incentives/keeper/distribute.go` |

### Ante decorator

`sdk50.Decorator` is placed right after the message filter decorator in `app/ante.go`, before fee deduction, the send block decorator and the smart account circuit breaker. So both signature paths (classic and smart account) are covered, and a rejected transaction costs nothing (baseapp discards the cached ante context on error). A transaction with no messages goes to the next decorator unchecked. A transaction that involves a listed address and has no fee information is refused.

### Allowed messages for a listed signer

Standard rules (`sdk50/rules.go`, `registerStandardRules`):

| Message | Condition |
| --- | --- |
| bank `MsgSend` | to the recovery address |
| bank `MsgMultiSend` | every output to the recovery address |
| staking `MsgUndelegate` | delegator is the signer; withdraw address is the wallet or the recovery address |
| distribution `MsgWithdrawDelegatorReward` | delegator is the signer; withdraw address is the wallet or the recovery address |
| distribution `MsgWithdrawValidatorCommission` | withdraw address of the operator account is the wallet or the recovery address |
| distribution `MsgSetWithdrawAddress` | delegator is the signer; new address is the wallet or the recovery address |

Osmosis rules (`plugin.go`, `RegisterRules`):

| Message | Condition |
| --- | --- |
| lockup `MsgBeginUnlocking`, `MsgBeginUnlockingAll` | owner is the signer (pays the owner) |
| concentrated liquidity `MsgWithdrawPosition`, `MsgCollectSpreadRewards`, `MsgCollectIncentives` | sender is the signer (pays the position owner) |
| lockup `MsgSetRewardReceiverAddress` | owner is the signer; receiver is the owner or the recovery address |
| concentrated liquidity `MsgTransferPositions` | sender is the signer; new owner is the recovery address |
| wasm `MsgExecuteContract` | sender is the signer, no funds, and the message is a contract exit (below) |

Not allowed: gamm `MsgExitPool`. It moves shares with `SendCoinsFromAccountToModule`, which the coin rule blocks, so it cannot succeed. gamm shares are bank coins and leave with `MsgSend` to the recovery address. Redelegation and cancel unbonding are not allowed messages either.

All rules only read state.

### Contract exits

`core/contract.go`, `CheckContractExit`. The message must be exactly one of:

```
{"transfer":{"recipient":<recovery>,"amount":"<uint>"}}             (cw20)
{"transfer_nft":{"recipient":<recovery>,"token_id":"<string>"}}     (cw721)
{"decrease_allowance":{"spender":<address>,"amount":"<uint>"}}      (cw20, no expires)
{"revoke":{"spender":<address>,"token_id":"<string>"}}              (cw721)
{"revoke_all":{"operator":<address>}}                               (cw721)
```

Only the message shape is checked, not which contract runs it; the recovery operator chooses the contract. `<recovery>` must equal the canonical recovery address byte for byte. The JSON is parsed strictly: duplicate keys, unknown or missing fields, other value types and trailing data are refused. `send`, `send_nft`, `approve`, `increase_allowance` and every other execute message are refused. The contract runs with the wallet as sender but cannot move the wallet's bank coins (no funds attached, and the contract's own submessages have the contract as sender).

### Fee payer rule

Osmosis has no feegrant module, so fee grants are not used. Listed wallets pay no fees: in every transaction that involves a listed address, the fee payer (`FeeTx.FeePayer()`: the `Fee.Payer` field, or the first signer when that field is empty) must be one of the configured fee payer accounts. The fee payer always signs the transaction, together with the listed wallet. In practice the recovery operator sets `Fee.Payer`. Listed wallets act only in top level messages of transactions that pass the ante decorator with a configured fee payer. A listed wallet as fee payer or fee granter makes the transaction involve a listed address, so it needs a configured fee payer and no granter.

### Router refusal

`sdk50.Router` wraps `bApp.MsgServiceRouter()` for the authz, ICA host and wasm keepers (keepers.go). Top level transaction messages do not pass through it; baseapp routes them with its own router after the ante handler. While active, the wrapper refuses every dispatched message with a listed signer at any depth ("listed accounts may act only in transactions paid by a fee payer"). Errors of the check itself (depth cap, undecodable inner messages) also refuse. Dispatch has no fee payer, so without this a contract or grantee holding an authz grant could act as a listed wallet. A fee payer transaction that wraps a listed signer's message in `MsgExec` therefore fails in delivery.

Not wrapped: the ICA controller keeper (it dispatches only `MsgChannelOpenInit` signed by the ICA module account), the gov keeper (signer is the gov module), the configurator, and `wasm.NewAppModule`, which stores the router in a field nothing reads (wasmd v0.54.10).

### Staking hooks

`sdk50.StakingHooks` is registered first in `NewMultiStakingHooks` (keepers.go), before the distribution hooks. While active, for a listed delegator whose withdraw address is neither the wallet nor the recovery address, `BeforeDelegationSharesModified` and `BeforeDelegationRemoved` delete the withdraw address entry, so it falls back to the wallet and the coin rule applies to the payout. This covers payouts no transaction controls, for example `SlashRedelegation` calling `Unbond`. `AfterValidatorRemoved` does the same for the operator account of a listed validator, whose commission the distribution hook pays out when the validator is deleted. One store read and at most one delete; deterministic.

### Lock incentives receiver

The `x/incentives` keeper has a `rewardReceiver` func, nil unless set. keepers.go sets it with `SetRewardReceiverFunc` right after `NewKeeper`, and `distributeInternal` calls it after the upstream "empty receiver means owner" default. Superfluid synthetic payouts use the same code. `recoverylock.LockRewardReceiver` returns the owner for a lock owned by a listed wallet unless the stored receiver is the owner or the recovery address. Before activation and after expiry it returns the receiver unchanged without reading state. The recovery operator can also reset the receiver with `MsgSetRewardReceiverAddress`.

## Build tags

- No tag (release, `make build`, goreleaser): `testBuild = false` (`testbuild_off.go`). Test values and known test addresses are refused, release checks apply.
- `-tags recoverylock_test`: `testBuild = true` (`testbuild_on.go`). Test values are accepted, release checks are skipped. Only for tests and local chains.

## Building a release and verifying it

1. Replace `listed.txt` and `params.json` with the release values (`test_values: false`), set `list_sha256` to `sha256sum listed.txt` and `list_count` to the number of addresses.
2. Build without tags: `go build ./cmd/osmosisd` (or `make build`). Start it once: it panics on any configuration problem.
3. Run the tests with the test tag (below) on the same commit.

Validators verify a release by:

- Checking that the diff against e37a6f4752b2cace9ab921004d03116ffd4a6be3 touches only `app/recoverylock`, the wiring and hook files listed in Layout, and `DESIGN-IMPL.md`.
- Computing `sha256sum app/recoverylock/listed.txt` and comparing it with `list_sha256` and the published value.
- Building from the release commit and comparing the `recovery lock config` startup line (mode release, activation height, expiry, recovery address, fee payer count, list count and hash) with the published values.

## Tests

Run: `go test -count=1 -tags recoverylock_test ./app/recoverylock/...`. `TestEmbeddedConfig` fails (not skips) without the tag. The app tests (`app_test.go`, `attack_paths_test.go`) run a full Osmosis app. Contracts used: cw20-base 1.0.1 from `tests/ibc-hooks/bytecode/cw20_base.wasm`, cw721-base 0.16.0 from `testdata/cw721_base_v0.16.0.wasm` (the `cw721_base.wasm` release asset of CosmWasm/cw-nfts v0.16.0, sha256 d5fa68c6dcf210fd14e556a21d2f7a4976933da33caa97348ad99e0ec4373d9f), and `wasmbinding/testdata/osmo_reflect.wasm` as an attacker contract.

Map of invariants to code and tests:

| Invariant | Code | Tests |
| --- | --- | --- |
| Inert before activation and after expiry | `sdk50/lock.go`: `Lock.active`; `plugin.go`: `LockRewardReceiver` | TestWindowAndListing, TestBeforeActivationNoop, TestExpiryByTime, TestNoopBeforeActivation, TestExpiryEdge, TestRedelegationSlashBeforeActivation, TestCommissionOnValidatorRemovalBeforeActivation, TestLockIncentivesPaidToOwner |
| CheckTx height and restart handling | `sdk50/lock.go`: `Lock.active` | TestCheckTxHeight, TestCheckTxAfterRestart |
| Coin rule | `sdk50/components.go`: `SendRestriction` | TestRestriction, TestCoinRuleAndStakingGuard |
| Staking guard | `sdk50/components.go`: `StakingBank` | TestStakingGuard, TestCoinRuleAndStakingGuard |
| The only call of a bank function that skips the send restriction with a user address as source is x/staking delegation (guarded by `StakingBank`); the other pinned calls send from module accounts | keepers.go wiring | TestBypassCallSites (pins every call site over every package linked into osmosisd) |
| Fee payer rule, granter refused, non-zero fee, both signature paths, rejects cost nothing | `sdk50/components.go`: `Decorator.AnteHandle` | TestAnte, TestEmptyTx, TestAnteFeesAndPaths, TestSmartAccountPathReached, TestListedFeeGranterRefused |
| Only registered message types are allowed for a listed signer | `core/rules.go`: `Registry.Check` | TestRegistry, TestAnte |
| Allowed messages pay only into the wallet or the recovery address | `sdk50/rules.go`; `plugin.go`: `RegisterRules` | TestExitsEndToEnd, TestAttackerWithdrawAddress, TestWithdrawValidatorCommission, TestListedOperatorWithdrawCommission |
| Contract exits parsed strictly | `core/contract.go`: `CheckContractExit` | TestContractExit, TestContractExits |
| No dispatch path acts for a listed signer | `sdk50/components.go`: `Router.wrap`; keepers.go | TestRouterDispatch, TestRoutersWired, TestContractExits, TestDispatchCannotFillUnbondingEntries, TestDispatchNestedGranteeRefused, TestDispatchContractExitWithoutOperatorRefused |
| Nesting depth cap | `sdk50/lock.go`: `MaxDepth`, `checkMsg` | TestNestingDepthCap |
| Reward and commission payouts outside transactions reach the wallet or the recovery address | `sdk50/hooks.go`: `StakingHooks` | TestStakingHooksReset, TestRedelegationSlashResetsWithdrawAddr, TestCommissionOnValidatorRemovalGoesToWallet |
| Lock incentives reach the owner or the recovery address | `plugin.go`: `LockRewardReceiver`; x/incentives hook | TestLockIncentivesPaidToOwner, TestLockIncentivesRecoveryReceiverKept |
| A wrong or test config never starts in a release build | `config.go`: `parseParams`, `build`, `checkEstimate`, `MustConfig`; `core/config.go`: `NewConfig` | TestConfigValidation, TestParamsStrict, TestEmbeddedConfig, TestKnownTestAddressesCoverEmbedded, TestKnownTestAddressesCoverLocalNetwork, TestValidation |

## Known limits

- Only listed addresses are protected. The list is compiled in; changing it needs a new binary.
- Before activation the attacker can still act, including filling unbonding entries; the expiry margin allows for that. Value moved out before activation is not covered.
- The rules check destinations and withdraw addresses, not amounts.
- While active, listed wallets cannot act through authz, contract dispatch or the ICA host at all, including legitimate use by the owner. A fee payer transaction that wraps a listed wallet's message in authz `MsgExec` passes the ante decorator and fails in delivery, so the fee is charged; the recovery operator signs listed messages directly.
- cw20 allowances and cw721 approvals granted before activation live in contract state, not in the bank, so the coin rule does not cover them. They stay usable until the recovery operator removes them with `decrease_allowance`, `revoke` or `revoke_all`; remove rights first.
- Positions held inside contracts (vaults, staking contracts) and superfluid positions have no exit messages.
- `TestBypassCallSites` pins the call sites of this commit's dependencies; any dependency change must pass it again.
