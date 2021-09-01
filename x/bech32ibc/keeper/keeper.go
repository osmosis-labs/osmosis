package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
)

type (
	Keeper struct {
		channelKeeper types.ChannelKeeper

		cdc      codec.Marshaler
		storeKey sdk.StoreKey

		tk types.TransferKeeper
	}
)

func NewKeeper(
	channelKeeper types.ChannelKeeper,
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	tk types.TransferKeeper,
) *Keeper {
	return &Keeper{
		channelKeeper: channelKeeper,
		cdc:           cdc,
		storeKey:      storeKey,
		tk:            tk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
