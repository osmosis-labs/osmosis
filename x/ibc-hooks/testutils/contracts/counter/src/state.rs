use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Coin};
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Counter {
    pub count: i32,
    pub total_funds: Vec<Coin>,
    pub owner: Addr,
}

pub const COUNTERS: Map<Addr, Counter> = Map::new("state");
