use thiserror::Error;

use cosmwasm_std::StdError;

/// Never is a placeholder to ensure we don't return any errors
#[derive(Error, Debug)]
pub enum Never {}

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("InvalidAuthenticatorParams")]
    InvalidAuthenticatorParams {},

    #[error("QueryError: {val}")]
    QueryError { val: String },

    #[error("InvalidPoolRoute: {reason}")]
    InvalidPoolRoute { reason: String },

    #[error("InvalidTwapString: {twap}")]
    InvalidTwapString { twap: String },

    #[error("InvalidTwapOperation: {operation}")]
    InvalidTwapOperation { operation: String },

    #[error("InvalidTwapDenom: {denom}")]
    TwapNotFound {
        denom: String,
        sell_denom: String,
        pool_id: u64,
    },
}
