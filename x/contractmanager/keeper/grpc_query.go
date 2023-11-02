package keeper

import (
	"github.com/osmosis-labs/osmosis/v20/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
