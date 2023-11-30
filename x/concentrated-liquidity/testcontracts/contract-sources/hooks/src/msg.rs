use cosmwasm_schema::cw_serde;
use cosmwasm_std::Coin;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum SudoMsg{
    BeforeCreatePosition {
        pool_id: u64,
        owner: String,
        tokens_provided: Vec<Coin>,
        amount_0_min: String,
        amount_1_min: String,
        lower_tick: i64,
        upper_tick: i64,
    },
    AfterCreatePosition {
        pool_id: u64,
        owner: String,
        tokens_provided: Vec<Coin>,
        amount_0_min: String,
        amount_1_min: String,
        lower_tick: i64,
        upper_tick: i64,
    },
    BeforeWithdrawPosition {
        pool_id: u64,
        owner: String,
        position_id: u64,
        amount_to_withdraw: String,
    },
    AfterWithdrawPosition {
        pool_id: u64,
        owner: String,
        position_id: u64,
        amount_to_withdraw: String,
    },
    BeforeSwapExactAmountIn {
        pool_id: u64,
        sender: String,
        token_in: Coin,
        token_out_denom: String,
        token_out_min_amount: String,
        spread_factor: String,
    },
    AfterSwapExactAmountIn {
        pool_id: u64,
        sender: String,
        token_in: Coin,
        token_out_denom: String,
        token_out_min_amount: String,
        spread_factor: String,
    },
    BeforeSwapExactAmountOut {
        pool_id: u64,
        sender: String,
        token_in_denom: String,
        token_in_max_amount: String,
        token_out: Coin,
        spread_factor: String,
    },
    AfterSwapExactAmountOut {
        pool_id: u64,
        sender: String,
        token_in_denom: String,
        token_in_max_amount: String,
        token_out: Coin,
        spread_factor: String,
    },
}

