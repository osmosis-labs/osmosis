package tradingtiers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v28/x/trading-tiers/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	stakingKeeper types.StakingKeeper
	txFeesKeeper  types.TxFeesKeeperI
	twapKeeper    types.TwapKeeperI
	epochsKeeper  types.EpochsKeeper

	cachedCurrentEpochNumber int64
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, stakingKeeper types.StakingKeeper, txFeesKeeper types.TxFeesKeeperI, twapKeeper types.TwapKeeperI, epochsKeeper types.EpochsKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:                 storeKey,
		paramSpace:               paramSpace,
		stakingKeeper:            stakingKeeper,
		txFeesKeeper:             txFeesKeeper,
		twapKeeper:               twapKeeper,
		epochsKeeper:             epochsKeeper,
		cachedCurrentEpochNumber: 0,
	}
}

// GetParams returns the total set of trading-tier parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of trading-tier parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// SetParam sets a specific trading-tier module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
}

// InitGenesis initializes the trading-tier module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the trading-tier module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

// SetTxFeesKeeper sets the txFeesKeeper
func (k *Keeper) SetTxFeesKeeper(txFeesKeeper types.TxFeesKeeperI) {
	k.txFeesKeeper = txFeesKeeper
}

// SetTwapKeeper sets the twapKeeper
func (k *Keeper) SetTwapKeeper(twapKeeper types.TwapKeeperI) {
	k.twapKeeper = twapKeeper
}
