package keeper

import (
	"github.com/c-osmosis/osmosis/x/claim/types"
)

var _ types.QueryServer = Keeper{}
