package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bech32ibc module sentinel errors
var (
	ErrInvalidHRP     = sdkerrors.Register(ModuleName, 1, "Invalid HRP")
	ErrInvalidIBCData = sdkerrors.Register(ModuleName, 2, "Invalid IBC Data")
	ErrRecordNotFound = sdkerrors.Register(ModuleName, 3, "No record found for requested HRP")
	ErrNoNativeHrp    = sdkerrors.Register(ModuleName, 4, "No native prefix was set")
)
