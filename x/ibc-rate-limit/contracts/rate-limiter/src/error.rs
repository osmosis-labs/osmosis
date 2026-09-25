use cosmwasm_std::{StdError, Timestamp, Uint256};
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("IBC Rate Limit exceeded for {channel}/{denom}. Tried to transfer {amount} which exceeds capacity on the '{quota_name}' quota ({used}/{max}). Try again after {reset:?}")]
    RateLimitExceded {
        channel: String,
        denom: String,
        amount: Uint256,
        quota_name: String,
        used: Uint256,
        max: Uint256,
        reset: Timestamp,
    },

    #[error("Quota {quota_id} not found for channel {channel_id}")]
    QuotaNotFound {
        quota_id: String,
        channel_id: String,
        denom: String,
    },
    #[error("{0}")]
    InvalidParameters(String),

    #[error("Channel {channel} has been blocked for denom {denom}")]
    ChannelBlocked { channel: String, denom: String },

    #[error("Arithmetic overflow: {0}")]
    Overflow(String),
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test] // The chain-side middleware keys on this prefix to tell a quota rejection from any other contract failure; keep it stable
    fn rate_limit_exceeded_message_keeps_the_marker_the_chain_matches() {
        let err = ContractError::RateLimitExceded {
            channel: "channel-0".to_string(),
            denom: "uosmo".to_string(),
            amount: Uint256::from(1_u32),
            quota_name: "daily".to_string(),
            used: Uint256::zero(),
            max: Uint256::zero(),
            reset: Timestamp::from_seconds(0),
        };
        assert!(err.to_string().starts_with("IBC Rate Limit exceeded for"));
    }
}
