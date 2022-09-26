use cosmwasm_std::{Addr, Uint128};
use cw_utils::Duration;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub required_denom: String,
    pub purchase_price: Uint128,
    pub transfer_price: Uint128,
    pub annual_rent_bps: Uint128,
    pub owner_grace_period: Duration,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Register { name: String, years: Uint128 },
    Transfer { name: String, to: String },
    SetName { name: String },
    AddBid { name: String },
    // RemoveBid { name: String, amount: Uint128 },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    // ResolveAddress returns the current address that the name resolves to
    ResolveRecord { name: String },
    ReverseResolveRecord { address: Addr },
    Config {},
}

// We define a custom struct for each query response
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct ResolveRecordResponse {
    pub address: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct ReverseResolveRecordResponse {
    pub name: Option<String>,
}
