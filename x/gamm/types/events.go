package types

const (
	TypeEvtPoolJoined    = "pool_joined"
	TypeEvtPoolExited    = "pool_exited"
	TypeEvtPoolCreated   = "pool_created"
	TypeEvtTokenSwapped  = "token_swapped"
	TypeEvtMigrateShares = "migrate_shares"

	AttributeValueCategory     = ModuleName
	AttributeKeyPoolId         = "pool_id"
	AttributeKeyPoolIdEntering = "pool_id_entering"
	AttributeKeyPoolIdLeaving  = "pool_id_leaving"
	AttributeKeySwapFee        = "swap_fee"
	AttributeKeyTokensIn       = "tokens_in"
	AttributeKeyTokensOut      = "tokens_out"
)
