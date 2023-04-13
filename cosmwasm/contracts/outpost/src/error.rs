use cosmwasm_std::{Coin, StdError};
use registry::RegistryError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    RegistryError(#[from] RegistryError),

    #[error("{0}")]
    Payment(#[from] cw_utils::PaymentError),

    #[error("{0}")]
    CrosschainSwaps(#[from] crosschain_swaps::ContractError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid Json: could not serialize msg into json: {error}")]
    InvalidJson { error: String },

    #[error("Invalid Funds: Should be exactly one token. Got: {funds:?}")]
    InvalidFunds { funds: Vec<Coin> },

    #[error("SwapAmountTooHigh: got {received}, max allowed: {max}")]
    SwapAmountTooHigh { received: u128, max: u128 },

    #[error("Invalid Crosschain Swpas Contract: {contract}")]
    InvalidCrosschainSwapsContract { contract: String },

    #[error("Custom Error val: {val:?}")]
    CustomError { val: String },
    // Add any other custom errors you like here.
    // Look at https://docs.rs/thiserror/1.0.21/thiserror/ for details.
}
