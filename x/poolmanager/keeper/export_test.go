package keeper

import (
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// SetPoolRoutesUnsafe sets the given routes to the poolmanager keeper
// to allow routing from a pool type to a certain swap module.
// For example, balancer -> gamm.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolRoutesUnsafe(routes map[types.PoolType]types.SwapI) {
	k.routes = routes
}
