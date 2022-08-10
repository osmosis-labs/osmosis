use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// Initialize the contract with the address of the IBC module and any existing channels.
/// Only the ibc module is allowed to execute actions on this contract
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub gov_module: Addr,
    pub ibc_module: Addr,
    pub channel_quotas: Vec<(String, u32)>,
}

/// The caller (IBC module) is responsibble for correctly calculating the funds
/// being sent through the channel
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    SendPacket {
        channel_id: String,
        channel_value: u128,
        funds: u128,
    },
    RecvPacket {
        channel_id: String,
        channel_value: u128,
        funds: u128,
    },
    AddChannel {},    // TODO: Who is allowed to do this?
    RemoveChannel {}, // TODO: Who is allowed to do this?
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {}
