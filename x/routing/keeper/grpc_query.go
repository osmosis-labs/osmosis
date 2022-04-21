package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/routing/types"
)

var _ types.QueryServer = Keeper{}
