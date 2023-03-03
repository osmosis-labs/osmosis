use cosmwasm_std::{coins, to_binary, wasm_execute, BankMsg, Empty, Timestamp};
use cosmwasm_std::{Addr, Coin, DepsMut, Response, SubMsg, SubMsgResponse, SubMsgResult};
use swaprouter::msg::ExecuteMsg as SwapRouterExecute;

use crate::checks::{check_is_contract_governor, ensure_key_missing, validate_receiver};
use crate::consts::{MsgReplyID, CALLBACK_KEY, PACKET_LIFETIME};
use crate::ibc::{MsgTransfer, MsgTransferResponse};
use crate::msg::{CrosschainSwapResponse, FailedDeliveryAction, SerializableJson};

use crate::state::{self, Config, CHANNEL_MAP, DISABLED_PREFIXES};
use crate::state::{
    ForwardMsgReplyState, ForwardTo, SwapMsgReplyState, CONFIG, FORWARD_REPLY_STATE,
    INFLIGHT_PACKETS, RECOVERY_STATES, SWAP_REPLY_STATE,
};
use crate::utils::{build_memo, parse_swaprouter_reply};
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
    swap_coin: Coin,
    output_denom: String,
    slippage: swaprouter::Slippage,
    receiver: &str,
    next_memo: Option<SerializableJson>,
    failed_delivery_action: FailedDeliveryAction,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("executing swap and forward"));
    let config = CONFIG.load(deps.storage)?;

    // Message to swap tokens in the underlying swaprouter contract
    let swap_msg = SwapRouterExecute::Swap {
        input_coin: swap_coin.clone(),
        output_denom,
        slippage,
    };
    let msg = wasm_execute(config.swap_contract, &swap_msg, vec![swap_coin])?;

    // Check that the received is valid and retrieve its channel
    let (valid_channel, valid_receiver) = validate_receiver(deps.as_ref(), receiver)?;
    // If there is a memo, check that it is valid (i.e. a valud json object that
    // doesn't contain the key that we will insert later)
    if let Some(memo) = &next_memo {
        // Ensure the json is an object ({...}) and that it does not contain the CALLBACK_KEY
        ensure_key_missing(memo.as_value(), CALLBACK_KEY)?;
    }

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
            block_time,
            contract_addr,
            forward_to: ForwardTo {
                channel: valid_channel,
                receiver: valid_receiver,
                next_memo,
                on_failed_delivery: failed_delivery_action,
            },
        },
    )?;

    Ok(Response::new().add_submessage(SubMsg::reply_on_success(msg, MsgReplyID::Swap.repr())))
}

// The swap has succeeded and we need to generate the forward IBC transfer
pub fn handle_swap_reply(
    deps: DepsMut,
    msg: cosmwasm_std::Reply,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("handle_swap_reply"));
    let swap_msg_state = SWAP_REPLY_STATE.load(deps.storage)?;
    SWAP_REPLY_STATE.remove(deps.storage);

    // Extract the relevant response from the swaprouter reply
    let swap_response = parse_swaprouter_reply(msg)?;

    // Build an IBC packet to forward the swap.
    let contract_addr = &swap_msg_state.contract_addr;
    let ts = swap_msg_state.block_time.plus_seconds(PACKET_LIFETIME);

    // If the memo is provided we want to include it in the IBC message. If not,
    // we default to an empty object. The resulting memo will always include the
    // callback so this contract can track the IBC send
    let memo = build_memo(swap_msg_state.forward_to.next_memo, contract_addr.as_str())?;

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
        memo,
    };

    // Base response
    let response = Response::new()
        .add_attribute("status", "ibc_message_created")
        .add_attribute("ibc_message", format!("{:?}", ibc_transfer));

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
    FORWARD_REPLY_STATE.save(
        deps.storage,
        &ForwardMsgReplyState {
            channel_id: swap_msg_state.forward_to.channel,
            to_address: swap_msg_state.forward_to.receiver.into(),
            amount: swap_response.amount.into(),
            denom: swap_response.token_out_denom,
            on_failed_delivery: swap_msg_state.forward_to.on_failed_delivery,
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
        on_failed_delivery: failed_delivery_action,
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
                sequence: response.sequence,
                amount,
                denom: denom.clone(),
                status: state::ibc::PacketLifecycleStatus::Sent,
            };

            // Save as in-flight to be able to manipulate when the ack/timeout is received
            INFLIGHT_PACKETS.save(deps.storage, (&channel_id, response.sequence), &recovery)?;
        }
    }

    // The response data
    let response_data =
        CrosschainSwapResponse::new(amount, &denom, &channel_id, &to_address, response.sequence);

    Ok(Response::new()
        .set_data(to_binary(&response_data)?)
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

/// Set a prefix->channel map in the registry
pub fn set_channel(
    deps: DepsMut,
    sender: Addr,
    prefix: String,
    channel: String,
) -> Result<Response, ContractError> {
    check_is_contract_governor(deps.as_ref(), sender)?;
    if CHANNEL_MAP.has(deps.storage, &prefix) {
        return Err(ContractError::PrefixAlreadyExists { prefix });
    }
    CHANNEL_MAP.save(deps.storage, &prefix, &channel)?;
    Ok(Response::new().add_attribute("method", "set_channel"))
}

/// Disable a prefix
pub fn disable_prefix(
    deps: DepsMut,
    sender: Addr,
    prefix: String,
) -> Result<Response, ContractError> {
    check_is_contract_governor(deps.as_ref(), sender)?;
    if !CHANNEL_MAP.has(deps.storage, &prefix) {
        return Err(ContractError::PrefixDoesNotExist { prefix });
    }
    DISABLED_PREFIXES.save(deps.storage, &prefix, &Empty {})?;
    Ok(Response::new().add_attribute("method", "disable_prefix"))
}

/// Re-enable a prefix
pub fn re_enable_prefix(
    deps: DepsMut,
    sender: Addr,
    prefix: String,
) -> Result<Response, ContractError> {
    check_is_contract_governor(deps.as_ref(), sender)?;
    if !DISABLED_PREFIXES.has(deps.storage, &prefix) {
        return Err(ContractError::PrefixNotDisabled { prefix });
    }
    DISABLED_PREFIXES.remove(deps.storage, &prefix);
    Ok(Response::new().add_attribute("method", "re_enable_prefix"))
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

    Ok(Response::new().add_attribute("method", "set_swaps_contract"))
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};

    use super::*;
    use crate::{contract, msg::InstantiateMsg, ExecuteMsg};

    static CREATOR_ADDRESS: &str = "creator";
    static SWAPCONTRACT_ADDRESS: &str = "swapcontract";

    static RECEIVER_ADDRESS1: &str = "prefix12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9";
    static RECEIVER_ADDRESS2: &str = "other12smx2wdlyttvyzvzg54y2vnqwq2qjatere840z";

    // test helper
    #[allow(unused_assignments)]
    fn initialize_contract(deps: DepsMut) -> Addr {
        let msg = InstantiateMsg {
            governor: String::from(CREATOR_ADDRESS),
            swap_contract: String::from(SWAPCONTRACT_ADDRESS),
            channels: vec![("prefix".to_string(), "channel1".to_string())],
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
        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "prefix").unwrap(),
            "channel1"
        );
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
    fn modify_channel_registry() {
        let mut deps = mock_dependencies();

        let governor = initialize_contract(deps.as_mut());
        let governor_info = mock_info(governor.as_str(), &vec![] as &Vec<Coin>);
        validate_receiver(deps.as_ref(), RECEIVER_ADDRESS1).unwrap();

        // and new channel
        let msg = ExecuteMsg::SetChannel {
            prefix: "other".to_string(),
            channel: "channel2".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info.clone(), msg).unwrap();

        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "prefix").unwrap(),
            "channel1"
        );
        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "other").unwrap(),
            "channel2"
        );
        validate_receiver(deps.as_ref(), RECEIVER_ADDRESS1).unwrap();
        validate_receiver(deps.as_ref(), RECEIVER_ADDRESS2).unwrap();

        // Can't override an existing channel
        let msg = ExecuteMsg::SetChannel {
            prefix: "prefix".to_string(),
            channel: "new_channel".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info.clone(), msg).unwrap_err();

        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "prefix").unwrap(),
            "channel1"
        );

        // remove channel
        let msg = ExecuteMsg::DisablePrefix {
            prefix: "prefix".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info.clone(), msg).unwrap();

        assert!(DISABLED_PREFIXES.load(&deps.storage, "prefix").is_ok());
        // The prefix no longer validates
        validate_receiver(deps.as_ref(), RECEIVER_ADDRESS1).unwrap_err();

        // Re enable the prefix
        let msg = ExecuteMsg::ReEnablePrefix {
            prefix: "prefix".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), governor_info.clone(), msg).unwrap();
        assert!(DISABLED_PREFIXES.load(&deps.storage, "prefix").is_err());

        // The prefix is allowed again
        validate_receiver(deps.as_ref(), RECEIVER_ADDRESS1).unwrap();
    }

    #[test]
    fn modify_channel_registry_unauthorized() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // A user other than the owner cannot modify the channel registry
        let other_info = mock_info("other_sender", &vec![] as &Vec<Coin>);
        let msg = ExecuteMsg::SetChannel {
            prefix: "other".to_string(),
            channel: "something_else".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), other_info.clone(), msg).unwrap_err();
        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "prefix").unwrap(),
            "channel1"
        );

        let msg = ExecuteMsg::DisablePrefix {
            prefix: "prefix".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), other_info, msg).unwrap_err();
        assert_eq!(
            CHANNEL_MAP.load(&deps.storage, "prefix").unwrap(),
            "channel1"
        );
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
        contract::execute(deps.as_mut(), mock_env(), governor_info.clone(), msg).unwrap();

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
