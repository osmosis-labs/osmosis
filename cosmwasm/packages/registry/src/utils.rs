use std::collections::BTreeMap;

use cosmwasm_std::StdError;
use serde_cw_value::Value;

use crate::RegistryError;

pub fn stringify(json: &serde_cw_value::Value) -> Result<String, RegistryError> {
    serde_json_wasm::to_string(&json).map_err(|_| {
        RegistryError::Std(StdError::generic_err(
            "invalid value".to_string(), // This shouldn't happen.
        ))
    })
}

pub fn extract_map(json: Value) -> Result<BTreeMap<Value, Value>, RegistryError> {
    match json {
        serde_cw_value::Value::Map(m) => Ok(m),
        _ => Err(RegistryError::InvalidJson {
            error: "invalid json: expected an object".to_string(),
            json: stringify(&json)?,
        }),
    }
}

pub fn merge_json(first: &str, second: &str) -> Result<String, RegistryError> {
    // replacing some potential empty values we want to accept with an empty object
    let first = match first {
        "" => "{}",
        "null" => "{}",
        _ => first,
    };
    let second = match second {
        "" => "{}",
        "null" => "{}",
        _ => second,
    };

    let first_val: Value = serde_json_wasm::from_str(first)?;
    let second_val: Value = serde_json_wasm::from_str(second)?;

    // checking potential "empty" values we want to accept

    let mut first_map = extract_map(first_val)?;
    let second_map = extract_map(second_val)?;

    first_map.extend(second_map);

    stringify(&Value::Map(first_map))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_merge_json() {
        // some examples
        assert_eq!(
            merge_json(r#"{"a": 1}"#, r#"{"b": 2}"#).unwrap(),
            r#"{"a":1,"b":2}"#
        );
        assert_eq!(
            merge_json(r#"{"a": 1}"#, r#"{"a": 2}"#).unwrap(),
            r#"{"a":2}"#
        );
        assert_eq!(
            merge_json(r#"{"a": 1}"#, r#"{"a": {"b": 2}}"#).unwrap(),
            r#"{"a":{"b":2}}"#
        );
        assert_eq!(
            merge_json(r#"{"a": {"b": 2}}"#, r#"{"a": 1}"#).unwrap(),
            r#"{"a":1}"#
        );
        assert_eq!(
            merge_json(r#"{"a": {"b": 2}}"#, r#"{"a": {"c": 3}}"#).unwrap(),
            r#"{"a":{"c":3}}"#
        );
        // Empties
        assert_eq!(merge_json(r#"{"a": 1}"#, r#""#).unwrap(), r#"{"a":1}"#);
        assert_eq!(merge_json(r#""#, r#"{"a": 1}"#).unwrap(), r#"{"a":1}"#);
        assert_eq!(merge_json(r#"{"a": 1}"#, r#"null"#).unwrap(), r#"{"a":1}"#);
        assert_eq!(merge_json(r#"null"#, r#"{"a": 1}"#).unwrap(), r#"{"a":1}"#);
        assert_eq!(merge_json(r#"{"a": 1}"#, r#"{}"#).unwrap(), r#"{"a":1}"#);
        assert_eq!(merge_json(r#"{}"#, r#"{"a": 1}"#).unwrap(), r#"{"a":1}"#);
    }
}
