package keeper

import (
	"authenticator/x/authenticator/types"
)

var _ types.QueryServer = Keeper{}
