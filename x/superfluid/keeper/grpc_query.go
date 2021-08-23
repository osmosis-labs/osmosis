package keeper

import (
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

var _ types.QueryServer = Keeper{}
