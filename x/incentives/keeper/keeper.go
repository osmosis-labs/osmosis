package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	cdc        codec.Marshaler
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
	hooks      types.IncentiveHooks
	ak         authkeeper.AccountKeeper
	bk         types.BankKeeper
	lk         types.LockupKeeper
	ek         types.EpochKeeper
}

func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, bk types.BankKeeper, lk types.LockupKeeper, ek types.EpochKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		bk:         bk,
		lk:         lk,
		ek:         ek,
	}
}

// Set the gamm hooks
func (k *Keeper) SetHooks(ih types.IncentiveHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set incentive hooks twice")
	}

	k.hooks = ih

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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
