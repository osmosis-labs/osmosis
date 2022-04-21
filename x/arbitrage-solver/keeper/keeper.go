package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/arbitrage-solver/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramSpace paramtypes.Subspace

	// keepers
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	gammKeeper    types.GammKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	gammKeeper types.GammKeeper) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: paramSpace,
		// keepers
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		gammKeeper:    gammKeeper,
	}
}

func (k *Keeper) RecordSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	// Store this in state
	// TODO: Reference docs, or we can talk separately, about how to put this state
}
