use cosmwasm_schema::cw_serde;

use crate::execute;

#[cw_serde]
pub struct InstantiateMsg {
    pub owner: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    // Contract Registry
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

    UnwrapCoin {
        receiver: String,
        into_chain: Option<String>,
        #[serde(default = "String::new")]
        with_memo: String,
    },
}

// Import the queries from the package to avoid cyclic dependencies
pub use registry::msg::QueryMsg;
pub use registry::msg::{
    GetAddressFromAliasResponse, GetChannelFromChainPairResponse,
    GetDestinationChainFromSourceChainViaChannelResponse,
    QueryGetBech32PrefixFromChainNameResponse,
};
