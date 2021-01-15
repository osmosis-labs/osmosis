package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/c-osmosis/osmosis/x/gamm/types"
)

func (k Keeper) UpdateSwapFee(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	newSwapFee sdk.Dec,
) (err error) {
	poolAcc, err := k.GetPool(ctx, poolId)
	if err != nil {
		return err
	}

	params := poolAcc.GetPoolParams()

	swapFeeGovernor, err := sdk.AccAddressFromBech32(params.SwapFeeGovernor)
	if err != nil {
		return err
	}

	if !sender.Equals(swapFeeGovernor) {
		return sdkerrors.Wrapf(types.ErrUnauthorizedGovernor, "unauthorized to change swapfee")
	}

	params.SwapFee = newSwapFee
	poolAcc.SetPoolParams(params)

	err = k.SetPool(ctx, poolAcc)
	if err != nil {
		return err
	}

	return nil
}
