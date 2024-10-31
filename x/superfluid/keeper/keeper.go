package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage module storage.
type Keeper struct {
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	ak   authkeeper.AccountKeeper
	bk   types.BankKeeper
	sk   types.StakingKeeper
	ck   types.CommunityPoolKeeper
	ek   types.EpochKeeper
	lk   types.LockupKeeper
	gk   types.GammKeeper
	ik   types.IncentivesKeeper
	clk  types.ConcentratedKeeper
	pmk  types.PoolManagerKeeper
	vspk types.ValSetPreferenceKeeper

	lms types.LockupMsgServer
}

var _ govtypes.StakingKeeper = (*Keeper)(nil)

// NewKeeper returns an instance of Keeper.
func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, dk types.CommunityPoolKeeper, ek types.EpochKeeper, lk types.LockupKeeper, gk types.GammKeeper, ik types.IncentivesKeeper, lms types.LockupMsgServer, clk types.ConcentratedKeeper, pmk types.PoolManagerKeeper, vspk types.ValSetPreferenceKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
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
		clk:        clk,
		pmk:        pmk,
		vspk:       vspk,

		lms: lms,
	}
}

// Logger returns a logger instance.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
