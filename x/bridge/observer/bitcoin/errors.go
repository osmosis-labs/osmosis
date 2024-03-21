package bitcoin

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidCfg       = errorsmod.Register(ModuleNameObserver, 1, "invalid configuration")
	ErrRpcClient        = errorsmod.Register(ModuleNameObserver, 2, "rpc client error")
	ErrBlockUnavailable = errorsmod.Register(ModuleNameObserver, 3, "block not available")
)
