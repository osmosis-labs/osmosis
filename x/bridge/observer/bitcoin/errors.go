package bitcoin

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidCfg           = errorsmod.Register(ModuleNameObserver, 1, "invalid configuration")
	ErrRpcClient            = errorsmod.Register(ModuleNameObserver, 2, "rpc client error")
	ErrTxInvalidDestination = errorsmod.Register(ModuleNameObserver, 3, "invalid destination in tx")
	ErrBlockFetch           = errorsmod.Register(ModuleNameObserver, 4, "failed to fetch block")
	ErrBlockUnavailable     = errorsmod.Register(ModuleNameObserver, 5, "block not available")
)
