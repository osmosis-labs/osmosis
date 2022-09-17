package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v12/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	ck types.CommunityPoolKeeper
	ek types.EpochKeeper
	lk types.LockupKeeper
	gk types.GammKeeper
	ik types.IncentivesKeeper

	lms types.LockupMsgServer
}

var _ govtypes.StakingKeeper = (*Keeper)(nil)

// NewKeeper returns an instance of Keeper.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, dk types.CommunityPoolKeeper, ek types.EpochKeeper, lk types.LockupKeeper, gk types.GammKeeper, ik types.IncentivesKeeper, lms types.LockupMsgServer) *Keeper {
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
		ck:         dk,
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
