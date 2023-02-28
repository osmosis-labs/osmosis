use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::exports::MultiHopDenom;

#[cw_serde]
pub struct InstantiateMsg {
    pub owner: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    // Contract Registry

    // Set a alias->address map in the registry
    SetContractAlias {
        // The alias to be used for the contract
        contract_alias: String,
        // The address of the contract
        contract_address: String,
    },
    // Change an existing alias->address map in the registry
    ChangeContractAlias {
        // The alias currently used by the contract
        current_contract_alias: String,
        // The new alias to be used bythe contract
        new_contract_alias: String,
    },
    // Remove an existing alias->address map in the registry
    RemoveContractAlias {
        // The alias to be removed
        contract_alias: String,
    },

    // Chain to Chain Channel Registry
    SetChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
        // The channel id
        channel_id: String,
    },

    ChangeChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
        // The new channel id
        new_channel_id: Option<String>,
        // The new destination chain
        new_destination_chain: Option<String>,
    },

    RemoveChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
    },

    // Osmosis Denom Registry

    // Set a native_denom->ibc_denom map in the registry
    SetNativeDenomToIbcDenom {
        // The native denom
        native_denom: String,
        // The ibc denom
        ibc_denom: String,
    },

    // Change an existing native_denom->ibc_denom map in the registry
    ChangeNativeDenomToIbcDenom {
        // The native denom
        native_denom: String,
        // The new ibc denom
        new_ibc_denom: String,
    },

    // Remove an existing native_denom->ibc_denom map in the registry
    RemoveNativeDenomToIbcDenom {
        // The native denom
        native_denom: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetAddressFromAliasResponse)]
    GetAddressFromAlias { contract_alias: String },

    #[returns(GetChainToChainChannelLinkResponse)]
    GetChainToChainChannelLink {
        source_chain: String,
        destination_chain: String,
    },

    #[returns(GetChainToChainChannelLinkResponse)]
    GetConnectedChainViaChannel {
        on_chain: String,
        via_channel: String,
    },

    #[returns(crate::helpers::QueryDenomTraceResponse)]
    GetDenomTrace { ibc_denom: String },

    #[returns(crate::helpers::QueryDenomTraceResponse)]
    UnwrapDenom { ibc_denom: String },
}

// Response for GetAddressFromAlias query
#[cw_serde]
pub struct GetAddressFromAliasResponse {
    pub address: String,
}

// Response for GetChainToChainChannelLink query
#[cw_serde]
pub struct GetChainToChainChannelLinkResponse {
    pub channel_id: String,
}

// Response for UnwrapDenom query
#[cw_serde]
pub struct UnwrapDenomResponse {
    pub hops: MultiHopDenom,
}
