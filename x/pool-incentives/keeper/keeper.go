package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	paramSpace paramtypes.Subspace

	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	incentivesKeeper types.IncentivesKeeper
	distrKeeper      types.DistrKeeper

	communityPoolName string // name of the Community pool ModuleAccount (Maybe the distribution module)
	feeCollectorName  string // name of the FeeCollector ModuleAccount
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, incentivesKeeper types.IncentivesKeeper, distrKeeper types.DistrKeeper, communityPoolName string, feeCollectorName string) Keeper {
	// ensure pool-incentives module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		paramSpace: paramSpace,

		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		incentivesKeeper: incentivesKeeper,
		distrKeeper:      distrKeeper,

		communityPoolName: communityPoolName,
		feeCollectorName:  feeCollectorName,
	}
}

func (k Keeper) CreatePoolPots(ctx sdk.Context, poolId uint64) error {
	// Create the same number of pots as there are LockableDurations
	for _, lockableDuration := range k.GetLockableDurations(ctx) {
		potId, err := k.incentivesKeeper.CreatePot(
			ctx,
			true,
			k.accountKeeper.GetModuleAddress(types.ModuleName),
			sdk.Coins{},
			lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         gammtypes.GetPoolShareDenom(poolId),
				Duration:      lockableDuration,
				Timestamp:     time.Time{},
			},
			// QUESTION: Should we set the startTime as the epoch start time that the modules share or the current block time?
			ctx.BlockTime(),
			1,
		)
		if err != nil {
			return err
		}

		k.SetPoolPotId(ctx, poolId, lockableDuration, potId)
	}

	return nil
}

func (k Keeper) SetPoolPotId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration, potId uint64) {
	key := types.GetPoolPotIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, sdk.Uint64ToBigEndian(potId))
}

func (k Keeper) GetPoolPotId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error) {
	key := types.GetPoolPotIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	if len(bz) == 0 {
		return 0, sdkerrors.Wrapf(types.ErrNoPotIdExist, "pot id for pool (%d) with duration (%s) not exist", poolId, lockableDuration.String())
	}

	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) SetLockableDurations(ctx sdk.Context, lockableDurations []time.Duration) {
	store := ctx.KVStore(k.storeKey)

	info := types.LockableDurationsInfo{LockableDurations: lockableDurations}

	store.Set(types.LockableDurationsKey, k.cdc.MustMarshalBinaryBare(&info))
}

func (k Keeper) GetLockableDurations(ctx sdk.Context) []time.Duration {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{}

	bz := store.Get(types.LockableDurationsKey)
	if len(bz) == 0 {
		panic("lockable durations not set")
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &info)

	return info.LockableDurations
}
