package poolmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v21/app/params"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v21/x/txfees/types"
)

// SetDenomPairTakerFee sets the taker fee for the given trading pair.
// If the taker fee for this denom pair matches the default taker fee, then
// it is deleted from state.
func (k Keeper) SetDenomPairTakerFee(ctx sdk.Context, denom0, denom1 string, takerFee osmomath.Dec) {
	store := ctx.KVStore(k.storeKey)
	// if given taker fee is equal to the default taker fee,
	// delete whatever we have in current state to use default taker fee.
	if takerFee.Equal(k.GetParams(ctx).TakerFeeParams.DefaultTakerFee) {
		store.Delete(types.FormatDenomTradePairKey(denom0, denom1))
		return
	} else {
		osmoutils.MustSetDec(store, types.FormatDenomTradePairKey(denom0, denom1), takerFee)
	}
}

// SenderValidationSetDenomPairTakerFee sets the taker fee for the given trading pair iff the sender's address
// also exists in the pool manager taker fee admin address list.
func (k Keeper) SenderValidationSetDenomPairTakerFee(ctx sdk.Context, sender, denom0, denom1 string, takerFee osmomath.Dec) error {
	adminAddresses := k.GetParams(ctx).TakerFeeParams.AdminAddresses
	isAdmin := false
	for _, admin := range adminAddresses {
		if admin == sender {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return fmt.Errorf("%s is not in the pool manager taker fee admin address list", sender)
	}

	k.SetDenomPairTakerFee(ctx, denom0, denom1, takerFee)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetDenomPairTakerFee,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender),
			sdk.NewAttribute(types.AttributeKeyDenom0, denom0),
			sdk.NewAttribute(types.AttributeKeyDenom1, denom1),
			sdk.NewAttribute(types.AttributeKeyTakerFee, takerFee.String()),
		),
	})

	return nil
}

// GetTradingPairTakerFee returns the taker fee for the given trading pair.
// If the trading pair does not exist, it returns the default taker fee.
func (k Keeper) GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatDenomTradePairKey(denom0, denom1)

	takerFee := &sdk.DecProto{}
	found, err := osmoutils.Get(store, key, takerFee)
	if err != nil {
		return osmomath.Dec{}, err
	}
	if !found {
		return k.GetParams(ctx).TakerFeeParams.DefaultTakerFee, nil
	}

	return takerFee.Dec, nil
}

// GetAllTradingPairTakerFees returns all the custom taker fees for trading pairs.
func (k Keeper) GetAllTradingPairTakerFees(ctx sdk.Context) ([]types.DenomPairTakerFee, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, types.DenomTradePairPrefix)
	defer iterator.Close()

	var takerFees []types.DenomPairTakerFee
	for ; iterator.Valid(); iterator.Next() {
		takerFee := &sdk.DecProto{}
		osmoutils.MustGet(store, iterator.Key(), takerFee)
		denom0, denom1, err := types.ParseDenomTradePairKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		takerFees = append(takerFees, types.DenomPairTakerFee{
			Denom0:   denom0,
			Denom1:   denom1,
			TakerFee: takerFee.Dec,
		})
	}

	return takerFees, nil
}

// chargeTakerFee extracts the taker fee from the given tokenIn and sends it to the appropriate
// module account. It returns the tokenIn after the taker fee has been extracted.
// If the sender is in the taker fee reduced whitelisted, it returns the tokenIn without extracting the taker fee.
// In the future, we might charge a lower taker fee as opposed to no fee at all.
func (k Keeper) chargeTakerFee(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, sender sdk.AccAddress, exactIn bool) (sdk.Coin, error) {
	feeCollectorForStakingRewardsName := txfeestypes.FeeCollectorForStakingRewardsName
	feeCollectorForCommunityPoolName := txfeestypes.FeeCollectorForCommunityPoolName
	defaultTakerFeeDenom := appparams.BaseCoinUnit
	poolManagerParams := k.GetParams(ctx)

	// Determine if eligible to bypass taker fee.
	if osmoutils.Contains(poolManagerParams.TakerFeeParams.ReducedFeeWhitelist, sender.String()) {
		return tokenIn, nil
	}

	takerFee, err := k.GetTradingPairTakerFee(ctx, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	var tokenInAfterTakerFee sdk.Coin
	var takerFeeCoin sdk.Coin
	if exactIn {
		tokenInAfterTakerFee, takerFeeCoin = CalcTakerFeeExactIn(tokenIn, takerFee)
	} else {
		tokenInAfterTakerFee, takerFeeCoin = CalcTakerFeeExactOut(tokenIn, takerFee)
	}

	// N.B. We truncate from the community pool calculation, then remove that from the total, and use the remaining for staking rewards.
	// If we truncate both, these can leave tokens in the users wallet when swapping and exact amount in, which is bad UX.

	// We determine the distributution of the taker fee based on its denom
	// If the denom is the base denom:
	takerFeeAmtRemaining := takerFeeCoin.Amount
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		// Community Pool:
		if poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool.GT(osmomath.ZeroDec()) {
			// Osmo community pool funds is a direct send
			osmoTakerFeeToCommunityPoolDec := takerFeeAmtRemaining.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool)
			osmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(defaultTakerFeeDenom, osmoTakerFeeToCommunityPoolDec.TruncateInt())
			err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(osmoTakerFeeToCommunityPoolCoin), sender)
			if err != nil {
				return sdk.Coin{}, err
			}
			k.IncreaseTakerFeeTrackerForCommunityPool(ctx, osmoTakerFeeToCommunityPoolCoin)
			takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(osmoTakerFeeToCommunityPoolCoin.Amount)
		}
		// Staking Rewards:
		if poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards.GT(osmomath.ZeroDec()) {
			// Osmo staking rewards funds are sent to the non native fee pool module account (even though its native, we want to distribute at the same time as the non native fee tokens)
			// We could stream these rewards via the fee collector account, but this is decision to be made by governance.
			osmoTakerFeeToStakingRewardsCoin := sdk.NewCoin(defaultTakerFeeDenom, takerFeeAmtRemaining)
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForStakingRewardsName, sdk.NewCoins(osmoTakerFeeToStakingRewardsCoin))
			if err != nil {
				return sdk.Coin{}, err
			}
			k.IncreaseTakerFeeTrackerForStakers(ctx, osmoTakerFeeToStakingRewardsCoin)
		}

		// If the denom is not the base denom:
	} else {
		// Community Pool:
		if poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool.GT(osmomath.ZeroDec()) {
			denomIsWhitelisted := isDenomWhitelisted(takerFeeCoin.Denom, poolManagerParams.AuthorizedQuoteDenoms)
			// If the non osmo denom is a whitelisted quote asset, we send to the community pool
			if denomIsWhitelisted {
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeAmtRemaining.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(nonOsmoTakerFeeToCommunityPoolCoin), sender)
				if err != nil {
					return sdk.Coin{}, err
				}
				k.IncreaseTakerFeeTrackerForCommunityPool(ctx, nonOsmoTakerFeeToCommunityPoolCoin)
				takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
			} else {
				// If the non osmo denom is not a whitelisted asset, we send to the non native fee pool for community pool module account.
				// At epoch, this account swaps the non native, non whitelisted assets for XXX and sends to the community pool.
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeAmtRemaining.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForCommunityPoolName, sdk.NewCoins(nonOsmoTakerFeeToCommunityPoolCoin))
				if err != nil {
					return sdk.Coin{}, err
				}
				k.IncreaseTakerFeeTrackerForCommunityPool(ctx, nonOsmoTakerFeeToCommunityPoolCoin)
				takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
			}
		}
		// Staking Rewards:
		if poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards.GT(osmomath.ZeroDec()) {
			// Non Osmo staking rewards are sent to the non native fee pool module account
			nonOsmoTakerFeeToStakingRewardsCoin := sdk.NewCoin(takerFeeCoin.Denom, takerFeeAmtRemaining)
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForStakingRewardsName, sdk.NewCoins(nonOsmoTakerFeeToStakingRewardsCoin))
			if err != nil {
				return sdk.Coin{}, err
			}
			k.IncreaseTakerFeeTrackerForStakers(ctx, nonOsmoTakerFeeToStakingRewardsCoin)
		}
	}

	return tokenInAfterTakerFee, nil
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func CalcTakerFeeExactIn(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	takerFeeFactor := osmomath.OneDec().SubMut(takerFee)
	// TODO: Remove .ToLegacyDec and instead do MulInt. Need to test state compat.
	amountInAfterSubTakerFee := tokenIn.Amount.ToLegacyDec().MulTruncate(takerFeeFactor)
	tokenInAfterSubTakerFee := sdk.Coin{Denom: tokenIn.Denom, Amount: amountInAfterSubTakerFee.TruncateInt()}
	takerFeeCoin := sdk.Coin{Denom: tokenIn.Denom, Amount: tokenIn.Amount.Sub(tokenInAfterSubTakerFee.Amount)}

	return tokenInAfterSubTakerFee, takerFeeCoin
}

func CalcTakerFeeExactOut(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	takerFeeFactor := osmomath.OneDec().SubMut(takerFee)
	amountInAfterAddTakerFee := tokenIn.Amount.ToLegacyDec().Quo(takerFeeFactor)
	tokenInAfterAddTakerFee := sdk.Coin{Denom: tokenIn.Denom, Amount: amountInAfterAddTakerFee.Ceil().TruncateInt()}
	takerFeeCoin := sdk.Coin{Denom: tokenIn.Denom, Amount: tokenInAfterAddTakerFee.Amount.Sub(tokenIn.Amount)}

	return tokenInAfterAddTakerFee, takerFeeCoin
}
