package types

const (
	TypeEvtCreatePosition   = "create_position"
	TypeEvtWithdrawPosition = "withdraw_position"
	TypeEvtCollectFees      = "collect_fees"

	AttributeValueCategory = ModuleName
	AttributeKeyPoolId     = "pool_id"
	AttributeAmount0       = "amount0"
	AttributeAmount1       = "amount1"
	AttributeKeySwapFee    = "swap_fee"
	AttributeKeyTokensIn   = "tokens_in"
	AttributeKeyTokensOut  = "tokens_out"
	AttributeLiquidity     = "liquidity"
	AttributeLowerTick     = "lower_tick"
	AttributeUpperTick     = "upper_tick"
	TypeEvtPoolJoined      = "pool_joined"
	TypeEvtPoolExited      = "pool_exited"
	TypeEvtPoolCreated     = "pool_created"
	TypeEvtTokenSwapped    = "token_swapped"
)
