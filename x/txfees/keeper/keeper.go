package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/x/txfees/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	poolManager        types.PoolManager
	protorevKeeper     types.ProtorevKeeper
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
	poolManager types.PoolManager,
	protorevKeeper types.ProtorevKeeper,
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
		poolManager:        poolManager,
		protorevKeeper:     protorevKeeper,
		distributionKeeper: distributionKeeper,
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

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.FeeTokensStorePrefix)
}

// GetParamsNoUnmarshal returns the current consensus parameters from the consensus params store as raw bytes.
func (k Keeper) GetParamsNoUnmarshal(ctx sdk.Context) []byte {
	return k.consensusKeeper.GetParamsNoUnmarshal(ctx)
}

// UnmarshalParamBytes unmarshals the consensus params bytes to the consensus params type.
func (k Keeper) UnmarshalParamBytes(ctx sdk.Context, bz []byte) (*tmproto.ConsensusParams, error) {
	return k.consensusKeeper.UnmarshalParamBytes(ctx, bz)
}
