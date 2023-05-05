package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrRateLimitExceeded = errorsmod.Register(ModuleName, 2, "rate limit exceeded")
	ErrBadMessage        = errorsmod.Register(ModuleName, 3, "bad message")
	ErrContractError     = errorsmod.Register(ModuleName, 4, "contract error")
)
