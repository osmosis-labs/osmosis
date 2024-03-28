package bitcoin

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidCfg       = errorsmod.Register(ModuleName, 1, "invalid configuration")
	ErrRpcClient        = errorsmod.Register(ModuleName, 2, "rpc client error")
	ErrBlockUnavailable = errorsmod.Register(ModuleName, 3, "block not available")
)
