package stableswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	TypeMsgCreateStableswapPool = "create_stableswap_pool"
)

var _ sdk.Msg = &MsgCreateStableswapPool{}
var _ types.CreatePoolMsg = &MsgCreateStableswapPool{}

func NewMsgCreateStableswapPool(
	sender sdk.AccAddress,
	poolParams PoolParams,
	initialLiquidity sdk.Coins,
	futurePoolGovernor string) MsgCreateStableswapPool {
	return MsgCreateStableswapPool{
		Sender:               sender.String(),
		PoolParams:           &poolParams,
		InitialPoolLiquidity: initialLiquidity,
		FuturePoolGovernor:   futurePoolGovernor,
	}
}

func (msg MsgCreateStableswapPool) Route() string { return types.RouterKey }
func (msg MsgCreateStableswapPool) Type() string  { return TypeMsgCreateStableswapPool }
func (msg MsgCreateStableswapPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = msg.PoolParams.Validate()
	if err != nil {
		return err
	}

	// TODO: Add validation for pool initial liquidity

	// validation for future owner
	if err = types.ValidateFutureGovernor(msg.FuturePoolGovernor); err != nil {
		return err
	}

	return nil
}

func (msg MsgCreateStableswapPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateStableswapPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

/// Implement the CreatePoolMsg interface

func (msg MsgCreateStableswapPool) PoolCreator() sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return sender
}
func (msg MsgCreateStableswapPool) Validate(ctx sdk.Context) error {
	return msg.ValidateBasic()
}

func (msg MsgCreateStableswapPool) InitialLiquidity() sdk.Coins {
	return msg.InitialPoolLiquidity
}

func (msg MsgCreateStableswapPool) CreatePool(ctx sdk.Context, poolID uint64) (types.PoolI, error) {
	return &Pool{}, types.ErrNotImplemented
}
