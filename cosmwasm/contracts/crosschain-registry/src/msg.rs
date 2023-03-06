use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::execute;
use crate::exports::MultiHopDenom;

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
}

#[cw_serde]
//#[derive(QueryResponses)]
pub enum QueryMsg {
    //#[returns(GetAddressFromAliasResponse)]
    GetAddressFromAlias {
        contract_alias: String,
    },

    //#[returns(GetChannelFromChainPairResponse)]
    GetChannelFromChainPair {
        source_chain: String,
        destination_chain: String,
    },

    //#[returns(GetDestinationChainFromSourceChainViaChannelResponse)]
    GetDestinationChainFromSourceChainViaChannel {
        on_chain: String,
        via_channel: String,
    },

    //#[returns(crate::helpers::QueryDenomTraceResponse)]
    GetDenomTrace {
        ibc_denom: String,
    },

    //#[returns(crate::helpers::QueryDenomTraceResponse)]
    UnwrapDenom {
        ibc_denom: String,
    },

    // this is for testing purposes only
    //#[returns(crosschain_swaps::ibc:::IbcMessage)]
    UnwrapDenomIntoMsg {
        ibc_denom: String,
    },
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

// Response for UnwrapDenom query
#[cw_serde]
pub struct UnwrapDenomResponse {
    pub hops: MultiHopDenom,
}
