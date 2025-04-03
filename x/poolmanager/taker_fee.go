package poolmanager

import (
	"bytes"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var zero = osmomath.ZeroInt()

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
// The order of the trading pair matters.
func (k Keeper) GetTradingPairTakerFee(ctx sdk.Context, tokenInDenom, tokenOutDenom string) (osmomath.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatDenomTradePairKey(tokenInDenom, tokenOutDenom)

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
		tokenInDenom, tokenOutDenom, err := types.ParseDenomTradePairKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		takerFees = append(takerFees, types.DenomPairTakerFee{
			TokenInDenom:  tokenInDenom,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee.Dec,
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
	panic("not supported")

	reducedFeeWhitelist := []string{}
	k.paramSpace.Get(ctx, types.KeyReducedTakerFeeByWhitelist, &reducedFeeWhitelist)

	// Determine if eligible to bypass taker fee.
	if osmoutils.Contains(reducedFeeWhitelist, sender.String()) {
		return tokenIn, sdk.Coin{Denom: tokenIn.Denom, Amount: zero}, nil
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
// The function first sorts the denominations lexicographically and then checks for denomShareAgreement and alloyedAssetShareAgreement denoms.
// DenomShareAgreement denoms represent a denom that has a taker fee share agreement with the Osmosis protocol, while alloyedAssetShareAgreement denoms represent a registered alloyed asset pool composed of one or more denoms with a denomShareAgreement.
// If there are one or more denomShareAgreement denoms, the function calculates the percentage of the taker fees that should be skimmed off and increases the accumulator for the denomShareAgreement denom / taker fee denomination pair.
// If there were no denomShareAgreement denoms but there are one or more alloyedAssetShareAgreement denoms, the function calculates the taker fee share for the alloyed asset for each underlying asset that has a taker fee share agreement.
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

	// Retrieve the share agreements for denoms.
	denomShareAgreements, alloyedAssetShareAgreements := k.getTakerFeeShareAgreements(denomsInvolvedInRoute)

	shareAgreementsToProcess := []types.TakerFeeShareAgreement{}
	if len(denomShareAgreements) > 0 {
		shareAgreementsToProcess = append(shareAgreementsToProcess, denomShareAgreements...)
	} else if len(alloyedAssetShareAgreements) > 0 {
		shareAgreementsToProcess = append(shareAgreementsToProcess, alloyedAssetShareAgreements...)
	}

	return k.processShareAgreements(ctx, shareAgreementsToProcess, totalTakerFees)
}

// getTakerFeeShareAgreements checks for individual denomShareAgreement and alloyedAssetShareAgreement denoms.
func (k Keeper) getTakerFeeShareAgreements(denomsInvolvedInRoute []string) ([]types.TakerFeeShareAgreement, []types.TakerFeeShareAgreement) {
	denomShareAgreements := []types.TakerFeeShareAgreement{}
	alloyedAssetShareAgreements := []types.TakerFeeShareAgreement{}

	for _, denom := range denomsInvolvedInRoute {
		// We first check if this denom has a taker fee share agreement.
		takerFeeShareAgreement, found := k.getTakerFeeShareAgreementFromDenom(denom)
		if found {
			// If the denom has a denomShareAgreement, add the denomShareAgreement to the denomShareAgreements slice.
			denomShareAgreements = append(denomShareAgreements, takerFeeShareAgreement)
		} else {
			// Check if the denom is an alloyedAssetShareAgreement denom.
			// If it is, add the alloyedAssetShareAgreement to the alloyedAssetShareAgreements slice.
			cachedAlloyContractState, found := k.getRegisteredAlloyedPoolFromDenom(denom)
			if found {
				alloyedAssetShareAgreements = append(alloyedAssetShareAgreements, cachedAlloyContractState.TakerFeeShareAgreements...)
			}
		}
	}

	return denomShareAgreements, alloyedAssetShareAgreements
}

// processShareAgreements processes share agreements by calculating the taker fee share for the alloyed asset for each underlying asset that has a taker fee share agreement.
func (k Keeper) processShareAgreements(ctx sdk.Context, shareAgreements []types.TakerFeeShareAgreement, totalTakerFees sdk.Coins) error {
	// Early return if there are no share agreements.
	if len(shareAgreements) == 0 {
		return nil
	}

	percentageOfTakerFeeToSkim := osmomath.ZeroDec()
	for _, agreement := range shareAgreements {
		// Add up the percentage of the taker fee that should be skimmed off.
		percentageOfTakerFeeToSkim.AddMut(agreement.SkimPercent)
	}

	// Validate the total percentage of taker fees to skim.
	if err := k.validatePercentage(percentageOfTakerFeeToSkim); err != nil {
		return err
	}

	// For each taker fee coin, calculate the amount to skim off and increase the accumulator for the underlying denomShareAgreement denom / taker fee denom pair.
	for _, takerFeeCoin := range totalTakerFees {
		for _, agreement := range shareAgreements {
			amountToSkim := osmomath.NewDecFromInt(takerFeeCoin.Amount).Mul(agreement.SkimPercent).TruncateInt()
			// Increase the accumulator for the underlying denomShareAgreement denom / taker fee denom pair.
			if err := k.increaseTakerFeeShareDenomsToAccruedValue(ctx, agreement.Denom, takerFeeCoin.Denom, amountToSkim); err != nil {
				return err
			}
		}
	}

	return nil
}

// validatePercentage validates the total percentage of taker fees to skim.
func (k Keeper) validatePercentage(percentage osmomath.Dec) error {
	if percentage.GT(types.OneDec) || percentage.LT(types.ZeroDec) {
		return types.InvalidTakerFeeSharePercentageError{Percentage: percentage}
	}
	return nil
}
