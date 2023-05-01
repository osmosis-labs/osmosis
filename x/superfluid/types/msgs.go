package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants.
const (
	TypeMsgSuperfluidDelegate                           = "superfluid_delegate"
	TypeMsgSuperfluidUndelegate                         = "superfluid_undelegate"
	TypeMsgSuperfluidRedelegate                         = "superfluid_redelegate"
	TypeMsgSuperfluidUnbondLock                         = "superfluid_unbond_underlying_lock"
	TypeMsgSuperfluidUndeledgateAndUnbondLock           = "superfluid_undelegate_and_unbond_lock"
	TypeMsgLockAndSuperfluidDelegate                    = "lock_and_superfluid_delegate"
	TypeMsgUnPoolWhitelistedPool                        = "unpool_whitelisted_pool"
	TypeMsgUnlockAndMigrateShares                       = "unlock_and_migrate_shares"
	TypeMsgCreateFullRangePositionAndSuperfluidDelegate = "create_full_range_position_and_delegate"
)

var _ sdk.Msg = &MsgSuperfluidDelegate{}

// NewMsgSuperfluidDelegate creates a message to do superfluid delegation.
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

// NewMsgSuperfluidUndelegate creates a message to do superfluid undelegation.
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

var _ sdk.Msg = &MsgSuperfluidUndelegateAndUnbondLock{}

// MsgSuperfluidUndelegateAndUnbondLock creates a message to unbond a lock underlying a superfluid undelegation position.
// Amount to unbond can be less than or equal to the amount locked.
func NewMsgSuperfluidUndelegateAndUnbondLock(sender sdk.AccAddress, lockID uint64, coin sdk.Coin) *MsgSuperfluidUndelegateAndUnbondLock {
	return &MsgSuperfluidUndelegateAndUnbondLock{
		Sender: sender.String(),
		LockId: lockID,
		Coin:   coin,
	}
}

func (m MsgSuperfluidUndelegateAndUnbondLock) Route() string { return RouterKey }
func (m MsgSuperfluidUndelegateAndUnbondLock) Type() string {
	return TypeMsgSuperfluidUndeledgateAndUnbondLock
}

func (m MsgSuperfluidUndelegateAndUnbondLock) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("sender should not be an empty address")
	}
	if m.LockId == 0 {
		return fmt.Errorf("lockID should be set")
	}
	if !m.Coin.IsValid() {
		return fmt.Errorf("cannot unlock a zero or negative amount")
	}

	return nil
}

func (m MsgSuperfluidUndelegateAndUnbondLock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgSuperfluidUndelegateAndUnbondLock) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgLockAndSuperfluidDelegate{}

// NewMsgLockAndSuperfluidDelegate creates a message to create a lockup lock and superfluid delegation.
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

var _ sdk.Msg = &MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition{}

func NewMsgUnlockAndMigrateSharesToFullRangeConcentratedPosition(sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) *MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition {
	return &MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition{
		Sender:          sender.String(),
		LockId:          lockId,
		SharesToMigrate: sharesToMigrate,
	}
}

func (msg MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) Route() string { return RouterKey }
func (msg MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) Type() string {
	return TypeMsgUnlockAndMigrateShares
}
func (msg MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	if msg.LockId <= 0 {
		return fmt.Errorf("Invalid lock ID (%d)", msg.LockId)
	}
	if msg.SharesToMigrate.IsNegative() {
		return fmt.Errorf("Invalid shares to migrate (%s)", msg.SharesToMigrate)
	}
	return nil
}

func (msg MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgCreateFullRangePositionAndSuperfluidDelegate{}

func NewMsgCreateFullRangePositionAndSuperfluidDelegate(sender sdk.AccAddress, coins sdk.Coins, valAddr string, poolId uint64) *MsgCreateFullRangePositionAndSuperfluidDelegate {
	return &MsgCreateFullRangePositionAndSuperfluidDelegate{
		Sender:  sender.String(),
		Coins:   coins,
		ValAddr: valAddr,
		PoolId:  poolId,
	}
}

func (msg MsgCreateFullRangePositionAndSuperfluidDelegate) Route() string { return RouterKey }
func (msg MsgCreateFullRangePositionAndSuperfluidDelegate) Type() string {
	return TypeMsgUnlockAndMigrateShares
}
func (msg MsgCreateFullRangePositionAndSuperfluidDelegate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = msg.Coins.Validate()
	if err != nil {
		return err
	}

	if msg.ValAddr == "" {
		return fmt.Errorf("ValAddr should not be empty")
	}

	if msg.PoolId < 1 {
		return fmt.Errorf("pool id must be positive")
	}
	return nil
}

func (msg MsgCreateFullRangePositionAndSuperfluidDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateFullRangePositionAndSuperfluidDelegate) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
