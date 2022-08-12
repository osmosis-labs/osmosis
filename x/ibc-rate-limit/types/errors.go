package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	RateLimitExceededMsg = "rate limit exceeded"
	ErrRateLimitExceeded = sdkerrors.Register(ModuleName, 2, RateLimitExceededMsg)
	BadMessage           = sdkerrors.Register(ModuleName, 3, "bad message")
)
