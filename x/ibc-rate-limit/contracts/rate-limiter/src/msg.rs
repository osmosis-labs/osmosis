use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::{packet::Packet, state::rbac::Roles};

// PathMsg contains a channel_id and denom to represent a unique identifier within ibc-go, and a list of rate limit quotas.
// Unknown fields are rejected so a misspelled key in a governance proposal fails
// instead of silently configuring something other than what was intended.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
#[serde(deny_unknown_fields)]
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

// QuotaMsg represents a rate limiting Quota when sent as a wasm msg.
// Unknown fields are rejected: every field here is a bound, and a misspelled
// one would otherwise be dropped and leave the quota looser than intended.
//
// Each direction (send, recv) needs at least one bound: a percentage of the
// denom's channel value at the start of the window, an absolute amount in the
// denom's base units, or both. When both are set a transfer must satisfy both.
// Percentages above 100 are allowed only together with an absolute bound in
// the same direction.
//
// `send_recv: [30, 30]` as written by every proposal so far still parses. A
// `null` entry means no percentage bound in that direction, and the absolute
// fields may be omitted.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
#[serde(deny_unknown_fields)]
pub struct QuotaMsg {
    pub name: String,
    pub duration: u64,
    #[serde(default)]
    pub send_recv: (Option<u32>, Option<u32>),
    #[serde(default)]
    pub max_absolute_send: Option<Uint256>,
    #[serde(default)]
    pub max_absolute_recv: Option<Uint256>,
}

impl QuotaMsg {
    /// A quota bounded by percentages of the channel value in each direction
    pub fn new(name: &str, seconds: u64, send_percentage: u32, recv_percentage: u32) -> Self {
        QuotaMsg {
            name: name.to_string(),
            duration: seconds,
            send_recv: (Some(send_percentage), Some(recv_percentage)),
            max_absolute_send: None,
            max_absolute_recv: None,
        }
    }

    /// A quota bounded only by absolute amounts, in the denom's base units
    pub fn absolute(
        name: &str,
        seconds: u64,
        max_send: Option<Uint256>,
        max_recv: Option<Uint256>,
    ) -> Self {
        QuotaMsg {
            name: name.to_string(),
            duration: seconds,
            send_recv: (None, None),
            max_absolute_send: max_send,
            max_absolute_recv: max_recv,
        }
    }

    /// Adds absolute bounds, in the denom's base units, on top of the percentages
    pub fn with_absolute(mut self, max_send: Option<Uint256>, max_recv: Option<Uint256>) -> Self {
        self.max_absolute_send = max_send;
        self.max_absolute_recv = max_recv;
        self
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
    SetDenomRestrictions {
        denom: String,
        allowed_channels: Vec<String>,
    },
    UnsetDenomRestrictions {
        denom: String,
    },
    /// Grants a role to the given signer
    GrantRole {
        signer: String,
        /// full list of roles to grant the signer
        roles: Vec<Roles>,
    },
    /// Removes the role that has been granted to the signer
    RevokeRole {
        signer: String,
        /// fill list of roles to revoke from the signer
        roles: Vec<Roles>,
    },
    /// Replaces the quota identified by QuotaMsg::Name
    EditPathQuota {
        channel_id: String,
        denom: String,
        /// similar to ResetPathQuota, but QuotaMsg::Name is used as the quota_id
        quota: QuotaMsg,
    },
    /// Used to remove a message from the message queue to prevent execution
    RemoveMessage {
        message_id: String,
    },
    /// Used to change the timelock delay for newly submitted messages
    SetTimelockDelay {
        /// the address to apply the timelock delay to
        signer: String,
        hours: std::primitive::u64,
    },
    /// Permissionless message that anyone can invoke to trigger execution
    /// of queued messages that have passed the timelock delay
    ///
    /// If both count and message_ids are some, message_ids is used. If both are None returns an error
    ProcessMessages {
        /// number of queued messages to process, a value of 0 will attempt to process all queued messages
        count: Option<u64>,
        message_ids: Option<Vec<String>>,
    },
    /// Permissionless. Removes tracker entries that hold no quotas, which
    /// versions before 0.2.0 wrote for every (channel, denom) pair they saw.
    /// Bounded, so state left behind can never make a call exceed its gas.
    /// Returns the last key scanned so the next call can continue from it.
    PurgeEmptyPaths {
        start_after: Option<(String, String)>,
        limit: u32,
    },
    /// Permissionless. Removes pending-send records whose retention has ended
    /// (their windows are over and the packet could no longer be re-sent).
    /// Walks the retention index; `start_after` is the `last_key` of the
    /// previous call as (channel, retain_until in nanoseconds, content key).
    PurgeStaleSends {
        start_after: Option<(String, u64, String)>,
        limit: u32,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Vec<crate::state::rate_limit::RateLimit>)]
    GetQuotas { channel_id: String, denom: String },
    /// Returns a vector of all addresses that have been allocated one or more roles
    #[returns(Vec<String>)]
    GetRoleOwners,
    /// Returns a vector of all roles that have been granted to `owner`
    #[returns(Vec<crate::state::rbac::Roles>)]
    GetRoles { owner: String },
    /// Returns a vector of queued message id's
    #[returns(Vec<String>)]
    GetMessageIds,
    /// Returns the queued message matching id
    #[returns(crate::state::rbac::QueuedMessage)]
    GetMessage { id: String },
    /// Returns the restrictions for a given denom
    #[returns(Vec<String>)]
    GetDenomRestrictions { denom: String },
}

/// Messages only the chain sends. SendPacket is called before the packet is
/// committed, so its sequence is 0 and the destination is not filled in;
/// UndoSend carries the committed packet. The contract matches the two on the
/// content that is the same in both (see Packet::send_key).
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
    /// Sent by the chain when a send fails (error acknowledgement or timeout).
    /// Refunds the send to each quota that is still in the window the send
    /// was counted in and settles its record; a packet with no record refunds
    /// nothing.
    UndoSend { packet: Packet },
}

#[cw_serde]
pub struct MigrateMsg {}

impl ExecuteMsg {
    /// Given an ExecuteMsg variant returns the required RBAC role
    /// that must be held by the address which is invoking the message.
    ///
    /// If no RBAC role is required, returns None
    pub fn required_permission(&self) -> Option<Roles> {
        match self {
            Self::AddPath { .. } => Some(Roles::AddRateLimit),
            Self::RemovePath { .. } => Some(Roles::RemoveRateLimit),
            Self::ResetPathQuota { .. } => Some(Roles::ResetPathQuota),
            Self::SetDenomRestrictions { .. } => Some(Roles::ManageDenomRestrictions),
            Self::UnsetDenomRestrictions { .. } => Some(Roles::ManageDenomRestrictions),
            Self::GrantRole { .. } => Some(Roles::GrantRole),
            Self::RevokeRole { .. } => Some(Roles::RevokeRole),
            Self::EditPathQuota { .. } => Some(Roles::EditPathQuota),
            Self::RemoveMessage { .. } => Some(Roles::RemoveMessage),
            Self::SetTimelockDelay { .. } => Some(Roles::SetTimelockDelay),
            Self::ProcessMessages { .. } => None,
            Self::PurgeEmptyPaths { .. } => None,
            Self::PurgeStaleSends { .. } => None,
        }
    }
    /// Checks to see if the message type is able to skip queueing.
    ///
    /// This is limited to the permissionless housekeeping messages
    pub fn skip_queue(&self) -> bool {
        matches!(
            self,
            Self::ProcessMessages { .. }
                | Self::PurgeEmptyPaths { .. }
                | Self::PurgeStaleSends { .. }
        )
    }
}
