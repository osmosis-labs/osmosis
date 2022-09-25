package types

// x/gov module sentinel errors
// In a subsequent release, we should switch all usages of legacygovtypes to use these.
var (
	ErrUnknownProposal         = "unknown proposal"
	ErrInactiveProposal        = "inactive proposal"
	ErrAlreadyActiveProposal   = "proposal already active"
	ErrInvalidProposalContent  = "invalid proposal content"
	ErrInvalidProposalType     = "invalid proposal type"
	ErrInvalidVote             = "invalid vote option"
	ErrInvalidGenesis          = "invalid genesis state"
	ErrNoProposalHandlerExists = "no handler exists for proposal type"
	ErrMinDepositTooSmall      = "minimum deposit is too small"
)
