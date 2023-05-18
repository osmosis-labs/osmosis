package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants.
const (
	TypeMsgCreatePosition          = "create-position"
	TypeAddToPosition              = "add-to-position"
	TypeMsgWithdrawPosition        = "withdraw-position"
	TypeMsgCollectFees             = "collect-fees"
	TypeMsgCollectIncentives       = "collect-incentives"
	TypeMsgCreateIncentive         = "create-incentive"
	TypeMsgFungifyChargedPositions = "fungify-charged-positions"
)

var _ sdk.Msg = &MsgCreatePosition{}

func (msg MsgCreatePosition) Route() string { return RouterKey }
func (msg MsgCreatePosition) Type() string  { return TypeMsgCreatePosition }
func (msg MsgCreatePosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if msg.LowerTick >= msg.UpperTick {
		return InvalidLowerUpperTickError{LowerTick: msg.LowerTick, UpperTick: msg.UpperTick}
	}

	if msg.TokensProvided.Empty() {
		return fmt.Errorf("Empty coins provided (%s)", msg.TokensProvided.String())
	}

	if !msg.TokensProvided.IsValid() {
		return fmt.Errorf("Invalid coins (%s)", msg.TokensProvided.String())
	}

	if len(msg.TokensProvided) > 2 {
		return CoinLengthError{Length: len(msg.TokensProvided), MaxLength: 2}
	}

	for _, coin := range msg.TokensProvided {
		if coin.Amount.LTE(sdk.ZeroInt()) {
			return NotPositiveRequireAmountError{Amount: coin.Amount.String()}
		}
	}

	if msg.TokenMinAmount0.IsNegative() {
		return NotPositiveRequireAmountError{Amount: msg.TokenMinAmount0.String()}
	}

	if msg.TokenMinAmount1.IsNegative() {
		return NotPositiveRequireAmountError{Amount: msg.TokenMinAmount1.String()}
	}

	return nil
}

func (msg MsgCreatePosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgAddToPosition{}

func (msg MsgAddToPosition) Route() string { return RouterKey }
func (msg MsgAddToPosition) Type() string  { return TypeAddToPosition }
func (msg MsgAddToPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if msg.PositionId <= 0 {
		return fmt.Errorf("Invalid position id (%s)", strconv.FormatUint(msg.PositionId, 10))
	}

	if !msg.TokenDesired0.IsValid() || !msg.TokenDesired0.IsPositive() {
		return fmt.Errorf("Invalid coins (%s)", msg.TokenDesired0.String())
	}

	if !msg.TokenDesired1.IsValid() || !msg.TokenDesired1.IsPositive() {
		return fmt.Errorf("Invalid coins (%s)", msg.TokenDesired1.String())
	}

	return nil
}

func (msg MsgAddToPosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAddToPosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgWithdrawPosition{}

func (msg MsgWithdrawPosition) Route() string { return RouterKey }
func (msg MsgWithdrawPosition) Type() string  { return TypeMsgWithdrawPosition }
func (msg MsgWithdrawPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if !msg.LiquidityAmount.IsPositive() {
		return NotPositiveRequireAmountError{Amount: msg.LiquidityAmount.String()}
	}

	return nil
}

func (msg MsgWithdrawPosition) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgWithdrawPosition) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgCollectFees{}

func (msg MsgCollectFees) Route() string { return RouterKey }
func (msg MsgCollectFees) Type() string  { return TypeMsgCollectFees }
func (msg MsgCollectFees) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgCollectFees) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCollectFees) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgCollectIncentives{}

func (msg MsgCollectIncentives) Route() string { return RouterKey }
func (msg MsgCollectIncentives) Type() string  { return TypeMsgCollectIncentives }
func (msg MsgCollectIncentives) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgCollectIncentives) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCollectIncentives) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgCreateIncentive{}

func (msg MsgCreateIncentive) Route() string { return RouterKey }
func (msg MsgCreateIncentive) Type() string  { return TypeMsgCreateIncentive }
func (msg MsgCreateIncentive) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if !msg.IncentiveCoin.IsValid() {
		return InvalidIncentiveCoinError{PoolId: msg.PoolId, IncentiveCoin: msg.IncentiveCoin}
	}

	if !msg.EmissionRate.IsPositive() {
		return NonPositiveEmissionRateError{PoolId: msg.PoolId, EmissionRate: msg.EmissionRate}
	}

	return nil
}

func (msg MsgCreateIncentive) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateIncentive) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgFungifyChargedPositions{}

func (msg MsgFungifyChargedPositions) Route() string { return RouterKey }
func (msg MsgFungifyChargedPositions) Type() string  { return TypeMsgFungifyChargedPositions }
func (msg MsgFungifyChargedPositions) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if len(msg.PositionIds) < 2 {
		return fmt.Errorf("Must provide at least 2 positions, got %d", len(msg.PositionIds))
	}

	return nil
}

func (msg MsgFungifyChargedPositions) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgFungifyChargedPositions) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
