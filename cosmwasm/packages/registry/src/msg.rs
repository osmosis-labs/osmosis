use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;
use schemars::JsonSchema;

use crate::RegistryError;

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

// Value does not implement JsonSchema, so we wrap it here. This can be removed
// if https://github.com/CosmWasm/serde-cw-value/pull/3 gets merged
#[derive(
    ::cosmwasm_schema::serde::Serialize,
    ::cosmwasm_schema::serde::Deserialize,
    ::std::clone::Clone,
    ::std::fmt::Debug,
    PartialEq,
    Eq,
)]
pub struct SerializableJson(pub serde_cw_value::Value);

impl JsonSchema for SerializableJson {
    fn schema_name() -> String {
        "JSON".to_string()
    }

    fn json_schema(_gen: &mut schemars::gen::SchemaGenerator) -> schemars::schema::Schema {
        schemars::schema::Schema::from(true)
    }
}

impl SerializableJson {
    pub fn as_value(&self) -> &serde_cw_value::Value {
        &self.0
    }
}

impl From<serde_cw_value::Value> for SerializableJson {
    fn from(value: serde_cw_value::Value) -> Self {
        Self(value)
    }
}

/// Information about which contract to call when the crosschain swap finishes
#[cw_serde]
pub struct Callback {
    pub contract: Addr,
    pub msg: SerializableJson,
}

impl Callback {
    pub fn try_string(&self) -> Result<String, RegistryError> {
        serde_json_wasm::to_string(self).map_err(|e| e.into())
    }

    pub fn to_json(&self) -> Result<SerializableJson, RegistryError> {
        Ok(SerializableJson(serde_json_wasm::from_str(
            &self.try_string()?,
        )?))
    }
}
