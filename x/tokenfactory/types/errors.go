package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrDenomExists              = sdkerrors.Register(ModuleName, 1, "denom already exists")
	ErrUnauthorized             = sdkerrors.Register(ModuleName, 2, "unauthorized account")
	ErrInvalidDenom             = sdkerrors.Register(ModuleName, 3, "invalid denom")
	ErrInvalidCreator           = sdkerrors.Register(ModuleName, 4, "invalid creator")
	ErrInvalidAuthorityMetadata = sdkerrors.Register(ModuleName, 5, "invalid authority metadata")
	ErrInvalidGenesis           = sdkerrors.Register(ModuleName, 6, "invalid genesis")
)
