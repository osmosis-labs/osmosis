package cosmwasmpool

import (
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (k *Keeper) ConvertToCosmwasmPool(poolI poolmanagertypes.PoolI) (types.CosmWasmExtension, error) {
	return k.convertToCosmwasmPool(poolI)
}
