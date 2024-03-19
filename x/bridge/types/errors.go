package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidAsset          = errorsmod.Register(ModuleName, 2, "invalid asset")
	ErrInvalidAssets         = errorsmod.Register(ModuleName, 3, "invalid assets")
	ErrInvalidAssetStatus    = errorsmod.Register(ModuleName, 4, "invalid asset status")
	ErrInvalidParams         = errorsmod.Register(ModuleName, 5, "invalid params")
	ErrInvalidDenom          = errorsmod.Register(ModuleName, 6, "invalid denom")
	ErrInvalidSourceChain    = errorsmod.Register(ModuleName, 7, "invalid source chain")
	ErrInvalidSigners        = errorsmod.Register(ModuleName, 8, "invalid signers")
	ErrCantChangeAssetStatus = errorsmod.Register(ModuleName, 9, "can't change asset status")
	ErrCantCreateAsset       = errorsmod.Register(ModuleName, 10, "can't create asset")
	ErrTokenfactory          = errorsmod.Register(ModuleName, 11, "tokenfactory error")
)
