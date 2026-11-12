package sdk50

import (
	"context"
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/core"
)

var (
	addrA  = sdk.AccAddress([]byte("listed_a____________")) // listed
	addrB  = sdk.AccAddress([]byte("other_b_____________")) // not listed
	addrR  = sdk.AccAddress([]byte("recovery____________"))
	addrP  = sdk.AccAddress([]byte("feepayer____________"))
	addrG  = sdk.AccAddress([]byte("grantee_____________"))
	addrW  = sdk.AccAddress([]byte("attacker_withdraw___"))
	addrP2 = sdk.AccAddress([]byte("feepayer_two________"))
)

const expiry = 1_900_000_000

var lastCommitted int64 = 0

type distrMock map[string]sdk.AccAddress

func (d distrMock) GetDelegatorWithdrawAddr(_ context.Context, a sdk.AccAddress) (sdk.AccAddress, error) {
	if w, ok := d[string(a)]; ok {
		return w, nil
	}
	return a, nil
}

func setup(t *testing.T) (*Lock, distrMock, sdk.Context) {
	t.Helper()
	sdk.GetConfig().SetBech32PrefixForAccount("osmo", "osmopub")
	reg, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          addresscodec.NewBech32Codec("osmo"),
			ValidatorAddressCodec: addresscodec.NewBech32Codec("osmovaloper"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	banktypes.RegisterInterfaces(reg)
	stakingtypes.RegisterInterfaces(reg)
	distrtypes.RegisterInterfaces(reg)
	authz.RegisterInterfaces(reg)
	cdc := codec.NewProtoCodec(reg)
	cfg, err2 := core.NewConfig([][]byte{addrA}, addrR, [][]byte{addrP, addrP2}, 10, expiry)
	if err2 != nil {
		t.Fatal(err2)
	}
	d := distrMock{}
	l := New(cfg, cdc, addresscodec.NewBech32Codec("osmo"), d, func() int64 { return lastCommitted })
	key := storetypes.NewKVStoreKey("test")
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("t")).WithBlockTime(time.Unix(expiry-100, 0))
	return l, d, ctx
}

type fakeTx struct {
	msgs    []sdk.Msg
	fee     sdk.Coins
	payer   []byte
	granter []byte
}

func (f fakeTx) GetMsgs() []sdk.Msg                    { return f.msgs }
func (f fakeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (f fakeTx) GetGas() uint64                        { return 100000 }
func (f fakeTx) GetFee() sdk.Coins                     { return f.fee }
func (f fakeTx) FeePayer() []byte                      { return f.payer }
func (f fakeTx) FeeGranter() []byte                    { return f.granter }

var okFee = sdk.NewCoins(sdk.NewCoin("uosmo", math.NewInt(5000)))

func send(from, to sdk.AccAddress) sdk.Msg {
	return &banktypes.MsgSend{FromAddress: from.String(), ToAddress: to.String(), Amount: okFee}
}

func exec(t *testing.T, grantee sdk.AccAddress, msgs ...sdk.Msg) sdk.Msg {
	t.Helper()
	m := authz.NewMsgExec(grantee, msgs)
	return &m
}

func runAnte(l *Lock, ctx sdk.Context, tx fakeTx) error {
	_, err := Decorator{L: l}.AnteHandle(ctx, tx, false, func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil })
	return err
}

func TestBeforeActivationNoop(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(5)
	tx := fakeTx{msgs: []sdk.Msg{send(addrA, addrB)}, fee: nil, payer: addrA}
	if err := runAnte(l, ctx, tx); err != nil {
		t.Fatalf("must be a no-op before activation: %v", err)
	}
	if to, err := l.SendRestriction()(ctx, addrA, addrB, okFee); err != nil || !to.Equals(addrB) {
		t.Fatal("restriction must be a no-op before activation")
	}
	if ctx.GasMeter().GasConsumed() != 0 {
		t.Fatal("no gas before activation")
	}
}

func TestAnte(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(10)
	cases := []struct {
		name string
		tx   fakeTx
		ok   bool
	}{
		{"send to recovery, fee payer pays", fakeTx{[]sdk.Msg{send(addrA, addrR)}, okFee, addrP, nil}, true},
		{"send elsewhere", fakeTx{[]sdk.Msg{send(addrA, addrB)}, okFee, addrP, nil}, false},
		{"whole balance as fee from listed", fakeTx{[]sdk.Msg{send(addrA, addrR)}, okFee, addrA, nil}, false},
		{"attacker pays fee", fakeTx{[]sdk.Msg{send(addrA, addrR)}, okFee, addrB, nil}, false},
		{"zero fee", fakeTx{[]sdk.Msg{send(addrA, addrR)}, nil, addrP, nil}, false},
		{"fee granter", fakeTx{[]sdk.Msg{send(addrA, addrR)}, okFee, addrP, addrB}, false},
		{"delegate", fakeTx{[]sdk.Msg{&stakingtypes.MsgDelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]}}, okFee, addrP, nil}, false},
		{"cancel unbonding", fakeTx{[]sdk.Msg{&stakingtypes.MsgCancelUnbondingDelegation{DelegatorAddress: addrA.String()}}, okFee, addrP, nil}, false},
		{"set withdraw address", fakeTx{[]sdk.Msg{&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: addrA.String(), WithdrawAddress: addrW.String()}}, okFee, addrP, nil}, false},
		{"undelegate", fakeTx{[]sdk.Msg{&stakingtypes.MsgUndelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]}}, okFee, addrP, nil}, true},
		{"withdraw reward to self", fakeTx{[]sdk.Msg{&distrtypes.MsgWithdrawDelegatorReward{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x"}}, okFee, addrP, nil}, true},
		{"second fee payer", fakeTx{[]sdk.Msg{send(addrA, addrR)}, okFee, addrP2, nil}, true},
		{"set withdraw address to self", fakeTx{[]sdk.Msg{&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: addrA.String(), WithdrawAddress: addrA.String()}}, okFee, addrP, nil}, true},
		{"set withdraw address to recovery", fakeTx{[]sdk.Msg{&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: addrA.String(), WithdrawAddress: addrR.String()}}, okFee, addrP, nil}, true},
		{"listed fee granter, unlisted payer", fakeTx{[]sdk.Msg{send(addrB, addrG)}, okFee, addrB, addrA}, false},
		{"listed fee payer for unlisted message", fakeTx{[]sdk.Msg{send(addrB, addrG)}, okFee, addrA, nil}, false},
		{"authz grantee delegates for listed", fakeTx{[]sdk.Msg{exec(t, addrG, &stakingtypes.MsgDelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]})}, okFee, addrG, nil}, false},
		{"authz grantee sends listed coins elsewhere", fakeTx{[]sdk.Msg{exec(t, addrG, send(addrA, addrB))}, okFee, addrG, nil}, false},
		{"authz grantee sends to recovery but pays itself", fakeTx{[]sdk.Msg{exec(t, addrG, send(addrA, addrR))}, okFee, addrG, nil}, false},
		{"nested exec", fakeTx{[]sdk.Msg{exec(t, addrG, exec(t, addrG, send(addrA, addrB)))}, okFee, addrG, nil}, false},
		{"multisend split", fakeTx{[]sdk.Msg{&banktypes.MsgMultiSend{Inputs: []banktypes.Input{{Address: addrA.String(), Coins: okFee}}, Outputs: []banktypes.Output{{Address: addrR.String(), Coins: okFee}, {Address: addrB.String(), Coins: okFee}}}}, okFee, addrP, nil}, false},
		{"unrelated tx untouched", fakeTx{[]sdk.Msg{send(addrB, addrG)}, nil, addrB, nil}, true},
	}
	for _, c := range cases {
		err := runAnte(l, ctx, c.tx)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v err=%v", c.name, c.ok, err)
		}
		if err != nil && !errorsmod.IsOf(err, ErrLocked) {
			t.Errorf("%s: rejected for the wrong reason: %v", c.name, err)
		}
	}
}

func TestAttackerWithdrawAddress(t *testing.T) {
	l, d, ctx := setup(t)
	ctx = ctx.WithBlockHeight(10)
	d[string(addrA)] = addrW
	for _, m := range []sdk.Msg{
		&distrtypes.MsgWithdrawDelegatorReward{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x"},
		&stakingtypes.MsgUndelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]},
	} {
		if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{m}, okFee, addrP, nil}); err == nil {
			t.Fatalf("%T with an attacker withdraw address must be refused", m)
		}
	}
	d[string(addrA)] = addrR
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{&stakingtypes.MsgUndelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]}}, okFee, addrP, nil}); err != nil {
		t.Fatalf("withdraw address = recovery must be allowed: %v", err)
	}
}

func TestEmptyTx(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(10)
	if err := runAnte(l, ctx, fakeTx{}); err != nil {
		t.Fatalf("empty tx must pass through to later decorators: %v", err)
	}
}

func TestExpiryByTime(t *testing.T) {
	l, _, ctx := setup(t)
	tx := fakeTx{[]sdk.Msg{send(addrA, addrB)}, okFee, addrP, nil}
	if err := runAnte(l, ctx.WithBlockHeight(50).WithBlockTime(time.Unix(expiry-1, 0)), tx); err == nil {
		t.Fatal("one second before expiry must be locked")
	}
	if err := runAnte(l, ctx.WithBlockHeight(50).WithBlockTime(time.Unix(expiry, 0)), tx); err != nil {
		t.Fatalf("at expiry must be unlocked: %v", err)
	}
}

func TestCheckTxAfterRestart(t *testing.T) {
	l, _, ctx := setup(t)
	tx := fakeTx{[]sdk.Msg{send(addrA, addrB)}, okFee, addrP, nil}
	c := ctx.WithIsCheckTx(true).WithBlockHeight(0).WithBlockTime(time.Time{})
	lastCommitted = 40
	defer func() { lastCommitted = 0 }()
	if err := runAnte(l, c, tx); err == nil {
		t.Fatal("CheckTx at height 0 after restart with committed height 40 must be locked")
	}
	lastCommitted = 5
	if err := runAnte(l, c, tx); err != nil {
		t.Fatalf("committed height 5 is before activation: %v", err)
	}
}

func TestRestriction(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(15)
	r := l.SendRestriction()
	if _, err := r(ctx, addrA, addrB, okFee); err == nil {
		t.Fatal("listed to other must fail")
	}
	if _, err := r(ctx, addrA, sdk.AccAddress([]byte("fee_collector_______")), okFee); err == nil {
		t.Fatal("listed to fee collector must fail")
	}
	if to, err := r(ctx, addrA, addrR, okFee); err != nil || !to.Equals(addrR) {
		t.Fatal("listed to recovery must pass")
	}
	if _, err := r(ctx, addrB, addrA, okFee); err != nil {
		t.Fatal("others unaffected")
	}
	if _, err := r(ctx.WithBlockTime(time.Unix(expiry, 0)), addrA, addrB, okFee); err != nil {
		t.Fatal("must be off at expiry")
	}
}

type bankMock struct {
	stakingtypes.BankKeeper
	called bool
}

func (b *bankMock) DelegateCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	b.called = true
	return nil
}

func TestStakingGuard(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(12)
	inner := &bankMock{}
	sb := StakingBank{BankKeeper: inner, L: l}
	if err := sb.DelegateCoinsFromAccountToModule(ctx, addrA, "bonded", okFee); err == nil || inner.called {
		t.Fatal("listed delegation must be rejected before reaching bank")
	}
	if err := sb.DelegateCoinsFromAccountToModule(ctx, addrB, "bonded", okFee); err != nil || !inner.called {
		t.Fatal("others pass through")
	}
}

func TestCheckTxHeight(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(9).WithIsCheckTx(true)
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{send(addrA, addrB)}, okFee, addrP, nil}); err == nil {
		t.Fatal("CheckTx at committed height 9 targets block 10 and must reject")
	}
}

type routerMock struct{}

func (routerMock) Handler(sdk.Msg) baseapp.MsgServiceHandler {
	return func(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }
}
func (routerMock) HandlerByTypeURL(string) baseapp.MsgServiceHandler {
	return routerMock{}.Handler(nil)
}

func TestRouterDispatch(t *testing.T) {
	l, _, ctx := setup(t)
	ctx = ctx.WithBlockHeight(11)
	r := Router{Inner: routerMock{}, L: l}
	del := &stakingtypes.MsgDelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]}
	if _, err := r.Handler(del)(ctx, del); err == nil {
		t.Fatal("contract or ICA dispatch of a listed delegate must fail")
	}
	// allowed messages too: dispatch has no fee payer, so a listed signer at
	// any depth is refused with the lock error
	for _, m := range []sdk.Msg{
		send(addrA, addrR),
		&stakingtypes.MsgUndelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]},
		exec(t, addrG, send(addrA, addrR)),
		exec(t, addrG, exec(t, addrG, &stakingtypes.MsgUndelegate{DelegatorAddress: addrA.String(), ValidatorAddress: "osmovaloper1x", Amount: okFee[0]})),
	} {
		_, err := r.HandlerByTypeURL(sdk.MsgTypeURL(m))(ctx, m)
		if err == nil || !errorsmod.IsOf(err, ErrLocked) {
			t.Fatalf("dispatch of %T with a listed signer must be refused with the lock error: %v", m, err)
		}
	}
	other := send(addrB, addrG)
	if _, err := r.Handler(other)(ctx, other); err != nil {
		t.Fatal("others unaffected")
	}
	if _, err := r.Handler(exec(t, addrG, other))(ctx, exec(t, addrG, other)); err != nil {
		t.Fatal("unlisted exec unaffected")
	}
	// an allowed message with a listed signer passes before activation and
	// after expiry
	ok := send(addrA, addrR)
	if _, err := r.Handler(ok)(ctx.WithBlockHeight(5), ok); err != nil {
		t.Fatal("no-op before activation")
	}
	if _, err := r.Handler(ok)(ctx.WithBlockTime(time.Unix(expiry, 0)), ok); err != nil {
		t.Fatal("no-op after expiry")
	}
}

func TestWithdrawValidatorCommission(t *testing.T) {
	l, d, ctx := setup(t)
	ctx = ctx.WithBlockHeight(10)
	valA, err := addresscodec.NewBech32Codec("osmovaloper").BytesToString(addrA)
	if err != nil {
		t.Fatal(err)
	}
	m := &distrtypes.MsgWithdrawValidatorCommission{ValidatorAddress: valA}
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{m}, okFee, addrP, nil}); err != nil {
		t.Fatalf("listed operator, own withdraw address: %v", err)
	}
	d[string(addrA)] = addrR
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{m}, okFee, addrP, nil}); err != nil {
		t.Fatalf("listed operator, recovery withdraw address: %v", err)
	}
	d[string(addrA)] = addrW
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{m}, okFee, addrP, nil}); err == nil {
		t.Fatal("listed operator with an attacker withdraw address must be refused")
	}
	delete(d, string(addrA))
	if err := runAnte(l, ctx, fakeTx{[]sdk.Msg{m}, okFee, addrB, nil}); err == nil {
		t.Fatal("commission withdrawal of a listed operator needs a fee payer")
	}
}

type wmock struct {
	m       map[string]sdk.AccAddress
	deleted int
}

func (w *wmock) GetDelegatorWithdrawAddr(_ context.Context, a sdk.AccAddress) (sdk.AccAddress, error) {
	if x, ok := w.m[string(a)]; ok {
		return x, nil
	}
	return a, nil
}
func (w *wmock) DeleteDelegatorWithdrawAddr(_ context.Context, a, _ sdk.AccAddress) error {
	delete(w.m, string(a))
	w.deleted++
	return nil
}

func TestStakingHooksReset(t *testing.T) {
	l, _, ctx := setup(t)
	w := &wmock{m: map[string]sdk.AccAddress{string(addrA): addrW, string(addrB): addrW}}
	h := StakingHooks{L: l, D: w}
	if err := h.BeforeDelegationSharesModified(ctx.WithBlockHeight(5), addrA, nil); err != nil || w.deleted != 0 {
		t.Fatal("must do nothing before activation")
	}
	ctx = ctx.WithBlockHeight(10)
	if err := h.BeforeDelegationSharesModified(ctx, addrB, nil); err != nil || w.deleted != 0 {
		t.Fatal("unlisted delegators untouched")
	}
	if err := h.BeforeDelegationSharesModified(ctx, addrA, nil); err != nil || w.deleted != 1 {
		t.Fatal("listed delegator with attacker withdraw address must be reset")
	}
	w.m[string(addrA)] = addrR
	if err := h.BeforeDelegationRemoved(ctx, addrA, nil); err != nil || w.deleted != 1 {
		t.Fatal("recovery withdraw address must be kept")
	}
	// validator removal pays commission to the operator account's withdraw address
	w.m[string(addrA)] = addrW
	if err := h.AfterValidatorRemoved(ctx.WithBlockHeight(5), nil, sdk.ValAddress(addrA)); err != nil || w.deleted != 1 {
		t.Fatal("validator removal: nothing before activation")
	}
	if err := h.AfterValidatorRemoved(ctx, nil, sdk.ValAddress(addrB)); err != nil || w.deleted != 1 {
		t.Fatal("validator removal: unlisted operator untouched")
	}
	if err := h.AfterValidatorRemoved(ctx, nil, sdk.ValAddress(addrA)); err != nil || w.deleted != 2 {
		t.Fatal("validator removal: listed operator with attacker withdraw address must be reset")
	}
}
