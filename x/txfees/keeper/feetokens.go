package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/txfees/types"
	// this line is used by starport scaffolding # ibc/keeper/import
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error) {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	feeToken, err := k.GetFeeToken(ctx, inputFee.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	spotPrice, err := k.spotPriceCalculator.CalculateSpotPrice(ctx, feeToken.PoolID, feeToken.Denom, baseDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	return sdk.NewCoin(baseDenom, spotPrice.MulInt(inputFee.Amount).Ceil().RoundInt()), nil
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
func (k Keeper) setBaseDenom(ctx sdk.Context, denom string) error {
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
// - The gamm pool exists
// - The gamm pool includes the base token
func (k Keeper) ValidateFeeToken(ctx sdk.Context, feeToken types.FeeToken) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}
	_, err = k.spotPriceCalculator.CalculateSpotPrice(ctx, feeToken.PoolID, feeToken.Denom, baseDenom)

	return err
}

// GetFeeToken returns the fee token record for a specific denom
func (k Keeper) GetFeeToken(ctx sdk.Context, denom string) (types.FeeToken, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.FeeTokensStorePrefix))
	if !prefixStore.Has([]byte(denom)) {
		return types.FeeToken{}, types.ErrInvalidFeeToken
	}
	bz := prefixStore.Get([]byte(denom))

	feeToken := types.FeeToken{}
	err := proto.Unmarshal(bz, &feeToken)
	if err != nil {
		return types.FeeToken{}, err
	}

	return feeToken, nil
}

// setFeeToken sets a new fee token record for a specific denom
func (k Keeper) setFeeToken(ctx sdk.Context, feeToken types.FeeToken) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.FeeTokensStorePrefix))

	if feeToken.PoolID == 0 {
		if prefixStore.Has([]byte(feeToken.Denom)) {
			prefixStore.Delete([]byte(feeToken.Denom))
		}
		return nil
	}

	bz, err := proto.Marshal(&feeToken)
	if err != nil {
		return err
	}

	prefixStore.Set([]byte(feeToken.Denom), bz)
	return nil
}

func (k Keeper) GetFeeTokens(ctx sdk.Context) (feetokens []types.FeeToken) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.FeeTokensStorePrefix))

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

func (k Keeper) setFeeTokens(ctx sdk.Context, feetokens []types.FeeToken) {
	for _, feeToken := range feetokens {
		k.setFeeToken(ctx, feeToken)
	}
}
