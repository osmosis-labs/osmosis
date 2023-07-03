package stableswap

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

const (
	TypeMsgCreateStableswapPool           = "create_stableswap_pool"
	TypeMsgStableSwapAdjustScalingFactors = "stable_swap_adjust_scaling_factors"
)

var (
	_ sdk.Msg                        = &MsgCreateStableswapPool{}
	_ poolmanagertypes.CreatePoolMsg = &MsgCreateStableswapPool{}
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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = msg.PoolParams.Validate()
	if err != nil {
		return err
	}

	// validation for scaling factors
	scalingFactors := msg.ScalingFactors
	// The message's scaling factors must be empty or a valid set of scaling factors
	if len(scalingFactors) != 0 {
		if err = validateScalingFactors(scalingFactors, len(msg.InitialPoolLiquidity)); err != nil {
			return err
		}
	} else {
		for i := 0; i < len(msg.InitialPoolLiquidity); i += 1 {
			scalingFactors = append(scalingFactors, 1)
		}
	}

	// validation for pool initial liquidity
	// The message's pool liquidity must have between 2 and 8 assets with at most 10B post-scaled units in each
	if err = validatePoolLiquidity(msg.InitialPoolLiquidity, scalingFactors); err != nil {
		return err
	}

	// validation for scaling factor owner
	if err = validateScalingFactorController(msg.ScalingFactorController); err != nil {
		return err
	}

	// validation for future governor
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

func (msg MsgCreateStableswapPool) CreatePool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error) {
	stableswapPool, err := NewStableswapPool(poolId, *msg.PoolParams, msg.InitialPoolLiquidity,
		msg.ScalingFactors, msg.ScalingFactorController, msg.FuturePoolGovernor)
	if err != nil {
		return nil, err
	}

	return &stableswapPool, nil
}

func (msg MsgCreateStableswapPool) GetPoolType() poolmanagertypes.PoolType {
	return poolmanagertypes.Stableswap
}

var _ sdk.Msg = &MsgStableSwapAdjustScalingFactors{}

// Implement sdk.Msg
func NewMsgStableSwapAdjustScalingFactors(
	sender string,
	poolID uint64,
	scalingFactors []uint64,
) MsgStableSwapAdjustScalingFactors {
	return MsgStableSwapAdjustScalingFactors{
		Sender:         sender,
		PoolID:         poolID,
		ScalingFactors: scalingFactors,
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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
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
