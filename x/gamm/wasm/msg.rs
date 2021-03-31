use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Binary, Coin};

use crate::math::Uint128;

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Swap { 
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
    SpotPrice {}
    OutGivenInt {}
    InGivenOut {}
}

