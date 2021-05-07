package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants
const (
	TypeMsgLockTokens            = "lock_tokens"
	TypeMsgBeginUnlocking        = "begin_unlock_tokens"
	TypeMsgUnlockTokens          = "unlock_tokens"
	TypeMsgBeginUnlockPeriodLock = "begin_unlock_period_lock"
	TypeMsgUnlockPeriodLock      = "unlock_period_lock"
)

var _ sdk.Msg = &MsgLockTokens{}

// NewMsgLockTokens creates a message to lock tokens
//nolint:interfacer
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
	return []sdk.AccAddress{m.Owner}
}

var _ sdk.Msg = &MsgBeginUnlocking{}

// NewMsgBeginUnlocking creates a message to begin unlocking tokens
//nolint:interfacer
func NewMsgBeginUnlocking(owner sdk.AccAddress) *MsgBeginUnlocking {
	return &MsgBeginUnlocking{
		Owner: owner.String(),
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
	return []sdk.AccAddress{m.Owner}
}

var _ sdk.Msg = &MsgUnlockTokens{}

// NewMsgUnlockTokens creates a message to begin unlocking all tokens of a user
//nolint:interfacer
func NewMsgUnlockTokens(owner sdk.AccAddress) *MsgUnlockTokens {
	return &MsgUnlockTokens{
		Owner: owner.String(),
	}
}

func (m MsgUnlockTokens) Route() string { return RouterKey }
func (m MsgUnlockTokens) Type() string  { return TypeMsgUnlockTokens }
func (m MsgUnlockTokens) ValidateBasic() error {
	return nil
}
func (m MsgUnlockTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgUnlockTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

var _ sdk.Msg = &MsgBeginUnlockPeriodLock{}

// NewMsgUnlockTokens creates a message to begin unlocking all tokens of a user
//nolint:interfacer
func NewMsgBeginUnlockPeriodLock(owner sdk.AccAddress, id uint64) *MsgBeginUnlockPeriodLock {
	return &MsgBeginUnlockPeriodLock{
		Owner: owner.String(),
		ID:    id,
	}
}

func (m MsgBeginUnlockPeriodLock) Route() string { return RouterKey }
func (m MsgBeginUnlockPeriodLock) Type() string  { return TypeMsgBeginUnlockPeriodLock }
func (m MsgBeginUnlockPeriodLock) ValidateBasic() error {
	return nil
}
func (m MsgBeginUnlockPeriodLock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgBeginUnlockPeriodLock) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

var _ sdk.Msg = &MsgUnlockPeriodLock{}

func (m MsgUnlockPeriodLock) Route() string { return RouterKey }
func (m MsgUnlockPeriodLock) Type() string  { return TypeMsgUnlockPeriodLock }
func (m MsgUnlockPeriodLock) ValidateBasic() error {
	return nil
}
func (m MsgUnlockPeriodLock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgUnlockPeriodLock) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}
