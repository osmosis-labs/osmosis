package types

const (
	TypeEvtPoolJoined    = "pool_joined"
	TypeEvtPoolExited    = "pool_exited"
	TypeEvtTokenSwapped  = "token_swapped"
	TypeEvtMigrateShares = "migrate_shares"

	AttributeValueCategory     = ModuleName
	AttributeKeyPoolId         = "pool_id"
	AttributeKeyPoolIdEntering = "pool_id_entering"
	AttributeKeyPoolIdLeaving  = "pool_id_leaving"
	AttributeKeySwapFee        = "swap_fee"
	AttributeKeyTokensIn       = "tokens_in"
	AttributeKeyTokensOut      = "tokens_out"

	AttributePositionId = "position_id"
	AttributeAmount0    = "amount0"
	AttributeAmount1    = "amount1"
	AttributeLiquidity  = "liquidity"
)
