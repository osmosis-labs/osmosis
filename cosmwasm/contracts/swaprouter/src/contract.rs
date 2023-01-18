#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Reply, Response, StdResult,
};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::execute::{handle_swap_reply, set_route, trade_with_slippage_limit, transfer_ownership};
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::query::{query_owner, query_route};
use crate::state::{State, STATE, SWAP_REPLY_STATES};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:swaprouter";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// Msg Reply IDs
pub const SWAP_REPLY_ID: u64 = 1u64;

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // set contract version
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    // validate owner address and save to state
    let owner = deps.api.addr_validate(&msg.owner)?;
    let state = State { owner };
    STATE.save(deps.storage, &state)?;

    // return OK
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SetRoute {
            input_denom,
            output_denom,
            pool_route,
        } => set_route(deps, info, input_denom, output_denom, pool_route),
        ExecuteMsg::Swap {
            input_coin,
            output_denom,
            slippage,
        } => trade_with_slippage_limit(deps, env, info, input_coin, output_denom, slippage),
        ExecuteMsg::TransferOwnership { new_owner } => transfer_ownership(deps, info, new_owner),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetOwner {} => to_binary(&query_owner(deps)?),
        QueryMsg::GetRoute {
            input_denom,
            output_denom,
        } => to_binary(&query_route(deps, &input_denom, &output_denom)?),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, msg: Reply) -> Result<Response, ContractError> {
    deps.api
        .debug(&format!("executing swaprouter reply: {msg:?}"));
    if msg.id == SWAP_REPLY_ID {
        // get intermediate swap reply state. Error if not found.
        let swap_msg_state = SWAP_REPLY_STATES.load(deps.storage, msg.id)?;

        // prune intermedate state
        SWAP_REPLY_STATES.remove(deps.storage, msg.id);

        // call reply function to handle the swap return
        handle_swap_reply(deps, msg, swap_msg_state)
    } else {
        Ok(Response::new())
    }
}
