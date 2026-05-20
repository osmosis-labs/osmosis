package keeper

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmoutils "github.com/osmosis-labs/osmosis/osmoutils"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount.
func (k Keeper) ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error) {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	if inputFee.Denom == baseDenom {
		return inputFee, nil
	}

	feeToken, err := k.GetFeeToken(ctx, inputFee.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	spotPrice, err := k.CalcFeeSpotPrice(ctx, feeToken.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Note: spotPrice truncation is done here for maintaining state-compatibility with v19.x
	// It should be changed to support full spot price precision before
	// https://github.com/osmosis-labs/osmosis/issues/6064 is complete
	return sdk.NewCoin(baseDenom, spotPrice.Dec().MulIntMut(inputFee.Amount).RoundInt()), nil
}

// CalcFeeSpotPrice converts the provided tx fees into their equivalent value in the base denomination.
// It first attempts the registered direct pool (FeeToken.PoolID). If that fails, it falls back to
// a 2-hop route discovered via protorev + FeeSwapIntermediaryDenomList, matching the route discovery
// used at epoch-end by swapNonNativeFeeToDenom. This keeps registration-time validation and
// runtime swap behaviour aligned on the same routing source of truth.
func (k Keeper) CalcFeeSpotPrice(ctx sdk.Context, inputDenom string) (osmomath.BigDec, error) {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	feeToken, err := k.GetFeeToken(ctx, inputDenom)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	// Direct route via the registered pool. Argument order matches the existing convention
	// (quote=baseDenom, base=feeToken.Denom) so the return is in units of base/fee.
	spotPrice, directErr := k.poolManager.RouteCalculateSpotPrice(ctx, feeToken.PoolID, baseDenom, feeToken.Denom)
	if directErr == nil {
		return spotPrice, nil
	}

	// Fall back to 2-hop discovery via protorev. This is the same path swapNonNativeFeeToDenom
	// takes, so validation passing here means the epoch swap will find a route too.
	return k.calc2HopSpotPrice(ctx, feeToken.Denom, baseDenom)
}

// GetFeeToken returns the fee token record for a specific denom,
// In our case the baseDenom is uosmo.
func (k Keeper) GetBaseDenom(ctx sdk.Context) (denom string, err error) {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.BaseDenomKey) {
		return "", types.ErrNoBaseDenom
	}

	bz := store.Get(types.BaseDenomKey)

	return string(bz), nil
}

// SetBaseDenom sets the base fee denom for the chain. Should only be used once.
func (k Keeper) SetBaseDenom(ctx sdk.Context, denom string) error {
	store := ctx.KVStore(k.storeKey)

	err := sdk.ValidateDenom(denom)
	if err != nil {
		return err
	}

	store.Set(types.BaseDenomKey, []byte(denom))
	return nil
}

// ValidateFeeToken validates that a fee token record is valid.
// It first tries the registered direct pool (FeeToken.PoolID). If that fails, it tries to
// discover a 2-hop route via protorev + FeeSwapIntermediaryDenomList. This mirrors the
// route-discovery used at epoch swap time in swapNonNativeFeeToDenom, so any token that
// validates here will also be swappable at epoch end.
func (k Keeper) ValidateFeeToken(ctx sdk.Context, feeToken types.FeeToken) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}
	if baseDenom == feeToken.Denom {
		return errorsmod.Wrap(types.ErrInvalidFeeToken, "cannot add basedenom as a whitelisted fee token")
	}

	// Try the registered direct pool first. Success here implies the pool exists and contains
	// both feeToken.Denom and baseDenom.
	_, directErr := k.poolManager.RouteCalculateSpotPrice(ctx, feeToken.PoolID, feeToken.Denom, baseDenom)
	if directErr == nil {
		return nil
	}

	// Direct route failed. Fall back to 2-hop discovery via protorev.
	if _, _, _, err := k.find2HopRoute(ctx, feeToken.Denom, baseDenom); err != nil {
		// Surface the original direct-route error when no fallback is available, since that
		// is what the caller almost always cares about.
		return directErr
	}
	return nil
}

// GetFeeToken returns a unique fee token record for a specific denom.
// If the denom doesn't exist, returns an error.
func (k Keeper) GetFeeToken(ctx sdk.Context, denom string) (types.FeeToken, error) {
	prefixStore := k.GetFeeTokensStore(ctx)
	if !prefixStore.Has([]byte(denom)) {
		return types.FeeToken{}, errorsmod.Wrapf(types.ErrInvalidFeeToken, "%s", denom)
	}
	bz := prefixStore.Get([]byte(denom))

	feeToken := types.FeeToken{}
	err := proto.Unmarshal(bz, &feeToken)
	if err != nil {
		return types.FeeToken{}, err
	}

	return feeToken, nil
}

// setFeeToken sets a new fee token record for a specific denom.
// PoolID is the direct swap pool for the fee token, OR any non-zero sentinel value
// if the registration relies on a 2-hop route discovered via FeeSwapIntermediaryDenomList.
// PoolID == 0 deletes the entry.
func (k Keeper) setFeeToken(ctx sdk.Context, feeToken types.FeeToken) error {
	prefixStore := k.GetFeeTokensStore(ctx)

	if feeToken.PoolID == 0 {
		if prefixStore.Has([]byte(feeToken.Denom)) {
			prefixStore.Delete([]byte(feeToken.Denom))
		}
		return nil
	}

	err := k.ValidateFeeToken(ctx, feeToken)
	if err != nil {
		return err
	}

	bz, err := proto.Marshal(&feeToken)
	if err != nil {
		return err
	}

	prefixStore.Set([]byte(feeToken.Denom), bz)
	return nil
}

func (k Keeper) GetFeeTokens(ctx sdk.Context) (feetokens []types.FeeToken) {
	prefixStore := k.GetFeeTokensStore(ctx)

	// this entire store just contains FeeTokens, so iterate over all entries.
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	feeTokens := []types.FeeToken{}

	for ; iterator.Valid(); iterator.Next() {
		feeToken := types.FeeToken{}

		err := proto.Unmarshal(iterator.Value(), &feeToken)
		if err != nil {
			panic(err)
		}

		feeTokens = append(feeTokens, feeToken)
	}
	return feeTokens
}

func (k Keeper) SetFeeTokens(ctx sdk.Context, feetokens []types.FeeToken) error {
	for _, feeToken := range feetokens {
		err := k.setFeeToken(ctx, feeToken)
		if err != nil {
			return err
		}
	}
	return nil
}

// SenderValidationSetFeeTokens first checks to see if the sender is whitelisted to set fee tokens.
// If the sender is whitelisted, it sets the fee tokens.
// If the sender is not whitelisted, it returns an error.
func (k Keeper) SenderValidationSetFeeTokens(ctx sdk.Context, sender string, feetokens []types.FeeToken) error {
	whitelistedAddresses := k.GetParams(ctx).WhitelistedFeeTokenSetters

	isWhitelisted := osmoutils.Contains(whitelistedAddresses, sender)
	if !isWhitelisted {
		return errorsmod.Wrapf(types.ErrNotWhitelistedFeeTokenSetter, "%s", sender)
	}

	return k.SetFeeTokens(ctx, feetokens)
}

// find2HopRoute searches for a 2-hop route from feeDenom to baseDenom using protorev,
// iterating over the whitelisted intermediary denoms in params.FeeSwapIntermediaryDenomList.
// This is the same discovery used by hooks.go build2HopsRoute / get2HopRoute, so a route
// found here is the same route the epoch swap will use. If you change the iteration shape
// (skip conditions, ordering, partial-success behaviour), mirror the change in build2HopsRoute
// or the validator and the epoch swap will diverge.
// Returns (hop1PoolID, hop2PoolID, intermediaryDenom, nil) on success.
func (k Keeper) find2HopRoute(ctx sdk.Context, feeDenom, baseDenom string) (uint64, uint64, string, error) {
	params := k.GetParams(ctx)
	for _, intermediary := range params.FeeSwapIntermediaryDenomList {
		if intermediary == feeDenom || intermediary == baseDenom {
			continue
		}
		pool1, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, feeDenom, intermediary)
		if err != nil {
			continue
		}
		pool2, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, intermediary, baseDenom)
		if err != nil {
			continue
		}
		return pool1, pool2, intermediary, nil
	}
	return 0, 0, "", errorsmod.Wrapf(types.ErrNoValidRoute, "no 2-hop route from %s to %s", feeDenom, baseDenom)
}

// calc2HopSpotPrice computes a compounded spot price for a 2-hop route discovered via protorev.
// Units: (intermediary/fee) * (base/intermediary) = base/fee, matching the direct-route return.
func (k Keeper) calc2HopSpotPrice(ctx sdk.Context, feeDenom, baseDenom string) (osmomath.BigDec, error) {
	pool1, pool2, intermediary, err := k.find2HopRoute(ctx, feeDenom, baseDenom)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	// hop1Price: intermediary per one feeDenom.
	hop1Price, err := k.poolManager.RouteCalculateSpotPrice(ctx, pool1, intermediary, feeDenom)
	if err != nil {
		return osmomath.BigDec{}, errorsmod.Wrapf(err, "first hop spot price %s -> %s via pool %d", feeDenom, intermediary, pool1)
	}

	// hop2Price: baseDenom per one intermediary.
	hop2Price, err := k.poolManager.RouteCalculateSpotPrice(ctx, pool2, baseDenom, intermediary)
	if err != nil {
		return osmomath.BigDec{}, errorsmod.Wrapf(err, "second hop spot price %s -> %s via pool %d", intermediary, baseDenom, pool2)
	}

	return hop1Price.Mul(hop2Price), nil
}
