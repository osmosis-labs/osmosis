package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	nonNativeFeeAddr := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorName)
	baseDenom, _ := k.GetBaseDenom(ctx)

	//get all balances of this module
	balances := sdk.Coins{}

	//swap all to dym
	for _, coinBalance := range balances {
		if coinBalance.Denom == baseDenom {
			continue
		}
		if coinBalance.Amount.IsZero() {
			continue
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			minAmountOut := sdk.ZeroInt()
			feetoken, err := k.GetFeeToken(ctx, coinBalance.Denom)
			if err != nil {
				return err
			}

			_, err = k.poolManager.SwapExactAmountIn(cacheCtx, nonNativeFeeAddr, feetoken.PoolID, coinBalance, baseDenom, minAmountOut)
			return err
		})
	}

	// Get all of the txfee payout denom in the module account
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeFeeAddr, baseDenom))

	//TODO: BURN HERE!!!

	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorName, txfeestypes.FeeCollectorName, baseDenomCoins)
		return err
	})

	return nil
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

//TODO: add pool hooks to register as accepted fee tokens
/*
func (k Keeper) HandleUpdateFeeTokenProposal(ctx sdk.Context, p *types.UpdateFeeTokenProposal) error {
	// setFeeToken internally calls ValidateFeeToken
	return k.setFeeToken(ctx, p.Feetoken)
}
*/
