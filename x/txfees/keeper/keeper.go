package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/txfees/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	poolManager         types.PoolManager
	spotPriceCalculator types.SpotPriceCalculator
	protorevKeeper      types.ProtorevKeeper
	distributionKeeper  types.DistributionKeeper
}

var _ types.TxFeesKeeper = (*Keeper)(nil)

func NewKeeper(
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	storeKey sdk.StoreKey,
	poolManager types.PoolManager,
	spotPriceCalculator types.SpotPriceCalculator,
	protorevKeeper types.ProtorevKeeper,
	distributionKeeper types.DistributionKeeper,
) Keeper {
	return Keeper{
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		storeKey:            storeKey,
		poolManager:         poolManager,
		spotPriceCalculator: spotPriceCalculator,
		protorevKeeper:      protorevKeeper,
		distributionKeeper:  distributionKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.FeeTokensStorePrefix)
}
