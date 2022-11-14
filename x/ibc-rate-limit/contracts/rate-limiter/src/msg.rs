use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[cfg(test)]
use cosmwasm_std::Uint256;

use crate::packet::Packet;

// PathMsg contains a channel_id and denom to represent a unique identifier within ibc-go, and a list of rate limit quotas
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct PathMsg {
    pub channel_id: String,
    pub denom: String,
    pub quotas: Vec<QuotaMsg>,
}

impl PathMsg {
    pub fn new(
        channel: impl Into<String>,
        denom: impl Into<String>,
        quotas: Vec<QuotaMsg>,
    ) -> Self {
        PathMsg {
            channel_id: channel.into(),
            denom: denom.into(),
            quotas,
        }
    }
}

// QuotaMsg represents a rate limiting Quota when sent as a wasm msg
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
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
#[cw_serde]
pub struct InstantiateMsg {
    pub gov_module: Addr,
    pub ibc_module: Addr,
    pub paths: Vec<PathMsg>,
}

/// The caller (IBC module) is responsible for correctly calculating the funds
/// being sent through the channel
#[cw_serde]
pub enum ExecuteMsg {
    AddPath {
        channel_id: String,
        denom: String,
        quotas: Vec<QuotaMsg>,
    },
    RemovePath {
        channel_id: String,
        denom: String,
    },
    ResetPathQuota {
        channel_id: String,
        denom: String,
        quota_id: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Vec<crate::state::RateLimit>)]
    GetQuotas { channel_id: String, denom: String },
}

#[cw_serde]
pub enum SudoMsg {
    SendPacket {
        packet: Packet,
        #[cfg(test)]
        channel_value_mock: Option<Uint256>,
    },
    RecvPacket {
        packet: Packet,
        #[cfg(test)]
        channel_value_mock: Option<Uint256>,
    },
    UndoSend {
        packet: Packet,
    },
}

#[cw_serde]
pub enum MigrateMsg {}
