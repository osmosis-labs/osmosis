use crate::state::CONTRACT_MAP;
use cosmwasm_std::{DepsMut, Response};

use crate::ContractError;

/// Set a alias->address map in the registry
pub fn set_contract_alias(
    deps: DepsMut,
    alias: String,
    address: String,
) -> Result<Response, ContractError> {
    if CONTRACT_MAP.has(deps.storage, &alias) {
        return Err(ContractError::AliasAlreadyExists { alias });
    }
    CONTRACT_MAP.save(deps.storage, &alias, &address)?;
    Ok(Response::new().add_attribute("method", "set_contract_alias"))
}

/// Change an existing alias->address map in the registry
pub fn change_contract_alias(
    deps: DepsMut,
    current_alias: String,
    new_alias: String,
) -> Result<Response, ContractError> {
    let address = CONTRACT_MAP
        .load(deps.storage, &current_alias)
        .map_err(|_| ContractError::AliasDoesNotExist { current_alias })?;
    CONTRACT_MAP.save(deps.storage, &new_alias, &address)?;
    Ok(Response::new().add_attribute("method", "change_contract_alias"))
}

/// Remove an existing alias->address map in the registry
pub fn remove_contract_alias(
    deps: DepsMut,
    current_alias: &str,
) -> Result<Response, ContractError> {
    CONTRACT_MAP
        .load(deps.storage, current_alias)
        .map_err(|_| ContractError::AliasDoesNotExist {
            current_alias: current_alias.to_string(),
        })?;
    CONTRACT_MAP.remove(deps.storage, &current_alias);
    Ok(Response::new().add_attribute("method", "remove_contract_alias"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::contract;
    use crate::msg::ExecuteMsg;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    static CREATOR_ADDRESS: &str = "creator";

    #[test]
    fn test_set_contract_alias_success() {
        let mut deps = mock_dependencies();
        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();

        // Set contract alias swap_router to osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9
        let msg = ExecuteMsg::SetContractAlias {
            contract_alias: alias.clone(),
            contract_address: address.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        assert_eq!(
            CONTRACT_MAP.load(&deps.storage, "swap_router").unwrap(),
            "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9"
        );
    }

    #[test]
    fn test_set_contract_alias_fail_existing_alias() {
        let mut deps = mock_dependencies();
        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();

        // Set contract alias swap_router to an address
        let msg = ExecuteMsg::SetContractAlias {
            contract_alias: alias.clone(),
            contract_address: address.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set contract alias swap_router to a different address
        let msg = ExecuteMsg::SetContractAlias {
            contract_alias: alias.clone(),
            contract_address: "osmo1fsdaf7dsfasndjklk3jndskajnfkdjsfjn3jka".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::AliasAlreadyExists { alias };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            CONTRACT_MAP.load(&deps.storage, "swap_router").unwrap(),
            "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9"
        );
    }

    #[test]
    fn test_change_contract_alias_success() {
        let mut deps = mock_dependencies();
        let current_alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();
        let new_alias = "new_swap_router".to_string();

        // Set contract alias swap_router to an address
        let msg = ExecuteMsg::SetContractAlias {
            contract_alias: current_alias.clone(),
            contract_address: address.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change the contract alias swap_router to new_swap_router
        let msg = ExecuteMsg::ChangeContractAlias {
            current_contract_alias: current_alias.clone(),
            new_contract_alias: new_alias.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that the contract alias has changed from swap_router to new_swap_router
        assert_eq!(
            CONTRACT_MAP.load(&deps.storage, "new_swap_router").unwrap(),
            "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9"
        );
    }

    #[test]
    fn test_change_contract_alias_fail_non_existing_alias() {
        let mut deps = mock_dependencies();
        let current_alias = "swap_router".to_string();
        let new_alias = "new_swap_router".to_string();

        // Attempt to change an alias that does not exist
        let msg = ExecuteMsg::ChangeContractAlias {
            current_contract_alias: current_alias.clone(),
            new_contract_alias: new_alias.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::AliasDoesNotExist { current_alias };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_contract_alias_success() {
        let mut deps = mock_dependencies();
        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();

        // Set contract alias swap_router to osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9
        let msg = ExecuteMsg::SetContractAlias {
            contract_alias: alias.clone(),
            contract_address: address.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Remove the alias
        let msg = ExecuteMsg::RemoveContractAlias {
            contract_alias: alias,
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Verify that the alias no longer exists
        assert!(!CONTRACT_MAP.has(&deps.storage, "swap_router"));
    }

    #[test]
    fn test_remove_contract_alias_fail_nonexistent_alias() {
        let mut deps = mock_dependencies();
        let current_alias = "swap_router".to_string();

        // Attempt to remove an alias that does not exist
        let msg = ExecuteMsg::RemoveContractAlias {
            contract_alias: current_alias.clone(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::AliasDoesNotExist { current_alias };
        assert_eq!(result.unwrap_err(), expected_error);
    }
}
