package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// constants.
const (
	TypeMsgCreatePosition          = "create-position"
	TypeAddToPosition              = "add-to-position"
	TypeMsgWithdrawPosition        = "withdraw-position"
	TypeMsgCollectSpreadRewards    = "collect-spread-rewards"
	TypeMsgCollectIncentives       = "collect-incentives"
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
		if coin.Amount.LTE(osmomath.ZeroInt()) {
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

	if msg.Amount0.IsNegative() {
		return fmt.Errorf("Amount 0 cannot be negative, given amount: %s", msg.Amount0.String())
	}
	if msg.Amount1.IsNegative() {
		return fmt.Errorf("Amount 1 cannot be negative, given amount: %s", msg.Amount1.String())
	}
	if msg.TokenMinAmount0.IsNegative() {
		return fmt.Errorf("Amount 0 cannot be negative, given token min amount: %s", msg.TokenMinAmount0.String())
	}
	if msg.TokenMinAmount1.IsNegative() {
		return fmt.Errorf("Amount 1 cannot be negative, given token min amount: %s", msg.TokenMinAmount1.String())
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

var _ sdk.Msg = &MsgCollectSpreadRewards{}

func (msg MsgCollectSpreadRewards) Route() string { return RouterKey }
func (msg MsgCollectSpreadRewards) Type() string  { return TypeMsgCollectSpreadRewards }
func (msg MsgCollectSpreadRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgCollectSpreadRewards) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCollectSpreadRewards) GetSigners() []sdk.AccAddress {
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
