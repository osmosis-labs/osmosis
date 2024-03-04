package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/osmomath"
)

var _ sdk.Msg = &MsgInboundTransfer{}

// NewMsgInboundTransfer creates a message to transfer tokens from the source chain
func NewMsgInboundTransfer(
	sender string,
	sourceChain Chain,
	destinationAddress string,
	subdenom string,
	amount osmomath.Dec,
) *MsgInboundTransfer {
	return &MsgInboundTransfer{
		Sender:             sender,
		SourceChain:        sourceChain,
		DestinationAddress: destinationAddress,
		Subdenom:           subdenom,
		Amount:             amount,
	}
}

func (m MsgInboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.DestinationAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.SourceChain.ValidateBasic()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidChain, err.Error())
	}

	if len(m.Subdenom) == 0 {
		return errorsmod.Wrap(ErrInvalidSubdenom, "Subdenom is empty")
	}

	if !m.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgInboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgOutboundTransfer{}

// NewMsgInboundTransfer creates a message to transfer tokens from the source chain
func NewMsgOutboundTransfer(
	sender string,
	destinationChain Chain,
	destinationAddress string,
	amount sdk.Coin,
) *MsgOutboundTransfer {
	return &MsgOutboundTransfer{
		Sender:             sender,
		DestinationChain:   destinationChain,
		DestinationAddress: destinationAddress,
		Amount:             amount,
	}
}

func (m MsgOutboundTransfer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.DestinationAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid destination address (%s)", err)
	}

	err = m.DestinationChain.ValidateBasic()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidChain, err.Error())
	}

	if !m.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgOutboundTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

func (m Chain) ValidateBasic() error {
	switch m {
	case Chain_CHAIN_BITCOIN,
		Chain_CHAIN_OSMOSIS:
	case Chain_CHAIN_UNSPECIFIED:
		return fmt.Errorf("invalid chain: %s", m)
	default:
		return fmt.Errorf("unknown chain: %s", m)
	}
	return nil
}
