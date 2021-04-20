package keeper

import (
	"github.com/c-osmosis/osmosis/x/epochs/types"
)

var _ types.QueryServer = Keeper{}
