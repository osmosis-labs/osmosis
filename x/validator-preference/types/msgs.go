package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants
const (
	TypeMsgCreateValSetPreference = "create_validator_set_preference"
)

var _ sdk.Msg = &MsgCreateValidatorSetPreference{}

// NewMsgCreateValidatorSetPreference creates a msg to create a validator-set preference.
func NewMsgCreateValidatorSetPreference(delegator sdk.AccAddress, preferences []ValidatorPreference) *MsgCreateValidatorSetPreference {
	return &MsgCreateValidatorSetPreference{
		Delegator:   delegator.String(),
		Preferences: preferences,
	}
}

func (m MsgCreateValidatorSetPreference) Route() string { return RouterKey }
func (m MsgCreateValidatorSetPreference) Type() string  { return TypeMsgCreateValSetPreference }
func (m MsgCreateValidatorSetPreference) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegator address (%s)", err)
	}

	for _, validator := range m.Preferences {
		_, err := sdk.ValAddressFromBech32(validator.ValOperAddress)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator address (%s)", err)
		}
	}

	return nil
}

func (m MsgCreateValidatorSetPreference) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes a create validator-set message and returns the delegator in a byte array.
func (m MsgCreateValidatorSetPreference) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}

// constants
const (
	TypeMsgStakeToValidatorSet = "stake_to_validator_set"
)

var _ sdk.Msg = &MsgStakeToValidatorSet{}

// NewMsgMsgStakeToValidatorSet creates a msg to stake to a validator.
func NewMsgMsgStakeToValidatorSet(delegator sdk.AccAddress, coin sdk.Coin) *MsgStakeToValidatorSet {
	return &MsgStakeToValidatorSet{
		Delegator: delegator.String(),
		Coin:      coin,
	}
}

func (m MsgStakeToValidatorSet) Route() string { return RouterKey }
func (m MsgStakeToValidatorSet) Type() string  { return TypeMsgCreateValSetPreference }
func (m MsgStakeToValidatorSet) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (m MsgStakeToValidatorSet) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgStakeToValidatorSet) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}
