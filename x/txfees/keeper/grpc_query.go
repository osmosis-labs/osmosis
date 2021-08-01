package keeper

import (
	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

var _ types.QueryServer = Keeper{}
