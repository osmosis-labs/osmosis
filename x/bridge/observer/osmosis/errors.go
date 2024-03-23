package osmosis

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrGrpcConnection = errorsmod.Register(ModuleNameClient, 1, "grpc connection error")
	ErrSignTx         = errorsmod.Register(ModuleNameClient, 2, "tx signing error")
	ErrBroadcastTx    = errorsmod.Register(ModuleNameClient, 3, "tx broadcast error")
	ErrQuery          = errorsmod.Register(ModuleNameClient, 4, "query error")
)
