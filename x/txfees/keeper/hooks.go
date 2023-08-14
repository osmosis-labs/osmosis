package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v17/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	baseDenom, _ := k.GetBaseDenom(ctx)

	nonNativeFeeCollectorForStakingRewardsAddr := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorForStakingRewardsName)
	nonNativeFeeCollectorForStakingRewardsBalance := k.bankKeeper.GetAllBalances(ctx, nonNativeFeeCollectorForStakingRewardsAddr)

	// Non-native fee collector for staking rewards get swapped entirely into base denom.
	for _, coin := range nonNativeFeeCollectorForStakingRewardsBalance {
		if coin.Denom == baseDenom {
			continue
		}

		coinBalance := k.bankKeeper.GetBalance(ctx, nonNativeFeeCollectorForStakingRewardsAddr, coin.Denom)
		if coinBalance.Amount.IsZero() {
			continue
		}

		poolId, err := k.protorevKeeper.GetPoolForDenomPair(ctx, baseDenom, coin.Denom)
		if err != nil {
			// The pool route either doesn't exist or is disabled in protorev.
			// It will just accrue in the non-native fee collector account.
			// Skip this denom and move on to the next one.
			continue
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. Theres not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := sdk.ZeroInt()

			// We swap without charging a taker fee / sending to the non native fee collector, since these are funds that
			// are accruing from the taker fee itself.
			_, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, nonNativeFeeCollectorForStakingRewardsAddr, poolId, coinBalance, baseDenom, minAmountOut)
			return err
		})
	}

	// Now that the rewards have been swapped, transfer any base denom existing in the non-native fee collector to the fee collector (indirectly distributing to stakers)
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeFeeCollectorForStakingRewardsAddr, baseDenom))
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorForStakingRewardsName, txfeestypes.FeeCollectorName, baseDenomCoins)
		return err
	})

	// Non-native fee collector for community pool get swapped entirely into denom specified in the pool manager params.

	poolManagerParams := k.poolManager.GetParams(ctx)
	denomToSwapTo := poolManagerParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo

	nonNativeFeeCollectorForCommunityPoolAddr := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorForCommunityPoolName)
	nonNativeFeeCollectorForCommunityPoolBalance := k.bankKeeper.GetAllBalances(ctx, nonNativeFeeCollectorForCommunityPoolAddr)

	// Only non whitelisted assets should exist here since we do direct community pool funds when calculating the taker fee if
	// the input is a whitelisted asset.
	for _, coin := range nonNativeFeeCollectorForCommunityPoolBalance {
		if coin.Denom == denomToSwapTo {
			continue
		}

		coinBalance := k.bankKeeper.GetBalance(ctx, nonNativeFeeCollectorForCommunityPoolAddr, coin.Denom)
		if coinBalance.Amount.IsZero() {
			continue
		}

		poolId, err := k.protorevKeeper.GetPoolForDenomPair(ctx, denomToSwapTo, coin.Denom)
		if err != nil {
			// The pool route either doesn't exist or is disabled in protorev.
			// It will just accrue in the non-native fee collector account.
			// Skip this denom and move on to the next one.
			continue
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. Theres not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := sdk.ZeroInt()

			// We swap without charging a taker fee / sending to the non native fee collector, since these are funds that
			// are accruing from the taker fee itself.
			_, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, nonNativeFeeCollectorForCommunityPoolAddr, poolId, coinBalance, denomToSwapTo, minAmountOut)
			return err
		})
	}

	// Now that the non whitelisted assets have been swapped, fund the community pool with the denom we swapped to.
	denomToSwapToCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeFeeCollectorForCommunityPoolAddr, denomToSwapTo))
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.distributionKeeper.FundCommunityPool(ctx, denomToSwapToCoins, nonNativeFeeCollectorForCommunityPoolAddr)
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
