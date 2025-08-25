use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("unauthorized")]
    Unauthorized {},

    #[error("insufficient funds sent")]
    InsufficientFunds {},

    #[error("invalid affiliate bps, must be between 0 and 10_000 inclusive")]
    InvalidAffiliateBps {},

    #[error("swap failed: {reason}")]
    FailedSwap { reason: String },
}
