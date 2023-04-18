use cosmwasm_std::{from_binary, Reply, SubMsgResponse, SubMsgResult};
use registry::msg::SerializableJson;
use swaprouter::msg::SwapResponse;

use crate::{consts::CALLBACK_KEY, ContractError};

/// Extract the relevant response from the swaprouter reply
pub fn parse_swaprouter_reply(msg: Reply) -> Result<SwapResponse, ContractError> {
    // If the swaprouter swap failed, return an error
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::FailedSwap {
            msg: format!("No data"),
        })
    };

    // Parse underlying response from the chain
    let parsed =
        cw_utils::parse_execute_response_data(&b).map_err(|e| ContractError::FailedSwap {
            msg: format!("failed to parse swaprouter response: {e}"),
        })?;
    let swap_response: SwapResponse = from_binary(&parsed.data.unwrap_or_default())?;
    Ok(swap_response)
}

/// Build a memo to be used in the forward IBC transfer.
///
/// The resulting memo will include {IBC_CALLBACK_KEY: contract_addr} and any
/// other keys provided by the sender
pub fn build_memo(
    next_memo: Option<SerializableJson>,
    contract_addr: &str,
) -> Result<String, ContractError> {
    // If the memo is provided we want to include it in the IBC message. If not,
    // we default to an empty object
    let memo = next_memo.unwrap_or_else(|| serde_json_wasm::from_str("{}").unwrap());

    // Include the callback key in the memo without modifying the rest of the
    // provided memo
    let memo = {
        let serde_cw_value::Value::Map(mut m) = memo.0 else { unreachable!() };
        m.insert(
            serde_cw_value::Value::String(CALLBACK_KEY.to_string()),
            serde_cw_value::Value::String(contract_addr.to_string()),
        );
        serde_cw_value::Value::Map(m)
    };

    // Serialize the memo. If it is an empty json object, set it to ""
    let mut memo_str =
        serde_json_wasm::to_string(&memo).map_err(|_e| ContractError::InvalidMemo {
            error: "could not serialize".to_string(),
            memo: format!("{memo:?}"),
        })?;

    // This is redundant, as the ibc_callback_key will always exist. We leave it
    // here preemptively so if we make the callback key optional in the future,
    // the memo gets completely removed.
    if memo_str == "{}" {
        memo_str = String::new();
    }
    Ok(memo_str)
}
