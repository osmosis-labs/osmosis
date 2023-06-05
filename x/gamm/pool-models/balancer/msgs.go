package balancer

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

const (
	TypeMsgCreateBalancerPool = "create_balancer_pool"
)

var (
	_ sdk.Msg                        = &MsgCreateBalancerPool{}
	_ poolmanagertypes.CreatePoolMsg = &MsgCreateBalancerPool{}
)

func NewMsgCreateBalancerPool(
	sender sdk.AccAddress,
	poolParams PoolParams,
	poolAssets []PoolAsset,
	futurePoolGovernor string,
) MsgCreateBalancerPool {
	return MsgCreateBalancerPool{
		Sender:             sender.String(),
		PoolParams:         &poolParams,
		PoolAssets:         poolAssets,
		FuturePoolGovernor: futurePoolGovernor,
	}
}

func (msg MsgCreateBalancerPool) Route() string { return types.RouterKey }
func (msg MsgCreateBalancerPool) Type() string  { return TypeMsgCreateBalancerPool }
func (msg MsgCreateBalancerPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = validateUserSpecifiedPoolAssets(msg.PoolAssets)
	if err != nil {
		return err
	}

	err = msg.PoolParams.Validate(msg.PoolAssets)
	if err != nil {
		return err
	}

	// validation for future owner
	if err = types.ValidateFutureGovernor(msg.FuturePoolGovernor); err != nil {
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

/// Implement the CreatePoolMsg interface

func (msg MsgCreateBalancerPool) PoolCreator() sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return sender
}

func (msg MsgCreateBalancerPool) Validate(ctx sdk.Context) error {
	return msg.ValidateBasic()
}

func (msg MsgCreateBalancerPool) InitialLiquidity() sdk.Coins {
	var coins sdk.Coins
	for _, asset := range msg.PoolAssets {
		coins = append(coins, asset.Token)
	}
	if coins == nil {
		panic("InitialLiquidity coins is equal to nil - this shouldn't happen")
	}
	coins = coins.Sort()
	return coins
}

func (msg MsgCreateBalancerPool) CreatePool(ctx sdk.Context, poolID uint64) (poolmanagertypes.PoolI, error) {
	poolI, err := NewBalancerPool(poolID, *msg.PoolParams, msg.PoolAssets, msg.FuturePoolGovernor, ctx.BlockTime())
	return &poolI, err
}

func (msg MsgCreateBalancerPool) GetPoolType() poolmanagertypes.PoolType {
	return poolmanagertypes.Balancer
}
