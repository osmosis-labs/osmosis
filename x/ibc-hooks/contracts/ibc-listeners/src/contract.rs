#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{EventType, ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg, SudoMsg};
use crate::state::LISTENERS;

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:{{project-name}}";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Handling contract instantiation
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

/// Handling contract execution
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Subscribe { sequence, event } => subscribe(deps, info.sender, sequence, event),
    }
}

fn subscribe(
    deps: DepsMut,
    contract: Addr,
    sequence: u64,
    event: EventType,
) -> Result<Response, ContractError> {
    LISTENERS.update(
        deps.storage,
        (sequence, &event.to_string()),
        |maybe_list| -> Result<Vec<Addr>, ContractError> {
            let Some(mut list) = maybe_list else {
                return Ok(vec![contract]);
            };
            list.push(contract);
            Ok(list)
        },
    )?;
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    // TODO: Make this more resiliant by making sure all variants are covered
    match msg {
        SudoMsg::UnSubscribeAll { sequence } => {
            LISTENERS.remove(
                deps.storage,
                (sequence, &EventType::Acknowledgement {}.to_string()),
            );
            LISTENERS.remove(deps.storage, (sequence, &EventType::Timeout {}.to_string()));
            Ok(Response::new())
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Listeners { sequence, event } => {
            let maybe_list: Option<Vec<Addr>> =
                LISTENERS.may_load(deps.storage, (sequence, &event.to_string()))?;
            to_binary(&match maybe_list {
                Some(list) => list,
                None => vec![],
            })
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(_deps: DepsMut, _env: Env, msg: MigrateMsg) -> Result<Response, ContractError> {
    match msg {}
}
