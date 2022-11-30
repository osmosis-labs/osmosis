package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v13/x/ibc-hooks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		storeKey sdk.StoreKey

		paramSpace paramtypes.Subspace
	}
)

// NewKeeper returns a new instance of the x/ibchooks keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
