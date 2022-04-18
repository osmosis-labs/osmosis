package balancer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	TypeMsgCreateBalancerPool = "create_balancer_pool"
)

var _ sdk.Msg = &MsgCreateBalancerPool{}

func (msg MsgCreateBalancerPool) Route() string { return types.RouterKey }
func (msg MsgCreateBalancerPool) Type() string  { return TypeMsgCreateBalancerPool }
func (msg MsgCreateBalancerPool) ValidateBasic() error {

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = types.ValidateUserSpecifiedPoolAssets(msg.PoolAssets)
	if err != nil {
		return err
	}

	err = msg.PoolParams.Validate(msg.PoolAssets)
	if err != nil {
		return err
	}

	// validation for future owner
	if err = ValidateFutureGovernor(msg.FuturePoolGovernor); err != nil {
		return err
	}

	return nil
}
func (msg MsgCreateBalancerPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgCreateBalancerPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
