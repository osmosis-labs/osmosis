package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/txfees/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	poolManager         types.PoolManager
	spotPriceCalculator types.SpotPriceCalculator
	protorevKeeper      types.ProtorevKeeper
	distributionKeeper  types.DistributionKeeper
	dataDir             string
}

var _ types.TxFeesKeeper = (*Keeper)(nil)

func NewKeeper(
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	storeKey storetypes.StoreKey,
	poolManager types.PoolManager,
	spotPriceCalculator types.SpotPriceCalculator,
	protorevKeeper types.ProtorevKeeper,
	distributionKeeper types.DistributionKeeper,
	dataDir string,
) Keeper {
	return Keeper{
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		storeKey:            storeKey,
		poolManager:         poolManager,
		spotPriceCalculator: spotPriceCalculator,
		protorevKeeper:      protorevKeeper,
		distributionKeeper:  distributionKeeper,
		dataDir:             dataDir,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.FeeTokensStorePrefix)
}
