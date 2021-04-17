use cosmwasm_std::{

}

use amm_std::{

}

// Math expression contracts only exposes calculation queries.
// Swap algorithm is fixed by the module side.
pub trait MathT {
    fn spot_price(&self, token_balance_in: Uint128, token_weight_in: Uint128, token_balance_out: Uint128, token_weight_out: Uint128, swap_fee: Uint128) -> StdResult<Uint128>
    fn out_given_in(&self, token_balance_in: Uint128, token_weight_in: Uint128, token_balance_out: Uint128, token_weight_out: Uint128, token_amount_in: Uint128, swap_fee: Uint128) -> StdResult<Uint128>
    fn in_given_out(&self, token_balance_in: Uint128, token_weight_in: Uint128, token_balance_out: Uint128, token_weight_out: Uint128, token_amount_out: Uint128, swap_fee: Uint128) -> StdResult<Uint128>
}

// AMM contracts should expose a struct that implements Pool trait.
// Entry point macro will expand using Singleton<PoolT>.
pub trait PoolT: MathT {
    fn lock(&self mut) -> StdResult<Response<()>>
    fn unlock(&self mut) -> StdResult<Response<()>>
    fn swap(&self mut, token_in: Coin, token_in_max: Uint128, token_out: Coin, token_out_max: Uint128, max_spot_price: Uint128) -> StdResult<Response<SwapResult>>
}

/*
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Record {
    pub weight: Uint128,
    pub token: Coin,
}

// TODO: find out more fluent way to access params by both module and contract
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct PoolParams {
    lock: bool,
    swap_fee: Uint128,
    exit_fee: Uint128,
    swap_fee_governor: str,
}

// Default PoolT implementation.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Pool {
    pub id: Uint128,
    pub params: PoolParams,
    pub total_weight: Uint128,
    pub total_share: Coin,
    pub records: Vec<Record>,
}

pub fn lock_pool(pool: Singleton<Pool>) -> StdResult<Response> {

}

pub fn swap(
    pool: Singleton<Pool>,
    token_in: Coin,
    token_in_max: Uint128,
    token_out: Coin,
    token_out_max: Uint128,
    max_spot_price: Uint128,
    ) -> StdResult<Response> {

    }

pub fn pool_params(pool: Singleton<Pool>) -> StdResult<Response> {

}

*/
