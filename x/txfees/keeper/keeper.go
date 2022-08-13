package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/txfees/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	accountKeeper             types.AccountKeeper
	bankKeeper                types.BankKeeper
	epochKeeper               types.EpochKeeper
	gammKeeper                types.GammKeeper
	spotPriceCalculator       types.SpotPriceCalculator
	feeCollectorName          string
	nonNativeFeeCollectorName string
}

func NewKeeper(
	cdc codec.Codec,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	epochKeeper types.EpochKeeper,
	storeKey sdk.StoreKey,
	gammKeeper types.GammKeeper,
	spotPriceCalculator types.SpotPriceCalculator,
	feeCollectorName string,
	nonNativeFeeCollectorName string,
) Keeper {
	return Keeper{
		accountKeeper:             accountKeeper,
		bankKeeper:                bankKeeper,
		epochKeeper:               epochKeeper,
		storeKey:                  storeKey,
		gammKeeper:                gammKeeper,
		spotPriceCalculator:       spotPriceCalculator,
		feeCollectorName:          feeCollectorName,
		nonNativeFeeCollectorName: nonNativeFeeCollectorName,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.FeeTokensStorePrefix)
}
