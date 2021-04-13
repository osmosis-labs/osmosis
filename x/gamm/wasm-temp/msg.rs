use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Binary, Coin};

use crate::math::Uint128;

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Swap { 
        pool_id: Uint128, // For future extensibility, fixed to 0
        token_in: Coin, 
        token_in_max: Uint128, 
        token_out: Coin, 
        token_out_max: Uint128,
        max_spot_price: Uint128,
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    SpotPrice {
        pool_id: Uin128,
        token_balance_in: Uint128,
        token_weight_in: Uint128,
        token_balance_out: Uint128,
        token_weight_out: Uint128,
        swap_fee: Uint128,
    }
    OutGivenIn {
        pool_id: Uint128,
        token_balance_in: Uint128,
        token_weight_in: Uint128,
        token_balance_out: Uint128,
        token_weight_out: Uint128,
        swap_fee: Uint128,
    }
    InGivenOut {
        pool_id: Uint128,
        token_balance_in: Uint128,
        token_weight_in: Uint128,
        token_balance_out: Uint128,
        token_weight_out: Uint128,
        swap_fee: Uint128,
    }
}

