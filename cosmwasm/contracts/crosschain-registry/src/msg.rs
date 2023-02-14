use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    /// Contract Registry
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

    /// Chain Channel Registry
    // Set a source_chain->destination_chain->channel_id map in the registry
    SetChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
        // The channel id
        channel_id: String,
    },
    // Change an existing source_chain->destination_chain->channel_id map in the registry
    ChangeChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
        // The new channel id
        new_channel_id: String,
    },
    // Remove an existing source_chain->destination_chain->channel_id map in the registry
    RemoveChainChannelLink {
        // The source chain
        source_chain: String,
        // The destination chain
        destination_chain: String,
    },

    /// Asset Mapping Registry
    // Set a native_denom mapping to a destination chain in the registry
    SetAssetMapping {
        // The native denom
        native_denom: String,
        // The destination chain
        destination_chain: String,
        // The native_denom on the destination chain
        destination_chain_denom: String,
    },
    // Change an existing native_denom mapping to a destination chain in the registry
    ChangeAssetMapping {
        // The native denom
        native_denom: String,
        // The destination chain
        destination_chain: String,
        // The new native_denom on the destination chain
        new_destination_chain_denom: String,
    },
    // Remove an existing native_denom mapping to a destination chain in the registry
    RemoveAssetMapping {
        // The native denom
        native_denom: String,
        // The destination chain
        destination_chain: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetAddressFromAliasResponse)]
    GetAddressFromAlias { contract_alias: String },

    #[returns(GetChainChannelLinkResponse)]
    GetChainChannelLink {
        source_chain: String,
        destination_chain: String,
    },

    #[returns(GetAssetMappingResponse)]
    GetAssetMapping {
        native_denom: String,
        destination_chain: String,
    },
}

// Response for GetAddressFromAlias query
#[cw_serde]
pub struct GetAddressFromAliasResponse {
    pub address: String,
}

// Response for GetChainChannelLink query
#[cw_serde]
pub struct GetChainChannelLinkResponse {
    pub channel_id: String,
}

// Response for GetAssetMapping query
#[cw_serde]
pub struct GetAssetMappingResponse {
    pub destination_chain_denom: String,
}
