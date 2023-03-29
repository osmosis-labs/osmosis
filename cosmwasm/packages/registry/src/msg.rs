use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetAddressFromAliasResponse)]
    GetAddressFromAlias { contract_alias: String },

    #[returns(GetChannelFromChainPairResponse)]
    GetChannelFromChainPair {
        source_chain: String,
        destination_chain: String,
    },

    #[returns(GetDestinationChainFromSourceChainViaChannelResponse)]
    GetDestinationChainFromSourceChainViaChannel {
        on_chain: String,
        via_channel: String,
    },

    #[returns(QueryGetBech32PrefixFromChainNameResponse)]
    GetBech32PrefixFromChainName { chain_name: String },

    #[returns(QueryGetChainNameFromBech32PrefixResponse)]
    GetChainNameFromBech32Prefix { prefix: String },

    #[returns(crate::proto::QueryDenomTraceResponse)]
    GetDenomTrace { ibc_denom: String },
}

// Response for GetAddressFromAlias query
#[cw_serde]
pub struct GetAddressFromAliasResponse {
    pub address: String,
}

// Response for GetChannelFromChainPair query
#[cw_serde]
pub struct GetChannelFromChainPairResponse {
    pub channel_id: String,
}

// Response for GetDestinationChainFromSourceChainViaChannel query
#[cw_serde]
pub struct GetDestinationChainFromSourceChainViaChannelResponse {
    pub destination_chain: String,
}

// Response for GetBech32PrefixFromChainName query
#[cw_serde]
pub struct QueryGetBech32PrefixFromChainNameResponse {
    pub bech32_prefix: String,
}

// Response for GetChainNameFromBech32Prefix query
#[cw_serde]
pub struct QueryGetChainNameFromBech32PrefixResponse {
    pub chain_name: String,
}
