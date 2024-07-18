package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	crypto "github.com/cometbft/cometbft/crypto/tmhash"
)

// Oracle Errors
var (
	ErrInvalidExchangeRate   = errorsmod.Register(ModuleName, 2, "invalid exchange rate")
	ErrNoPrevote             = errorsmod.Register(ModuleName, 3, "no prevote")
	ErrNoVote                = errorsmod.Register(ModuleName, 4, "no vote")
	ErrNoVotingPermission    = errorsmod.Register(ModuleName, 5, "unauthorized voter")
	ErrInvalidHash           = errorsmod.Register(ModuleName, 6, "invalid hash")
	ErrInvalidHashLength     = errorsmod.Register(ModuleName, 7, fmt.Sprintf("invalid hash length; should equal %d", crypto.TruncatedSize))
	ErrVerificationFailed    = errorsmod.Register(ModuleName, 8, "hash verification failed")
	ErrRevealPeriodMissMatch = errorsmod.Register(ModuleName, 9, "reveal period of submitted vote do not match with registered prevote")
	ErrInvalidSaltLength     = errorsmod.Register(ModuleName, 10, "invalid salt length; should be 1~4")
	ErrNoAggregatePrevote    = errorsmod.Register(ModuleName, 11, "no aggregate prevote")
	ErrNoAggregateVote       = errorsmod.Register(ModuleName, 12, "no aggregate vote")
	ErrNoTobinTax            = errorsmod.Register(ModuleName, 13, "no tobin tax")
	ErrUnknownDenom          = errorsmod.Register(ModuleName, 14, "unknown denom")
)
