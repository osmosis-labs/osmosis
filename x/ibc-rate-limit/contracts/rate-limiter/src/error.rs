use cosmwasm_std::{StdError, Timestamp};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("IBC Rate Limit exceded for channel {channel:?} and denom {denom:?}. Tried to transfer {amount} and at most {channel_value} is allowed.Try again after {reset:?}")]
    RateLimitExceded {
        channel: String,
        denom: String,
        amount: u128,
        channel_value: u128,
        reset: Timestamp,
    },

    #[error("Quota {quota_id} not found for channel {channel_id}")]
    QuotaNotFound {
        quota_id: String,
        channel_id: String,
        denom: String,
    },
}
