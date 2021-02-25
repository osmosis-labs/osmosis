package keeper

import (
	"github.com/c-osmosis/osmosis/x/incentives/types"
)

var _ types.QueryServer = Keeper{}
