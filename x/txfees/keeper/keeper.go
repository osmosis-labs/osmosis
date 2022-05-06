package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v6/x/txfees/types"
)

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey

		spotPriceCalculator types.SpotPriceCalculator
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	spotPriceCalculator types.SpotPriceCalculator,
) Keeper {
	return Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		spotPriceCalculator: spotPriceCalculator,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeTokensStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, (types.FeeTokensStorePrefix))
}
