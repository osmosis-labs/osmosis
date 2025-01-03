use cosmwasm_std::{coins, to_binary, wasm_execute, BankMsg, Env, MessageInfo};
use cosmwasm_std::{Addr, Coin, DepsMut, Response, SubMsg, SubMsgResponse, SubMsgResult};
use osmosis_std::types::osmosis::poolmanager::v1beta1::SwapAmountInRoute;
use registry::msg::{Callback, SerializableJson};
use registry::RegistryError;
use swaprouter::msg::ExecuteMsg as SwapRouterExecute;

use crate::checks::{
    check_is_contract_governor, ensure_key_missing, get_registry, validate_receiver,
};
use crate::consts::{MsgReplyID, CALLBACK_KEY};
use crate::msg::{CrosschainSwapResponse, FailedDeliveryAction};
use registry::proto::MsgTransferResponse;

use crate::state::{
    Config, ForwardMsgReplyState, ForwardTo, SwapMsgReplyState, CONFIG, FORWARD_REPLY_STATE,
    INFLIGHT_PACKETS, RECOVERY_STATES, SWAP_REPLY_STATE,
};
use crate::utils::{build_memo, parse_swaprouter_reply};
use crate::ContractError;
use crate::{state, ExecuteMsg};

use std::fmt::Debug;

// Helper to add consistent events on ibc message submission
fn ibc_message_event<T: Debug>(context: &str, message: T) -> cosmwasm_std::Event {
    cosmwasm_std::Event::new("ibc_message_added")
        .add_attribute("context", context)
        .add_attribute("ibc_message", format!("{message:?}"))
}

/// This function takes any token. If it's already something we can work with
/// (either native to osmosis or native to a chain connected to osmosis via a
/// valid channel), it will just proceed to swap and forward. If it's not, then
/// it will send an IBC message to unwrap it first and provide a callback to
/// ensure the right swap_and_forward gets called after the unwrap succeeds
#[allow(clippy::too_many_arguments)]
pub fn unwrap_or_swap_and_forward(
    ctx: (DepsMut, Env, MessageInfo),
    output_denom: String,
    slippage: swaprouter::Slippage,
    receiver: &str,
    next_memo: Option<SerializableJson>,
    final_memo: Option<SerializableJson>,
    failed_delivery_action: FailedDeliveryAction,
    route: Option<Vec<SwapAmountInRoute>>,
) -> Result<Response, ContractError> {
    let (deps, env, info) = ctx;
    let swap_coin = cw_utils::one_coin(&info)?;

    deps.api
        .debug(&format!("executing unwrap or swap and forward"));
    let registry = get_registry(deps.as_ref())?;

    // Check the path that the coin took to get to the current chain.
    // Each element in the path is an IBC hop.
    let path = registry.unwrap_denom_path(&swap_coin.denom)?;
    if path.is_empty() {
        return Err(RegistryError::InvalidDenomTracePath {
            path: String::new(),
            denom: swap_coin.denom,
        }
        .into());
    }

    let amount: u128 = swap_coin.amount.into();

    // If the path is larger than 2, we need to unwrap this token first
    if path.len() > 2 {
        let registry = get_registry(deps.as_ref())?;
        let ibc_transfer = registry.unwrap_coin_into(
            swap_coin,
            env.contract.address.to_string(),
            None,
            env.contract.address.to_string(),
            env.block.time,
            build_memo(None, env.contract.address.as_str())?,
            String::new(),
            Some(Callback {
                contract: env.contract.address.clone(),
                msg: serde_cw_value::to_value(&ExecuteMsg::OsmosisSwap {
                    output_denom,
                    receiver: receiver.to_string(),
                    slippage,
                    next_memo,
                    final_memo,
                    on_failed_delivery: failed_delivery_action.clone(),
                    route,
                })?
                .try_into()?,
            }),
            false,
        )?;

        // Ensure the state is properly setup to handle a reply from the ibc_message
        save_forward_reply_state(
            deps,
            ForwardMsgReplyState {
                channel_id: ibc_transfer.source_channel.clone(),
                to_address: env.contract.address.to_string(),
                amount,
                denom: String::new(),
                on_failed_delivery: failed_delivery_action,
                is_swap: false,
            },
        )?;

        // Here we should add a response for the sender with the packet
        // sequence, but that would require habdling the reply. This will be
        // unncecessary once async acks lands, so we should wait for that
        return Ok(Response::new()
            //.set_data(data)
            .add_attribute("action", "unwrap_before_swap")
            .add_event(ibc_message_event(
                "pre-swap unwinding ibc message created",
                &ibc_transfer,
            ))
            .add_submessage(SubMsg::reply_on_success(
                ibc_transfer,
                MsgReplyID::Forward.repr(),
            )));
    }

    // If the denom is either native or only one hop, we swap it directly
    swap_and_forward(
        (deps, env, info),
        swap_coin,
        output_denom,
        slippage,
        receiver,
        next_memo,
        final_memo,
        failed_delivery_action,
        route,
    )
}

/// This function takes token "known to the chain", swaps it, and then forwards
/// the result to the receiver.
///
///
#[allow(clippy::too_many_arguments)]
pub fn swap_and_forward(
    ctx: (DepsMut, Env, MessageInfo),
    swap_coin: Coin,
    output_denom: String,
    slippage: swaprouter::Slippage,
    receiver: &str,
    next_memo: Option<SerializableJson>,
    final_memo: Option<SerializableJson>,
    failed_delivery_action: FailedDeliveryAction,
    route: Option<Vec<SwapAmountInRoute>>,
) -> Result<Response, ContractError> {
    let (deps, env, _) = ctx;

    deps.api.debug(&format!("executing swap and forward"));
    let config = CONFIG.load(deps.storage)?;

    // Check that the received is valid and retrieve its channel
    let (valid_chain, valid_receiver) = validate_receiver(deps.as_ref(), receiver)?;

    // If there are memos, check that they are valid (i.e. a valid json object that
    // doesn't contain the key that we will insert later)
    let s_next_memo = validate_user_provided_memo(&deps, next_memo.as_ref())?;
    let s_final_memo = validate_user_provided_memo(&deps, final_memo.as_ref())?;

    // Validate that the swapped token can be unwrapped. If it can't, abort
    // early to avoid swapping unnecessarily
    let registry = get_registry(deps.as_ref())?;
    registry.unwrap_coin_into(
        Coin::new(1, output_denom.clone()),
        valid_receiver.to_string(),
        Some(&valid_chain),
        env.contract.address.to_string(),
        env.block.time,
        s_next_memo,
        s_final_memo,
        None,
        false,
    )?;

    // Message to swap tokens in the underlying swaprouter contract
    let swap_msg = SwapRouterExecute::Swap {
        input_coin: swap_coin.clone(),
        output_denom,
        slippage,
        route,
    };
    let msg = wasm_execute(config.swap_contract, &swap_msg, vec![swap_coin])?;

    // Check that there isn't anything stored in SWAP_REPLY_STATES. If there is,
    // it means that the contract is already waiting for a reply and should not
    // override the stored state. This should only happen if a contract we call
    // calls back to this one. This is likely a malicious attempt modify the
    // contract's state before it has replied.
    if SWAP_REPLY_STATE.may_load(deps.storage)?.is_some() {
        return Err(ContractError::ContractLocked {
            msg: "Already waiting for a reply".to_string(),
        });
    }

    // Store information about the original message to be used in the reply
    SWAP_REPLY_STATE.save(
        deps.storage,
        &SwapMsgReplyState {
            swap_msg,
            block_time: env.block.time,
            contract_addr: env.contract.address,
            forward_to: ForwardTo {
                chain: valid_chain,
                receiver: valid_receiver,
                next_memo,
                final_memo,
                on_failed_delivery: failed_delivery_action,
            },
        },
    )?;

    Ok(Response::new()
        .add_attribute("action", "swap_and_forward")
        .add_submessage(SubMsg::reply_on_success(msg, MsgReplyID::Swap.repr())))
}

fn save_forward_reply_state(
    deps: DepsMut,
    forward_reply_state: ForwardMsgReplyState,
) -> Result<(), ContractError> {
    // Check that there isn't anything stored in FORWARD_REPLY_STATES. If there
    // is, it means that the contract is already waiting for a reply and should
    // not override the stored state. This should never happen here, but adding
    // the check for safety. If this happens there is likely a malicious attempt
    // modify the contract's state before it has replied.
    if FORWARD_REPLY_STATE.may_load(deps.storage)?.is_some() {
        return Err(ContractError::ContractLocked {
            msg: "Already waiting for a reply".to_string(),
        });
    }
    // Store the ibc send information and the user's failed delivery preference
    // so that it can be handled by the response
    FORWARD_REPLY_STATE.save(deps.storage, &forward_reply_state)?;
    Ok(())
}

// The swap has succeeded and we need to generate the forward IBC transfer
pub fn handle_swap_reply(
    deps: DepsMut,
    env: Env,
    msg: cosmwasm_std::Reply,
) -> Result<Response, ContractError> {
    let swap_msg_state = SWAP_REPLY_STATE.load(deps.storage)?;
    SWAP_REPLY_STATE.remove(deps.storage);

    // Extract the relevant response from the swaprouter reply
    let swap_response = parse_swaprouter_reply(msg)?;

    // Build an IBC packet to forward the swap.
    let contract_addr = &swap_msg_state.contract_addr;

    // If the memo is provided we want to include it in the IBC message. If not,
    // we default to an empty object. The resulting memo will always include the
    // callback so this contract can track the IBC send
    let next_memo = build_memo(swap_msg_state.forward_to.next_memo, contract_addr.as_str())?;
    let final_memo = swap_msg_state
        .forward_to
        .final_memo
        .map(|final_memo| serde_json_wasm::to_string(&final_memo))
        .transpose()?
        .unwrap_or_default();

    let registry = get_registry(deps.as_ref())?;
    let ibc_transfer = registry.unwrap_coin_into(
        Coin::new(
            swap_response.amount.into(),
            swap_response.token_out_denom.clone(),
        ),
        swap_msg_state.forward_to.receiver.clone().to_string(),
        Some(&swap_msg_state.forward_to.chain),
        env.contract.address.to_string(),
        env.block.time,
        next_memo,
        final_memo,
        None,
        false,
    )?;
    deps.api.debug(&format!("IBC transfer: {ibc_transfer:?}"));

    // Base response
    let response = Response::new()
        .add_attribute("action", "handle_swap_reply")
        .add_event(ibc_message_event(
            "forward ibc message added",
            &ibc_transfer,
        ));

    // Ensure the state is properly setup to handle a reply from the ibc_message
    save_forward_reply_state(
        deps,
        ForwardMsgReplyState {
            channel_id: ibc_transfer.source_channel.clone(),
            to_address: swap_msg_state.forward_to.receiver.into(),
            amount: swap_response.amount.into(),
            denom: swap_response.token_out_denom,
            on_failed_delivery: swap_msg_state.forward_to.on_failed_delivery,
            is_swap: true,
        },
    )?;

    Ok(response.add_submessage(SubMsg::reply_on_success(
        ibc_transfer,
        MsgReplyID::Forward.repr(),
    )))
}

// Included here so it's closer to the trait that needs it.
use ::prost::Message; // Proveides ::decode() for MsgTransferResponse

// The ibc transfer has been "sent" successfully. We create an inflight packet
// in storage for potential recovery.
// If recovery is set to "do_nothing", we just return a response.
pub fn handle_forward_reply(
    deps: DepsMut,
    msg: cosmwasm_std::Reply,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("handle_forward_reply"));
    // Parse the result from the underlying chain call (IBC send)
    let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result else {
        return Err(ContractError::FailedIBCTransfer {
            msg: format!("failed reply: {:?}", msg.result),
        });
    };

    // The response contains the packet sequence. This is needed to be able to
    // ensure that, if there is a delivery failure, the packet that failed is
    // the same one that we stored recovery information for
    let transfer_response =
        MsgTransferResponse::decode(&b[..]).map_err(|_e| ContractError::FailedIBCTransfer {
            msg: format!("could not decode response: {b}"),
        })?;

    let ForwardMsgReplyState {
        channel_id,
        to_address,
        amount,
        denom,
        on_failed_delivery: failed_delivery_action,
        is_swap,
    } = FORWARD_REPLY_STATE.load(deps.storage)?;
    FORWARD_REPLY_STATE.remove(deps.storage);

    // If a recovery address was provided, store sent IBC transfer so that it
    // can later be recovered by that addr.
    match failed_delivery_action {
        FailedDeliveryAction::DoNothing => {}
        FailedDeliveryAction::LocalRecoveryAddr(recovery_addr) => {
            let recovery = state::ibc::IBCTransfer {
                recovery_addr,
                channel_id: channel_id.clone(),
                sequence: transfer_response.sequence,
                amount,
                denom: denom.clone(),
                status: state::ibc::PacketLifecycleStatus::Sent,
            };

            // Save as in-flight to be able to manipulate when the ack/timeout is received
            INFLIGHT_PACKETS.save(
                deps.storage,
                (&channel_id, transfer_response.sequence),
                &recovery,
            )?;
        }
    }

    let response = Response::new()
        .add_attribute("action", "handle_forward_reply")
        .add_attribute("status", "ibc_message_successfully_submitted")
        .add_attribute("channel", &channel_id)
        .add_attribute("receiver", &to_address)
        .add_attribute(
            "packet_sequence",
            format!("{:?}", transfer_response.sequence),
        );

    if !is_swap {
        return Ok(response);
    }

    // Add the information about the swap
    let response_data = CrosschainSwapResponse::new(
        amount,
        &denom,
        &channel_id,
        &to_address,
        transfer_response.sequence,
    );

    Ok(response
        .set_data(to_binary(&response_data)?)
        .add_attribute("amount", amount.to_string())
        .add_attribute("denom", denom))
}

/// Transfers any tokens stored in RECOVERY_STATES[sender] to the sender.
pub fn recover(deps: DepsMut, sender: Addr) -> Result<Response, ContractError> {
    let recoveries = RECOVERY_STATES.load(deps.storage, &sender)?;
    // Remove the recoveries from the store. If the sends fail, the whole tx should be reverted.
    RECOVERY_STATES.remove(deps.storage, &sender);
    let msgs = recoveries.into_iter().map(|r| BankMsg::Send {
        to_address: r.recovery_addr.into_string(),
        amount: coins(r.amount, r.denom),
    });
    Ok(Response::new()
        .add_attribute("action", "recover")
        .add_messages(msgs))
}

// Transfer ownership of this contract
pub fn transfer_ownership(
    deps: DepsMut,
    sender: Addr,
    new_governor: String,
) -> Result<Response, ContractError> {
    // only owner can transfer
    check_is_contract_governor(deps.as_ref(), sender)?;
    let new_governor = deps.api.addr_validate(&new_governor)?;

    CONFIG.update(
        deps.storage,
        |mut config| -> Result<Config, ContractError> {
            config.governor = new_governor;
            Ok(config)
        },
    )?;

    Ok(Response::new().add_attribute("action", "transfer_ownership"))
}

/// Set the address of the swap contract to use
pub fn set_swap_contract(
    deps: DepsMut,
    sender: Addr,
    new_contract: String,
) -> Result<Response, ContractError> {
    check_is_contract_governor(deps.as_ref(), sender)?;
    let new_contract = deps.api.addr_validate(&new_contract)?;

    CONFIG.update(
        deps.storage,
        |mut config| -> Result<Config, ContractError> {
            config.swap_contract = new_contract;
            Ok(config)
        },
    )?;

    Ok(Response::new().add_attribute("action", "set_swaps_contract"))
}

fn validate_user_provided_memo(
    deps: &DepsMut,
    user_memo: Option<&SerializableJson>,
) -> Result<String, ContractError> {
    Ok(if let Some(memo) = user_memo {
        // Ensure the json is an object ({...}) and that it does not contain the CALLBACK_KEY
        deps.api.debug(&format!("checking memo: {memo:?}"));
        ensure_key_missing(memo.as_value(), CALLBACK_KEY)?;
        serde_json_wasm::to_string(&memo)?
    } else {
        String::new()
    })
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};

    use super::*;
    use crate::{contract, msg::InstantiateMsg, ExecuteMsg};

    static CREATOR_ADDRESS: &str = "creator";
    static SWAPCONTRACT_ADDRESS: &str = "swapcontract";
    static REGISTRY_ADDRESS: &str = "registrycontract";

    // test helper
    #[allow(unused_assignments)]
    fn initialize_contract(deps: DepsMut) -> Addr {
        let msg = InstantiateMsg {
            governor: String::from(CREATOR_ADDRESS),
            swap_contract: String::from(SWAPCONTRACT_ADDRESS),
            registry_contract: String::from(REGISTRY_ADDRESS),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);

        contract::instantiate(deps, mock_env(), info.clone(), msg).unwrap();

        info.sender
    }

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();

        let governor = initialize_contract(deps.as_mut());
        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.governor, governor);
    }

    #[test]
    fn transfer_ownership() {
        let mut deps = mock_dependencies();

        let governor = initialize_contract(deps.as_mut());
        let governor_info = mock_info(governor.as_str(), &vec![] as &Vec<Coin>);

        let new_governor = "new_owner".to_string();
        // The owner can transfer ownership
        let msg = ExecuteMsg::TransferOwnership {
            new_governor: new_governor.clone(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info, msg).unwrap();

        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(new_governor, config.governor);
    }

    #[test]
    fn transfer_ownership_unauthorized() {
        let mut deps = mock_dependencies();

        let governor = initialize_contract(deps.as_mut());

        let other_info = mock_info("other_sender", &vec![] as &Vec<Coin>);

        // An unauthorized user cannot transfer ownership
        let msg = ExecuteMsg::TransferOwnership {
            new_governor: "new_owner".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), other_info, msg).unwrap_err();

        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(governor, config.governor);
    }

    #[test]
    fn set_swap_contract() {
        let mut deps = mock_dependencies();

        let governor = initialize_contract(deps.as_mut());
        let governor_info = mock_info(governor.as_str(), &vec![] as &Vec<Coin>);

        // and new channel
        let msg = ExecuteMsg::SetSwapContract {
            new_contract: "new_swap_contract".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info, msg).unwrap();

        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.swap_contract, "new_swap_contract".to_string());
    }

    #[test]
    fn set_swap_contract_unauthorized() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // A user other than the owner cannot modify the channel registry
        let other_info = mock_info("other_sender", &vec![] as &Vec<Coin>);
        let msg = ExecuteMsg::SetSwapContract {
            new_contract: "new_swap_contract".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), other_info, msg).unwrap_err();
        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.swap_contract, SWAPCONTRACT_ADDRESS.to_string());
    }
}
