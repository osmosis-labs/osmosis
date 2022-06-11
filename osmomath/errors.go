package osmomath

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

const codespace = "osmomath"

var (
	ErrInvalidBaseMustBeUnderTwo = sdkerrors.Register(codespace, 1, "base must be lesser than two")
	ErrInvalidBaseMustBePositive = sdkerrors.Register(codespace, 2, "base must be positive")
)
