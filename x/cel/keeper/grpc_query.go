package keeper

import (
	"github.com/c-osmosis/osmosis/x/cel/types"
)

var _ types.QueryServer = Keeper{}
