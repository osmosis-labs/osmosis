package keeper

import (
	"github.com/c-osmosis/osmosis/x/lockup/types"
)

var _ types.QueryServer = Keeper{}
