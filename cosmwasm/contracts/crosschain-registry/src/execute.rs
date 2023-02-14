use crate::helpers::*;
use crate::state::{ASSET_MAP, CHAIN_CHANNEL_MAP, CONTRACT_MAP};
use cosmwasm_std::{DepsMut, Response};

use crate::ContractError;

/// Contract Registry

// Set a alias->address map in the registry
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

// Change an existing alias->address map in the registry
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

// Remove an existing alias->address map in the registry
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

/// Chain/Channel Registry

// Set a alias->address map in the registry
pub fn set_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
    channel_id: String,
) -> Result<Response, ContractError> {
    let key = make_chain_channel_key(&source_chain, &destination_chain);
    if CHAIN_CHANNEL_MAP.has(deps.storage, &key) {
        return Err(ContractError::ChainChannelLinkAlreadyExists {
            source_chain,
            destination_chain,
        });
    }
    CHAIN_CHANNEL_MAP.save(deps.storage, &key, &channel_id)?;
    Ok(Response::new().add_attribute("method", "set_chain_channel_link"))
}

// Change an existing alias->address map in the registry
pub fn change_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
    new_channel_id: String,
) -> Result<Response, ContractError> {
    let key = make_chain_channel_key(&source_chain, &destination_chain);
    CHAIN_CHANNEL_MAP.load(deps.storage, &key).map_err(|_| {
        ContractError::ChainChannelLinkDoesNotExist {
            source_chain,
            destination_chain,
        }
    })?;
    CHAIN_CHANNEL_MAP.save(deps.storage, &key, &new_channel_id)?;
    Ok(Response::new().add_attribute("method", "change_chain_channel_link"))
}

// Remove an existing alias->address map in the registry
pub fn remove_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
) -> Result<Response, ContractError> {
    let key = make_chain_channel_key(&source_chain, &destination_chain);
    CHAIN_CHANNEL_MAP.load(deps.storage, &key).map_err(|_| {
        ContractError::ChainChannelLinkDoesNotExist {
            source_chain,
            destination_chain,
        }
    })?;
    CHAIN_CHANNEL_MAP.remove(deps.storage, &key);
    Ok(Response::new().add_attribute("method", "remove_chain_channel_link"))
}

/// Asset Registry

// Set a mapping of a native denom on another chain to its corresponding denom on the destination chain
pub fn set_asset_map(
    deps: DepsMut,
    native_denom: String,
    destination_chain: String,
    destination_chain_denom: String,
) -> Result<Response, ContractError> {
    let key = make_asset_key(&native_denom, &destination_chain);
    if ASSET_MAP.has(deps.storage, &key) {
        return Err(ContractError::AssetMapLinkAlreadyExists {
            native_denom,
            destination_chain,
        });
    }
    ASSET_MAP.save(deps.storage, &key, &destination_chain_denom)?;
    Ok(Response::new().add_attribute("method", "set_asset_map"))
}

// Change an existing asset map in the registry
pub fn change_asset_map(
    deps: DepsMut,
    native_denom: String,
    destination_chain: String,
    new_destination_chain_denom: String,
) -> Result<Response, ContractError> {
    let key = make_asset_key(&native_denom, &destination_chain);
    ASSET_MAP
        .load(deps.storage, &key)
        .map_err(|_| ContractError::AssetMapLinkDoesNotExist {
            native_denom,
            destination_chain,
        })?;
    ASSET_MAP.save(deps.storage, &key, &new_destination_chain_denom)?;
    Ok(Response::new().add_attribute("method", "change_asset_map"))
}

// Remove an existing asset map in the registry
pub fn remove_asset_map(
    deps: DepsMut,
    native_denom: String,
    destination_chain: String,
) -> Result<Response, ContractError> {
    let key = make_asset_key(&native_denom, &destination_chain);
    ASSET_MAP
        .load(deps.storage, &key)
        .map_err(|_| ContractError::AssetMapLinkDoesNotExist {
            native_denom,
            destination_chain,
        })?;
    ASSET_MAP.remove(deps.storage, &key);
    Ok(Response::new().add_attribute("method", "remove_asset_map"))
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

        // Set contract alias swap_router to an address
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

        // Verify that the contract alias has changed from "swap_router" to "new_swap_router"
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

        // Set contract alias swap_router to an address
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

    #[test]
    fn test_set_chain_channel_link_success() {
        let mut deps = mock_dependencies();

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        let key = make_chain_channel_key("osmosis", "cosmos");

        assert_eq!(
            CHAIN_CHANNEL_MAP
                .load(&deps.storage, &key.to_string())
                .unwrap(),
            "channel-0"
        );
    }

    #[test]
    fn test_set_chain_channel_link_fail_existing_link() {
        let mut deps = mock_dependencies();
        let key = make_chain_channel_key("osmosis", "cosmos");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set the canonical channel link between osmosis and cosmos to channel-150
        // This should fail because the link already exists
        let msg = ExecuteMsg::SetChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-150".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::ChainChannelLinkAlreadyExists {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            CHAIN_CHANNEL_MAP
                .load(&deps.storage, &key.to_string())
                .unwrap(),
            "channel-0"
        );
    }

    #[test]
    fn test_change_chain_channel_link_success() {
        let mut deps = mock_dependencies();
        let key = make_chain_channel_key("osmosis", "cosmos");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change the canonical channel link between osmosis and cosmos to channel-150
        let msg = ExecuteMsg::ChangeChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            new_channel_id: "channel-150".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that the channel between osmosis and cosmos has changed from channel-0 to channel-150
        assert_eq!(
            CHAIN_CHANNEL_MAP
                .load(&deps.storage, &key.to_string())
                .unwrap(),
            "channel-150"
        );
    }

    #[test]
    fn test_change_chain_channel_link_fail_non_existing_link() {
        let mut deps = mock_dependencies();

        // Attempt to change a channel link that does not exist
        let msg = ExecuteMsg::ChangeChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            new_channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::ChainChannelLinkDoesNotExist {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_chain_channel_link_success() {
        let mut deps = mock_dependencies();
        let key = make_chain_channel_key("osmosis", "cosmos");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Remove the link
        let msg = ExecuteMsg::RemoveChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Verify that the link no longer exists
        assert!(!CHAIN_CHANNEL_MAP.has(&deps.storage, &key.to_string()));
    }

    #[test]
    fn test_remove_chain_channel_link_fail_nonexistent_link() {
        let mut deps = mock_dependencies();

        // Attempt to remove a link that does not exist
        let msg = ExecuteMsg::RemoveChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::ChainChannelLinkDoesNotExist {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_set_denom_map_success() {
        let mut deps = mock_dependencies();

        // Set contract alias swap_router to an address
        let msg = ExecuteMsg::SetAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            destination_chain_denom:
                "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        let key = make_asset_key("ustars", "osmosis");

        assert_eq!(
            ASSET_MAP.load(&deps.storage, &key.to_string()).unwrap(),
            "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4"
        );
    }

    #[test]
    fn test_set_denom_map_fail_existing_link() {
        let mut deps = mock_dependencies();
        let key = make_asset_key("ustars", "osmosis");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            destination_chain_denom:
                "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set the canonical channel link between osmosis and cosmos to channel-150
        // This should fail because the link already exists
        let msg = ExecuteMsg::SetAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            destination_chain_denom:
                "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::AssetMapLinkAlreadyExists {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            ASSET_MAP.load(&deps.storage, &key.to_string()).unwrap(),
            "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4"
        );
    }

    #[test]
    fn test_change_denom_map_success() {
        let mut deps = mock_dependencies();
        let key = make_asset_key("ustars", "osmosis");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            destination_chain_denom:
                "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change the canonical channel link between osmosis and cosmos to channel-150
        let msg = ExecuteMsg::ChangeAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            new_destination_chain_denom:
                "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that the channel between osmosis and cosmos has changed from channel-0 to channel-150
        assert_eq!(
            ASSET_MAP.load(&deps.storage, &key.to_string()).unwrap(),
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
        );
    }

    #[test]
    fn test_change_denom_map_fail_non_existing_link() {
        let mut deps = mock_dependencies();

        // Attempt to change a channel link that does not exist
        let msg = ExecuteMsg::ChangeAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            new_destination_chain_denom:
                "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::AssetMapLinkDoesNotExist {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_asset_map_success() {
        let mut deps = mock_dependencies();
        let key = make_asset_key("ustars", "osmosis");

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
            destination_chain_denom:
                "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Remove the link
        let msg = ExecuteMsg::RemoveAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Verify that the link no longer exists
        assert!(!ASSET_MAP.has(&deps.storage, &key.to_string()));
    }

    #[test]
    fn test_remove_asset_map_fail_nonexistent_link() {
        let mut deps = mock_dependencies();

        // Attempt to remove a link that does not exist
        let msg = ExecuteMsg::RemoveAssetMapping {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::AssetMapLinkDoesNotExist {
            native_denom: "ustars".to_string(),
            destination_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }
}
