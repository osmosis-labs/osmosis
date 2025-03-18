use cosmwasm_schema::cw_serde;

use crate::execute;

#[cw_serde]
pub struct InstantiateMsg {
    pub owner: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    // Contract Registry
    ModifyDenomAlias {
        operations: Vec<execute::DenomAliasInput>,
    },

    ModifyContractAlias {
        operations: Vec<execute::ContractAliasInput>,
    },

    // Chain to Chain Channel Registry
    ModifyChainChannelLinks {
        operations: Vec<execute::ConnectionInput>,
    },

    // Bech32 Prefix Registry
    ModifyBech32Prefixes {
        operations: Vec<execute::ChainToBech32PrefixInput>,
    },

    // Authorized Address Registry
    ModifyAuthorizedAddresses {
        operations: Vec<execute::AuthorizedAddressInput>,
    },

    // Transfer Ownership
    TransferOwnership {
        new_owner: String,
    },

    // Add PFM to the registry
    #[serde(rename = "propose_pfm")]
    ProposePFM {
        chain: String,
    },
    #[serde(rename = "validate_pfm")]
    ValidatePFM {
        chain: String,
    },

    UnwrapCoin {
        receiver: String,
        into_chain: Option<String>,
        #[serde(default = "String::new")]
        with_memo: String,
        #[serde(default = "String::new")]
        final_memo: String,
    },
}

// Import the queries from the package to avoid cyclic dependencies
pub use registry::msg::QueryMsg;
pub use registry::msg::{
    GetAddressFromAliasResponse, GetChannelFromChainPairResponse,
    GetDestinationChainFromSourceChainViaChannelResponse,
    QueryGetBech32PrefixFromChainNameResponse,
};

#[cw_serde]
pub enum IBCLifecycleComplete {
    #[serde(rename = "ibc_ack")]
    IBCAck {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
        /// String encoded version of the ack as seen by OnAcknowledgementPacket(..)
        ack: String,
        /// Whether an ack is a success of failure according to the transfer spec
        success: bool,
    },
    #[serde(rename = "ibc_timeout")]
    IBCTimeout {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
    },
}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    #[serde(rename = "ibc_lifecycle_complete")]
    IBCLifecycleComplete(IBCLifecycleComplete),
}
