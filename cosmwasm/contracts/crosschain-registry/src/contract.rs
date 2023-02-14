#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::CONTRACT_MAP;
use crate::{execute};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:crosschain-registry";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SetContractAlias { contract_alias, contract_address } => {
            execute::set_contract_alias(deps, contract_alias, contract_address)
        }
        ExecuteMsg::ChangeContractAlias { current_contract_alias, new_contract_alias } => {
            execute::change_contract_alias(deps, current_contract_alias, new_contract_alias)
        }
        ExecuteMsg::RemoveContractAlias { contract_alias } => {
            execute::remove_contract_alias(deps, contract_alias)
        }
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetAddressFromAlias { contract_alias } => to_binary(
            &CONTRACT_MAP
                .load(deps.storage, &contract_alias)?
        ),
    }
}

#[cfg(test)]
mod tests {}
