package keeper

// This file handles

import (
	"github.com/cosmos/gogoproto/proto"
	"github.com/osmosis-labs/osmosis/osmomath"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetSuperfluidAsset(ctx sdk.Context, asset types.SuperfluidAsset) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixSuperfluidAsset)
	bz, err := proto.Marshal(&asset)
	if err != nil {
		panic(err)
	}
	prefixStore.Set([]byte(asset.Denom), bz)
}

func (k Keeper) DeleteSuperfluidAsset(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixSuperfluidAsset)
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetSuperfluidAsset(ctx sdk.Context, denom string) (types.SuperfluidAsset, error) {
	asset := types.SuperfluidAsset{}
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixSuperfluidAsset)
	found, err := osmoutils.Get(prefixStore, []byte(denom), &asset)
	if err != nil {
		return types.SuperfluidAsset{}, err
	}
	if !found {
		return types.SuperfluidAsset{}, errorsmod.Wrapf(types.ErrNonSuperfluidAsset, "denom: %s", denom)
	}
	return asset, nil
}

func (k Keeper) GetAllSuperfluidAssets(ctx sdk.Context) []types.SuperfluidAsset {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixSuperfluidAsset)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	assets := []types.SuperfluidAsset{}
	for ; iterator.Valid(); iterator.Next() {
		asset := types.SuperfluidAsset{}

		err := proto.Unmarshal(iterator.Value(), &asset)
		if err != nil {
			panic(err)
		}

		assets = append(assets, asset)
	}
	return assets
}

func (k Keeper) SetDenomRiskFactor(ctx sdk.Context, denom string, riskFactor osmomath.Dec) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixDenomRiskFactor)
	riskFactorRecord := types.DenomRiskFactor{
		RiskFactor: riskFactor,
	}
	bz, err := proto.Marshal(&riskFactorRecord)
	if err != nil {
		return err
	}
	prefixStore.Set([]byte(denom), bz)
	return nil
}

func (k Keeper) DeleteDenomRiskFactor(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixDenomRiskFactor)
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetDenomRiskFactor(ctx sdk.Context, denom string) (osmomath.Dec, bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixDenomRiskFactor)
	bz := prefixStore.Get([]byte(denom))
	if bz == nil {
		return osmomath.ZeroDec(), false
	}
	denomRiskFactor := types.DenomRiskFactor{}
	err := proto.Unmarshal(bz, &denomRiskFactor)
	if err != nil {
		return osmomath.ZeroDec(), false
	}
	return denomRiskFactor.RiskFactor, true
}

func (k Keeper) GetAllDenomRiskFactors(ctx sdk.Context) []types.DenomRiskFactor {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixDenomRiskFactor)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	denomRiskFactors := []types.DenomRiskFactor{}
	for ; iterator.Valid(); iterator.Next() {
		denomRiskFactor := types.DenomRiskFactor{}

		err := proto.Unmarshal(iterator.Value(), &denomRiskFactor)
		if err != nil {
			panic(err)
		}

		denomRiskFactors = append(denomRiskFactors, denomRiskFactor)
	}
	return denomRiskFactors
}
