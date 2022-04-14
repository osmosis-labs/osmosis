package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage module storage.
type Keeper struct {
	cdc        codec.Codec
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace

	ak authkeeper.AccountKeeper
	bk types.BankKeeper
	sk types.StakingKeeper
	dk types.DistrKeeper
	ek types.EpochKeeper
	lk types.LockupKeeper
	gk types.GammKeeper
	ik types.IncentivesKeeper

	lms types.LockupMsgServer
}

// NewKeeper returns an instance of Keeper.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, dk types.DistrKeeper, ek types.EpochKeeper, lk types.LockupKeeper, gk types.GammKeeper, ik types.IncentivesKeeper, lms types.LockupMsgServer) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		bk:         bk,
		sk:         sk,
		dk:         dk,
		ek:         ek,
		lk:         lk,
		gk:         gk,
		ik:         ik,

		lms: lms,
	}
}

// Logger returns a logger instance.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
