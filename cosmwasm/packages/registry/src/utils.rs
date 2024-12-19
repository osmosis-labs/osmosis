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
