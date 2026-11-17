// Package attackerfreeze freezes a fixed list of attacker accounts from the
// v32.0.1 activation height on, with no expiry.
//
// From the activation height:
//   - no coins leave a frozen account (bank send restriction, which covers
//     fees, sends, multi-send, IBC escrow, module deposits and contract funds);
//   - a frozen account cannot delegate (x/staking moves delegated coins with
//     DelegateCoins, which skips the send restriction);
//   - a transaction in which a frozen account signs a message at any depth, or
//     is the fee payer or fee granter, is refused in the ante handler;
//   - a message dispatched by authz, CosmWasm or the ICA host with a frozen
//     signer at any depth is refused.
//
// Coins can still arrive at a frozen account. Below the activation height
// every part returns at once without reading state, using gas or emitting
// events, so v32.0.1 replays earlier blocks exactly as v32.0.0.
//
// The list is compiled in from frozen.txt. Changing it needs a new binary, and
// later binaries must carry it (or move the funds in an upgrade handler).
package attackerfreeze

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:embed frozen.txt
var frozenFile string

// MaxDepth caps unpacking of nested authz MsgExec. Deeper nesting is refused
// from the activation height, because a level that is not unpacked could hide
// a frozen signer.
const MaxDepth = 6

// ErrFrozen is returned for everything the freeze refuses.
var ErrFrozen = errorsmod.Register("attackerfreeze", 1, "account frozen")

// Freeze holds the frozen set and the activation height.
type Freeze struct {
	frozen map[string]struct{}
	height int64
	cdc    codec.Codec
	// lastHeight returns the latest committed height; used only in CheckTx
	// when the context height is 0 (right after a restart).
	lastHeight func() int64
}

// New builds a freeze for the given account addresses (raw bytes).
func New(frozen [][]byte, activationHeight int64, cdc codec.Codec, lastHeight func() int64) *Freeze {
	f := &Freeze{frozen: make(map[string]struct{}, len(frozen)), height: activationHeight, cdc: cdc, lastHeight: lastHeight}
	for _, a := range frozen {
		f.frozen[string(a)] = struct{}{}
	}
	return f
}

// MustLoad parses the embedded frozen.txt and panics on any problem, so a
// binary with a bad list never starts: every line must be a 20 byte osmo
// account address that is not a module or blocked account, with no
// duplicates. Empty lines and lines starting with # are skipped.
func MustLoad(ac address.Codec, blocked map[string]bool) [][]byte {
	out, err := parse(frozenFile, ac, blocked)
	if err != nil {
		panic("attacker freeze list: " + err.Error())
	}
	return out
}

func parse(file string, ac address.Codec, blocked map[string]bool) ([][]byte, error) {
	var out [][]byte
	seen := map[string]int{}
	for i, line := range strings.Split(file, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		b, err := ac.StringToBytes(line)
		if err != nil {
			return nil, fmt.Errorf("line %d %q: %w", i+1, line, err)
		}
		if len(b) != 20 {
			return nil, fmt.Errorf("line %d %q has %d bytes, want 20", i+1, line, len(b))
		}
		canon, err := ac.BytesToString(b)
		if err != nil {
			return nil, err
		}
		if blocked[canon] {
			return nil, fmt.Errorf("line %d %q is a module or blocked account", i+1, line)
		}
		if prev, dup := seen[string(b)]; dup {
			return nil, fmt.Errorf("line %d duplicates line %d", i+1, prev)
		}
		seen[string(b)] = i + 1
		out = append(out, b)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return out, nil
}

// IsFrozen reports whether addr is a frozen account.
func (f *Freeze) IsFrozen(addr []byte) bool {
	_, ok := f.frozen[string(addr)]
	return ok
}

// Count returns the number of frozen accounts.
func (f *Freeze) Count() int { return len(f.frozen) }

// active reports whether the freeze applies. In block execution it uses the
// block height. In CheckTx it checks the next block (context height + 1, or
// the last committed height + 1 right after a restart), which only makes the
// mempool stricter, never consensus.
func (f *Freeze) active(ctx sdk.Context) bool {
	h := ctx.BlockHeight()
	if ctx.IsCheckTx() {
		if h == 0 && f.lastHeight != nil {
			h = f.lastHeight()
		}
		h++
	}
	return h >= f.height
}

func frozenf(format string, a ...any) error {
	return errorsmod.Wrap(ErrFrozen, fmt.Sprintf(format, a...))
}

// SendRestriction is the bank send restriction: no coins leave a frozen
// account, to any destination.
func (f *Freeze) SendRestriction() banktypes.SendRestrictionFn {
	return func(ctx context.Context, from, to sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
		if !f.active(sdk.UnwrapSDKContext(ctx)) || !f.IsFrozen(from) {
			return to, nil
		}
		return nil, frozenf("coins of %s are frozen", from)
	}
}

// StakingBank wraps the bank keeper given to x/staking: DelegateCoins skips
// the send restriction, so delegation from a frozen account is refused here.
type StakingBank struct {
	stakingtypes.BankKeeper
	F *Freeze
}

// DelegateCoinsFromAccountToModule refuses frozen delegators while active.
func (b StakingBank) DelegateCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, module string, amt sdk.Coins) error {
	if b.F.active(sdk.UnwrapSDKContext(ctx)) && b.F.IsFrozen(sender) {
		return frozenf("%s may not delegate", sender)
	}
	return b.BankKeeper.DelegateCoinsFromAccountToModule(ctx, sender, module, amt)
}

// checkMsg refuses msg if a frozen account signs it or any message it wraps
// in an authz MsgExec.
func (f *Freeze) checkMsg(msg sdk.Msg, depth int) error {
	if depth > MaxDepth {
		return frozenf("message nesting deeper than %d", MaxDepth)
	}
	signers, _, err := f.cdc.GetMsgV1Signers(msg)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	for _, s := range signers {
		if f.IsFrozen(s) {
			return frozenf("%s is frozen", sdk.AccAddress(s))
		}
	}
	if exec, ok := msg.(*authz.MsgExec); ok {
		inner, err := exec.GetMessages()
		if err != nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		for _, m := range inner {
			if err := f.checkMsg(m, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// Decorator is the ante decorator. Place it before fee deduction and before
// any signature path split, so a refused transaction costs nothing.
type Decorator struct{ F *Freeze }

// AnteHandle refuses a transaction in which a frozen account signs a message
// at any depth, or is the fee payer or fee granter.
func (d Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	f := d.F
	if !f.active(ctx) {
		return next(ctx, tx, simulate)
	}
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		if err := f.checkMsg(m, 0); err != nil {
			return ctx, err
		}
	}
	// FeePayer() panics on a transaction without signers; baseapp refuses
	// transactions without messages before the ante handler runs.
	if feeTx, ok := tx.(sdk.FeeTx); ok && len(msgs) > 0 {
		if f.IsFrozen(feeTx.FeePayer()) || f.IsFrozen(feeTx.FeeGranter()) {
			return ctx, frozenf("frozen accounts may not pay or grant fees")
		}
	}
	return next(ctx, tx, simulate)
}

// Router wraps the message router held by the authz, CosmWasm and ICA host
// keepers and refuses every dispatched message with a frozen signer at any
// depth. Top level transaction messages are checked by the Decorator.
type Router struct {
	Inner baseapp.MessageRouter
	F     *Freeze
}

// Handler returns the inner handler wrapped with the freeze check.
func (r Router) Handler(msg sdk.Msg) baseapp.MsgServiceHandler { return r.wrap(r.Inner.Handler(msg)) }

// HandlerByTypeURL returns the inner handler wrapped with the freeze check.
func (r Router) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	return r.wrap(r.Inner.HandlerByTypeURL(typeURL))
}

func (r Router) wrap(h baseapp.MsgServiceHandler) baseapp.MsgServiceHandler {
	if h == nil {
		return nil
	}
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		if r.F.active(ctx) {
			if err := r.F.checkMsg(msg, 0); err != nil {
				return nil, err
			}
		}
		return h(ctx, msg)
	}
}
