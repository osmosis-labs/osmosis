package stableswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

const (
	TypeMsgCreateStableswapPool           = "create_stableswap_pool"
	TypeMsgStableSwapAdjustScalingFactors = "stable_swap_adjust_scaling_factors"
)

var (
	_ sdk.Msg             = &MsgCreateStableswapPool{}
	_ types.CreatePoolMsg = &MsgCreateStableswapPool{}
)

func NewMsgCreateStableswapPool(
	sender sdk.AccAddress,
	poolParams PoolParams,
	initialLiquidity sdk.Coins,
	scalingFactors []uint64,
	futurePoolGovernor string,
) MsgCreateStableswapPool {
	return MsgCreateStableswapPool{
		Sender:               sender.String(),
		PoolParams:           &poolParams,
		InitialPoolLiquidity: initialLiquidity,
		ScalingFactors:       scalingFactors,
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

	// validation for pool initial liquidity
	// TO DO: expand this check to accommodate multi-asset pools for stableswap
	if len(msg.InitialPoolLiquidity) < 2 {
		return types.ErrTooFewPoolAssets
	} else if len(msg.InitialPoolLiquidity) > 2 {
		return types.ErrTooManyPoolAssets
	}
	// valid scaling factor lengths are 0, or one factor for each asset
	if len(msg.ScalingFactors) != 0 && len(msg.ScalingFactors) != len(msg.InitialPoolLiquidity) {
		return types.ErrInvalidScalingFactors
	}

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

func (msg MsgCreateStableswapPool) CreatePool(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
	stableswapPool, err := NewStableswapPool(poolId, *msg.PoolParams, msg.InitialPoolLiquidity, msg.ScalingFactors, msg.FuturePoolGovernor)
	if err != nil {
		return nil, err
	}

	return &stableswapPool, nil
}

var _ sdk.Msg = &MsgStableSwapAdjustScalingFactors{}

// Implement sdk.Msg
func NewMsgStableSwapAdjustScalingFactors(
	sender string,
	poolID uint64,
) MsgStableSwapAdjustScalingFactors {
	return MsgStableSwapAdjustScalingFactors{
		Sender: sender,
		PoolID: poolID,
	}
}

func (msg MsgStableSwapAdjustScalingFactors) Route() string {
	return types.RouterKey
}

func (msg MsgStableSwapAdjustScalingFactors) Type() string { return TypeMsgCreateStableswapPool }
func (msg MsgStableSwapAdjustScalingFactors) ValidateBasic() error {
	if msg.Sender == "" {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgStableSwapAdjustScalingFactors) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStableSwapAdjustScalingFactors) GetSigners() []sdk.AccAddress {
	scalingFactorGovernor, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{scalingFactorGovernor}
}
