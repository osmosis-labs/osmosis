use std::str::FromStr;

use cosmwasm_std::{
    coins, Addr, BankMsg, Coin, DepsMut, Env, MessageInfo, Reply, Response, SubMsg, SubMsgResponse,
    SubMsgResult, Uint128,
};
use osmosis_std::types::osmosis::poolmanager::v1beta1::{
    MsgSwapExactAmountIn, MsgSwapExactAmountInResponse, SwapAmountInRoute,
};

use crate::error::ContractError;
use crate::msg::{SwapResponse};
use crate::state::{Config, SwapReplyState, CONFIG, SWAP_REPLY_STATE};

pub const SWAP_REPLY_ID: u64 = 1u64;

fn assert_owner(deps: &DepsMut, sender: &Addr) -> Result<(), ContractError> {
    let cfg = CONFIG.load(deps.storage)?;
    if cfg.owner != *sender {
        return Err(ContractError::Unauthorized {});
    }
    Ok(())
}

pub fn update_affiliate(
    deps: DepsMut,
    info: MessageInfo,
    affiliate_addr: String,
    affiliate_bps: u16,
) -> Result<Response, ContractError> {
    assert_owner(&deps, &info.sender)?;
    if affiliate_bps > 10_000 {
        return Err(ContractError::InvalidAffiliateBps {});
    }
    let addr = deps.api.addr_validate(&affiliate_addr)?;
    CONFIG.update(deps.storage, |mut cfg| -> Result<Config, ContractError> {
        cfg.affiliate_addr = addr;
        cfg.affiliate_bps = affiliate_bps;
        Ok(cfg)
    })?;
    Ok(Response::new().add_attribute("action", "update_affiliate"))
}

pub fn transfer_ownership(
    deps: DepsMut,
    info: MessageInfo,
    new_owner: String,
) -> Result<Response, ContractError> {
    assert_owner(&deps, &info.sender)?;
    let new_owner = deps.api.addr_validate(&new_owner)?;
    CONFIG.update(deps.storage, |mut cfg| -> Result<Config, ContractError> {
        cfg.owner = new_owner;
        Ok(cfg)
    })?;
    Ok(Response::new().add_attribute("action", "transfer_ownership"))
}

pub fn swap_with_fee(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    input_coin: Coin,
    output_denom: String,
    min_output_amount: Uint128,
    route: Vec<SwapAmountInRoute>,
) -> Result<Response, ContractError> {
    // Require the input funds to match
    if !info.funds.iter().any(|c| c.denom == input_coin.denom && c.amount == input_coin.amount) {
        return Err(ContractError::InsufficientFunds {});
    }

    let min_out = Coin::new(min_output_amount.u128(), output_denom.clone());

    let swap_msg = MsgSwapExactAmountIn {
        sender: env.contract.address.into_string(),
        routes: route,
        token_in: Some(input_coin.clone().into()),
        token_out_min_amount: min_out.amount.to_string(),
    };

    SWAP_REPLY_STATE.save(
        deps.storage,
        &SwapReplyState {
            original_sender: info.sender.clone(),
            swap_msg: swap_msg.clone(),
        },
    )?;

    Ok(Response::new()
        .add_attribute("action", "swap_with_fee")
        .add_submessage(SubMsg::reply_on_success(swap_msg, SWAP_REPLY_ID)))
}

pub fn handle_swap_reply(
    deps: DepsMut,
    msg: Reply,
) -> Result<Response, ContractError> {
    let state = SWAP_REPLY_STATE.load(deps.storage)?;
    // Clear saved state
    SWAP_REPLY_STATE.remove(deps.storage);

    if let SubMsgResult::Ok(SubMsgResponse { data: Some(b), .. }) = msg.result.clone() {
        let res: MsgSwapExactAmountInResponse = b.try_into().map_err(ContractError::Std)?;
        let amount = Uint128::from_str(&res.token_out_amount)?;

        // Determine output denom from last route element
        let token_out_denom = state
            .swap_msg
            .routes
            .last()
            .map(|r| r.token_out_denom.clone())
            .unwrap_or_default();

        let cfg = CONFIG.load(deps.storage)?;

        let affiliate_amount = amount.multiply_ratio(cfg.affiliate_bps as u128, 10_000u128);
        let user_amount = amount.checked_sub(affiliate_amount).unwrap();

        let mut msgs = vec![];
        if !affiliate_amount.is_zero() {
            msgs.push(BankMsg::Send {
                to_address: cfg.affiliate_addr.into_string(),
                amount: coins(affiliate_amount.u128(), token_out_denom.clone()),
            }.into());
        }
        if !user_amount.is_zero() {
            msgs.push(BankMsg::Send {
                to_address: state.original_sender.into_string(),
                amount: coins(user_amount.u128(), token_out_denom.clone()),
            }.into());
        }

        let response = SwapResponse {
            original_sender: state.original_sender.into_string(),
            token_out_denom,
            amount_sent_to_user: user_amount,
            amount_sent_to_affiliate: affiliate_amount,
        };

        return Ok(Response::new()
            .add_messages(msgs)
            .set_data(cosmwasm_std::to_binary(&response)?)
            .add_attribute("token_out_amount", amount)
            .add_attribute("affiliate_bps", cfg.affiliate_bps.to_string()));
    }

    Err(ContractError::FailedSwap {
        reason: format!("{:?}", msg.result.unwrap_err()),
    })
}
