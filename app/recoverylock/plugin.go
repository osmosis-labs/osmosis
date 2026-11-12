package recoverylock

import (
	"bytes"
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/core"
	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/sdk50"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	cltypes "github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
)

// LazyDistr lets the lock read the distribution keeper, which Osmosis creates
// after the staking keeper (and the staking keeper needs the lock first).
type LazyDistr struct{ K *distrkeeper.Keeper }

// GetDelegatorWithdrawAddr implements sdk50.DistrReader.
func (l *LazyDistr) GetDelegatorWithdrawAddr(ctx context.Context, a sdk.AccAddress) (sdk.AccAddress, error) {
	return l.K.GetDelegatorWithdrawAddr(ctx, a)
}

// RegisterRules adds the Osmosis exit messages for listed signers. Each one
// pays only into the signer's own wallet (in the Osmosis keepers, lockup
// unlock pays the lock owner and CL withdraw and collect pay the position
// owner), where the coin rule then applies, moves value only to the recovery
// address, or moves no value (reward receiver reset, removal of contract
// rights).
//
// gamm MsgExitPool is deliberately not allowed: BurnPoolShareFromAccount
// (x/gamm/keeper/share.go) moves the shares with SendCoinsFromAccountToModule,
// which the coin rule blocks, so it could never succeed. gamm shares are bank
// coins and are rescued with MsgSend to the recovery address.
func RegisterRules(l *sdk50.Lock) {
	self := func(field func(any) string) core.Rule {
		return func(_ *core.Config, _ any, msg any, signer []byte) error {
			return sdk50.SameAddr(l, field(msg), signer)
		}
	}
	l.Rules.Register(sdk.MsgTypeURL(&lockuptypes.MsgBeginUnlocking{}), self(func(m any) string { return m.(*lockuptypes.MsgBeginUnlocking).Owner }))
	l.Rules.Register(sdk.MsgTypeURL(&lockuptypes.MsgBeginUnlockingAll{}), self(func(m any) string { return m.(*lockuptypes.MsgBeginUnlockingAll).Owner }))
	l.Rules.Register(sdk.MsgTypeURL(&cltypes.MsgWithdrawPosition{}), self(func(m any) string { return m.(*cltypes.MsgWithdrawPosition).Sender }))
	l.Rules.Register(sdk.MsgTypeURL(&cltypes.MsgCollectSpreadRewards{}), self(func(m any) string { return m.(*cltypes.MsgCollectSpreadRewards).Sender }))
	l.Rules.Register(sdk.MsgTypeURL(&cltypes.MsgCollectIncentives{}), self(func(m any) string { return m.(*cltypes.MsgCollectIncentives).Sender }))
	// Lock incentives are paid to the lock's reward receiver. The recovery
	// operator resets an attacker-set receiver to the owner or the recovery
	// address by transaction.
	l.Rules.Register(sdk.MsgTypeURL(&lockuptypes.MsgSetRewardReceiverAddress{}), func(_ *core.Config, _ any, msg any, signer []byte) error {
		m := msg.(*lockuptypes.MsgSetRewardReceiverAddress)
		if err := sdk50.SameAddr(l, m.Owner, signer); err != nil {
			return err
		}
		return sdk50.SelfOrRecovery(l, m.RewardReceiver, signer)
	})
	// Contract exit: a cw20 balance or a cw721 token held in contract state
	// moves to the recovery address with exactly
	// {"transfer":{"recipient":<recovery>,"amount":"<uint>"}} or
	// {"transfer_nft":{"recipient":<recovery>,"token_id":"<string>"}}, and
	// rights granted earlier can be revoked with exactly
	// {"decrease_allowance":{"spender":<address>,"amount":"<uint>"}} (no
	// expires), {"revoke":{"spender":<address>,"token_id":"<string>"}} or
	// {"revoke_all":{"operator":<address>}}; parsed strictly
	// (core.CheckContractExit), with no funds attached. Only the message shape
	// is checked, not which contract runs it. send, send_nft,
	// approve, increase_allowance and every other execute are refused. The
	// contract runs with the wallet as sender but cannot move the wallet's
	// bank coins (no funds; its own submessages have the contract as sender).
	// Only in a top level message of a transaction paid by a fee payer: the
	// router wrapper refuses every dispatched message with a listed signer.
	l.Rules.Register(sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}), func(c *core.Config, _ any, msg any, signer []byte) error {
		m := msg.(*wasmtypes.MsgExecuteContract)
		if err := sdk50.SameAddr(l, m.Sender, signer); err != nil {
			return err
		}
		if len(m.Funds) != 0 {
			return fmt.Errorf("contract execute with funds is not allowed")
		}
		rec, err := l.AddrString(c.Recovery())
		if err != nil {
			return err
		}
		return core.CheckContractExit(m.Msg, rec, func(a string) error {
			_, err := l.AddrBytes(a)
			return err
		})
	})
	// CL positions move to the recovery address as a whole, with uncollected
	// rewards (transferPositions moves no coins, x/concentrated-liquidity
	// position.go). It refuses positions with an active underlying lock.
	l.Rules.Register(sdk.MsgTypeURL(&cltypes.MsgTransferPositions{}), func(_ *core.Config, _ any, msg any, signer []byte) error {
		m := msg.(*cltypes.MsgTransferPositions)
		if err := sdk50.SameAddr(l, m.Sender, signer); err != nil {
			return err
		}
		return sdk50.Recovery(l, m.NewOwner)
	})
}

// LockRewardReceiver is the reward receiver rule for x/incentives lock
// distributions (also superfluid synthetic lock payouts, which use the same
// code). While the lock is active, rewards of a lock owned by a listed wallet
// go to the owner unless the stored receiver is the owner or the recovery
// address; the owner's wallet is where the coin rule applies. Before
// activation and after expiry it returns the receiver unchanged, without
// reading state. Deterministic, writes nothing.
func LockRewardReceiver(l *sdk50.Lock) func(ctx sdk.Context, owner, receiver string) string {
	return func(ctx sdk.Context, owner, receiver string) string {
		if !l.Active(ctx) || receiver == owner {
			return receiver
		}
		o, err := l.AddrBytes(owner)
		if err != nil || !l.Cfg.IsListed(o) {
			return receiver
		}
		if r, err := l.AddrBytes(receiver); err == nil && (bytes.Equal(r, o) || l.Cfg.IsRecovery(r)) {
			return receiver
		}
		return owner
	}
}
