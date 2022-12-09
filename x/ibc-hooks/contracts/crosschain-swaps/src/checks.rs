use cosmwasm_std::{Addr, Deps};

use crate::{consts::CALLBACK_KEY, state::CHANNEL_MAP, ContractError};

// Validate that the receiver address is a valid address for the destination chain.
// This will prevent IBC transfers from failing after forwarding
pub fn validate_receiver(deps: Deps, receiver: Addr) -> Result<(String, Addr), ContractError> {
    let Ok((prefix, _, _)) = bech32::decode(receiver.as_str()) else {
        return Err(ContractError::CustomError { val: format!("invalid receiver {receiver}") })
    };

    let channel =
        CHANNEL_MAP
            .load(deps.storage, &prefix)
            .map_err(|_| ContractError::CustomError {
                val: "invalid receiver {receiver}".to_string(),
            })?;

    Ok((channel, receiver))
}

pub fn validate_memo(memo: &str) -> Result<(), ContractError> {
    let maybe_value: Result<serde_cw_value::Value, _> = serde_json_wasm::from_str(&memo);
    if maybe_value.is_err() {
        return Err(ContractError::CustomError {
            val: format!("invalid memo: {memo}"),
        });
    }

    let Ok(serde_cw_value::Value::Map(m)) = maybe_value else {
        return Err(ContractError::CustomError {
            val: format!("invalid memo: memo must be a json obect. Got: {memo}"),
        });
    };
    if m.get(&serde_cw_value::Value::String(CALLBACK_KEY.to_string()))
        .is_some()
    {
        return Err(ContractError::CustomError {
            val: format!("invalid memo: callback key not allowed. Got: {memo}"),
        });
    };

    Ok(())
}
