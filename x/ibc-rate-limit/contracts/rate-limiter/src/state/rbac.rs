use cosmwasm_std::{Addr, Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use std::cmp;

use crate::{
    msg::{ExecuteMsg, QuotaMsg},
    ContractError,
};

/// Roles defines the available permissions that can be assigned to addresses as part of the RBAC system
#[derive(Serialize, Deserialize, Clone, Copy, Debug, PartialEq, Eq, JsonSchema, PartialOrd, Ord, Hash)]
pub enum Roles {
    /// Has the ability to add a new rate limit
    AddRateLimit,
    /// Has the ability to complete remove a configured rate limit
    RemoveRateLimit,
    /// Has the ability to reset tracked quotas
    ResetPathQuota,
    /// Has the ability to edit existing quotas
    EditPathQuota,
    /// Has the ability to grant roles to an address
    GrantRole,
    /// Has the ability to revoke granted roles to an address
    RevokeRole,
    /// Has the ability to remove queued messages
    RemoveMessage,
    /// Has the ability to alter timelock delay's
    SetTimelockDelay,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct QueuedMessage {
    /// the message that submitted to the contract after a sucessful governance proposal
    pub message: ExecuteMsg,
    /// the time which the message was processed by the contract
    pub submitted_at: Timestamp,
    /// the timelock delay that was in place when the message was queued for execution
    pub timelock_delay: u64,
    /// Constructed using format!("{}_{}", Env::BlockInfo::Height Env::Transaction::Index)
    ///
    /// Can be used to remove a message from the queue without processing it
    pub message_id: String,
}

impl Roles {
    /// helper function that returns a vec containing all variants of the Roles enum
    pub fn all_roles() -> Vec<Roles> {
        vec![
            Roles::AddRateLimit,
            Roles::RemoveRateLimit,
            Roles::ResetPathQuota,
            Roles::EditPathQuota,
            Roles::GrantRole,
            Roles::RevokeRole,
            Roles::RemoveMessage,
            Roles::SetTimelockDelay,
        ]
    }
}


#[cfg(test)]
mod test {
    use super::*;
    #[test]
    fn test_all_roles() {
        let roles = Roles::all_roles();
        assert!(
            roles.contains(&Roles::AddRateLimit)
        );
        assert!(
            roles.contains(&Roles::RemoveRateLimit)
        );
        assert!(
            roles.contains(&Roles::ResetPathQuota)
        );
        assert!(
            roles.contains(&Roles::EditPathQuota)
        );
        assert!(
            roles.contains(&Roles::GrantRole)
        );
        assert!(
            roles.contains(&Roles::RevokeRole)
        );
        assert!(
            roles.contains(&Roles::RemoveMessage)
        );
        assert!(
            roles.contains(&Roles::SetTimelockDelay)
        );
    }
}