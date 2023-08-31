package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramSpace paramtypes.Subspace
	listeners  types.ConcentratedLiquidityListeners

	// keepers
	poolmanagerKeeper    types.PoolManagerKeeper
	accountKeeper        types.AccountKeeper
	bankKeeper           types.BankKeeper
	gammKeeper           types.GAMMKeeper
	poolIncentivesKeeper types.PoolIncentivesKeeper
	incentivesKeeper     types.IncentivesKeeper
	lockupKeeper         types.LockupKeeper
	communityPoolKeeper  types.CommunityPoolKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, gammKeeper types.GAMMKeeper, poolIncentivesKeeper types.PoolIncentivesKeeper, incentivesKeeper types.IncentivesKeeper, lockupKeeper types.LockupKeeper, communityPoolKeeper types.CommunityPoolKeeper, paramSpace paramtypes.Subspace) *Keeper {
	// ParamSubspace must be initialized within app/keepers/keepers.go
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		storeKey:             storeKey,
		paramSpace:           paramSpace,
		cdc:                  cdc,
		accountKeeper:        accountKeeper,
		bankKeeper:           bankKeeper,
		gammKeeper:           gammKeeper,
		poolIncentivesKeeper: poolIncentivesKeeper,
		incentivesKeeper:     incentivesKeeper,
		lockupKeeper:         lockupKeeper,
		communityPoolKeeper:  communityPoolKeeper,
	}
}

// GetParams returns the total set of concentrated-liquidity module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the concentrated-liquidity module's parameters with the provided parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// Set the poolmanager keeper.
func (k *Keeper) SetPoolManagerKeeper(poolmanagerKeeper types.PoolManagerKeeper) {
	k.poolmanagerKeeper = poolmanagerKeeper
}

// Set the gamm keeper.
func (k *Keeper) SetGammKeeper(gammKeeper types.GAMMKeeper) {
	k.gammKeeper = gammKeeper
}

// Set the pool incentives keeper.
func (k *Keeper) SetPoolIncentivesKeeper(poolIncentivesKeeper types.PoolIncentivesKeeper) {
	k.poolIncentivesKeeper = poolIncentivesKeeper
}

// Set the incentives keeper.
func (k *Keeper) SetIncentivesKeeper(incentivesKeeper types.IncentivesKeeper) {
	k.incentivesKeeper = incentivesKeeper
}

// GetNextPositionId returns the next position id.
func (k Keeper) GetNextPositionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPositionId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPositionId, &nextPositionId)
	return nextPositionId.Value
}

// SetNextPositionId sets next position Id.
func (k Keeper) SetNextPositionId(ctx sdk.Context, positionId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPositionId, &gogotypes.UInt64Value{Value: positionId})
}

// GetNextIncentiveRecordId returns the next incentive record ID.
func (k Keeper) GetNextIncentiveRecordId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextIncentiveRecord := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalIncentiveRecordId, &nextIncentiveRecord)
	return nextIncentiveRecord.Value
}

// SetNextIncentiveRecordId sets next incentive record ID.
func (k Keeper) SetNextIncentiveRecordId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalIncentiveRecordId, &gogotypes.UInt64Value{Value: id})
}

// Set the concentrated-liquidity listeners.
func (k *Keeper) SetListeners(listeners types.ConcentratedLiquidityListeners) *Keeper {
	if k.listeners != nil {
		panic("cannot set concentrated liquidity listeners twice")
	}

	k.listeners = listeners

	return k
}

// ValidatePermissionlessPoolCreationEnabled returns nil if permissionless pool creation in the module is enabled.
// Otherwise, returns an error.
func (k Keeper) ValidatePermissionlessPoolCreationEnabled(ctx sdk.Context) error {
	if !k.GetParams(ctx).IsPermissionlessPoolCreationEnabled {
		return types.ErrPermissionlessPoolCreationDisabled
	}
	return nil
}

// GetAuthorizedQuoteDenoms gets the authorized quote denoms from the poolmanager keeper.
// This method is meant to be used for getting access to x/poolmanager params
// for use in sim_msgs.go for the CL module.
func (k Keeper) GetAuthorizedQuoteDenoms(ctx sdk.Context) []string {
	return k.poolmanagerKeeper.GetParams(ctx).AuthorizedQuoteDenoms
}

// SetAuthorizedQuoteDenoms sets the authorized quote denoms in the poolmanager keeper.
// This method is meant to be used for getting access to x/poolmanager params
// for use in sim_msgs.go for the CL module.
func (k Keeper) SetAuthorizedQuoteDenoms(ctx sdk.Context, authorizedQuoteDenoms []string) {
	params := k.poolmanagerKeeper.GetParams(ctx)
	params.AuthorizedQuoteDenoms = authorizedQuoteDenoms
	k.poolmanagerKeeper.SetParams(ctx, params)
}
