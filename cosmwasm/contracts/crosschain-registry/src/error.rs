use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum RegistryError {
    #[error("{0}")]
    Std(#[from] StdError),

    // Validation errors
    #[error("Invalid channel id: {0}")]
    InvalidChannelId(String),

    #[error("error {action} {addr}")]
    Bech32Error {
        action: String,
        addr: String,
        #[source]
        source: bech32::Error,
    },

    #[error("serialization error: {error}")]
    SerialiaztionError { error: String },

    #[error("denom {denom:?} is not an IBC denom")]
    InvalidIBCDenom { denom: String },

    #[error("No deom trace found for: {denom:?}")]
    NoDenomTrace { denom: String },

    #[error("Invalid denom trace: {error}")]
    InvalidDenomTrace { error: String },

    #[error("Invalid path {path:?} for denom {denom:?}")]
    InvalidDenomTracePath { path: String, denom: String },

    #[error("Invalid transfer port {port:?}")]
    InvalidTransferPort { port: String },

    #[error("Invalid multihop length {length:?}. Must be >={min}")]
    InvalidMultiHopLengthMin { length: usize, min: usize },

    #[error("Invalid multihop length {length:?}. Must be <={max}")]
    InvalidMultiHopLengthMax { length: usize, max: usize },

    // Registry loading errors
    #[error("contract alias does not exist: {alias:?}")]
    AliasDoesNotExist { alias: String },

    #[error("chain channel link does not exist: {source_chain:?} -> {destination_chain:?}")]
    ChainChannelLinkDoesNotExist {
        source_chain: String,
        destination_chain: String,
    },

    #[error("channel chain link does not exist: {channel_id:?} on {source_chain:?} -> chain")]
    ChannelChainLinkDoesNotExist {
        channel_id: String,
        source_chain: String,
    },

    #[error("channel chain link does not exist: {channel_id:?} on {source_chain:?} -> chain")]
    ChannelToChainChainLinkDoesNotExist {
        channel_id: String,
        source_chain: String,
    },
    #[error("native denom link does not exist: {native_denom:?}")]
    NativeDenomLinkDoesNotExist { native_denom: String },
}

impl From<RegistryError> for StdError {
    fn from(e: RegistryError) -> Self {
        StdError::generic_err(e.to_string())
    }
}

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    RegistryError(#[from] RegistryError),

    #[error("{0}")]
    Payment(#[from] cw_utils::PaymentError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("contract alias already exists: {alias:?}")]
    AliasAlreadyExists { alias: String },

    #[error("chain channel link already exists: {source_chain:?} -> {destination_chain:?}")]
    ChainToChainChannelLinkAlreadyExists {
        source_chain: String,
        destination_chain: String,
    },

    #[error("channel chain link already exists: {channel_id:?} -> {source_chain:?}")]
    ChannelToChainChainLinkAlreadyExists {
        channel_id: String,
        source_chain: String,
    },

    #[error("native denom link already exists: {native_denom:?}")]
    NativeDenomLinkAlreadyExists { native_denom: String },

    #[error("input not valid: {message:?}")]
    InvalidInput { message: String },

    #[error("missing field: {field:?}")]
    MissingField { field: String },

    #[error("custom error: {msg:?}")]
    CustomError { msg: String },
}
