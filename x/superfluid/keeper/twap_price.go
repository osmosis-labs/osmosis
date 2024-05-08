package keeper

import (
	"github.com/cosmos/gogoproto/proto"
	cltypes "github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/types"
	"strings"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This function calculates the osmo equivalent worth of an LP share.
// It is intended to eventually use the TWAP of the worth of an LP share
// once that is exposed from the gamm module.
func (k Keeper) calculateOsmoBackingPerShare(pool gammtypes.CFMMPoolI, osmoInPool osmomath.Int) osmomath.Dec {
	twap := osmoInPool.ToLegacyDec().Quo(pool.GetTotalShares().ToLegacyDec())
	return twap
}

func (k Keeper) SetOsmoEquivalentMultiplier(ctx sdk.Context, epoch int64, denom string, multiplier osmomath.Dec) {
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

func isNative(denom string) bool {
	if strings.HasPrefix(denom, cltypes.ConcentratedLiquidityTokenPrefix) {
		return false
	}
	if strings.HasPrefix(denom, gammtypes.GAMMTokenPrefix) {
		return false
	}
	return true
}

func (k Keeper) GetSuperfluidOSMOTokens(ctx sdk.Context, denom string, amount osmomath.Int, includeNative bool) (osmomath.Int, error) {
	if !includeNative && isNative(denom) {
		return osmomath.ZeroInt(), nil
	}

	multiplier := k.GetOsmoEquivalentMultiplier(ctx, denom)
	if multiplier.IsZero() {
		return osmomath.ZeroInt(), nil
	}

	decAmt := multiplier.Mul(amount.ToLegacyDec())
	_, err := k.GetSuperfluidAsset(ctx, denom)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	// TODO: here! change the risk adjusted value based on the type/denom
	return k.GetRiskAdjustedOsmoValue(ctx, decAmt.RoundInt(), denom), nil
}

func (k Keeper) DeleteOsmoEquivalentMultiplier(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetOsmoEquivalentMultiplier(ctx sdk.Context, denom string) osmomath.Dec {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenMultiplier)
	bz := prefixStore.Get([]byte(denom))
	if bz == nil {
		return osmomath.ZeroDec()
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
