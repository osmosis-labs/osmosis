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
#[repr(transparent)]
pub struct SerializableJson(
    #[serde(deserialize_with = "deserialize_cw_value")] serde_cw_value::Value,
);

impl JsonSchema for SerializableJson {
    fn schema_name() -> String {
        "JSON".to_string()
    }

    fn json_schema(_gen: &mut schemars::gen::SchemaGenerator) -> schemars::schema::Schema {
        schemars::schema::Schema::from(true)
    }
}

impl SerializableJson {
    pub fn into_value(self) -> serde_cw_value::Value {
        self.0
    }

    pub fn new(mut value: serde_cw_value::Value) -> Result<Self, RegistryError> {
        flatten_cw_value(&mut value);
        match &value {
            serde_cw_value::Value::Map(_) => Ok(Self(value)),
            serde_cw_value::Value::Unit | serde_cw_value::Value::Option(None) => Ok(Self::empty()),
            json => Err(RegistryError::InvalidJson {
                error: "invalid json: expected an object".to_string(),
                json: stringify(json)?,
            }),
        }
    }

    pub fn empty() -> Self {
        Self(serde_cw_value::Value::Map(BTreeMap::new()))
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
            serde_cw_value::Value::Unit | serde_cw_value::Value::Option(None) => BTreeMap::new(),
            json => {
                return Err(RegistryError::InvalidJson {
                    error: "invalid json: expected an object".to_string(),
                    json: stringify(&json)?,
                })
            }
        };
        let second_map = match other.0 {
            serde_cw_value::Value::Map(m) => m,
            serde_cw_value::Value::Unit | serde_cw_value::Value::Option(None) => BTreeMap::new(),
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

        Ok(SerializableJson(serde_cw_value::Value::Map(first_map)))
    }
}

impl From<SerializableJson> for serde_cw_value::Value {
    fn from(SerializableJson(value): SerializableJson) -> Self {
        value
    }
}

impl TryFrom<serde_cw_value::Value> for SerializableJson {
    type Error = RegistryError;

    fn try_from(value: serde_cw_value::Value) -> Result<Self, RegistryError> {
        Self::new(value)
    }
}

impl TryFrom<Memo> for SerializableJson {
    type Error = RegistryError;

    fn try_from(memo: Memo) -> Result<Self, RegistryError> {
        let value = serde_cw_value::to_value(&memo)?;
        Self::new(value)
    }
}

impl TryFrom<String> for SerializableJson {
    type Error = RegistryError;

    fn try_from(value: String) -> Result<Self, RegistryError> {
        Self::new(from_str(&value)?)
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
        SerializableJson::new(serde_json_wasm::from_str(&self.try_string()?)?)
    }
}

fn flatten_cw_value(v: &mut serde_cw_value::Value) {
    use std::mem;
    use std::ops::DerefMut;

    use serde_cw_value::Value::*;

    match v {
        Bool(_) | U8(_) | U16(_) | U32(_) | U64(_) | I8(_) | I16(_) | I32(_) | I64(_) | Char(_)
        | String(_) | Unit | Bytes(_) => {}
        Option(opt) => {
            *v = opt.take().map_or(Unit, |mut value| {
                flatten_cw_value(&mut value);
                *value
            });
        }
        Newtype(value) => {
            flatten_cw_value(value);
            *v = mem::replace(value.deref_mut(), Unit);
        }
        Seq(seq) => {
            for value in seq.iter_mut() {
                flatten_cw_value(value);
            }
        }
        Map(map) => {
            let old_map = mem::take(map);

            for (mut key, mut value) in old_map {
                flatten_cw_value(&mut key);
                flatten_cw_value(&mut value);
                map.insert(key, value);
            }
        }
    }
}

fn deserialize_cw_value<'de, D>(deserializer: D) -> Result<serde_cw_value::Value, D::Error>
where
    D: ::cosmwasm_schema::serde::Deserializer<'de>,
{
    use ::cosmwasm_schema::serde::Deserialize;
    let value = serde_cw_value::Value::deserialize(deserializer)?;
    let object =
        SerializableJson::new(value).map_err(::cosmwasm_schema::serde::de::Error::custom)?;
    Ok(object.into_value())
}

#[cfg(test)]
mod registry_msg_tests {
    use serde_cw_value::Value;

    use super::*;
    use crate::registry::{ChannelId, ForwardingMemo, Memo};

    #[test]
    fn test_deserialize_empty() {
        let memo: SerializableJson = serde_json_wasm::from_str("{}").unwrap();
        assert!(matches!(memo.into_value(), serde_cw_value::Value::Map(m) if m.is_empty()));

        let memo: SerializableJson = serde_json_wasm::from_str("null").unwrap();
        println!("{memo:#?}");
        assert!(matches!(memo.into_value(), serde_cw_value::Value::Map(m) if m.is_empty()));
    }

    #[test]
    fn test_from_memo() {
        let next_next_memo = SerializableJson(Value::Seq(vec![
            Value::U64(1),
            Value::U64(5),
            Value::U64(1234),
            Value::Map({
                let mut m = BTreeMap::new();
                m.insert(Value::String("a".to_owned()), Value::U64(5));
                m.insert(Value::String("b".to_owned()), Value::U64(2));
                m
            }),
        ]));

        let memo: SerializableJson = Memo {
            callback: None,
            forward: Some(ForwardingMemo {
                receiver: "abc1abc".to_owned(),
                port: "transfer".to_owned(),
                channel: ChannelId::new("channel-0").unwrap(),
                next: Some(Box::new(
                    Memo {
                        callback: None,
                        forward: Some(
                            ForwardingMemo {
                                receiver: "def1def".to_owned(),
                                port: "transfer".to_owned(),
                                channel: ChannelId::new("channel-1").unwrap(),
                                next: Some(Box::new(next_next_memo)),
                            }
                            .try_into()
                            .unwrap(),
                        ),
                    }
                    .try_into()
                    .unwrap(),
                )),
            }),
        }
        .try_into()
        .unwrap();

        let expected_memo_json = map([(
            "forward",
            map([
                ("receiver", Value::String("abc1abc".to_owned())),
                ("port", Value::String("transfer".to_owned())),
                ("channel", Value::String("channel-0".to_owned())),
                (
                    "next",
                    map([(
                        "forward",
                        map([
                            ("receiver", Value::String("def1def".to_owned())),
                            ("port", Value::String("transfer".to_owned())),
                            ("channel", Value::String("channel-1".to_owned())),
                            (
                                "next",
                                seq([
                                    Value::U64(1),
                                    Value::U64(5),
                                    Value::U64(1234),
                                    map([("a", Value::U64(5)), ("b", Value::U64(2))]).into_value(),
                                ]),
                            ),
                        ])
                        .into_value(),
                    )])
                    .into_value(),
                ),
            ]),
        )]);

        assert_eq!(memo, expected_memo_json);
    }

    #[test]
    fn test_deserialize_json() {
        let input = r#"
        {
            "test": [
                1,
                5,
                1234,
                {"a": 5, "b": 2}
            ]
        }
        "#;
        let expected = map([(
            "test",
            Value::Seq(vec![
                Value::U64(1),
                Value::U64(5),
                Value::U64(1234),
                Value::Map({
                    let mut m = BTreeMap::new();
                    m.insert(Value::String("a".to_owned()), Value::U64(5));
                    m.insert(Value::String("b".to_owned()), Value::U64(2));
                    m
                }),
            ]),
        )]);

        let parsed_input: SerializableJson = from_str(input).unwrap();
        assert_eq!(parsed_input, expected);
    }

    #[test]
    fn test_flatten_cw_value() {
        let input = Value::Newtype(Box::new(
            map([(
                "test",
                Value::Newtype(Box::new(Value::Newtype(Box::new(Value::Seq(vec![
                    Value::Newtype(Box::new(Value::U8(1))),
                    Value::Newtype(Box::new(Value::U32(5))),
                    Value::U64(1234),
                    Value::Map({
                        let mut m = BTreeMap::new();
                        m.insert(
                            Value::Newtype(Box::new(Value::String("a".to_owned()))),
                            Value::Newtype(Box::new(Value::U32(5))),
                        );
                        m.insert(
                            Value::String("b".to_owned()),
                            Value::Newtype(Box::new(Value::U8(2))),
                        );
                        m
                    }),
                ]))))),
            )])
            .into_value(),
        ));
        let expected = map([(
            "test",
            Value::Seq(vec![
                Value::U8(1),
                Value::U32(5),
                Value::U64(1234),
                Value::Map({
                    let mut m = BTreeMap::new();
                    m.insert(Value::String("a".to_owned()), Value::U32(5));
                    m.insert(Value::String("b".to_owned()), Value::U8(2));
                    m
                }),
            ]),
        )])
        .into_value();

        let input = SerializableJson::new(input).unwrap();
        assert_eq!(input.into_value(), expected);
    }

    #[test]
    fn test_merge_json() {
        // some examples
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(map([("b", Value::U64(2))]))
                .unwrap(),
            map([("a", Value::U64(1)), ("b", Value::U64(2))]),
        );
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(map([("a", Value::U64(2))]))
                .unwrap_err(),
            RegistryError::DuplicateKeyError,
        );
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(map([("b", map([("b", Value::U64(2))]))]))
                .unwrap(),
            map([
                ("a", Value::U64(1)),
                ("b", map([("b", Value::U64(2))]).into_value())
            ]),
        );
        assert_eq!(
            map([("a", map([("b", Value::U64(2))]))])
                .merge(map([("b", Value::U64(1))]))
                .unwrap(),
            map([
                ("a", map([("b", Value::U64(2))]).into_value()),
                ("b", Value::U64(1))
            ]),
        );
        assert_eq!(
            map([("a", map([("b", Value::U64(1))]))])
                .merge(map([("b", map([("b", Value::U64(2))]))]))
                .unwrap(),
            map([
                ("a", map([("b", Value::U64(1))])),
                ("b", map([("b", Value::U64(2))])),
            ]),
        );

        // non-empty + empty
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(SerializableJson::new(Value::Unit).unwrap())
                .unwrap(),
            map([("a", Value::U64(1))])
        );
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(SerializableJson::new(Value::Option(None)).unwrap())
                .unwrap(),
            map([("a", Value::U64(1))])
        );
        assert_eq!(
            map([("a", Value::U64(1))])
                .merge(SerializableJson::empty())
                .unwrap(),
            map([("a", Value::U64(1))])
        );

        // empty + non-empty
        assert_eq!(
            SerializableJson::new(Value::Unit)
                .unwrap()
                .merge(map([("a", Value::U64(1))]))
                .unwrap(),
            map([("a", Value::U64(1))])
        );
        assert_eq!(
            SerializableJson::new(Value::Option(None))
                .unwrap()
                .merge(map([("a", Value::U64(1))]))
                .unwrap(),
            map([("a", Value::U64(1))])
        );
        assert_eq!(
            SerializableJson::empty()
                .merge(map([("a", Value::U64(1))]))
                .unwrap(),
            map([("a", Value::U64(1))])
        );
    }

    fn map<K, V, I>(kvpairs: I) -> SerializableJson
    where
        K: Into<String>,
        V: Into<Value>,
        I: IntoIterator<Item = (K, V)>,
    {
        SerializableJson::new(Value::Map(BTreeMap::from_iter(
            kvpairs
                .into_iter()
                .map(|(k, v)| (Value::String(k.into()), v.into())),
        )))
        .unwrap()
    }

    fn seq<V, I>(vals: I) -> Value
    where
        V: Into<Value>,
        I: IntoIterator<Item = V>,
    {
        Value::Seq(
            vals.into_iter()
                .map(|value| {
                    let mut value = value.into();
                    flatten_cw_value(&mut value);
                    value
                })
                .collect(),
        )
    }
}
