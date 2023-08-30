package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/gov module sentinel errors
var (
	ErrUnknownProposal         = sdkerrors.Register(ModuleName, 102, "unknown proposal")
	ErrInactiveProposal        = sdkerrors.Register(ModuleName, 103, "inactive proposal")
	ErrAlreadyActiveProposal   = sdkerrors.Register(ModuleName, 104, "proposal already active")
	ErrInvalidProposalContent  = sdkerrors.Register(ModuleName, 105, "invalid proposal content")
	ErrInvalidProposalType     = sdkerrors.Register(ModuleName, 106, "invalid proposal type")
	ErrInvalidVote             = sdkerrors.Register(ModuleName, 107, "invalid vote option")
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 108, "invalid genesis state")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 109, "no handler exists for proposal type")
	ErrMinDepositTooSmall      = sdkerrors.Register(ModuleName, 110, "minimum deposit is too small")
)
