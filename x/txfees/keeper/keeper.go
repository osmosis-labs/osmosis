package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	oracleKeeper       types.OracleKeeper
	marketKeeper       types.MarketKeeper
	distributionKeeper types.DistributionKeeper
	consensusKeeper    types.ConsensusKeeper
	dataDir            string

	paramSpace paramtypes.Subspace
}

var _ types.TxFeesKeeper = (*Keeper)(nil)

func NewKeeper(
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	storeKey storetypes.StoreKey,
	marketKeeper types.MarketKeeper,
	oracleKeeper types.OracleKeeper,
	distributionKeeper types.DistributionKeeper,
	consensusKeeper types.ConsensusKeeper,
	dataDir string,
	paramSpace paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		storeKey:           storeKey,
		distributionKeeper: distributionKeeper,
		marketKeeper:       marketKeeper,
		oracleKeeper:       oracleKeeper,
		consensusKeeper:    consensusKeeper,
		dataDir:            dataDir,
		paramSpace:         paramSpace,
	}
}

// GetParams returns the total set of txfees parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of txfees parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// SetParam sets a specific txfees module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetConsParams returns the current consensus parameters from the consensus params store.
func (k Keeper) GetConsParams(ctx sdk.Context) (*consensustypes.QueryParamsResponse, error) {
	return k.consensusKeeper.Params(ctx, &consensustypes.QueryParamsRequest{})
}
