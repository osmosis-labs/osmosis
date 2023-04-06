package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	poolManager         types.PoolManager
	spotPriceCalculator types.SpotPriceCalculator
}

var _ types.TxFeesKeeper = (*Keeper)(nil)

func NewKeeper(
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	storeKey sdk.StoreKey,
	poolManager types.PoolManager,
	spotPriceCalculator types.SpotPriceCalculator,
) Keeper {
	return Keeper{
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
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
