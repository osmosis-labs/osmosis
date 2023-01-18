use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid Pool Route: {reason:?}")]
    InvalidPoolRoute { reason: String },

    #[error("Failed Swap: {reason:?}")]
    FailedSwap { reason: String },

    #[error("Insufficient Funds")]
    InsufficientFunds {},

    #[error("Query Error: {val:?}")]
    QueryError { val: String },

    #[error("TwapNotFound: Twap price not found for {denom} in {sell_denom} via pool {pool_id}")]
    TwapNotFound {
        denom: String,
        sell_denom: String,
        pool_id: u64,
    },

    #[error(
        "InvalidTwap: Invalid twap value received from the chain: {twap}. Should be a Decimal"
    )]
    InvalidTwapString { twap: String },

    #[error("InvalidTwap: Invalid value for twap price: {operation}.")]
    InvalidTwapOperation { operation: String },

    #[error("Custom Error: {val:?}")]
    CustomError { val: String },
    // Add any other custom errors you like here.
    // Look at https://docs.rs/thiserror/1.0.21/thiserror/ for details.
}
