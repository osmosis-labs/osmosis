package keeper

import (
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
