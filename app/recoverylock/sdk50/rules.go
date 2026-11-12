package sdk50

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/core"
)

// DistrReader is the part of the distribution keeper the rules need.
type DistrReader interface {
	GetDelegatorWithdrawAddr(ctx context.Context, delAddr sdk.AccAddress) (sdk.AccAddress, error)
}

func registerStandardRules(l *Lock, distr DistrReader) {
	l.Rules.Register(sdk.MsgTypeURL(&banktypes.MsgSend{}), func(c *core.Config, _ any, msg any, _ []byte) error {
		m := msg.(*banktypes.MsgSend)
		to, err := l.AddrBytes(m.ToAddress)
		if err != nil {
			return err
		}
		if !c.IsRecovery(to) {
			return lockedf("send to %s: only the recovery address is allowed", m.ToAddress)
		}
		return nil
	})
	l.Rules.Register(sdk.MsgTypeURL(&banktypes.MsgMultiSend{}), func(c *core.Config, _ any, msg any, _ []byte) error {
		for _, o := range msg.(*banktypes.MsgMultiSend).Outputs {
			to, err := l.AddrBytes(o.Address)
			if err != nil {
				return err
			}
			if !c.IsRecovery(to) {
				return lockedf("multi-send output %s: only the recovery address is allowed", o.Address)
			}
		}
		return nil
	})
	// Undelegate and reward withdrawal pay rewards to the withdraw address,
	// so both require it to be the wallet or the recovery address. The
	// recovery operator fixes an attacker withdraw address first with
	// MsgSetWithdrawAddress (allowed below).
	l.Rules.Register(sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}), func(c *core.Config, ctx any, msg any, signer []byte) error {
		if err := sameAddr(l, msg.(*stakingtypes.MsgUndelegate).DelegatorAddress, signer); err != nil {
			return err
		}
		return safeWithdrawAddr(c, ctx, distr, signer)
	})
	l.Rules.Register(sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}), func(c *core.Config, ctx any, msg any, signer []byte) error {
		if err := sameAddr(l, msg.(*distrtypes.MsgWithdrawDelegatorReward).DelegatorAddress, signer); err != nil {
			return err
		}
		return safeWithdrawAddr(c, ctx, distr, signer)
	})
	// Commission of a listed validator operator is paid to the operator
	// account's withdraw address, so the same rule applies. The only signer
	// field of the message is validator_address, so the signer bytes are the
	// validator address bytes (sdk.AccAddress(valAddr)).
	l.Rules.Register(sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}), func(c *core.Config, ctx any, _ any, signer []byte) error {
		return safeWithdrawAddr(c, ctx, distr, signer)
	})
	l.Rules.Register(sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}), func(c *core.Config, _ any, msg any, signer []byte) error {
		m := msg.(*distrtypes.MsgSetWithdrawAddress)
		if err := sameAddr(l, m.DelegatorAddress, signer); err != nil {
			return err
		}
		w, err := l.AddrBytes(m.WithdrawAddress)
		if err != nil {
			return err
		}
		if !bytes.Equal(w, signer) && !c.IsRecovery(w) {
			return lockedf("withdraw address may only be the wallet or the recovery address")
		}
		return nil
	})
}

func safeWithdrawAddr(c *core.Config, ctx any, distr DistrReader, signer []byte) error {
	w, err := distr.GetDelegatorWithdrawAddr(ctx.(sdk.Context), signer)
	if err != nil {
		return err
	}
	if !bytes.Equal(w, signer) && !c.IsRecovery(w) {
		return lockedf("withdraw address %s is neither the wallet nor the recovery address; set it first", w)
	}
	return nil
}

// SelfOrRecovery requires a bech32 field to be the signer or the recovery address.
func SelfOrRecovery(l *Lock, field string, signer []byte) error {
	b, err := l.AddrBytes(field)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, signer) && !l.Cfg.IsRecovery(b) {
		return lockedf("address %s is neither the wallet nor the recovery address", field)
	}
	return nil
}

// Recovery requires a bech32 field to be the recovery address.
func Recovery(l *Lock, field string) error {
	b, err := l.AddrBytes(field)
	if err != nil {
		return err
	}
	if !l.Cfg.IsRecovery(b) {
		return lockedf("address %s is not the recovery address", field)
	}
	return nil
}

// sameAddr requires a bech32 field of the message to be the signer itself.
func sameAddr(l *Lock, field string, signer []byte) error {
	b, err := l.AddrBytes(field)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, signer) {
		return lockedf("address %s is not the signer", field)
	}
	return nil
}

// SameAddr is exported for chain specific rules.
func SameAddr(l *Lock, field string, signer []byte) error { return sameAddr(l, field, signer) }
