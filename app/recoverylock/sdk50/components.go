package sdk50

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SendRestriction returns the coin rule for bank.AppendSendRestriction: coins
// leaving a listed address may only go to the recovery address. This covers
// fees, sends, multi-send, IBC escrow, module deposits and contract funds,
// because they all go through SendCoins or InputOutputCoins.
func (l *Lock) SendRestriction() banktypes.SendRestrictionFn {
	return func(ctx context.Context, from, to sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
		sctx := sdk.UnwrapSDKContext(ctx)
		if !l.active(sctx) || !l.Cfg.IsListed(from) || l.Cfg.IsRecovery(to) {
			return to, nil
		}
		return nil, lockedf("coins of %s may only go to the recovery address", from)
	}
}

// StakingBank wraps the bank keeper given to x/staking. DelegateCoins does not
// apply the send restriction, so a new delegation from the balance of a
// listed address is rejected here, whatever path reached staking. It covers only coins moving from the balance into staking:
// redelegation and cancel unbonding move no coins from the balance and are
// stopped by the message rules (neither is an allowed message), which apply
// to transactions (Decorator) and to dispatched messages (Router).
type StakingBank struct {
	stakingtypes.BankKeeper
	L *Lock
}

// DelegateCoinsFromAccountToModule rejects listed delegators while active.
func (b StakingBank) DelegateCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, module string, amt sdk.Coins) error {
	if l := b.L; l.active(sdk.UnwrapSDKContext(ctx)) && l.Cfg.IsListed(sender) {
		return lockedf("%s may not delegate", sender)
	}
	return b.BankKeeper.DelegateCoinsFromAccountToModule(ctx, sender, module, amt)
}

// Router wraps the message router held by the authz, CosmWasm and ICA host
// keepers (authz MsgExec inner messages, contract dispatch, ICA host
// execution). Top level transaction messages do not pass through it: baseapp
// routes them with its own router after the ante handler, where the Decorator
// applies the rules and the fee payer requirement.
//
// While the lock is active, the wrapper refuses every dispatched message that
// has a listed signer at any depth, even one the rules would allow. Dispatch
// has no fee payer: with an authz grant made before activation, a contract or
// any grantee could otherwise run allowed messages as a listed wallet without
// the recovery operator (for example fill all unbonding entries with tiny
// undelegations so the operator's undelegation fails). Listed wallets
// therefore act only in top level messages of transactions that pass the
// Decorator. Messages without a listed signer at any depth pass, except that
// errors of the check itself (nesting deeper than MaxDepth, undecodable
// wrapped messages) still refuse.
type Router struct {
	Inner baseapp.MessageRouter
	L     *Lock
}

// Handler returns the inner handler wrapped with the lock check.
func (r Router) Handler(msg sdk.Msg) baseapp.MsgServiceHandler { return r.wrap(r.Inner.Handler(msg)) }

// HandlerByTypeURL returns the inner handler wrapped with the lock check.
func (r Router) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	return r.wrap(r.Inner.HandlerByTypeURL(typeURL))
}

func (r Router) wrap(h baseapp.MsgServiceHandler) baseapp.MsgServiceHandler {
	if h == nil {
		return nil
	}
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		if r.L.active(ctx) {
			involved, err := r.L.checkMsg(ctx, msg, 0)
			if err != nil {
				return nil, err
			}
			if involved {
				return nil, lockedf("listed accounts may act only in transactions paid by a fee payer")
			}
		}
		return h(ctx, msg)
	}
}

// Decorator is the ante decorator. Place it before any fee deduction and
// before any signature path split (for example Osmosis smart accounts). A
// rejected ante costs nothing: baseapp runs the ante handler in a cached
// context and discards it on error.
type Decorator struct{ L *Lock }

// AnteHandle rejects a transaction that involves a listed address (as signer
// of a message at any depth, as fee payer or as fee granter) unless every
// message with a listed signer is allowed and the fee is paid by the
// configured fee payer set, without a fee granter, and is not zero.
func (d Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	l := d.L
	if !l.active(ctx) {
		return next(ctx, tx, simulate)
	}
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		// Nothing to check, and FeePayer() would panic on an empty signer
		// list. baseapp refuses transactions without messages before the
		// ante handler runs.
		return next(ctx, tx, simulate)
	}
	involved := false
	for _, m := range msgs {
		sub, err := l.checkMsg(ctx, m, 0)
		if err != nil {
			return ctx, err
		}
		involved = involved || sub
	}
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		if involved {
			return ctx, lockedf("transaction without fee information")
		}
		return next(ctx, tx, simulate)
	}
	if l.Cfg.IsListed(feeTx.FeePayer()) || l.Cfg.IsListed(feeTx.FeeGranter()) {
		involved = true
	}
	if involved {
		if !l.Cfg.IsFeePayer(feeTx.FeePayer()) {
			return ctx, lockedf("fee must be paid by one of the configured fee payers")
		}
		if len(feeTx.FeeGranter()) != 0 {
			return ctx, lockedf("fee grants are not allowed")
		}
		if feeTx.GetFee().IsZero() {
			return ctx, lockedf("fee must not be zero")
		}
	}
	return next(ctx, tx, simulate)
}
