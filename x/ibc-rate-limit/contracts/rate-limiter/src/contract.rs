#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg, SudoMsg};
use crate::state::{FlowType, Path, GOVMODULE, IBCMODULE};
use crate::{execute, query, sudo};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:rate-limiter";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    IBCMODULE.save(deps.storage, &msg.ibc_module)?;
    GOVMODULE.save(deps.storage, &msg.gov_module)?;

    execute::add_new_paths(deps, msg.paths, env.block.time)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("ibc_module", msg.ibc_module.to_string())
        .add_attribute("gov_module", msg.gov_module.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::AddPath {
            channel_id,
            denom,
            quotas,
        } => execute::try_add_path(deps, info.sender, channel_id, denom, quotas, env.block.time),
        ExecuteMsg::RemovePath { channel_id, denom } => {
            execute::try_remove_path(deps, info.sender, channel_id, denom)
        }
        ExecuteMsg::ResetPathQuota {
            channel_id,
            denom,
            quota_id,
        } => execute::try_reset_path_quota(
            deps,
            info.sender,
            channel_id,
            denom,
            quota_id,
            env.block.time,
        ),
    }
}

#[entry_point]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::SendPacket {
            channel_id,
            channel_value,
            funds,
            denom,
        } => sudo::try_transfer(
            deps,
            &Path::new(&channel_id, &denom),
            channel_value,
            funds,
            FlowType::Out,
            env.block.time,
        ),
        SudoMsg::RecvPacket {
            channel_id,
            channel_value,
            funds,
            denom,
        } => sudo::try_transfer(
            deps,
            &Path::new(&channel_id, &denom),
            channel_value,
            funds,
            FlowType::In,
            env.block.time,
        ),
        SudoMsg::UndoSend {
            channel_id,
            denom,
            funds,
        } => sudo::undo_send(deps, &Path::new(&channel_id, &denom), funds),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetQuotas { channel_id, denom } => query::get_quotas(deps, channel_id, denom),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(_deps: DepsMut, _env: Env, _msg: MigrateMsg) -> Result<Response, ContractError> {
    unimplemented!()
}
