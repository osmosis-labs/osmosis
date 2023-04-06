package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants.
const (
	TypeMsgLockTokens        = "lock_tokens"
	TypeMsgBeginUnlockingAll = "begin_unlocking_all"
	TypeMsgBeginUnlocking    = "begin_unlocking"
	TypeMsgExtendLockup      = "edit_lockup"
	TypeForceUnlock          = "force_unlock"
)

var _ sdk.Msg = &MsgLockTokens{}

// NewMsgLockTokens creates a message to lock tokens.
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
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	if m.Duration <= 0 {
		return fmt.Errorf("duration should be positive: %d < 0", m.Duration)
	}

	// we only allow locks with one denom for now
	if m.Coins.Len() != 1 {
		return fmt.Errorf("lockups can only have one denom per lock ID, got %v", m.Coins)
	}

	if !m.Coins.IsAllPositive() {
		return fmt.Errorf("cannot lock up a zero or negative amount")
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

// NewMsgBeginUnlockingAll creates a message to begin unlocking tokens.
func NewMsgBeginUnlockingAll(owner sdk.AccAddress) *MsgBeginUnlockingAll {
	return &MsgBeginUnlockingAll{
		Owner: owner.String(),
	}
}

func (m MsgBeginUnlockingAll) Route() string { return RouterKey }
func (m MsgBeginUnlockingAll) Type() string  { return TypeMsgBeginUnlockingAll }
func (m MsgBeginUnlockingAll) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}
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

// NewMsgBeginUnlocking creates a message to begin unlocking the tokens of a specific lock.
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
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	if m.ID == 0 {
		return fmt.Errorf("invalid lockup ID, got %v", m.ID)
	}

	// only allow unlocks with a single denom or empty
	if m.Coins.Len() > 1 {
		return fmt.Errorf("can only unlock one denom per lock ID, got %v", m.Coins)
	}

	if !m.Coins.Empty() && !m.Coins.IsAllPositive() {
		return fmt.Errorf("cannot unlock a zero or negative amount")
	}

	return nil
}

func (m MsgBeginUnlocking) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgBeginUnlocking) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

// NewMsgExtendLockup creates a message to edit the properties of existing locks
func NewMsgExtendLockup(owner sdk.AccAddress, id uint64, duration time.Duration) *MsgExtendLockup {
	return &MsgExtendLockup{
		Owner:    owner.String(),
		ID:       id,
		Duration: duration,
	}
}

func (m MsgExtendLockup) Route() string { return RouterKey }
func (m MsgExtendLockup) Type() string  { return TypeMsgExtendLockup }
func (m MsgExtendLockup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}
	if m.ID == 0 {
		return fmt.Errorf("id is empty")
	}
	if m.Duration <= 0 {
		return fmt.Errorf("duration should be positive: %d < 0", m.Duration)
	}
	return nil
}

func (m MsgExtendLockup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON((&m)))
}

func (m MsgExtendLockup) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgForceUnlock{}

// NewMsgBeginUnlockingAll creates a message to begin unlocking tokens.
func NewMsgForceUnlock(owner sdk.AccAddress, id uint64, coins sdk.Coins) *MsgForceUnlock {
	return &MsgForceUnlock{
		Owner: owner.String(),
		ID:    id,
		Coins: coins,
	}
}

func (m MsgForceUnlock) Route() string { return RouterKey }
func (m MsgForceUnlock) Type() string  { return TypeMsgBeginUnlockingAll }
func (m MsgForceUnlock) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	if m.ID <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lock id should be bigger than 1 (%s)", err)
	}

	if !m.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, m.Coins.String())
	}
	return nil
}

func (m MsgForceUnlock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgForceUnlock) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}
