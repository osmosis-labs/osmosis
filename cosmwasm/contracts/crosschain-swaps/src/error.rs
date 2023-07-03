use cosmwasm_std::StdError;
use registry::RegistryError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    JsonSerialization(#[from] serde_json_wasm::ser::Error),

    #[error("{0}")]
    JsonDeserialization(#[from] serde_json_wasm::de::Error),

    #[error("{0}")]
    ValueSerialization(#[from] serde_cw_value::SerializerError),

    #[error("{0}")]
    RegistryError(#[from] RegistryError),

    #[error("{0}")]
    Payment(#[from] cw_utils::PaymentError),

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

    #[error("prefix already exists: {prefix:?}")]
    PrefixAlreadyExists { prefix: String },

    #[error("prefix does not exist: {prefix:?}")]
    PrefixDoesNotExist { prefix: String },

    #[error("prefix not disabled: {prefix:?}")]
    PrefixNotDisabled { prefix: String },

    #[error("custom error: {msg:?}")]
    CustomError { msg: String },
}
