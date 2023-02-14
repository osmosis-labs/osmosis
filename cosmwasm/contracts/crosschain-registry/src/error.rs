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

    #[error("chain channel link already exists: {source_chain:?} -> {destination_chain:?}")]
    ChainChannelLinkAlreadyExists {
        source_chain: String,
        destination_chain: String,
    },

    #[error("chain channel link does not exist: {source_chain:?} -> {destination_chain:?}")]
    ChainChannelLinkDoesNotExist {
        source_chain: String,
        destination_chain: String,
    },

    #[error("asset map link already exists: {native_denom:?} -> {destination_chain:?}")]
    AssetMapLinkAlreadyExists {
        native_denom: String,
        destination_chain: String,
    },

    #[error("asset map link does not exist: {native_denom:?} -> {destination_chain:?}")]
    AssetMapLinkDoesNotExist {
        native_denom: String,
        destination_chain: String,
    },
}
