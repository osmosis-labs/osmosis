package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// This function calculates the osmo equivalent worth of an LP share.
// It is intended to eventually use the TWAP of the worth of an LP share
// once that is exposed from the gamm module.
// CONTRACT: Passed in lpShareDenom MUST be a valid denom for a real gamm pool with OSMO in it
func (k Keeper) CalculateOsmoBackingPerShare(ctx sdk.Context, lpShareDenom string) (sdk.Dec, error) {
	// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
	poolId := gammtypes.MustGetPoolIdFromShareDenom(lpShareDenom)
	pool, err := k.gk.GetPool(ctx, poolId)
	if err != nil {
		// Pool has been unexpectedly deleted
		k.Logger(ctx).Error(err.Error())
		k.BeginUnwindSuperfluidAsset(ctx, 0, types.NewSuperfluidAsset(types.SuperfluidAssetTypeLPShare, lpShareDenom))
		return sdk.ZeroDec(), err
	}

	// get OSMO amount
	bondDenom := k.sk.BondDenom(ctx)
	osmoPoolAsset, err := pool.GetPoolAsset(bondDenom)
	if err != nil {
		// Pool has unexpectedly removed Osmo from its assets.
		k.Logger(ctx).Error(err.Error())
		k.BeginUnwindSuperfluidAsset(ctx, 0, types.NewSuperfluidAsset(types.SuperfluidAssetTypeLPShare, lpShareDenom))
		return sdk.ZeroDec(), err
	}

	twap := osmoPoolAsset.Token.Amount.ToDec().Quo(pool.GetTotalShares().Amount.ToDec())
	return twap, nil
}

func (k Keeper) SetOsmoEquivalentMultiplier(ctx sdk.Context, epoch int64, denom string, multiplier sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	priceRecord := types.OsmoEquivalentMultiplierRecord{
		EpochNumber: epoch,
		Denom:       denom,
		Multiplier:  multiplier,
	}
	bz, err := proto.Marshal(&priceRecord)
	if err != nil {
		panic(err)
	}
	prefixStore.Set([]byte(denom), bz)
}

func (k Keeper) GetSuperfluidOSMOTokens(ctx sdk.Context, denom string, amount sdk.Int) sdk.Int {
	multiplier := k.GetOsmoEquivalentMultiplier(ctx, denom)
	if multiplier.IsZero() {
		return sdk.ZeroInt()
	}

	decAmt := multiplier.Mul(amount.ToDec())
	asset := k.GetSuperfluidAsset(ctx, denom)
	return k.GetRiskAdjustedOsmoValue(ctx, asset, decAmt.RoundInt())
}

func (k Keeper) DeleteOsmoEquivalentMultiplier(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetOsmoEquivalentMultiplier(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	bz := prefixStore.Get([]byte(denom))
	if bz == nil {
		return sdk.ZeroDec()
	}
	priceRecord := types.OsmoEquivalentMultiplierRecord{}
	err := proto.Unmarshal(bz, &priceRecord)
	if err != nil {
		panic(err)
	}
	return priceRecord.Multiplier
}

func (k Keeper) GetAllOsmoEquivalentMultipliers(ctx sdk.Context) []types.OsmoEquivalentMultiplierRecord {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	priceRecords := []types.OsmoEquivalentMultiplierRecord{}
	for ; iterator.Valid(); iterator.Next() {
		priceRecord := types.OsmoEquivalentMultiplierRecord{}

		err := proto.Unmarshal(iterator.Value(), &priceRecord)
		if err != nil {
			panic(err)
		}

		priceRecords = append(priceRecords, priceRecord)
	}
	return priceRecords
}
