// Package sdk50 adapts the recovery lock core to cosmos-sdk v0.50 and forks
// with the same APIs (for example osmosis-labs/cosmos-sdk v0.50.x-osmo).
//
// Components:
//   - SendRestriction: bank send restriction, the coin rule
//   - StakingBank: bank keeper wrapper for x/staking, the staking guard
//   - Router: message router wrapper for authz, wasm and ICA host dispatch;
//     refuses any dispatched message with a listed signer
//   - Decorator: ante decorator, early rejection at no cost
//   - Rules: allowed messages for listed signers (rules only read state;
//     there is no cleanup step)
//   - StakingHooks: reset an unsafe withdraw address of a listed delegator
//     before automatic reward and commission payouts
//
// Before the activation height and from the expiry time on, every component
// returns immediately without reading state, consuming gas or emitting events.
package sdk50

import (
	"fmt"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/core"
)

// MaxDepth caps recursive unpacking of wrapper messages (authz MsgExec).
// While the lock is active, a message nested deeper than this is refused with
// the lock error for every transaction, also one without any listed signer:
// a level that is not unpacked could hide a listed signer. This only changes
// which transactions the chain accepts while active (deterministic); no
// normal use nests MsgExec this deep.
const MaxDepth = 6

// ErrLocked is returned for everything the lock rejects by rule. Messages
// whose signers or wrapped messages cannot be decoded are refused with
// ErrInvalidRequest.
var ErrLocked = errorsmod.Register("recoverylock", 1, "recovery lock")

// Lock bundles configuration, rules and the codecs it needs.
type Lock struct {
	Cfg   *core.Config
	Rules *core.Registry
	cdc   codec.Codec
	addr  address.Codec
	// lastHeight returns the latest committed height (baseapp
	// LastBlockHeight). Used only in CheckTx when the context height is 0,
	// which happens right after a node restart.
	lastHeight func() int64
}

// New builds a lock with the standard rules registered (bank send and
// multi-send to the recovery address, undelegate, withdraw delegator reward,
// withdraw validator commission, set withdraw address). Chains add their own
// exit rules on top.
func New(cfg *core.Config, cdc codec.Codec, accAddr address.Codec, distr DistrReader, lastHeight func() int64) *Lock {
	l := &Lock{Cfg: cfg, Rules: core.NewRegistry(), cdc: cdc, addr: accAddr, lastHeight: lastHeight}
	registerStandardRules(l, distr)
	return l
}

// AddrBytes decodes a bech32 account address with the chain's codec.
func (l *Lock) AddrBytes(s string) ([]byte, error) { return l.addr.StringToBytes(s) }

// AddrString encodes account address bytes with the chain's codec.
func (l *Lock) AddrString(b []byte) (string, error) { return l.addr.BytesToString(b) }

// Active reports whether the lock applies in this context. Exported for
// chain specific hooks (for example a reward receiver rule); same semantics
// as the check used by every component here.
func (l *Lock) Active(ctx sdk.Context) bool { return l.active(ctx) }

// active is the single entry check used by every component. In block
// execution it uses the block height and block time. In CheckTx the context
// carries the last committed height (0 right after a restart, then the
// committed height is used), so the next block is that height + 1; if the
// block time is unknown, expiry is ignored. Both only make the mempool
// stricter, never consensus.
func (l *Lock) active(ctx sdk.Context) bool {
	if !ctx.IsCheckTx() {
		return l.Cfg.Active(ctx.BlockHeight(), ctx.BlockTime().Unix())
	}
	h := ctx.BlockHeight()
	if h == 0 && l.lastHeight != nil {
		h = l.lastHeight()
	}
	h++
	if ctx.BlockTime().IsZero() {
		return l.Cfg.HeightActive(h)
	}
	return l.Cfg.Active(h, ctx.BlockTime().Unix())
}

// checkMsg applies the rules to msg and, recursively, to the messages wrapped
// in an authz MsgExec.
// It returns whether any listed signer was involved.
func (l *Lock) checkMsg(ctx sdk.Context, msg sdk.Msg, depth int) (bool, error) {
	if depth > MaxDepth {
		return false, errorsmod.Wrapf(ErrLocked, "message nesting deeper than %d", MaxDepth)
	}
	signers, _, err := l.cdc.GetMsgV1Signers(msg)
	if err != nil {
		return false, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	involved := false
	for _, s := range signers {
		if !l.Cfg.IsListed(s) {
			continue
		}
		involved = true
		if err := l.Rules.Check(l.Cfg, ctx, sdk.MsgTypeURL(msg), msg, s); err != nil {
			return true, errorsmod.Wrap(ErrLocked, err.Error())
		}
	}
	if exec, ok := msg.(*authz.MsgExec); ok {
		inner, err := exec.GetMessages()
		if err != nil {
			return involved, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		for _, m := range inner {
			sub, err := l.checkMsg(ctx, m, depth+1)
			involved = involved || sub
			if err != nil {
				return involved, err
			}
		}
	}
	return involved, nil
}

func lockedf(format string, a ...any) error {
	return errorsmod.Wrap(ErrLocked, fmt.Sprintf(format, a...))
}
