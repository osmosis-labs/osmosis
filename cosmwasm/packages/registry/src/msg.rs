use std::collections::BTreeMap;

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde_json_wasm::from_str;

use crate::registry::Memo;
use crate::utils::stringify;
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

    #[returns(bool)]
    HasPacketForwarding { chain: String },

    #[returns(QueryAliasForDenomPathResponse)]
    GetAliasForDenomPath { denom_path: String },

    #[returns(QueryDenomPathForAliasResponse)]
    GetDenomPathForAlias { alias: String },
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

// Response for GetAliasForDenomPath query
#[cw_serde]
pub struct QueryAliasForDenomPathResponse {
    pub alias: String,
}

// Response for GetDenomPathForAlias query
#[cw_serde]
pub struct QueryDenomPathForAliasResponse {
    pub denom_path: String,
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
    pub const fn empty() -> Self {
        SerializableJson(serde_cw_value::Value::Map(BTreeMap::new()))
    }

    pub fn is_empty(&self) -> bool {
        match &self.0 {
            serde_cw_value::Value::Map(m) => m.is_empty(),
            _ => true,
        }
    }

    pub fn as_value(&self) -> &serde_cw_value::Value {
        &self.0
    }

    /// Merge two [`SerializableJson`] instances together. Fail in case
    /// the same top-level key is found twice, or in case any of the two
    /// JSON structs are not objects.
    pub fn merge(self, other: SerializableJson) -> Result<Self, RegistryError> {
        let mut first_map = match self.0 {
            serde_cw_value::Value::Map(m) => m,
            serde_cw_value::Value::Unit => BTreeMap::new(),
            json => {
                return Err(RegistryError::InvalidJson {
                    error: "invalid json: expected an object".to_string(),
                    json: stringify(&json)?,
                })
            }
        };
        let second_map = match other.0 {
            serde_cw_value::Value::Map(m) => m,
            serde_cw_value::Value::Unit => BTreeMap::new(),
            json => {
                return Err(RegistryError::InvalidJson {
                    error: "invalid json: expected an object".to_string(),
                    json: stringify(&json)?,
                })
            }
        };

        for (key, value) in second_map {
            if first_map.insert(key, value).is_some() {
                return Err(RegistryError::DuplicateKeyError);
            }
        }

        return Ok(SerializableJson(serde_cw_value::Value::Map(first_map)));
    }
}

impl From<serde_cw_value::Value> for SerializableJson {
    fn from(value: serde_cw_value::Value) -> Self {
        Self(value)
    }
}

impl TryFrom<Memo> for SerializableJson {
    type Error = RegistryError;

    fn try_from(memo: Memo) -> Result<Self, RegistryError> {
        let value = serde_cw_value::to_value(&memo)?;
        Ok(Self(value))
    }
}

impl TryFrom<String> for SerializableJson {
    type Error = RegistryError;

    fn try_from(value: String) -> Result<Self, RegistryError> {
        Ok(Self(from_str(&value)?))
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
