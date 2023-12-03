package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

// Hooks is the wrapper struct for the txfees keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}
var _ gammtypes.GammHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

/* -------------------------------------------------------------------------- */
/*                                 epoch hooks                                */
/* -------------------------------------------------------------------------- */
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// at the end of each epoch, swap all non-DYM fees into DYM and burn them
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(txfeestypes.ModuleName)
	baseDenom, _ := k.GetBaseDenom(ctx)

	//get all balances of this module
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddr)

	//swap all to dym
	for _, coinBalance := range balances {
		if coinBalance.Denom == baseDenom {
			continue
		}
		if coinBalance.Amount.IsZero() {
			continue
		}

		feetoken, err := k.GetFeeToken(ctx, coinBalance.Denom)
		if err != nil {
			return err
		}

		// Do the swap of this fee token denom to base denom.
		route := []poolmanagertypes.SwapAmountInRoute{
			{
				PoolId:        feetoken.PoolID,
				TokenOutDenom: baseDenom,
			},
		}
		_, err = k.poolManager.RouteExactAmountIn(ctx, moduleAddr, route, coinBalance, sdk.ZeroInt())
		if err != nil {
			return err
		}
	}

	// Get all of the txfee payout denom in the module account
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, moduleAddr, baseDenom))
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, baseDenomCoins)
	if err != nil {
		return err
	}

	return nil
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

/* -------------------------------------------------------------------------- */
/*                                 pool hooks                                 */
/* -------------------------------------------------------------------------- */

// AfterPoolCreated creates a gauge for each pool’s lockable duration.
func (h Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	//check if base denom included in the pool
	baseDenom, err := h.k.GetBaseDenom(ctx)
	if err != nil {
		h.k.Logger(ctx).Error("failed to get base denom", "error", err)
		return
	}
	denoms, err := h.k.spotPriceCalculator.GetPoolDenoms(ctx, poolId)
	if err != nil {
		h.k.Logger(ctx).Error("failed to get pool denoms", "error", err)
		return
	}

	if len(denoms) != 2 {
		h.k.Logger(ctx).Debug("expecting pools of 2 assets", "denoms", denoms)
		return
	}
	if !contains(denoms, baseDenom) {
		h.k.Logger(ctx).Debug("base denom not included in the pool. skipping", "baseDenom", baseDenom, "denoms", denoms)
		return
	}

	//get the non-native denom
	var nonNativeDenom string
	if denoms[0] == baseDenom {
		nonNativeDenom = denoms[1]
	} else {
		nonNativeDenom = denoms[0]
	}

	_, err = h.k.GetFeeToken(ctx, nonNativeDenom)
	if err == nil {
		h.k.Logger(ctx).Error("fee token already exists", "denom", nonNativeDenom)
		return
	}

	feeToken := txfeestypes.FeeToken{
		PoolID: poolId,
		Denom:  nonNativeDenom,
	}
	err = h.k.setFeeToken(ctx, feeToken)
	if err != nil {
		h.k.Logger(ctx).Error("failed to set fee token", "error", err)
		return
	}
}

// AfterJoinPool hook is a noop.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
}

// AfterExitPool hook is a noop.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
}

// AfterSwap hook is a noop.
func (h Hooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
}

func contains(strarr []string, str string) bool {
	for _, v := range strarr {
		if v == str {
			return true
		}
	}

	return false
}
