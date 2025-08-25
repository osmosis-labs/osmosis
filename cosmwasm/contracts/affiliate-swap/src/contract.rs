#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Reply, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::execute::{handle_swap_reply, swap_with_fee, transfer_ownership, update_affiliate};
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::query::query_config;
use crate::state::{Config, CONFIG};
use crate::execute::SWAP_REPLY_ID;

const CONTRACT_NAME: &str = "crates.io:affiliate-swap";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    if msg.affiliate_bps > 10_000 {
        return Err(ContractError::InvalidAffiliateBps {});
    }
    let owner = deps.api.addr_validate(&msg.owner)?;
    let affiliate_addr = deps.api.addr_validate(&msg.affiliate_addr)?;
    let cfg = Config {
        owner,
        affiliate_addr,
        affiliate_bps: msg.affiliate_bps,
    };
    CONFIG.save(deps.storage, &cfg)?;
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
        ExecuteMsg::SwapWithFee { input_coin, output_denom, min_output_amount, route } => {
            swap_with_fee(deps, env, info, input_coin, output_denom, min_output_amount, route)
        }
        ExecuteMsg::UpdateAffiliate { affiliate_addr, affiliate_bps } => {
            update_affiliate(deps, info, affiliate_addr, affiliate_bps)
        }
        ExecuteMsg::TransferOwnership { new_owner } => transfer_ownership(deps, info, new_owner),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {} => to_binary(&query_config(deps)?),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, msg: Reply) -> Result<Response, ContractError> {
    if msg.id == SWAP_REPLY_ID {
        handle_swap_reply(deps, msg)
    } else {
        Ok(Response::new())
    }
}

