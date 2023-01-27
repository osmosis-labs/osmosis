use cosmwasm_std::{Addr, Deps, DepsMut};

use crate::{state::CHANNEL_MAP, ContractError};

// Validate that the receiver address is a valid address for the destination chain.
// This will prevent IBC transfers from failing after forwarding
pub fn validate_receiver(deps: Deps, receiver: Addr) -> Result<(String, Addr), ContractError> {
    let Ok((prefix, _, _)) = bech32::decode(receiver.as_str()) else {
        return Err(ContractError::InvalidReceiver { receiver: receiver.to_string() })
    };

    let channel =
        CHANNEL_MAP
            .load(deps.storage, &prefix)
            .map_err(|_| ContractError::InvalidReceiver {
                receiver: receiver.to_string(),
            })?;

    Ok((channel, receiver))
}

fn stringify(json: &serde_cw_value::Value) -> Result<String, ContractError> {
    serde_json_wasm::to_string(&json).map_err(|_| ContractError::CustomError {
        msg: "invalid value".to_string(), // This shouldn't happen.
    })
}

pub fn ensure_map(json: &serde_cw_value::Value) -> Result<(), ContractError> {
    match json {
        serde_cw_value::Value::Map(_) => Ok(()),
        _ => Err(ContractError::InvalidJson {
            error: format!("invalid json: expected an object"),
            json: stringify(json)?,
        }),
    }
}

pub fn ensure_key_missing(
    json_object: &serde_cw_value::Value,
    key: &str,
) -> Result<(), ContractError> {
    ensure_map(json_object)?;
    let serde_cw_value::Value::Map(m) = json_object else {
        unreachable!()
    };

    if m.get(&serde_cw_value::Value::String(key.to_string()))
        .is_some()
    {
        Err(ContractError::InvalidJson {
            error: format!("invalid json: {key} key not allowed"),
            json: stringify(json_object)?,
        })
    } else {
        Ok(())
    }
}
