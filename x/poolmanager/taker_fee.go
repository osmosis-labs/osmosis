package poolmanager

import (
	"bytes"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v25/x/txfees/types"
)

func (k *Keeper) GetDefaultTakerFee(ctx sdk.Context) osmomath.Dec {
	defaultTakerFeeBz := k.paramSpace.GetRaw(ctx, types.KeyDefaultTakerFee)
	if !bytes.Equal(defaultTakerFeeBz, k.defaultTakerFeeBz) {
		var defaultTakerFeeValue osmomath.Dec
		err := json.Unmarshal(defaultTakerFeeBz, &defaultTakerFeeValue)
		if err != nil {
			defaultTakerFeeValue = osmomath.ZeroDec()
		}
		k.defaultTakerFeeBz = defaultTakerFeeBz
		k.defaultTakerFeeVal = defaultTakerFeeValue
	}
	return k.defaultTakerFeeVal
}

// SetDenomPairTakerFee sets the taker fee for the given trading pair.
// If the taker fee for this denom pair matches the default taker fee, then
// it is deleted from state.
func (k Keeper) SetDenomPairTakerFee(ctx sdk.Context, denom0, denom1 string, takerFee osmomath.Dec) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatDenomTradePairKey(denom0, denom1)
	// if given taker fee is equal to the default taker fee,
	// delete whatever we have in current state to use default taker fee.
	// TODO: This logic is actually wrong imo, where it can be valid to set an override over the default.
	if takerFee.Equal(k.GetDefaultTakerFee(ctx)) {
		store.Delete(key)
		return
	} else {
		osmoutils.MustSetDec(store, key, takerFee)
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
		return k.GetDefaultTakerFee(ctx), nil
	}

	return takerFee.Dec, nil
}

// GetAllTradingPairTakerFees returns all the custom taker fees for trading pairs.
func (k Keeper) GetAllTradingPairTakerFees(ctx sdk.Context) ([]types.DenomPairTakerFee, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, types.DenomTradePairPrefix)
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
// TODO: Gas optimize this function, its expensive in both gas and CPU.
func (k Keeper) chargeTakerFee(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, sender sdk.AccAddress, exactIn bool) (sdk.Coin, sdk.Coin, error) {
	takerFeeModuleAccountName := txfeestypes.TakerFeeCollectorName

	reducedFeeWhitelist := []string{}
	k.paramSpace.Get(ctx, types.KeyReducedTakerFeeByWhitelist, &reducedFeeWhitelist)

	// Determine if eligible to bypass taker fee.
	if osmoutils.Contains(reducedFeeWhitelist, sender.String()) {
		return tokenIn, sdk.Coin{}, nil
	}

	takerFee, err := k.GetTradingPairTakerFee(ctx, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	var tokenInAfterTakerFee sdk.Coin
	var takerFeeCoin sdk.Coin
	if exactIn {
		tokenInAfterTakerFee, takerFeeCoin = CalcTakerFeeExactIn(tokenIn, takerFee)
	} else {
		tokenInAfterTakerFee, takerFeeCoin = CalcTakerFeeExactOut(tokenIn, takerFee)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, takerFeeModuleAccountName, sdk.NewCoins(takerFeeCoin))
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	return tokenInAfterTakerFee, takerFeeCoin, nil
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func CalcTakerFeeExactIn(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	takerFeeFactor := osmomath.OneDec().SubMut(takerFee)
	amountInAfterSubTakerFee := takerFeeFactor.MulIntMut(tokenIn.Amount)
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

// TakerFeeSkim calculates the taker fee share for each denomination involved in a route and increases the accumulator for the respective denomination pair.
// The function first sorts the denominations lexicographically and then checks for tier 1 and tier 2 denominations.
// Tier 1 denominations represent a bridge provider that has a taker fee share agreement, while tier 2 denominations represent the alloyed assets themselves.
// If there are one or more tier 1 share agreements, the function calculates the percentage of the taker fees that should be skimmed off and increases the accumulator for the tier 1 denomination / taker fee denomination pair.
// If there were no tier 1 denominations and there are tier 2 denominations, the function calculates the taker fee share for the alloyed asset for each underlying asset that has a taker fee share agreement.
// The function returns an error if the total taker fee share percentage is greater than 1.
//
// Parameters:
// - ctx: The context of the function call.
// - denomsInvolvedInRoute: A slice of strings representing the denominations involved in the route.
// - totalTakerFees: The total taker fees from the swap represented as sdk.Coins.
//
// Returns:
// - An error if the total taker fee share percentage is greater than 1, or if there's an error in increasing the accumulator for the denomination pair.
func (k Keeper) TakerFeeSkim(ctx sdk.Context, denomsInvolvedInRoute []string, totalTakerFees sdk.Coins) error {
	// Sort the denoms involved in the route lexicographically.
	osmoutils.SortSlice(denomsInvolvedInRoute)

	// Check for tier 1 and tier 2 denoms,
	// Tier 1 denoms are denoms that represent a bridge provider that has a taker fee share agreement.
	// Tier 2 denoms are denoms that represent the alloyed assets themselves.
	tier1ShareAgreements := []types.TakerFeeShareAgreement{}
	tier2ShareAgreements := []types.TakerFeeShareAgreement{}
	for _, denom := range denomsInvolvedInRoute {
		// We first check if this denom has a taker fee share agreement with a bridge provider (tier 1).
		takerFeeShareAgreement, found := k.GetTakerFeeShareAgreementFromDenom(ctx, denom)

		if found {
			// If the denom is a tier 1 denom, add it to the tier 1 share agreements slice.
			tier1ShareAgreements = append(tier1ShareAgreements, takerFeeShareAgreement)
		} else if len(tier1ShareAgreements) == 0 {
			// If there are no tier 1 share agreements in the tier 1 share agreements slice, continue to filter this denom to determine if it is a tier 2 denom.
			// If there are 1 or more tier 1 share agreements in the tier 1 share agreements slice, we don't need to check for tier 2 denoms anymore, since the taker fee share
			// only goes to tier 2 denoms IFF there are no tier 1 denoms in the route.
			// Check if denom is tier 2
			// If it is, add it to the tier 2 share agreements slice.
			cachedAlloyContractState, found := k.GetRegisteredAlloyedPoolFromDenom(ctx, denom)
			if found {
				tier2ShareAgreements = append(tier2ShareAgreements, cachedAlloyContractState.TakerFeeShareAgreements...)
			}
		}
	}

	// Filtering complete

	// If there are 1 or more tier 1 share agreements, add up the percentage of the taker fees that should be skimmed off.
	// If the total of taker fee share is greater than 1, return an error.
	// Then, for each taker fee coin, calculate the amount to skim off and increase the accumulator for the tier 1 denom / taker fee denom pair.
	if len(tier1ShareAgreements) > 1 {
		percentageOfTakerFeeToSkim := osmomath.ZeroDec()
		for _, takerFeeShareAgreement := range tier1ShareAgreements {
			// Add up the percentage of the taker fee that should be skimmed off.
			percentageOfTakerFeeToSkim = percentageOfTakerFeeToSkim.Add(takerFeeShareAgreement.SkimPercent)
		}
		if percentageOfTakerFeeToSkim.GT(osmomath.OneDec()) {
			return fmt.Errorf("total taker fee share percentage is greater than 1")
		}
		for _, takerFeeCoin := range totalTakerFees {
			for _, takerFeeShareAgreement := range tier1ShareAgreements {
				amountToSkim := osmomath.NewDecFromInt(takerFeeCoin.Amount).Mul(takerFeeShareAgreement.SkimPercent).TruncateInt()

				// Increase the accumulator for the tier 1 denom / taker fee denom pair.
				err := k.IncreaseTakerFeeShareDenomsToAccruedValue(ctx, takerFeeShareAgreement.Denom, takerFeeCoin.Denom, amountToSkim)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// IFF there were no tier 1 denoms and there are tier 2 denoms, we calculate the taker fee share for the alloyed asset for each underlying asset that has a taker fee share agreement.
	if len(tier1ShareAgreements) == 0 && len(tier2ShareAgreements) > 0 {
		percentageOfTakerFeeToSkim := osmomath.ZeroDec()
		for _, takerFeeShareAgreement := range tier2ShareAgreements {
			// Add up the percentage of the taker fee that should be skimmed off.
			percentageOfTakerFeeToSkim = percentageOfTakerFeeToSkim.Add(takerFeeShareAgreement.SkimPercent)
		}
		if percentageOfTakerFeeToSkim.GT(osmomath.OneDec()) {
			return fmt.Errorf("total taker fee share percentage is greater than 1")
		}
		for _, takerFeeCoin := range totalTakerFees {
			for _, takerFeeShareAgreement := range tier2ShareAgreements {
				amountToSkim := osmomath.NewDecFromInt(takerFeeCoin.Amount).Mul(takerFeeShareAgreement.SkimPercent).TruncateInt()

				// Increase the accumulator for the underlying tier 1 denom / taker fee denom pair.
				// TODO
				err := k.IncreaseTakerFeeShareDenomsToAccruedValue(ctx, takerFeeShareAgreement.Denom, takerFeeCoin.Denom, amountToSkim)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}

func (k Keeper) BeginBlock(ctx sdk.Context) {
	// Set the caches if they are empty
	if len(k.cachedTakerFeeShareAgreement) == 0 {
		err := k.SetTakerFeeShareAgreementsMapCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting taker fee share agreements map cached: %w", err).Error())
		}
	}
	if len(k.cachedRegisteredAlloyPoolToState) == 0 {
		err := k.SetAllRegisteredAlloyedPoolsCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting all registered alloyed pools cached: %w", err).Error())
		}
	}
	if len(k.cachedRegisteredAlloyedPoolId) == 0 {
		err := k.SetAllRegisteredAlloyedPoolsIdCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting all registered alloyed pools id cached: %w", err).Error())
		}
	}
}

func (k Keeper) EndBlock(ctx sdk.Context) {
	// get changed pools grabs all altered pool ids from the twap transient store.
	// 'altered pool ids' gets automatically cleared on commit by being a transient store
	changedPoolIds := k.twapKeeper.GetChangedPools(ctx)

	for _, id := range changedPoolIds {
		_, found := k.cachedRegisteredAlloyedPoolId[id]

		if found {
			err := k.recalculateAndSetTakerFeeShareAlloyComposition(ctx, id)
			if err != nil {
				ctx.Logger().Error(fmt.Errorf(
					"error in setting registered alloyed pool for pool id %d: %w", id, err,
				).Error())
			}
		}
	}
}
