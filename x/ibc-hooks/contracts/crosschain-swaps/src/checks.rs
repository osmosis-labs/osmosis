use cosmwasm_std::{Addr, Deps};

use crate::{consts::CALLBACK_KEY, state::CHANNEL_MAP, ContractError};

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

pub fn validate_memo(memo: &str) -> Result<(), ContractError> {
    let maybe_value: Result<serde_cw_value::Value, _> = serde_json_wasm::from_str(memo);
    if let Err(err) = maybe_value {
        return Err(ContractError::InvalidMemo {
            error: format!("failed to parse: {err}"),
            memo: memo.to_string(),
        });
    }

    let Ok(serde_cw_value::Value::Map(m)) = maybe_value else {
        return Err(ContractError::InvalidMemo {
            error: format!("memo must be a json obect"),
            memo: memo.to_string(),
        });
    };
    if m.get(&serde_cw_value::Value::String(CALLBACK_KEY.to_string()))
        .is_some()
    {
        return Err(ContractError::InvalidMemo {
            error: format!("callback key not allowed"),
            memo: memo.to_string(),
        });
    };

    Ok(())
}
