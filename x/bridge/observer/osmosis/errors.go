package osmosis

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrGrpcConnection = errorsmod.Register(ModuleName, 1, "grpc connection error")
	ErrRpcClient      = errorsmod.Register(ModuleName, 2, "rpc client error")
	ErrSignTx         = errorsmod.Register(ModuleName, 3, "tx signing error")
	ErrBroadcastTx    = errorsmod.Register(ModuleName, 4, "tx broadcast error")
	ErrQuery          = errorsmod.Register(ModuleName, 5, "query error")
)
