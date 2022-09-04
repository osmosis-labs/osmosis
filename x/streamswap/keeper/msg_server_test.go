package keeper

import (
	"github.com/osmosis-labs/osmosis/v11/x/streamswap/types"
)

var _ types.MsgServer = Keeper{}
