package keeper

import (
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	return sdk.NewCoin(baseDenom, spotPrice.MulInt(inputFee.Amount).RoundInt()), nil
}

func (k Keeper) CalcFeeSpotPrice(ctx sdk.Context, inputDenom string) (sdk.Dec, error) {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return sdk.Dec{}, err
	}

	feeToken, err := k.GetFeeToken(ctx, inputDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	spotPrice, err := k.spotPriceCalculator.CalculateSpotPrice(ctx, feeToken.PoolID, baseDenom, feeToken.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	return spotPrice, nil
}

// GetTotalSwapFee gets the total swap fee for a swap message along a route of pool ids
func (k Keeper) getTotalSwapFee(ctx sdk.Context, poolIds []uint64, denomPath []string) (sdk.Dec, error) {
	prefixStore := k.GetFeeTokensStore(ctx)
	swapFees := sdk.ZeroDec()

	// Join/Exit pool support
	if len(denomPath) == 1 {
		return k.gammKeeper.GetSwapFee(ctx, poolIds[0])
	}
	// Get swap fees from pools
	for i := range poolIds {
		// Get swap fee
		swapFee, err := k.gammKeeper.GetSwapFee(ctx, poolIds[i])
		if err != nil {
			return sdk.Dec{}, err
		}

		// if either of the denoms for the pool is a fee token, swap fee can be added
		if prefixStore.Has([]byte(denomPath[i])) || prefixStore.Has([]byte(denomPath[i+1])) {
			// add to existing swap fees
			swapFees = swapFees.Add(swapFee)
		}
	}
	return swapFees, nil
}

func (k Keeper) feeTokenExists(ctx sdk.Context, denom string) bool {
	return k.GetFeeTokensStore(ctx).Has([]byte(denom))
}

// getFeesPaid returns a token representing the fees paid along the route of swaps identified by the pool Ids
func (k Keeper) getFeesPaid(ctx sdk.Context, poolIds []uint64, denomPath []string, token sdk.Coin) (sdk.Coin, error) {
	// Only tokens with fee record can pay swap fees
	prefixStore := k.GetFeeTokensStore(ctx)
	if !prefixStore.Has([]byte(token.Denom)) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidFeeToken, "%s", token.Denom)
	}

	// Get total swap fees
	swapFees, err := k.getTotalSwapFee(ctx, poolIds, denomPath)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Convert token to baseDenom
	token, err = k.ConvertToBaseToken(ctx, token)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Appy swap fee to token amount = ceil(swapFees * token.Amount)
	feedAmount := swapFees.Mul(token.Amount.ToDec()).Ceil().RoundInt()

	// return coin of fee amount in base denom
	return sdk.NewCoin(token.Denom, feedAmount), nil
}

// GetFeeToken returns the fee token record for a specific denom
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

// ValidateFeeToken validates that a fee token record is valid
// It checks:
// - The denom exists
// - The denom is not the base denom
// - The gamm pool exists
// - The gamm pool includes the base token and fee token.
func (k Keeper) ValidateFeeToken(ctx sdk.Context, feeToken types.FeeToken) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}
	if baseDenom == feeToken.Denom {
		return sdkerrors.Wrap(types.ErrInvalidFeeToken, "cannot add basedenom as a whitelisted fee token")
	}
	// This not returning an error implies that:
	// - feeToken.Denom exists
	// - feeToken.PoolID exists
	// - feeToken.PoolID has both feeToken.Denom and baseDenom
	_, err = k.spotPriceCalculator.CalculateSpotPrice(ctx, feeToken.PoolID, feeToken.Denom, baseDenom)

	return err
}

// GetFeeToken returns the fee token record for a specific denom.
// If the denom doesn't exist, returns an error.
func (k Keeper) GetFeeToken(ctx sdk.Context, denom string) (types.FeeToken, error) {
	prefixStore := k.GetFeeTokensStore(ctx)
	if !prefixStore.Has([]byte(denom)) {
		return types.FeeToken{}, sdkerrors.Wrapf(types.ErrInvalidFeeToken, "%s", denom)
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
// If the feeToken pool ID is 0, deletes the fee Token entry.
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
