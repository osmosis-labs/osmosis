package poolmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v19/app/params"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v19/x/txfees/types"
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

// chargeTakerFee extracts the taker fee from the given tokenIn and sends it to the appropriate
// module account. It returns the tokenIn after the taker fee has been extracted.
func (k Keeper) chargeTakerFee(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, sender sdk.AccAddress, exactIn bool) (sdk.Coin, error) {
	feeCollectorForStakingRewardsName := txfeestypes.FeeCollectorForStakingRewardsName
	feeCollectorForCommunityPoolName := txfeestypes.FeeCollectorForCommunityPoolName
	defaultTakerFeeDenom := appparams.BaseCoinUnit
	poolManagerParams := k.GetParams(ctx)

	takerFee, err := k.GetTradingPairTakerFee(ctx, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	var tokenInAfterTakerFee sdk.Coin
	var takerFeeCoin sdk.Coin
	if exactIn {
		tokenInAfterTakerFee, takerFeeCoin = k.calcTakerFeeExactIn(tokenIn, takerFee)
	} else {
		tokenInAfterTakerFee, takerFeeCoin = k.calcTakerFeeExactOut(tokenIn, takerFee)
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
			osmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(defaultTakerFeeDenom, osmoTakerFeeToCommunityPoolDec.TruncateInt()))
			err := k.communityPoolKeeper.FundCommunityPool(ctx, osmoTakerFeeToCommunityPoolCoins, sender)
			if err != nil {
				return sdk.Coin{}, err
			}
			takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(osmoTakerFeeToCommunityPoolCoins.AmountOf(defaultTakerFeeDenom))
		}
		// Staking Rewards:
		if poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards.GT(osmomath.ZeroDec()) {
			// Osmo staking rewards funds are sent to the non native fee pool module account (even though its native, we want to distribute at the same time as the non native fee tokens)
			// We could stream these rewards via the fee collector account, but this is decision to be made by governance.
			osmoTakerFeeToStakingRewardsCoins := sdk.NewCoins(sdk.NewCoin(defaultTakerFeeDenom, takerFeeAmtRemaining))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForStakingRewardsName, osmoTakerFeeToStakingRewardsCoins)
			if err != nil {
				return sdk.Coin{}, err
			}
		}

		// If the denom is not the base denom:
	} else {
		// Community Pool:
		if poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool.GT(osmomath.ZeroDec()) {
			denomIsWhitelisted := isDenomWhitelisted(takerFeeCoin.Denom, poolManagerParams.AuthorizedQuoteDenoms)
			// If the non osmo denom is a whitelisted quote asset, we send to the community pool
			if denomIsWhitelisted {
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeAmtRemaining.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt()))
				err := k.communityPoolKeeper.FundCommunityPool(ctx, nonOsmoTakerFeeToCommunityPoolCoins, sender)
				if err != nil {
					return sdk.Coin{}, err
				}
				takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(nonOsmoTakerFeeToCommunityPoolCoins.AmountOf(tokenIn.Denom))
			} else {
				// If the non osmo denom is not a whitelisted asset, we send to the non native fee pool for community pool module account.
				// At epoch, this account swaps the non native, non whitelisted assets for XXX and sends to the community pool.
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeAmtRemaining.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt()))
				err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForCommunityPoolName, nonOsmoTakerFeeToCommunityPoolCoins)
				if err != nil {
					return sdk.Coin{}, err
				}
				takerFeeAmtRemaining = takerFeeAmtRemaining.Sub(nonOsmoTakerFeeToCommunityPoolCoins.AmountOf(tokenIn.Denom))
			}
		}
		// Staking Rewards:
		if poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards.GT(osmomath.ZeroDec()) {
			// Non Osmo staking rewards are sent to the non native fee pool module account
			nonOsmoTakerFeeToStakingRewardsCoins := sdk.NewCoins(sdk.NewCoin(takerFeeCoin.Denom, takerFeeAmtRemaining))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, feeCollectorForStakingRewardsName, nonOsmoTakerFeeToStakingRewardsCoins)
			if err != nil {
				return sdk.Coin{}, err
			}
		}
	}

	return tokenInAfterTakerFee, nil
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) calcTakerFeeExactIn(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterSubTakerFee := tokenIn.Amount.ToLegacyDec().MulTruncate(osmomath.OneDec().Sub(takerFee))
	tokenInAfterSubTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterSubTakerFee.TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenIn.Amount.Sub(tokenInAfterSubTakerFee.Amount))

	return tokenInAfterSubTakerFee, takerFeeCoin
}

func (k Keeper) calcTakerFeeExactOut(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterAddTakerFee := tokenIn.Amount.ToLegacyDec().Quo(osmomath.OneDec().Sub(takerFee))
	tokenInAfterAddTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterAddTakerFee.Ceil().TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenInAfterAddTakerFee.Amount.Sub(tokenIn.Amount))

	return tokenInAfterAddTakerFee, takerFeeCoin
}
