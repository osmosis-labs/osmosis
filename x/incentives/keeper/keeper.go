package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage incentives module storage.
type Keeper struct {
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
	hooks      types.IncentiveHooks
	bk         types.BankKeeper
	lk         types.LockupKeeper
	ek         types.EpochKeeper
	ck         types.CommunityPoolKeeper
	tk         types.TxFeesKeeper
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, bk types.BankKeeper, lk types.LockupKeeper, ek types.EpochKeeper, ck types.CommunityPoolKeeper, txfk types.TxFeesKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		bk:         bk,
		lk:         lk,
		ek:         ek,
		ck:         ck,
		tk:         txfk,
	}
}

// SetHooks sets the incentives hooks.
func (k *Keeper) SetHooks(ih types.IncentiveHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set incentive hooks twice")
	}

	k.hooks = ih

	return k
}

// Logger returns a logger instance for the incentives module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetLockableDurations sets which lockable durations will be incentivized.
func (k Keeper) SetLockableDurations(ctx sdk.Context, lockableDurations []time.Duration) {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{LockableDurations: lockableDurations}
	osmoutils.MustSet(store, types.LockableDurationsKey, &info)
}

// GetLockableDurations returns all incentivized lockable durations.
func (k Keeper) GetLockableDurations(ctx sdk.Context) []time.Duration {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{}
	osmoutils.MustGet(store, types.LockableDurationsKey, &info)
	return info.LockableDurations
}
