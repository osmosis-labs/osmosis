use cosmwasm_std::{coins, from_binary, to_binary, wasm_execute, BankMsg, Timestamp};
use cosmwasm_std::{Addr, Coin, DepsMut, Response, SubMsg, SubMsgResponse, SubMsgResult};
use swaprouter;
use swaprouter::msg::{ExecuteMsg as SwapRouterExecute, SwapResponse};

use crate::checks::{ensure_key_missing, parse_json, validate_receiver};
use crate::consts::{MsgReplyID, CALLBACK_KEY, PACKET_LIFETIME};
use crate::ibc::{MsgTransfer, MsgTransferResponse};
use crate::msg::{CrosschainSwapResponse, Recovery};

use crate::state;
use crate::state::{
    ForwardMsgReplyState, ForwardTo, SwapMsgReplyState, CONFIG, FORWARD_REPLY_STATES,
    INFLIGHT_PACKETS, RECOVERY_STATES, SWAP_REPLY_STATES,
};
use crate::ContractError;

/// This is the main execute call of this contract.
///
/// It's objective is to trigger a swap between the supplied pairs
///
#[allow(clippy::too_many_arguments)]
pub fn swap_and_forward(
    deps: DepsMut,
    block_time: Timestamp,
    contract_addr: Addr,
    input_coin: Coin,
    output_denom: String,
    slippage: swaprouter::Slippage,
    receiver: Addr,
    next_memo: Option<String>,
    failed_delivery: Option<Recovery>,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("executing swap and forward"));
    let config = CONFIG.load(deps.storage)?;

    // Message to swap tokens in the underlying swaprouter contract
    let swap_msg = SwapRouterExecute::Swap {
        input_coin: input_coin.clone(),
        output_denom,
        slippage,
    };
    let msg = wasm_execute(config.swap_contract, &swap_msg, vec![input_coin])?;

    // Check that the received is valid and retrieve its channel
    let (valid_channel, valid_receiver) = validate_receiver(deps.as_ref(), receiver)?;
    // If there is a memo, check that it is valid
    if let Some(memo) = &next_memo {
        // Parse the string as valid json
        let json = parse_json(memo)?;
        // Ensure the json is an object ({...}) and that it does not contain the CALLBACK_KEY
        ensure_key_missing(&json, CALLBACK_KEY)?;
    }

    // Check that there isn't anything stored in SWAP_REPLY_STATES. If there is,
    // it means that the contract is already waiting for a reply and should not
    // override the stored state. This should only happen if a contract we call
    // calls back to this one. This is likely a malicious attempt modify the
    // contract's state before it has replied.
    if SWAP_REPLY_STATES.may_load(deps.storage)?.is_some() {
        return Err(ContractError::ContractLocked {
            msg: "Already waiting for a reply".to_string(),
        });
    }
    // Store information about the original message to be used in the reply
    SWAP_REPLY_STATES.save(
        deps.storage,
        &SwapMsgReplyState {
            swap_msg,
            block_time,
            contract_addr,
            forward_to: ForwardTo {
                channel: valid_channel,
                receiver: valid_receiver,
                next_memo,
                failed_delivery,
            },
        },
    )?;

    Ok(Response::new().add_submessage(SubMsg::reply_on_success(msg, MsgReplyID::Swap.repr())))
}

pub fn handle_swap_reply(
    deps: DepsMut,
    msg: cosmwasm_std::Reply,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("handle_swap_reply"));
    let swap_msg_state = SWAP_REPLY_STATES.load(deps.storage)?;
    SWAP_REPLY_STATES.remove(deps.storage);

    // If the swaprouter swap failed, return an error
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::FailedSwap {
            msg: format!("No data"),
        })
    };

    // Parse underlying response from the chain
    let parsed =
        cw_utils::parse_execute_response_data(&b).map_err(|e| ContractError::FailedSwap {
            msg: format!("failed to parse: {e}"),
        })?;
    let swap_response: SwapResponse = from_binary(&parsed.data.unwrap_or_default())?;

    // Build an IBC packet to forward the swap.
    let contract_addr = &swap_msg_state.contract_addr;
    let ts = swap_msg_state.block_time.plus_seconds(PACKET_LIFETIME);
    let config = CONFIG.load(deps.storage)?;

    // If the memo is provided we want to include it in the IBC message
    let memo: serde_cw_value::Value = if let Some(memo) = &swap_msg_state.forward_to.next_memo {
        serde_json_wasm::from_str(&memo.to_string()).map_err(|_e| ContractError::InvalidMemo {
            error: format!("this should be unreachable"),
            memo: memo.to_string(),
        })?
    } else {
        serde_json_wasm::from_str("{}").unwrap()
    };

    // If tracking callbacks, we want to include the callback key in the memo
    // without otherwise modifying the provided one
    let memo = match config.track_ibc_callbacks {
        true => {
            let serde_cw_value::Value::Map(mut m) = memo else { unreachable!() };
            m.insert(
                serde_cw_value::Value::String(CALLBACK_KEY.to_string()),
                serde_cw_value::Value::String(contract_addr.to_string()),
            );
            serde_cw_value::Value::Map(m)
        }
        false => memo,
    };

    // Serialize the memo. If it is an empty json object, set it to ""
    let mut memo_str =
        serde_json_wasm::to_string(&memo).map_err(|_e| ContractError::InvalidMemo {
            error: "could not serialize".to_string(),
            memo: format!("{:?}", swap_msg_state.forward_to.next_memo),
        })?;
    if memo_str == "{}" {
        memo_str = String::new();
    }

    // Cosmwasm's  IBCMsg::Transfer  does not support memo.
    // To build and send the packet properly, we need to send it using stargate messages.
    // See https://github.com/CosmWasm/cosmwasm/issues/1477
    let ibc_transfer = MsgTransfer {
        source_port: "transfer".to_string(),
        source_channel: swap_msg_state.forward_to.channel.clone(),
        token: Some(
            Coin::new(
                swap_response.amount.into(),
                swap_response.token_out_denom.clone(),
            )
            .into(),
        ),
        sender: contract_addr.to_string(),
        receiver: swap_msg_state.forward_to.receiver.clone().into(),
        timeout_height: None,
        timeout_timestamp: Some(ts.nanos()),
        memo: memo_str,
    };

    // Base response
    let response = Response::new()
        .add_attribute("status", "ibc_message_created")
        .add_attribute("ibc_message", format!("{:?}", ibc_transfer));

    if !config.track_ibc_callbacks || swap_msg_state.forward_to.failed_delivery.is_none() {
        // If we're not tracking callbacks, or there isn't any recovery addres,
        // then there's no need to listen to the response of the send.

        // The response data needs to be added for consistency. it would
        // normally be added in the next message (after the forward succeeds)
        let amount = swap_response.amount;
        let denom = swap_response.token_out_denom;
        let channel_id = swap_msg_state.forward_to.channel;
        let to_address = swap_msg_state.forward_to.receiver;
        let data = CrosschainSwapResponse {
            msg: format!("Sent {amount}{denom} to {channel_id}/{to_address}"),
        };

        return Ok(response
            .set_data(to_binary(&data)?)
            .add_message(ibc_transfer));
    }

    // Check that there isn't anything stored in FORWARD_REPLY_STATES. If there
    // is, it means that the contract is already waiting for a reply and should
    // not override the stored state. This should never happen here, but adding
    // the check for safety. If this happens there is likely a malicious attempt
    // modify the contract's state before it has replied.
    if FORWARD_REPLY_STATES.may_load(deps.storage)?.is_some() {
        return Err(ContractError::ContractLocked {
            msg: "Already waiting for a reply".to_string(),
        });
    }
    // Store the ibc send information and the user's failed delivery preference
    // so that it can be handled by the response
    FORWARD_REPLY_STATES.save(
        deps.storage,
        &ForwardMsgReplyState {
            channel_id: swap_msg_state.forward_to.channel,
            to_address: swap_msg_state.forward_to.receiver.into(),
            amount: swap_response.amount.into(),
            denom: swap_response.token_out_denom,
            failed_delivery: swap_msg_state.forward_to.failed_delivery,
        },
    )?;

    Ok(response.add_submessage(SubMsg::reply_on_success(
        ibc_transfer,
        MsgReplyID::Forward.repr(),
    )))
}

use ::prost::Message; // Proveides ::decode() for MsgTransferResponse

pub fn handle_forward_reply(
    deps: DepsMut,
    msg: cosmwasm_std::Reply,
) -> Result<Response, ContractError> {
    // Parse the result from the underlying chain call (IBC send)
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::FailedIBCTransfer { msg: format!("failed reply: {:?}", msg.result) })
    };

    // The response contains the packet sequence. This is needed to be able to
    // ensure that, if there is a delivery failure, the packet that failed is
    // the same one that we stored recovery information for
    let response =
        MsgTransferResponse::decode(&b[..]).map_err(|_e| ContractError::FailedIBCTransfer {
            msg: format!("could not decode response: {b}"),
        })?;

    let ForwardMsgReplyState {
        channel_id,
        to_address,
        amount,
        denom,
        failed_delivery,
    } = FORWARD_REPLY_STATES.load(deps.storage)?;
    FORWARD_REPLY_STATES.remove(deps.storage);

    // If a recovery address was provided, store sent IBC transfer so that it
    // can later be recovered by that addr.
    if let Some(Recovery { recovery_addr }) = failed_delivery {
        let recovery = state::ibc::IBCTransfer {
            recovery_addr,
            channel_id: channel_id.clone(),
            sequence: response.sequence,
            amount,
            denom: denom.clone(),
            status: state::ibc::Status::Sent,
        };

        // Save as in-flight to be able to manipulate when the ack/timeout is received
        INFLIGHT_PACKETS.save(deps.storage, (&channel_id, response.sequence), &recovery)?;
    };

    // The response data
    let response = CrosschainSwapResponse {
        msg: format!("Sent {amount}{denom} to {channel_id}/{to_address}"),
    };

    Ok(Response::new()
        .set_data(to_binary(&response)?)
        .add_attribute("status", "ibc_message_created")
        .add_attribute("amount", amount.to_string())
        .add_attribute("denom", denom)
        .add_attribute("channel", channel_id)
        .add_attribute("receiver", to_address))
}

/// Transfers any tokens stored in RECOVERY_STATES[sender] to the sender.
pub fn recover(deps: DepsMut, sender: Addr) -> Result<Response, ContractError> {
    let recoveries = RECOVERY_STATES.load(deps.storage, &sender)?;
    let msgs = recoveries.into_iter().map(|r| BankMsg::Send {
        to_address: r.recovery_addr.into(),
        amount: coins(r.amount, r.denom),
    });
    Ok(Response::new().add_messages(msgs))
}
