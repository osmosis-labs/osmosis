package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage incentives module storage.
type Keeper struct {
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace
	hooks      types.IncentiveHooks
	ak         types.AccountKeeper
	bk         types.BankKeeper
	lk         types.LockupKeeper
	ek         types.EpochKeeper
	ck         types.CommunityPoolKeeper
	tk         types.TxFeesKeeper
	clk        types.ConcentratedLiquidityKeeper
	pmk        types.PoolManagerKeeper
	pik        types.PoolIncentiveKeeper
	prk        types.ProtorevKeeper
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, ak types.AccountKeeper, bk types.BankKeeper, lk types.LockupKeeper, ek types.EpochKeeper, ck types.CommunityPoolKeeper, txfk types.TxFeesKeeper, clk types.ConcentratedLiquidityKeeper, pmk types.PoolManagerKeeper, pik types.PoolIncentiveKeeper, prk types.ProtorevKeeper) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		bk:         bk,
		lk:         lk,
		ek:         ek,
		ck:         ck,
		pik:        pik,
		tk:         txfk,
		pmk:        pmk,
		clk:        clk,
		prk:        prk,
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

// SetPoolIncentivesKeeper sets pool incentives keeper
func (k *Keeper) SetPoolIncentivesKeeper(poolIncentiveKeeper types.PoolIncentiveKeeper) {
	k.pik = poolIncentiveKeeper
}
