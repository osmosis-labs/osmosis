use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("invalid reply id: {id}")]
    InvalidReplyID { id: u64 },

    #[error("invalid receiver: {receiver}")]
    InvalidReceiver { receiver: String },

    #[error("invalid json: {error}. Got: {json}")]
    InvalidJson { error: String, json: String },

    #[error("invalid memo: {error}. Got: {memo}")]
    InvalidMemo { error: String, memo: String },

    #[error("contract locked: {msg}")]
    ContractLocked { msg: String },

    #[error("failed swap: {msg}")]
    FailedSwap { msg: String },

    #[error("failed ibc transfer: {msg:?}")]
    FailedIBCTransfer { msg: String },

    #[error("custom error: {msg:?}")]
    CustomError { msg: String },
}
