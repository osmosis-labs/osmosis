package keeper

import (
	"github.com/osmosis-labs/osmosis/v10/x/launchpad/types"
)

var _ types.MsgServer = Keeper{}
