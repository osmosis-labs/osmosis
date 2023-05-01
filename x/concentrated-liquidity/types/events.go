package types

const (
	TypeEvtCreatePosition         = "create_position"
	TypeEvtWithdrawPosition       = "withdraw_position"
	TypeEvtTotalCollectFees       = "total_collect_fees"
	TypeEvtCollectFees            = "collect_fees"
	TypeEvtTotalCollectIncentives = "total_collect_incentives"
	TypeEvtCollectIncentives      = "collect_incentives"
	TypeEvtCreateIncentive        = "create_incentive"
	TypeEvtFungifyChargedPosition = "fungify_charged_position"

	AttributeValueCategory         = ModuleName
	AttributeKeyPositionId         = "position_id"
	AttributeKeyPoolId             = "pool_id"
	AttributeAmount0               = "amount0"
	AttributeAmount1               = "amount1"
	AttributeKeySwapFee            = "swap_fee"
	AttributeKeyTokensIn           = "tokens_in"
	AttributeKeyTokensOut          = "tokens_out"
	AttributeKeyForfeitedTokens    = "forfeited_tokens"
	AttributeLiquidity             = "liquidity"
	AttributeJoinTime              = "join_time"
	AttributeLowerTick             = "lower_tick"
	AttributeUpperTick             = "upper_tick"
	TypeEvtPoolJoined              = "pool_joined"
	TypeEvtPoolExited              = "pool_exited"
	TypeEvtTokenSwapped            = "token_swapped"
	AttributeIncentiveDenom        = "incentive_denom"
	AttributeIncentiveAmount       = "incentive_amount"
	AttributeIncentiveEmissionRate = "incentive_emission_rate"
	AttributeIncentiveStartTime    = "incentive_start_time"
	AttributeIncentiveMinUptime    = "incentive_min_uptime"
	AttributeInputPositionIds      = "input_position_ids"
	AttributeOutputPositionId      = "output_position_id"
)
