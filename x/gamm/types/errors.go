package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/gamm module sentinel errors
var (
	ErrPoolNotFound = sdkerrors.Register(ModuleName, 2, "pool not found")
	ErrMathApprox   = sdkerrors.Register(ModuleName, 3, "math approx error")
	ErrLimitExceed  = sdkerrors.Register(ModuleName, 4, "limit exceeded")

	ErrNotBound      = sdkerrors.Register(ModuleName, 100, "ERR_NOT_BOUND")
	ErrMaxInRatio    = sdkerrors.Register(ModuleName, 101, "ERR_MAX_IN_RATIO")
	ErrMaxOutRatio   = sdkerrors.Register(ModuleName, 101, "ERR_MAX_OUT_RATIO")
	ErrBadLimitPrice = sdkerrors.Register(ModuleName, 102, "ERR_BAD_LIMIT_PRICE")
	ErrLimitIn       = sdkerrors.Register(ModuleName, 103, "ERR_LIMIT_IN")
	ErrLimitOut      = sdkerrors.Register(ModuleName, 103, "ERR_LIMIT_OUT")
	ErrLimitPrice    = sdkerrors.Register(ModuleName, 105, "ERR_LIMIT_PRICE")
	//ErrMathApprox    = sdkerrors.Register(ModuleName, 104, "ERR_MATH_APPROX")
)
