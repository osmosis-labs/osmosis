package keeper

import (
	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
)

var _ types.QueryServer = Keeper{}
