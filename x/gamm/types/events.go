package types

const (
	TypeEvtPoolJoined   = "gamm_pool_joined"
	TypeEvtPoolExited   = "gamm_pool_exited"
	TypeEvtPoolCreated  = "gamm_pool_created"
	TypeEvtTokenSwapped = "gamm_token_swapped"

	AttributeValueCategory = ModuleName
	AttributeKeyPoolId     = "pool_id"
	AttributeKeySwapFee    = "swap_fee"
	AttributeKeyTokensIn   = "tokens_in"
	AttributeKeyTokensOut  = "tokens_out"
)
