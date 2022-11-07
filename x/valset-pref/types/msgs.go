package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
)

// constants
const (
	TypeMsgSetValidatorSetPreference = "set_validator_set_preference"
)

var _ sdk.Msg = &MsgSetValidatorSetPreference{}

// NewMsgCreateValidatorSetPreference creates a msg to create a validator-set preference.
func NewMsgSetValidatorSetPreference(delegator sdk.AccAddress, preferences []ValidatorPreference) *MsgSetValidatorSetPreference {
	return &MsgSetValidatorSetPreference{
		Delegator:   delegator.String(),
		Preferences: preferences,
	}
}

func (m MsgSetValidatorSetPreference) Type() string { return TypeMsgSetValidatorSetPreference }
func (m MsgSetValidatorSetPreference) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegator address (%s)", err)
	}

	totalWeight := sdk.ZeroDec()
	validatorAddrs := []string{}
	for _, validator := range m.Preferences {
		_, err := sdk.ValAddressFromBech32(validator.ValOperAddress)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator address (%s)", err)
		}

		totalWeight = totalWeight.Add(validator.Weight)
		validatorAddrs = append(validatorAddrs, validator.ValOperAddress)
	}

	// check that all the validator address are unique
	containsDuplicate := osmoutils.ContainsDuplicate(validatorAddrs)
	if containsDuplicate {
		return fmt.Errorf("The validator operator address are duplicated")
	}

	// check if the total validator distribution weights equal 1
	if !totalWeight.Equal(sdk.OneDec()) {
		return fmt.Errorf("The weights allocated to the validators do not add up to 1, Got: %d", totalWeight)
	}

	return nil
}

func (m MsgSetValidatorSetPreference) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes a create validator-set message and returns the delegator in a byte array.
func (m MsgSetValidatorSetPreference) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}

// constants
const (
	TypeMsgDelegateToValidatorSet = "delegate_to_validator_set"
)

var _ sdk.Msg = &MsgDelegateToValidatorSet{}

// NewMsgMsgStakeToValidatorSet creates a msg to stake to a validator set.
func NewMsgMsgStakeToValidatorSet(delegator sdk.AccAddress, coin sdk.Coin) *MsgDelegateToValidatorSet {
	return &MsgDelegateToValidatorSet{
		Delegator: delegator.String(),
		Coin:      coin,
	}
}

func (m MsgDelegateToValidatorSet) Type() string { return TypeMsgDelegateToValidatorSet }
func (m MsgDelegateToValidatorSet) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if !m.Coin.IsValid() {
		return fmt.Errorf("The stake coin is not valid")
	}

	return nil
}

func (m MsgDelegateToValidatorSet) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgDelegateToValidatorSet) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}

// constants
const (
	TypeMsgUndelegateFromValidatorSet = "undelegate_from_validator_set"
)

var _ sdk.Msg = &MsgUndelegateFromValidatorSet{}

// NewMsgMsgStakeToValidatorSet creates a msg to stake to a validator.
func NewMsgUndelegateFromValidatorSet(delegator sdk.AccAddress, coin sdk.Coin) *MsgUndelegateFromValidatorSet {
	return &MsgUndelegateFromValidatorSet{
		Delegator: delegator.String(),
		Coin:      coin,
	}
}

func (m MsgUndelegateFromValidatorSet) Type() string { return TypeMsgUndelegateFromValidatorSet }
func (m MsgUndelegateFromValidatorSet) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if !m.Coin.IsValid() {
		return fmt.Errorf("The stake coin is not valid")
	}

	return nil
}

func (m MsgUndelegateFromValidatorSet) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgUndelegateFromValidatorSet) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}

// constants
const (
	TypeMsgWithdrawDelegationRewards = "withdraw_delegation_rewards"
)

var _ sdk.Msg = &MsgWithdrawDelegationRewards{}

// NewMsgMsgStakeToValidatorSet creates a msg to stake to a validator.
func NewMsgWithdrawDelegationRewards(delegator sdk.AccAddress) *MsgWithdrawDelegationRewards {
	return &MsgWithdrawDelegationRewards{
		Delegator: delegator.String(),
	}
}

func (m MsgWithdrawDelegationRewards) Type() string { return TypeMsgWithdrawDelegationRewards }
func (m MsgWithdrawDelegationRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (m MsgWithdrawDelegationRewards) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgWithdrawDelegationRewards) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(m.Delegator)
	return []sdk.AccAddress{delegator}
}
