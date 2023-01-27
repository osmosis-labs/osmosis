use crate::{
    checks::validate_input_amount,
    msg::{Callback, ExecuteMsg, Wasm, WasmHookExecute},
    state::CONFIG,
    ContractError,
};
use cosmwasm_std::{Addr, Coin, DepsMut, Response, Timestamp};

// IBC timeout
pub const PACKET_LIFETIME: u64 = 604_800u64; // One week in seconds

//#[cfg(feature = "callbacks")]
fn build_callback_memo(
    callback: Option<Callback>,
) -> Result<crosschain_swaps::msg::SerializableJson, ContractError> {
    match callback {
        Some(callback) => callback.try_string(),
        None => Ok(String::new()),
    }
}

pub fn execute_swap(
    deps: DepsMut,
    own_addr: Addr,
    now: Timestamp,
    coin: Coin,
    user_msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    let ExecuteMsg::OsmosisSwap {
        swap_amount,
        output_denom,
        receiver,
        slippage,
        on_failed_delivery,
        #[cfg(feature = "callbacks")]
        callback,
    } = user_msg;
    let config = CONFIG.load(deps.storage)?;

    // If the callbacks feature is not active, the variable won't exist. Create it here with the default
    #[cfg(not(feature = "callbacks"))]
    let callback = None;

    let next_memo = build_callback_memo(callback)?;
    // Wrap in an option, as expected by MsgTransfer bellow
    let next_memo = if next_memo.is_empty() {
        None
    } else {
        Some(next_memo)
    };

    if swap_amount > coin.amount.into() {
        return Err(ContractError::SwapAmountTooHigh {
            received: swap_amount,
            max: coin.amount.into(),
        });
    }

    validate_input_amount(swap_amount, coin.amount)?;

    // note that this is not the same osmosis swap as the one above (which is
    // defined in this create). The one in crosschain_swaps doesn't accept a
    // callback. They share the same name because that's the name we want to
    // expose to the user
    let instruction = crosschain_swaps::ExecuteMsg::OsmosisSwap {
        swap_amount,
        output_denom,
        receiver,
        slippage,
        next_memo: None,
        on_failed_delivery,
    };

    let msg = WasmHookExecute {
        wasm: Wasm {
            contract: config.crosschain_swaps_contract.clone(),
            msg: instruction,
        },
    };
    let memo = serde_json_wasm::to_string(&msg).map_err(|e| ContractError::InvalidJson {
        error: e.to_string(),
    })?;

    let ibc_transfer_msg = crosschain_swaps::ibc::MsgTransfer {
        source_port: "transfer".to_string(),
        source_channel: "channel-0".to_string(),
        token: Some(Coin::new(coin.amount.into(), coin.denom).into()),
        sender: own_addr.to_string(),
        receiver: config.crosschain_swaps_contract,
        timeout_height: None,
        timeout_timestamp: Some(now.plus_seconds(PACKET_LIFETIME).nanos()),
        memo,
    };
    Ok(Response::default().add_message(ibc_transfer_msg))
}
