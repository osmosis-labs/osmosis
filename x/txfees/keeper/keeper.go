package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	accountKeeper       types.AccountKeeper
	epochKeeper         types.EpochKeeper
	bankKeeper          types.BankKeeper
	poolManager         types.PoolManager
	spotPriceCalculator types.SpotPriceCalculator
}

var _ types.TxFeesKeeper = (*Keeper)(nil)

func NewKeeper(
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	epochKeeper types.EpochKeeper,
	bankKeeper types.BankKeeper,
	poolManager types.PoolManager,
	spotPriceCalculator types.SpotPriceCalculator,
) Keeper {
	return Keeper{
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		epochKeeper:         epochKeeper,
		storeKey:            storeKey,
		poolManager:         poolManager,
		spotPriceCalculator: spotPriceCalculator,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.FeeTokensStorePrefix)
}
