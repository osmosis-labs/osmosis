package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmkeeper "github.com/c-osmosis/osmosis/x/farm/keeper"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	// WARNING: Currently param space is used, but param changes by governance has not been considered
	paramSpace paramtypes.Subspace

	// keepers
	accountKeeper types.AccountKeeper
	farmKeeper    farmkeeper.Keeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, farmkeeper farmkeeper.Keeper) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		paramSpace: paramSpace,

		farmKeeper: farmkeeper,
	}
}

func (k Keeper) CreatePoolFarms(ctx sdk.Context, poolId uint64) error {
	// Create the same number of farms as there are LockableDurations
	params := k.GetParams(ctx)
	for _, lockableDuration := range params.LockableDurations {
		farm, err := k.farmKeeper.NewFarm(ctx)
		if err != nil {
			return err
		}

		k.SetPoolFarmId(ctx, poolId, lockableDuration, farm.FarmId)
	}

	return nil
}

func (k Keeper) SetPoolFarmId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration, farmId uint64) {
	key := types.GetPoolFarmIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, sdk.Uint64ToBigEndian(farmId))
}

func (k Keeper) GetPoolFarmId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) uint64 {
	key := types.GetPoolFarmIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	return sdk.BigEndianToUint64(bz)
}

// GetParams returns the total set of yield parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}
