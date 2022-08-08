package keeper

import (
	"github.com/osmosis-labs/osmosis/v10/x/streamswap/types"
)

var _ types.MsgServer = Keeper{}
