package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/gamm module sentinel errors
var (
	ErrPoolNotFound   = sdkerrors.Register(ModuleName, 2, "pool not found")
	ErrMathApprox     = sdkerrors.Register(ModuleName, 3, "math approx error")
	ErrLimitExceed    = sdkerrors.Register(ModuleName, 4, "limit exceeded")
	ErrInvalidRequest = sdkerrors.Register(ModuleName, 5, "bad request")
	ErrNotBound       = sdkerrors.Register(ModuleName, 6, "not bound")
)
