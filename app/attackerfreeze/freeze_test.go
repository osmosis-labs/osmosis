package attackerfreeze

import (
	"context"
	"testing"

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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const height = 100

var (
	addrF = sdk.AccAddress([]byte("frozen_f____________"))
	addrB = sdk.AccAddress([]byte("other_b_____________"))
	addrG = sdk.AccAddress([]byte("grantee_____________"))
	fee   = sdk.NewCoins(sdk.NewCoin("uosmo", math.NewInt(5000)))
)

func setup(t *testing.T) (*Freeze, sdk.Context) {
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
	authz.RegisterInterfaces(reg)
	f := New([][]byte{addrF}, height, codec.NewProtoCodec(reg), func() int64 { return 0 })
	key := storetypes.NewKVStoreKey("test")
	return f, testutil.DefaultContext(key, storetypes.NewTransientStoreKey("t"))
}

type fakeTx struct {
	msgs    []sdk.Msg
	payer   []byte
	granter []byte
}

func (f fakeTx) GetMsgs() []sdk.Msg                    { return f.msgs }
func (f fakeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (f fakeTx) GetGas() uint64                        { return 100000 }
func (f fakeTx) GetFee() sdk.Coins                     { return fee }
func (f fakeTx) FeePayer() []byte                      { return f.payer }
func (f fakeTx) FeeGranter() []byte                    { return f.granter }

func send(from, to sdk.AccAddress) sdk.Msg {
	return &banktypes.MsgSend{FromAddress: from.String(), ToAddress: to.String(), Amount: fee}
}

func exec(grantee sdk.AccAddress, msgs ...sdk.Msg) sdk.Msg {
	m := authz.NewMsgExec(grantee, msgs)
	return &m
}

func runAnte(f *Freeze, ctx sdk.Context, tx fakeTx) error {
	_, err := Decorator{F: f}.AnteHandle(ctx, tx, false, func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil })
	return err
}

type routerMock struct{}

func (routerMock) Handler(sdk.Msg) baseapp.MsgServiceHandler { return ok }
func (routerMock) HandlerByTypeURL(string) baseapp.MsgServiceHandler {
	return ok
}

func ok(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }

type bankMock struct{ stakingtypes.BankKeeper }

func (bankMock) DelegateCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}

func TestInactiveBelowHeight(t *testing.T) {
	f, ctx := setup(t)
	ctx = ctx.WithBlockHeight(height - 1)
	if to, err := f.SendRestriction()(ctx, addrF, addrB, fee); err != nil || !to.Equals(addrB) {
		t.Fatal("send restriction must be a no-op below the activation height")
	}
	if err := runAnte(f, ctx, fakeTx{msgs: []sdk.Msg{send(addrF, addrB)}, payer: addrF}); err != nil {
		t.Fatalf("ante must be a no-op below the activation height: %v", err)
	}
	m := send(addrF, addrB)
	if _, err := (Router{Inner: routerMock{}, F: f}).Handler(m)(ctx, m); err != nil {
		t.Fatal("router must be a no-op below the activation height")
	}
	if err := (StakingBank{BankKeeper: bankMock{}, F: f}).DelegateCoinsFromAccountToModule(ctx, addrF, "bonded", fee); err != nil {
		t.Fatal("delegation must be allowed below the activation height")
	}
	if ctx.GasMeter().GasConsumed() != 0 {
		t.Fatal("no gas below the activation height")
	}
	// CheckTx checks the next block
	if err := runAnte(f, ctx.WithIsCheckTx(true), fakeTx{msgs: []sdk.Msg{send(addrF, addrB)}, payer: addrB}); !errorsmod.IsOf(err, ErrFrozen) {
		t.Fatalf("CheckTx one block before activation must refuse: %v", err)
	}
}

func TestFrozen(t *testing.T) {
	f, ctx := setup(t)
	ctx = ctx.WithBlockHeight(height)

	for _, to := range []sdk.AccAddress{addrB, addrG, addrF} {
		if _, err := f.SendRestriction()(ctx, addrF, to, fee); !errorsmod.IsOf(err, ErrFrozen) {
			t.Fatalf("coins of a frozen account must not leave it (to %s): %v", to, err)
		}
	}
	if _, err := f.SendRestriction()(ctx, addrB, addrF, fee); err != nil {
		t.Fatal("coins may still arrive at a frozen account")
	}
	if err := (StakingBank{BankKeeper: bankMock{}, F: f}).DelegateCoinsFromAccountToModule(ctx, addrF, "bonded", fee); !errorsmod.IsOf(err, ErrFrozen) {
		t.Fatalf("a frozen account must not delegate: %v", err)
	}

	cases := []struct {
		name string
		tx   fakeTx
		ok   bool
	}{
		{"frozen signer, other fee payer", fakeTx{msgs: []sdk.Msg{send(addrF, addrB)}, payer: addrB}, false},
		{"frozen signer in authz exec", fakeTx{msgs: []sdk.Msg{exec(addrG, send(addrF, addrB))}, payer: addrG}, false},
		{"frozen signer nested two levels", fakeTx{msgs: []sdk.Msg{exec(addrG, exec(addrG, send(addrF, addrB)))}, payer: addrG}, false},
		{"frozen fee payer", fakeTx{msgs: []sdk.Msg{send(addrB, addrG)}, payer: addrF}, false},
		{"frozen fee granter", fakeTx{msgs: []sdk.Msg{send(addrB, addrG)}, payer: addrB, granter: addrF}, false},
		{"unrelated tx", fakeTx{msgs: []sdk.Msg{send(addrB, addrG)}, payer: addrB}, true},
		{"send to a frozen account", fakeTx{msgs: []sdk.Msg{send(addrB, addrF)}, payer: addrB}, true},
	}
	for _, c := range cases {
		err := runAnte(f, ctx, c.tx)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v err=%v", c.name, c.ok, err)
		}
		if err != nil && !errorsmod.IsOf(err, ErrFrozen) {
			t.Errorf("%s: refused for the wrong reason: %v", c.name, err)
		}
	}

	r := Router{Inner: routerMock{}, F: f}
	for _, m := range []sdk.Msg{send(addrF, addrB), exec(addrG, send(addrF, addrB))} {
		if _, err := r.HandlerByTypeURL(sdk.MsgTypeURL(m))(ctx, m); !errorsmod.IsOf(err, ErrFrozen) {
			t.Fatalf("dispatch of %T with a frozen signer must be refused: %v", m, err)
		}
	}
	if m := send(addrB, addrG); func() error { _, err := r.Handler(m)(ctx, m); return err }() != nil {
		t.Fatal("unrelated dispatch must pass")
	}
}

func TestEmbeddedList(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("osmo", "osmopub")
	ac := addresscodec.NewBech32Codec("osmo")
	list := MustLoad(ac, map[string]bool{})
	if len(list) != 48 {
		t.Fatalf("frozen.txt has %d accounts, want 48", len(list))
	}
	if _, err := parse("osmo1lzxp35uxugvns092592d3w9zlpycmrpjlny7tp\nosmo1lzxp35uxugvns092592d3w9zlpycmrpjlny7tp\n", ac, nil); err == nil {
		t.Fatal("duplicates must be refused")
	}
	if _, err := parse("osmo1lzxp35uxugvns092592d3w9zlpycmrpjlny7tp\n", ac, map[string]bool{"osmo1lzxp35uxugvns092592d3w9zlpycmrpjlny7tp": true}); err == nil {
		t.Fatal("blocked accounts must be refused")
	}
}
