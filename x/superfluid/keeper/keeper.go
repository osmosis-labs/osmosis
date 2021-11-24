package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper provides a way to manage module storage
type Keeper struct {
	cdc        codec.Marshaler
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
}

// NewKeeper returns an instance of Keeper
func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, dk types.DistrKeeper, ek types.EpochKeeper, lk types.LockupKeeper, gk types.GammKeeper, ik types.IncentivesKeeper) *Keeper {
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
	}
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
