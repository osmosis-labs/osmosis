package types

const (
	TypeEvtPoolJoined     = "pool_joined"
	TypeEvtPoolExited     = "pool_exited"
	TypeEvtPoolCreated    = "pool_created"
	TypeEvtTokenSwapped   = "token_swapped"
	TypeEvtMultiHopAmtIn  = "multihop_tokenin"
	TypeEvtMultiHopAmtOut = "multihop_tokenout"

	AttributeValueCategory = ModuleName
	AttributeKeyPoolId     = "pool_id"
	AttributeKeySwapFee    = "swap_fee"
	AttributeKeyTokensIn   = "tokens_in"
	AttributeKeyTokensOut  = "tokens_out"
)
