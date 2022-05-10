package stableswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	TypeMsgCreateStableswapPool = "create_stableswap_pool"
)

var (
	_ sdk.Msg             = &MsgCreateStableswapPool{}
	_ types.CreatePoolMsg = &MsgCreateStableswapPool{}
)

func NewMsgCreateStableswapPool(
	sender sdk.AccAddress,
	poolParams PoolParams,
	poolAssets []PoolAsset,
	futurePoolGovernor string,
) MsgCreateStableswapPool {
	return MsgCreateStableswapPool{
		Sender:             sender.String(),
		PoolParams:         &poolParams,
		PoolAssets:         poolAssets,
		FuturePoolGovernor: futurePoolGovernor,
	}
}

func (msg MsgCreateStableswapPool) Route() string { return types.RouterKey }
func (msg MsgCreateStableswapPool) Type() string  { return TypeMsgCreateStableswapPool }
func (msg MsgCreateStableswapPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	if err := msg.PoolParams.Validate(); err != nil {
		return err
	}
	if err := ValidatePoolAssets(msg.PoolAssets); err != nil {
		return err
	}
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
	var coins sdk.Coins
	for _, asset := range msg.PoolAssets {
		coins = append(coins, asset.Token)
	}
	if coins == nil {
		panic("Shouldn't happen")
	}
	coins = coins.Sort()
	return coins
}

func (msg MsgCreateStableswapPool) CreatePool(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
	stableswapPool, err := NewStableswapPool(poolId, *msg.PoolParams, msg.PoolAssets, msg.FuturePoolGovernor)
	if err != nil {
		return nil, err
	}

	return stableswapPool, nil
}

func ValidatePoolAssets(poolAssets []PoolAsset) error {
	// validation for pool initial liquidity
	// TODO: expand this check to accommodate multi-asset pools for stableswap
	if len(poolAssets) < 2 {
		return types.ErrTooFewPoolAssets
	}
	if len(poolAssets) > 2 {
		return types.ErrTooManyPoolAssets
	}

	if err := validatePoolAssetsAgainstDuplicates(poolAssets); err != nil {
		return err
	}

	for _, pa := range poolAssets {
		if err := pa.Validate(); err != nil {
			return err
		}
	}
	return nil
}
