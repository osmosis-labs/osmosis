package recoverylock_test

// Tests for attack paths through contracts, authz, validator removal, lock
// reward receivers, fee grants and nesting depth. Each attack is run and must
// fail while the lock is active; the BeforeActivation tests show the same
// path succeeding before activation.
// The reflect contract is wasmbinding/testdata/osmo_reflect.wasm (part of the
// Osmosis repository): it dispatches any stargate message with the contract as
// signer, which is how an attacker contract reaches the wrapped router.

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gammtypes "github.com/osmosis-labs/osmosis/v32/x/gamm/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v32/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
)

const routerRefusal = "listed accounts may act only in transactions paid by a fee payer"

// isRouterRefusal accepts the lock error, also when the wasm keeper re-wraps
// a dispatch error as text.
func isRouterRefusal(err error) bool {
	return isLocked(err) || (err != nil && strings.Contains(err.Error(), routerRefusal))
}

// reflect deploys osmo_reflect with the attacker as owner.
func (s *LockSuite) reflect(owner sdk.AccAddress) sdk.AccAddress {
	gk := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	gov := authtypes.NewModuleAddress(govtypes.ModuleName)
	code, err := os.ReadFile("../../wasmbinding/testdata/osmo_reflect.wasm")
	s.Require().NoError(err)
	id, _, err := gk.Create(s.Ctx, gov, code, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody})
	s.Require().NoError(err)
	c, _, err := wasmkeeper.NewDefaultPermissionKeeper(s.App.WasmKeeper).Instantiate(s.Ctx, id, owner, nil, []byte(`{}`), "reflect", nil)
	s.Require().NoError(err)
	return c
}

// reflectMsg is the execute message that makes the reflect contract dispatch m.
func (s *LockSuite) reflectMsg(m sdk.Msg) string {
	raw, err := s.App.AppCodec().Marshal(m)
	s.Require().NoError(err)
	return fmt.Sprintf(`{"reflect_msg":{"msgs":[{"stargate":{"type_url":"%s","value":"%s"}}]}}`, sdk.MsgTypeURL(m), base64.StdEncoding.EncodeToString(raw))
}

func (s *LockSuite) attackerAccount() (*secp256k1.PrivKey, sdk.AccAddress) {
	k := key("attacker-test-key")
	a := addrOf(k)
	s.FundAcc(a, sdk.NewCoins(sdk.NewCoin("uosmo", math.NewInt(1_000_000_000_000)), fee[0].AddAmount(math.NewInt(1_000_000_000_000))))
	return k, a
}

// TestDispatchCannotFillUnbondingEntries: with an authz grant made before
// activation, an attacker contract tries to fill all 7 unbonding entries of a
// listed wallet with 1 unit undelegations (no fee payer involved), which would
// make the recovery operator's undelegation fail with max entries. The router
// wrapper refuses every dispatched message with a listed signer.
func (s *LockSuite) TestDispatchCannotFillUnbondingEntries() {
	val := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	kA, A := s.attackerAccount()
	C := s.reflect(A)
	exp := s.Ctx.BlockTime().Add(100_000 * time.Hour)
	s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, C, s.L, authz.NewGenericAuthorization(sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{})), &exp))
	attack := func() error {
		ex := authz.NewMsgExec(C, []sdk.Msg{&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1))}})
		return s.deliver(s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kA}, wexec(A, C, s.reflectMsg(&ex), nil)))
	}
	// control: before activation the path works (one entry), so the refusals
	// below come from the lock and not from a broken test setup
	s.Require().NoError(attack())
	s.activate()
	t0 := s.Ctx.BlockTime()
	for i := 0; i < 7; i++ {
		s.Ctx = s.Ctx.WithBlockHeight(act + int64(i)).WithBlockTime(t0.Add(time.Duration(i) * 5 * time.Second))
		err := attack()
		s.Require().True(isRouterRefusal(err), "attacker block %d: %v", i, err)
	}
	ubd, err := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, s.L, val)
	s.Require().NoError(err)
	s.Require().Len(ubd.Entries, 1, "no entry created by the attacker while active")
	// the recovery operator's full undelegation works
	s.Ctx = s.Ctx.WithBlockHeight(act + 7).WithBlockTime(t0.Add(40 * time.Second))
	s.Require().NoError(s.deliver(s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL},
		&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(4_999_999))})))
	ubd, err = s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, s.L, val)
	s.Require().NoError(err)
	s.Require().Len(ubd.Entries, 2)
	s.Require().Equal(math.NewInt(4_999_999), ubd.Entries[1].Balance)
}

// TestDispatchNestedGranteeRefused: an EOA grantee holding a grant from the
// listed wallet grants MsgExec onward to a contract (nested MsgExec).
func (s *LockSuite) TestDispatchNestedGranteeRefused() {
	val := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	kA, A := s.attackerAccount()
	G := s.X
	C := s.reflect(A)
	exp := s.Ctx.BlockTime().Add(100_000 * time.Hour)
	s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, G, s.L, authz.NewGenericAuthorization(sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{})), &exp))
	s.activate()
	s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, C, G, authz.NewGenericAuthorization(sdk.MsgTypeURL(&authz.MsgExec{})), &exp))
	inner := authz.NewMsgExec(G, []sdk.Msg{&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1))}})
	outer := authz.NewMsgExec(C, []sdk.Msg{&inner})
	err := s.deliver(s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kA}, wexec(A, C, s.reflectMsg(&outer), nil)))
	s.Require().True(isRouterRefusal(err), "%v", err)
	_, err = s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, s.L, val)
	s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation)
}

// TestDispatchContractExitWithoutOperatorRefused: the same path cannot run a
// contract exit without the recovery operator.
func (s *LockSuite) TestDispatchContractExitWithoutOperatorRefused() {
	cw20 := s.storeAndInstantiate("../../tests/ibc-hooks/bytecode/cw20_base.wasm",
		`{"name":"Test","symbol":"TST","decimals":6,"initial_balances":[{"address":"`+s.L.String()+`","amount":"1000"}]}`)
	kA, A := s.attackerAccount()
	C := s.reflect(A)
	exp := s.Ctx.BlockTime().Add(100_000 * time.Hour)
	s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, C, s.L, authz.NewGenericAuthorization(sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{})), &exp))
	// control: before activation the attacker moves the tokens anywhere
	ex := authz.NewMsgExec(C, []sdk.Msg{wexec(s.L, cw20, `{"transfer":{"recipient":"`+A.String()+`","amount":"1"}}`, nil)})
	s.Require().NoError(s.deliver(s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kA}, wexec(A, C, s.reflectMsg(&ex), nil))))
	s.activate()
	for _, to := range []sdk.AccAddress{s.R, A} {
		ex := authz.NewMsgExec(C, []sdk.Msg{wexec(s.L, cw20, `{"transfer":{"recipient":"`+to.String()+`","amount":"10"}}`, nil)})
		err := s.deliver(s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kA}, wexec(A, C, s.reflectMsg(&ex), nil)))
		s.Require().True(isRouterRefusal(err), "to %s: %v", to, err)
	}
	s.Require().Equal(`{"balance":"999"}`, s.query(cw20, `{"balance":{"address":"`+s.L.String()+`"}}`))
	s.Require().Equal(`{"balance":"0"}`, s.query(cw20, `{"balance":{"address":"`+s.R.String()+`"}}`))
}

// listedValidator creates a validator operated by the listed wallet, with a
// delegation from another account and rewards, then removes the operator's
// self delegation (all before activation, as the attacker could).
func (s *LockSuite) listedValidator() (sdk.ValAddress, sdk.AccAddress) {
	valL := sdk.ValAddress(s.L)
	cv, err := stakingtypes.NewMsgCreateValidator(valL.String(), ed25519.GenPrivKey().PubKey(), sdk.NewCoin("stake", math.NewInt(1_000_000)),
		stakingtypes.Description{Moniker: "listed"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyOneDec(), math.LegacyNewDecWithPrec(1, 2)), math.OneInt())
	s.Require().NoError(err)
	_, err = stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper).CreateValidator(s.Ctx, cv)
	s.Require().NoError(err)
	D := s.TestAccs[2]
	s.FundAcc(D, coins("stake", 10_000_000))
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: D.String(), ValidatorAddress: valL.String(), Amount: sdk.NewCoin("stake", math.NewInt(3_000_000))}))
	s.AllocateRewardsToValidator(valL, math.NewInt(1_000_000))
	s.Require().NoError(s.App.DistrKeeper.SetDelegatorWithdrawAddr(s.Ctx, s.L, s.X))
	s.Require().NoError(s.handle(&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: valL.String(), Amount: sdk.NewCoin("stake", math.NewInt(1_000_000))}))
	comm, err := s.App.DistrKeeper.GetValidatorAccumulatedCommission(s.Ctx, valL)
	s.Require().NoError(err)
	s.Require().True(comm.Commission.AmountOf("stake").IsPositive())
	return valL, D
}

// TestCommissionOnValidatorRemovalGoesToWallet: removing a listed operator's
// validator pays the accumulated commission out.
// StakingHooks.AfterValidatorRemoved resets an attacker-set withdraw address
// to the wallet first.
func (s *LockSuite) TestCommissionOnValidatorRemovalGoesToWallet() {
	valL, D := s.listedValidator()
	s.activate()
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake")
	walletBefore := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake")
	// the last delegator leaving removes the validator
	s.Require().NoError(s.handle(&stakingtypes.MsgUndelegate{DelegatorAddress: D.String(), ValidatorAddress: valL.String(), Amount: sdk.NewCoin("stake", math.NewInt(3_000_000))}))
	_, err := s.App.StakingKeeper.GetValidator(s.Ctx, valL)
	s.Require().ErrorIs(err, stakingtypes.ErrNoValidatorFound, "validator removed")
	s.Require().Equal(before, s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake"), "attacker withdraw address gets nothing")
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake").Amount.GT(walletBefore.Amount), "commission lands in the wallet")
	w, err := s.App.DistrKeeper.GetDelegatorWithdrawAddr(s.Ctx, s.L)
	s.Require().NoError(err)
	s.Require().Equal(s.L, w)
}

// TestCommissionOnValidatorRemovalBeforeActivation: the same removal before
// activation pays the attacker-set withdraw address, exactly as upstream.
func (s *LockSuite) TestCommissionOnValidatorRemovalBeforeActivation() {
	valL, D := s.listedValidator()
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake")
	s.Require().NoError(s.handle(&stakingtypes.MsgUndelegate{DelegatorAddress: D.String(), ValidatorAddress: valL.String(), Amount: sdk.NewCoin("stake", math.NewInt(3_000_000))}))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake").Amount.GT(before.Amount), "upstream behaviour before activation")
}

// TestListedOperatorWithdrawCommission: a listed operator may withdraw
// commission in a fee payer transaction once its withdraw address is the
// wallet or the recovery address; refused while it is the attacker's, and
// refused through the router.
func (s *LockSuite) TestListedOperatorWithdrawCommission() {
	valL := sdk.ValAddress(s.L)
	cv, err := stakingtypes.NewMsgCreateValidator(valL.String(), ed25519.GenPrivKey().PubKey(), sdk.NewCoin("stake", math.NewInt(1_000_000)),
		stakingtypes.Description{Moniker: "listed"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyOneDec(), math.LegacyNewDecWithPrec(1, 2)), math.OneInt())
	s.Require().NoError(err)
	_, err = stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper).CreateValidator(s.Ctx, cv)
	s.Require().NoError(err)
	s.AllocateRewardsToValidator(valL, math.NewInt(1_000_000))
	s.Require().NoError(s.App.DistrKeeper.SetDelegatorWithdrawAddr(s.Ctx, s.L, s.X))
	s.activate()
	m := &distrtypes.MsgWithdrawValidatorCommission{ValidatorAddress: valL.String()}
	tx := func(msgs ...sdk.Msg) error {
		return s.deliver(s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL}, msgs...))
	}
	s.Require().True(isLocked(tx(m)), "attacker withdraw address")
	s.Require().NoError(tx(&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: s.L.String(), WithdrawAddress: s.L.String()}))
	s.Require().True(isLocked(s.dispatch(m)), "router")
	xBefore := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake")
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake")
	s.Require().NoError(tx(m))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake").Amount.GT(before.Amount), "commission lands in the wallet")
	s.Require().Equal(xBefore, s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake"))
}

// lockWithAttackerReceiver locks gamm shares of the listed wallet for the
// first lockable duration and sets the attacker as the lock's reward receiver
// (before activation).
func (s *LockSuite) lockWithAttackerReceiver() (string, time.Duration) {
	dur := s.App.IncentivesKeeper.GetLockableDurations(s.Ctx)[0]
	poolId := s.PrepareBalancerPool()
	shares := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], gammtypes.GetPoolShareDenom(poolId))
	s.Require().NoError(s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[0], s.L, sdk.NewCoins(sdk.NewCoin(shares.Denom, shares.Amount.QuoRaw(2)))))
	l1, err := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper).LockTokens(s.Ctx, lockuptypes.NewMsgLockTokens(s.L, dur, sdk.NewCoins(sdk.NewCoin(shares.Denom, shares.Amount.QuoRaw(4)))))
	s.Require().NoError(err)
	s.Require().NoError(s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, l1.ID, s.L, s.X.String()))
	ip := s.App.IncentivesKeeper.GetParams(s.Ctx)
	ip.MinValueForDistribution = sdk.NewCoin("uosmo", math.ZeroInt())
	s.App.IncentivesKeeper.SetParams(s.Ctx, ip)
	return shares.Denom, dur
}

// distributeOnce creates a one epoch gauge of 1,000,000 uosmo for the lock's
// denom and runs the epoch end; returns what the attacker and the wallet got.
func (s *LockSuite) distributeOnce(denom string, dur time.Duration) (attacker, wallet math.Int) {
	s.FundAcc(s.TestAccs[0], coins("uosmo", 1_000_000))
	_, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, s.TestAccs[0], coins("uosmo", 1_000_000),
		lockuptypes.QueryCondition{LockQueryType: lockuptypes.ByDuration, Denom: denom, Duration: dur}, s.Ctx.BlockTime(), 1, 0)
	s.Require().NoError(err)
	x0 := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "uosmo").Amount
	l0 := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "uosmo").Amount
	s.Require().NoError(s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier, 1))
	return s.App.BankKeeper.GetBalance(s.Ctx, s.X, "uosmo").Amount.Sub(x0), s.App.BankKeeper.GetBalance(s.Ctx, s.L, "uosmo").Amount.Sub(l0)
}

// TestLockIncentivesPaidToOwner: while active, lock incentives for a lock with
// an attacker-set reward receiver are paid to the lock owner by the
// x/incentives hook. Before activation and after expiry the receiver is paid
// as upstream.
func (s *LockSuite) TestLockIncentivesPaidToOwner() {
	denom, dur := s.lockWithAttackerReceiver()
	// before activation: upstream, the attacker receiver is paid
	x, w := s.distributeOnce(denom, dur)
	s.Require().True(x.IsPositive(), "before activation the receiver is paid")
	s.Require().True(w.IsZero())
	// active: the owner is paid
	s.activate()
	x, w = s.distributeOnce(denom, dur)
	s.Require().True(x.IsZero(), "attacker receiver gets nothing while active, got %s", x)
	s.Require().True(w.IsPositive(), "the owner is paid")
	s.T().Logf("active: owner got %s uosmo, attacker receiver 0", w)
	// after expiry: upstream again
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(expiryUnix, 0))
	x, w = s.distributeOnce(denom, dur)
	s.Require().True(x.IsPositive(), "after expiry the receiver is paid")
	s.Require().True(w.IsZero())
}

// TestLockIncentivesRecoveryReceiverKept: a receiver set to the recovery
// address is kept while active.
func (s *LockSuite) TestLockIncentivesRecoveryReceiverKept() {
	denom, dur := s.lockWithAttackerReceiver()
	locks := s.App.LockupKeeper.GetAccountPeriodLocks(s.Ctx, s.L)
	s.Require().Len(locks, 1)
	s.Require().NoError(s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, locks[0].ID, s.L, s.R.String()))
	s.activate()
	r0 := s.App.BankKeeper.GetBalance(s.Ctx, s.R, "uosmo").Amount
	x, w := s.distributeOnce(denom, dur)
	s.Require().True(x.IsZero() && w.IsZero())
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.R, "uosmo").Amount.GT(r0), "recovery receiver is paid")
}

// TestListedFeeGranterRefused: a listed wallet as fee granter of an attacker
// transaction is refused before any fee is taken.
func (s *LockSuite) TestListedFeeGranterRefused() {
	_, A := s.attackerAccount()
	s.activate()
	b := s.App.GetTxConfig().NewTxBuilder()
	s.Require().NoError(b.SetMsgs(s.send(A, s.X, coins("uosmo", 1))))
	b.SetFeeAmount(fee)
	b.SetGasLimit(500_000)
	b.SetFeeGranter(s.L)
	cc, _ := s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	_, err := s.App.GetBaseApp().AnteHandler()(cc, b.GetTx(), false)
	s.Require().True(isLocked(err), "%v", err)
}

// TestNestingDepthCap checks the nesting cap: while active, a transaction
// without any listed signer but with MsgExec nested deeper than the cap is
// refused (a level that is not unpacked could hide a listed signer).
func (s *LockSuite) TestNestingDepthCap() {
	kA, A := s.attackerAccount()
	var m sdk.Msg = s.send(A, s.X, coins("uosmo", 1))
	for i := 0; i < 8; i++ {
		e := authz.NewMsgExec(A, []sdk.Msg{m})
		m = &e
	}
	tx := s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kA}, m)
	cc, _ := s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	_, err := s.App.GetBaseApp().AnteHandler()(cc, tx, false)
	s.Require().False(isLocked(err), "before activation the lock does not look: %v", err)
	s.activate()
	cc, _ = s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	_, err = s.App.GetBaseApp().AnteHandler()(cc, tx, false)
	s.Require().True(isLocked(err), "%v", err)
}
