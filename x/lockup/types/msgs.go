package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants
const (
	TypeMsgLockTokens            = "lock_tokens"
	TypeMsgBeginUnlockTokens     = "begin_unlock_tokens"
	TypeMsgUnlockTokens          = "unlock_tokens"
	TypeMsgBeginUnlockPeriodLock = "begin_unlock_period_lock"
	TypeMsgUnlockPeriodLock      = "unlock_period_lock"
)

var _ sdk.Msg = &MsgLockTokens{}

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

var _ sdk.Msg = &MsgBeginUnlockTokens{}

func (m MsgBeginUnlockTokens) Route() string { return RouterKey }
func (m MsgBeginUnlockTokens) Type() string  { return TypeMsgBeginUnlockTokens }
func (m MsgBeginUnlockTokens) ValidateBasic() error {
	return nil
}
func (m MsgBeginUnlockTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgBeginUnlockTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

var _ sdk.Msg = &MsgUnlockTokens{}

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
