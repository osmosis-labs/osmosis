package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants.
const (
	TypeMsgSetFeeTokens = "set-fee-tokens"
)

var _ sdk.Msg = &MsgSetFeeTokens{}

func (msg MsgSetFeeTokens) Route() string { return RouterKey }
func (msg MsgSetFeeTokens) Type() string  { return TypeMsgSetFeeTokens }
func (msg MsgSetFeeTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}

	if len(msg.FeeTokens) == 0 {
		return fmt.Errorf("Fee tokens must not be empty")
	}

	return nil
}

func (msg MsgSetFeeTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSetFeeTokens) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
