package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants
const (
	TypeMsgSuperfluidDelegate        = "superfluid_delegate"
	TypeMsgSuperfluidUndelegate      = "superfluid_undelegate"
	TypeMsgSuperfluidRedelegate      = "superfluid_redelegate"
	TypeMsgSuperfluidUnbondLock      = "superfluid_unbond_underlying_lock"
	TypeMsgLockAndSuperfluidDelegate = "lock_and_superfluid_delegate"
	TypeMsgUnPoolWhitelistedPool     = "unpool_whitelisted_pool"
)

var _ sdk.Msg = &MsgSuperfluidDelegate{}

// NewMsgSuperfluidDelegate creates a message to do superfluid delegation
func NewMsgSuperfluidDelegate(sender sdk.AccAddress, lockId uint64, valAddr sdk.ValAddress) *MsgSuperfluidDelegate {
	return &MsgSuperfluidDelegate{
		Sender:  sender.String(),
		LockId:  lockId,
		ValAddr: valAddr.String(),
	}
}

func (m MsgSuperfluidDelegate) Route() string { return RouterKey }
func (m MsgSuperfluidDelegate) Type() string  { return TypeMsgSuperfluidDelegate }
func (m MsgSuperfluidDelegate) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("sender should not be an empty address")
	}
	if m.LockId == 0 {
		return fmt.Errorf("lock id should be positive: %d < 0", m.LockId)
	}
	if m.ValAddr == "" {
		return fmt.Errorf("ValAddr should not be empty")
	}
	return nil
}
func (m MsgSuperfluidDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgSuperfluidDelegate) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSuperfluidUndelegate{}

// NewMsgSuperfluidUndelegate creates a message to do superfluid undelegation
func NewMsgSuperfluidUndelegate(sender sdk.AccAddress, lockId uint64) *MsgSuperfluidUndelegate {
	return &MsgSuperfluidUndelegate{
		Sender: sender.String(),
		LockId: lockId,
	}
}

func (m MsgSuperfluidUndelegate) Route() string { return RouterKey }
func (m MsgSuperfluidUndelegate) Type() string  { return TypeMsgSuperfluidUndelegate }
func (m MsgSuperfluidUndelegate) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("sender should not be an empty address")
	}
	if m.LockId == 0 {
		return fmt.Errorf("lock id should be positive: %d < 0", m.LockId)
	}
	return nil
}
func (m MsgSuperfluidUndelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgSuperfluidUndelegate) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

// var _ sdk.Msg = &MsgSuperfluidRedelegate{}

// // NewMsgSuperfluidRedelegate creates a message to do superfluid redelegation
// func NewMsgSuperfluidRedelegate(sender sdk.AccAddress, lockId uint64, newValAddr sdk.ValAddress) *MsgSuperfluidRedelegate {
// 	return &MsgSuperfluidRedelegate{
// 		Sender:     sender.String(),
// 		LockId:     lockId,
// 		NewValAddr: newValAddr.String(),
// 	}
// }

// func (m MsgSuperfluidRedelegate) Route() string { return RouterKey }
// func (m MsgSuperfluidRedelegate) Type() string  { return TypeMsgSuperfluidRedelegate }
// func (m MsgSuperfluidRedelegate) ValidateBasic() error {
// 	if m.Sender == "" {
// 		return fmt.Errorf("sender should not be an empty address")
// 	}
// 	if m.LockId == 0 {
// 		return fmt.Errorf("lock id should be positive: %d < 0", m.LockId)
// 	}
// 	if m.NewValAddr == "" {
// 		return fmt.Errorf("NewValAddr should not be empty")
// 	}
// 	return nil
// }
// func (m MsgSuperfluidRedelegate) GetSignBytes() []byte {
// 	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
// }
// func (m MsgSuperfluidRedelegate) GetSigners() []sdk.AccAddress {
// 	sender, _ := sdk.AccAddressFromBech32(m.Sender)
// 	return []sdk.AccAddress{sender}
// }

var _ sdk.Msg = &MsgSuperfluidUnbondLock{}

// MsgSuperfluidUnbondLock creates a message to unbond a lock underlying a superfluid undelegation position.
func NewMsgSuperfluidUnbondLock(sender sdk.AccAddress, lockID uint64) *MsgSuperfluidUnbondLock {
	return &MsgSuperfluidUnbondLock{
		Sender: sender.String(),
		LockId: lockID,
	}
}

func (m MsgSuperfluidUnbondLock) Route() string { return RouterKey }
func (m MsgSuperfluidUnbondLock) Type() string {
	return TypeMsgSuperfluidUnbondLock
}
func (m MsgSuperfluidUnbondLock) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("sender should not be an empty address")
	}
	if m.LockId == 0 {
		return fmt.Errorf("lockID should be set")
	}
	return nil
}
func (m MsgSuperfluidUnbondLock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgSuperfluidUnbondLock) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgLockAndSuperfluidDelegate{}

// NewMsgLockAndSuperfluidDelegate creates a message to create a lockup lock and superfluid delegation
func NewMsgLockAndSuperfluidDelegate(sender sdk.AccAddress, coins sdk.Coins, valAddr sdk.ValAddress) *MsgLockAndSuperfluidDelegate {
	return &MsgLockAndSuperfluidDelegate{
		Sender:  sender.String(),
		Coins:   coins,
		ValAddr: valAddr.String(),
	}
}

func (m MsgLockAndSuperfluidDelegate) Route() string { return RouterKey }
func (m MsgLockAndSuperfluidDelegate) Type() string  { return TypeMsgLockAndSuperfluidDelegate }
func (m MsgLockAndSuperfluidDelegate) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("sender should not be an empty address")
	}

	if m.Coins.Len() != 1 {
		return ErrMultipleCoinsLockupNotSupported
	}

	if m.ValAddr == "" {
		return fmt.Errorf("ValAddr should not be empty")
	}
	return nil
}
func (m MsgLockAndSuperfluidDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
func (m MsgLockAndSuperfluidDelegate) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgUnPoolWhitelistedPool{}

// NewMsgUnPoolWhitelistedPool creates a message to create a lockup lock and superfluid delegation
func NewMsgUnPoolWhitelistedPool(sender sdk.AccAddress, poolID uint64) *MsgUnPoolWhitelistedPool {
	return &MsgUnPoolWhitelistedPool{
		Sender: sender.String(),
		PoolId: poolID,
	}
}

func (msg MsgUnPoolWhitelistedPool) Route() string { return RouterKey }
func (msg MsgUnPoolWhitelistedPool) Type() string  { return TypeMsgUnPoolWhitelistedPool }
func (msg MsgUnPoolWhitelistedPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgUnPoolWhitelistedPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgUnPoolWhitelistedPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
