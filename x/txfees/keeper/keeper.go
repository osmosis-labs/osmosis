package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/txfees/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

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
) Keeper {
	return Keeper{
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		storeKey:           storeKey,
		poolManager:        poolManager,
		protorevKeeper:     protorevKeeper,
		distributionKeeper: distributionKeeper,
		consensusKeeper:    consensusKeeper,
		dataDir:            dataDir,
	}
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
