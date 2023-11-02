package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRemoveInterchainQueryRequest{}

func NewMsgRemoveInterchainQuery(sender string, queryID uint64) MsgRemoveInterchainQueryRequest {
	return MsgRemoveInterchainQueryRequest{
		QueryId: queryID,
		Sender:  sender,
	}
}

func (msg MsgRemoveInterchainQueryRequest) Route() string {
	return RouterKey
}

func (msg MsgRemoveInterchainQueryRequest) Type() string {
	return "remove-interchain-query"
}

func (msg MsgRemoveInterchainQueryRequest) ValidateBasic() error {
	if msg.GetQueryId() == 0 {
		return sdkerrors.Wrap(ErrInvalidQueryID, "query_id cannot be empty or equal to 0")
	}

	if strings.TrimSpace(msg.Sender) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	return nil
}

func (msg MsgRemoveInterchainQueryRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRemoveInterchainQueryRequest) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
