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

    #[error(
        "receiver prefix for {receiver} must match the bech32 prefix of the destination chain {chain}"
    )]
    InvalidReceiverPrefix { receiver: String, chain: String },

    // Registry loading errors
    #[error("contract alias does not exist: {alias:?}")]
    AliasDoesNotExist { alias: String },

    #[error("no authorized address found for source chain: {source_chain:?}")]
    ChainAuthorizedAddressDoesNotExist { source_chain: String },

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

    #[error("bech32 prefix does not exist for chain: {chain}")]
    Bech32PrefixDoesNotExist { chain: String },
}

impl From<RegistryError> for StdError {
    fn from(e: RegistryError) -> Self {
        match e {
            RegistryError::Std(e) => e,
            _ => StdError::generic_err(e.to_string()),
        }
    }
}
