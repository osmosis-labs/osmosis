package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/twap/types"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	transientKey *sdk.TransientStoreKey

	paramSpace paramtypes.Subspace

	ammkeeper types.AmmInterface
}

const pruneEpochIdentifier = "day"

func NewKeeper(storeKey sdk.StoreKey, transientKey *sdk.TransientStoreKey, paramSpace paramtypes.Subspace, ammKeeper types.AmmInterface) *Keeper {
	return &Keeper{storeKey: storeKey, transientKey: transientKey, paramSpace: paramSpace, ammkeeper: ammKeeper}
}

// TODO: make this read from a parameter, or hardcode it.
func (k *Keeper) PruneEpochIdentifier(ctx sdk.Context) string {
	return pruneEpochIdentifier
}
