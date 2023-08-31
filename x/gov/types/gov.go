package types

import (
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// NOTE: Since we don't use the sdk gov module anymore it's necessary
// to add these proposal types to the allowed proposal in this module.
func init() {
	RegisterProposalType(distrtypes.ProposalTypeCommunityPoolSpend)
	RegisterProposalTypeCodec(&distrtypes.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal")
	RegisterProposalType(paramtypes.ProposalTypeChange)
	RegisterProposalTypeCodec(&paramtypes.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal")
}
