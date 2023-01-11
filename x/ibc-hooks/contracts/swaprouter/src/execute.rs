use std::convert::TryInto;
use std::str::FromStr;

use cosmwasm_std::{
    coin, coins, has_coins, to_binary, BankMsg, Coin, DepsMut, Env, MessageInfo, Reply, Response,
    SubMsg, SubMsgResponse, SubMsgResult, Uint128,
};
use osmosis_std::types::osmosis::gamm::v1beta1::{MsgSwapExactAmountInResponse, SwapAmountInRoute};

use crate::contract::SWAP_REPLY_ID;
use crate::error::ContractError;
use crate::helpers::{
    calculate_min_output_from_twap, check_is_contract_owner, generate_swap_msg, validate_pool_route,
};
use crate::msg::{Slippage, SwapResponse};
use crate::state::{State, SwapMsgReplyState, ROUTING_TABLE, STATE, SWAP_REPLY_STATES};

pub fn set_route(
    deps: DepsMut,
    info: MessageInfo,
    input_denom: String,
    output_denom: String,
    pool_route: Vec<SwapAmountInRoute>,
) -> Result<Response, ContractError> {
    // only owner
    check_is_contract_owner(deps.as_ref(), info.sender)?;

    validate_pool_route(
        deps.as_ref(),
        input_denom.clone(),
        output_denom.clone(),
        pool_route.clone(),
    )?;

    ROUTING_TABLE.save(deps.storage, (&input_denom, &output_denom), &pool_route)?;

    Ok(Response::new().add_attribute("action", "set_route"))
}

pub fn transfer_ownership(
    deps: DepsMut,
    info: MessageInfo,
    new_owner: String,
) -> Result<Response, ContractError> {
    // only owner can transfer
    println!("{new_owner}");
    check_is_contract_owner(deps.as_ref(), info.sender)?;
    let owner = deps.api.addr_validate(&new_owner)?;

    println!("{owner}");

    STATE.save(deps.storage, &State { owner })?;

    Ok(Response::new().add_attribute("action", "transfer_ownership"))
}

pub fn trade_with_slippage_limit(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    input_token: Coin,
    output_denom: String,
    slippage: Slippage,
) -> Result<Response, ContractError> {
    if !has_coins(&info.funds, &input_token) {
        return Err(ContractError::InsufficientFunds {});
    }

    let min_output_token = match slippage {
        Slippage::Twap {
            window_seconds,
            slippage_percentage,
        } => calculate_min_output_from_twap(
            deps.as_ref(),
            input_token.clone(),
            output_denom,
            env.block.time,
            window_seconds,
            slippage_percentage,
        )?,
        Slippage::MinOutputAmount(minimum_output_amount) => {
            coin(minimum_output_amount.u128(), output_denom)
        }
    };

    // generate the swap_msg
    let swap_msg = generate_swap_msg(
        deps.as_ref(),
        env.contract.address,
        input_token,
        min_output_token,
    )?;

    // save intermediate state for reply
    SWAP_REPLY_STATES.save(
        deps.storage,
        SWAP_REPLY_ID,
        &SwapMsgReplyState {
            original_sender: info.sender,
            swap_msg: swap_msg.clone(),
        },
    )?;

    Ok(Response::new()
        .add_attribute("action", "trade_with_slippage_limit")
        .add_submessage(SubMsg::reply_on_success(swap_msg, SWAP_REPLY_ID)))
}

pub fn handle_swap_reply(
    _deps: DepsMut,
    msg: Reply,
    swap_msg_reply_state: SwapMsgReplyState,
) -> Result<Response, ContractError> {
    if let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result {
        let res: MsgSwapExactAmountInResponse = b.try_into().map_err(ContractError::Std)?;

        let amount = Uint128::from_str(&res.token_out_amount)?;

        let token_out_denom = &swap_msg_reply_state
            .swap_msg
            .routes
            .last()
            .unwrap()
            .token_out_denom;

        let bank_msg = BankMsg::Send {
            to_address: swap_msg_reply_state.original_sender.clone().into_string(),
            amount: coins(amount.u128(), token_out_denom),
        };

        let response = SwapResponse {
            original_sender: swap_msg_reply_state.original_sender.into_string(),
            token_out_denom: token_out_denom.to_string(),
            amount: amount.u128().into(),
        };

        return Ok(Response::new()
            .add_message(bank_msg)
            .set_data(to_binary(&response)?)
            .add_attribute("token_out_amount", amount));
    }

    Err(ContractError::FailedSwap {
        reason: msg.result.unwrap_err(),
    })
}
