#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::CONTRACT_NAMES;

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:cannonical-names";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    Ok(Response::default())
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SetName { name, address } => {
            CONTRACT_NAMES.save(deps.storage, &name, &address)?;
            set_names(deps, vec![(name, address)])?;
            Ok(Response::default().add_attribute("stuff", "value".to_string()))
        }
        _ => Err(ContractError::Unauthorized {}),
    }
}

fn set_names(deps: DepsMut, names: Vec<(String, String)>) -> Result<i32, ContractError> {
    for (name, address) in names {
        CONTRACT_NAMES.save(deps.storage, &name, &address)?;
        if name == "foo" {
            return Err(ContractError::Unauthorized {});
        }
    }
    Ok(32)
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: QueryMsg) -> StdResult<Binary> {
    unimplemented!()
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, from_binary, CosmosMsg, WasmMsg};

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {};
        let info = mock_info("creator", &coins(1000, "earth"));
        let env = mock_env();

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), env.clone(), info.clone(), msg)
            .expect("couldn't instantiate contract");
    }

    #[test]
    fn testing_the_tests() {
        assert_eq!(1, 1);
    }
}
