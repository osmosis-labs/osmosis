package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSetDenomMetadata = "set_contract_param"

var _ sdk.Msg = &MsgSetContractParam{}

// NewMsgSetContractParam creates a message to create a lockup lock and superfluid delegation
func NewMsgSetContractParam(address sdk.AccAddress) *MsgSetContractParam {
	return &MsgSetContractParam{
		Address: address.String(),
	}
}

func (msg MsgSetContractParam) Route() string { return RouterKey }
func (msg MsgSetContractParam) Type() string  { return TypeMsgSetDenomMetadata }
func (msg MsgSetContractParam) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return fmt.Errorf("invalid address (%s)", msg.Address)
	}

	return nil
}

func (msg MsgSetContractParam) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSetContractParam) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
