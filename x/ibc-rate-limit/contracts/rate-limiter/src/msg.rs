use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Channel {
    pub name: String,
    pub quotas: Vec<QuotaMsg>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct QuotaMsg {
    pub name: String,
    pub duration: u64,
    pub send_recv: (u32, u32),
}

impl QuotaMsg {
    pub fn new(name: &str, seconds: u64, send_percentage: u32, recv_percentage: u32) -> Self {
        QuotaMsg {
            name: name.to_string(),
            duration: seconds,
            send_recv: (send_percentage, recv_percentage),
        }
    }
}

/// Initialize the contract with the address of the IBC module and any existing channels.
/// Only the ibc module is allowed to execute actions on this contract
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub gov_module: Addr,
    pub ibc_module: Addr,
    pub channels: Vec<Channel>,
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
    AddChannel {
        channel_id: String,
        quotas: Vec<QuotaMsg>,
    },
    RemoveChannel {
        channel_id: String,
    },
    ResetChannelQuota {
        channel_id: String,
        quota_id: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {}
