package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	paramSpace paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	farmKeeper    types.FarmKeeper

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, farmkeeper types.FarmKeeper, feeCollectorName string) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		paramSpace: paramSpace,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		farmKeeper:    farmkeeper,

		feeCollectorName: feeCollectorName,
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
