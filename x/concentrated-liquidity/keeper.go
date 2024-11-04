package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
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
	contractKeeper       types.ContractKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, gammKeeper types.GAMMKeeper, poolIncentivesKeeper types.PoolIncentivesKeeper, incentivesKeeper types.IncentivesKeeper, lockupKeeper types.LockupKeeper, communityPoolKeeper types.CommunityPoolKeeper, contractKeeper types.ContractKeeper, paramSpace paramtypes.Subspace) *Keeper {
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
		contractKeeper:       contractKeeper,
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

// SetParam sets a specific concentrated-liquidity module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
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

// Set the contract keeper.
func (k *Keeper) SetContractKeeper(contractKeeper types.ContractKeeper) {
	k.contractKeeper = contractKeeper
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

// IsPermissionlessPoolCreationEnabled returns true if permissionless pool creation in the module is enabled.
// Otherwise, returns false
func (k Keeper) IsPermissionlessPoolCreationEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).IsPermissionlessPoolCreationEnabled
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

func (k Keeper) GetWhitelistedAddresses(ctx sdk.Context) []string {
	return k.GetParams(ctx).UnrestrictedPoolCreatorWhitelist
}
