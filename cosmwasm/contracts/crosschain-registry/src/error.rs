use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("contract alias already exists: {alias:?}")]
    AliasAlreadyExists { alias: String },

    #[error("contract alias does not exist: {current_alias:?}")]
    AliasDoesNotExist { current_alias: String },
}
