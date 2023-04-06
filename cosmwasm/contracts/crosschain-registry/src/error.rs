use cosmwasm_std::StdError;
use registry::RegistryError;
use thiserror::Error;

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

    #[error("authorized address already exists for source chain: {source_chain:?}")]
    ChainAuthorizedAddressAlreadyExists { source_chain: String },

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
