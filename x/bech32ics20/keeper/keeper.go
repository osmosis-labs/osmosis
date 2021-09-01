package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/x/bech32ics20/types"
)

type (
	Keeper struct {
		bk                     bankkeeper.Keeper
		tk                     types.TransferKeeper
		hrpToChannelMapper     types.Bech32HrpToSourceChannelMap
		ics20TransferMsgServer types.ICS20TransferMsgServer
		cdc                    codec.Marshaler
		storeKey               sdk.StoreKey
		memKey                 sdk.StoreKey
	}
)

func NewKeeper(
	bk bankkeeper.Keeper,
	tk types.TransferKeeper,
	hrpToChannelMapper types.Bech32HrpToSourceChannelMap,
	cdc codec.Marshaler,
) *Keeper {
	return &Keeper{
		bk:                 bk,
		tk:                 tk,
		hrpToChannelMapper: hrpToChannelMapper,
		cdc:                cdc,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
