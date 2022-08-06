package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrRateLimitExceeded = sdkerrors.Register(ModuleName, 2, "rate limit exceeded")
)
