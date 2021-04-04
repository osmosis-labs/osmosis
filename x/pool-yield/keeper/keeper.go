package keeper

import (
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmkeeper "github.com/c-osmosis/osmosis/x/farm/keeper"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	farmKeeper farmkeeper.Keeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, farmkeeper farmkeeper.Keeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		farmKeeper: farmkeeper,
	}
}

func (k Keeper) CreatePoolFarms(ctx sdk.Context, poolId uint64) error {
	// Create the same number of farms as there are LockableDurations
	for _, lockableDuration := range k.GetGenesisState(ctx).LockableDurations {
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

func (k Keeper) GetPoolFarmId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error) {
	key := types.GetPoolFarmIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	if len(bz) == 0 {
		return 0, sdkerrors.Wrapf(types.ErrNoFarmIdExist, "farm id for pool (%d) with duration (%s) not exist", poolId, lockableDuration.String())
	}

	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) SetGenesisState(ctx sdk.Context, genState *types.GenesisState) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GenesisStateKey, types.ModuleCdc.MustMarshalBinaryBare(genState))
}

func (k Keeper) GetGenesisState(ctx sdk.Context) types.GenesisState {
	store := ctx.KVStore(k.storeKey)
	genState := types.GenesisState{}

	bz := store.Get(types.GenesisStateKey)
	if len(bz) == 0 {
		panic("genesis state not set")
	}

	types.ModuleCdc.MustUnmarshalBinaryBare(bz, &genState)

	return genState
}
