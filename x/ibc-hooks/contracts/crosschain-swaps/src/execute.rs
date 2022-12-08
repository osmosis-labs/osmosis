use cosmwasm_std::{coins, from_binary, to_binary, wasm_execute, BankMsg, Deps, Reply, Timestamp};
use cosmwasm_std::{Addr, Coin, DepsMut, Response, SubMsg, SubMsgResponse, SubMsgResult};
use swaprouter::msg::{ExecuteMsg as SwapRouterExecute, Slippage, SwapResponse};

use crate::consts::{FORWARD_REPLY_ID, PACKET_LIFETIME, SWAP_REPLY_ID};
use crate::ibc::{MsgTransfer, MsgTransferResponse};
use crate::msg::{CrosschainSwapResponse, Recovery};

use crate::state::{
    ForwardMsgReplyState, ForwardTo, IBCTransfer, Status, SwapMsgReplyState, CHANNEL_MAP, CONFIG,
    FORWARD_REPLY_STATES, INFLIGHT_PACKETS, RECOVERY_STATES, SWAP_REPLY_STATES,
};
use crate::ContractError;

// Validate that the receiver address is a valid address for the destination chain.
// This will prevent IBC transfers from failing after forwarding
fn validate_receiver(deps: Deps, receiver: Addr) -> Result<(String, Addr), ContractError> {
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

/// This is the main execute call of this contract.
///
/// It's objective is to trigger a swap between the supplied pairs
///
pub fn swap_and_forward(
    deps: DepsMut,
    block_time: Timestamp,
    contract_addr: Addr,
    input_coin: Coin,
    output_denom: String,
    slippage: Slippage,
    receiver: Addr,
    failed_delivery: Option<Recovery>,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;

    // Message to swap tokens in the underlying swaprouter contract
    let swap_msg = SwapRouterExecute::Swap {
        input_coin: input_coin.clone(),
        output_denom,
        slippage,
    };
    let msg = wasm_execute(config.swap_contract, &swap_msg, vec![input_coin])?;

    let (valid_channel, valid_receiver) = validate_receiver(deps.as_ref(), receiver)?;

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
                failed_delivery,
            },
        },
    )?;

    Ok(Response::new().add_submessage(SubMsg::reply_on_success(msg, SWAP_REPLY_ID)))
}

pub fn handle_swap_reply(deps: DepsMut, msg: Reply) -> Result<Response, ContractError> {
    deps.api.debug(&format!("handle_swap_reply"));
    // TODO: Warning! This may be succeptible to "reentrancy". We are assuming
    //       that the swaprouter this contract was initialized with is the
    //       correct one and that it doesn't call back into this contract (which
    //       could modify the item stored in SWAP_REPLY_STATES).
    //
    //       Review this. Though this may not be an issue at all, because: if
    //       the underlying swaprouter contract is compromised, they will
    //       already have the funds and can just send them without having to do
    //       contract call trickery.
    //
    //       Alternative: replace the item with a "stack" and add complexity.
    let swap_msg_state = SWAP_REPLY_STATES.load(deps.storage)?;
    SWAP_REPLY_STATES.remove(deps.storage);

    // If the swaprouter swap failed, return an error
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::CustomError {
            val: format!("Failed Swap"),
        })
    };

    // Parse underlying response from the chain
    let parsed = cw_utils::parse_execute_response_data(&b)
        .map_err(|e| ContractError::CustomError { val: e.to_string() })?;
    let swap_response: SwapResponse = from_binary(&parsed.data.unwrap())?;

    // Build an IBC packet to forward the swap.
    let contract_addr = &swap_msg_state.contract_addr;
    let ts = swap_msg_state.block_time.plus_seconds(PACKET_LIFETIME);
    let config = CONFIG.load(deps.storage)?;
    let memo = match config.track_ibc_callbacks {
        true => format!(r#"{{"callback": "{contract_addr}"}}"#),
        false => format!(""),
    };

    // Cosmwasm's  IBCMsg::Transfer  does not support memo.
    // To build and send the packet properly, we need to send it using stargate messages.
    // See https://github.com/CosmWasm/cosmwasm/issues/1477
    let ibc_transfer = MsgTransfer {
        source_port: "transfer".to_string(),
        source_channel: swap_msg_state.forward_to.channel.clone(),
        token: Some(
            Coin::new(
                swap_response.amount.clone().into(),
                swap_response.token_out_denom.clone(),
            )
            .into(),
        ),
        sender: contract_addr.to_string(),
        receiver: swap_msg_state.forward_to.receiver.clone().into(),
        timeout_height: None,
        timeout_timestamp: Some(ts.nanos()),
        memo,
    };

    // Base response
    let response = Response::new()
        .add_attribute("status", "ibc_message_created")
        .add_attribute("ibc_message", format!("{:?}", ibc_transfer));

    if !config.track_ibc_callbacks || swap_msg_state.forward_to.failed_delivery.is_none() {
        // If we're not tracking callbacks, or there isn't any recovery addres,
        // then there's no need to listen to the response of the send.
        // The response data

        // Add the response data since it won't be added in the reply.
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

    Ok(response.add_submessage(SubMsg::reply_on_success(ibc_transfer, FORWARD_REPLY_ID)))
}

use ::prost::Message; // Proveides ::decode() for MsgTransferResponse

pub fn handle_forward_reply(deps: DepsMut, msg: Reply) -> Result<Response, ContractError> {
    // Parse the result from the underlying chain call (IBC send)
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::CustomError { val: "invalid reply".to_string() })
    };

    // The response contains the packet sequence. This is needed to be able to
    // ensure that, if there is a delivery failure, the packet that failed is
    // the same one that we stored recovery information for
    let response =
        MsgTransferResponse::decode(&b[..]).map_err(|_e| ContractError::CustomError {
            val: "could not decode response".to_string(),
        })?;

    // Similar consideration as the warning above. Is it safe for this to be an Item?
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
        let recovery = IBCTransfer {
            recovery_addr: recovery_addr.clone(),
            channel_id: channel_id.clone(),
            sequence: response.sequence,
            amount,
            denom: denom.clone(),
            status: Status::Sent,
        };

        // Save as in-flight to be able to manipulate when the ack/timeout is received
        INFLIGHT_PACKETS.save(
            deps.storage,
            (&channel_id.clone(), response.sequence),
            &recovery,
        )?;
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
