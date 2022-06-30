package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

var _ types.MsgServer = Keeper{}
