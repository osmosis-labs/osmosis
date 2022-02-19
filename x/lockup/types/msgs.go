package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants
const (
	TypeMsgLockTokens        = "lock_tokens"
	TypeMsgBeginUnlockingAll = "begin_unlocking_all"
	TypeMsgBeginUnlocking    = "begin_unlocking"
)

var _ sdk.Msg = &MsgLockTokens{}

// NewMsgLockTokens creates a message to lock tokens
func NewMsgLockTokens(owner sdk.AccAddress, duration time.Duration, coins sdk.Coins) *MsgLockTokens {
	return &MsgLockTokens{
		Owner:    owner.String(),
		Duration: duration,
		Coins:    coins,
	}
}

func (m MsgLockTokens) Route() string { return RouterKey }
func (m MsgLockTokens) Type() string  { return TypeMsgLockTokens }
func (m MsgLockTokens) ValidateBasic() error {
	if m.Duration <= 0 {
		return fmt.Errorf("duration should be positive: %d < 0", m.Duration)
	}
	return nil
}
func (m MsgLockTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgLockTokens) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgBeginUnlockingAll{}

// NewMsgBeginUnlockingAll creates a message to begin unlocking tokens
func NewMsgBeginUnlockingAll(owner sdk.AccAddress) *MsgBeginUnlockingAll {
	return &MsgBeginUnlockingAll{
		Owner: owner.String(),
	}
}

func (m MsgBeginUnlockingAll) Route() string { return RouterKey }
func (m MsgBeginUnlockingAll) Type() string  { return TypeMsgBeginUnlockingAll }
func (m MsgBeginUnlockingAll) ValidateBasic() error {
	return nil
}
func (m MsgBeginUnlockingAll) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgBeginUnlockingAll) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgBeginUnlocking{}

// NewMsgBeginUnlocking creates a message to begin unlocking the tokens of a specific lock
func NewMsgBeginUnlocking(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgBeginUnlocking {
	return &MsgBeginUnlocking{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

func (m MsgBeginUnlocking) Route() string { return RouterKey }
func (m MsgBeginUnlocking) Type() string  { return TypeMsgBeginUnlocking }
func (m MsgBeginUnlocking) ValidateBasic() error {
	return nil
}
func (m MsgBeginUnlocking) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgBeginUnlocking) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}
