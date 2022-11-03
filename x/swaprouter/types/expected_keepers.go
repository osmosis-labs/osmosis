package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// GammKeeper defines the expected interface needed for swaprouter module
type GammKeeper interface {
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.TraditionalAmmInterface, error)

	GetNextPoolId(ctx sdk.Context) uint64
}
