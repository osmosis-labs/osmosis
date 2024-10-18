package types

import errorsmod "cosmossdk.io/errors"

var (
	DefaultCodespace                = ModuleName
	ErrContractNotFound             = errorsmod.Register(DefaultCodespace, 2, "contract with given address not found")
	ErrCallbackHeightNotInFuture    = errorsmod.Register(DefaultCodespace, 3, "callback request height is not in the future")
	ErrUnauthorized                 = errorsmod.Register(DefaultCodespace, 4, "sender not authorized to register callback")
	ErrCallbackNotFound             = errorsmod.Register(DefaultCodespace, 5, "callback with given job id does not exist for given height")
	ErrInsufficientFees             = errorsmod.Register(DefaultCodespace, 6, "insufficient fees to register callback")
	ErrCallbackExists               = errorsmod.Register(DefaultCodespace, 7, "callback with given job id already exists for given height")
	ErrCallbackHeightTooFarInFuture = errorsmod.Register(DefaultCodespace, 8, "callback request height is too far in the future")
	ErrBlockFilled                  = errorsmod.Register(DefaultCodespace, 9, "block filled with max capacity of callbacks")
)
