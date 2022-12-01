package swaprouter

import (
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// SetPoolRoutesUnsafe sets the given routes to the swaprouter keeper
// to allow routing from a pool type to a certain swap module.
// For example, balancer -> gamm.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolRoutesUnsafe(routes map[types.PoolType]types.SwapI) {
	k.routes = routes
}
