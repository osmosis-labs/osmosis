package keeper

import (
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

var _ types.QueryServer = Keeper{}
