package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgAllocateAssets  = "allocate_assets"
	TypeMsgWithdrawRewards = "withdraw_rewards"
)

var _ sdk.Msg = &MsgAllocateAssets{}

func NewMsgAllocateAssets(from sdk.AccAddress, farmId uint64, assets sdk.Coins) *MsgAllocateAssets {
	return &MsgAllocateAssets{
		FromAddress: from.String(),
		FarmId:      farmId,
		Assets:      assets,
	}
}

func (msg MsgAllocateAssets) Route() string { return RouterKey }

func (msg MsgAllocateAssets) Type() string { return TypeMsgAllocateAssets }

func (msg MsgAllocateAssets) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if !msg.Assets.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Assets.String())
	}

	if !msg.Assets.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Assets.String())
	}

	return nil
}

func (msg MsgAllocateAssets) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAllocateAssets) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

var _ sdk.Msg = &MsgWithdrawRewards{}

func NewMsgWithdrawRewards(from sdk.AccAddress, farmId uint64) *MsgWithdrawRewards {
	return &MsgWithdrawRewards{
		FromAddress: from.String(),
		FarmId:      farmId,
	}
}

func (msg MsgWithdrawRewards) Route() string { return RouterKey }

func (msg MsgWithdrawRewards) Type() string { return TypeMsgWithdrawRewards }

func (msg MsgWithdrawRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgWithdrawRewards) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgWithdrawRewards) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
