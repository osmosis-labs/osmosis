use crate::helpers::*;
use crate::state::{
    CHAIN_TO_CHAIN_CHANNEL_MAP, CHANNEL_TO_CHAIN_CHAIN_MAP, CONTRACT_ALIAS_MAP,
    NATIVE_DENOM_TO_IBC_DENOM_MAP,
};
use cosmwasm_std::{Deps, DepsMut, Response, StdError};

use crate::ContractError;

// Contract Registry

// Set a alias->address map in the registry
pub fn set_contract_alias(
    deps: DepsMut,
    alias: String,
    address: String,
) -> Result<Response, ContractError> {
    if CONTRACT_ALIAS_MAP.has(deps.storage, &alias) {
        return Err(ContractError::AliasAlreadyExists { alias });
    }
    CONTRACT_ALIAS_MAP.save(deps.storage, &alias, &address)?;
    Ok(Response::new().add_attribute("method", "set_contract_alias"))
}

// Change an existing alias->address map in the registry
pub fn change_contract_alias(
    deps: DepsMut,
    current_alias: String,
    new_alias: String,
) -> Result<Response, ContractError> {
    let address = CONTRACT_ALIAS_MAP
        .load(deps.storage, &current_alias)
        .map_err(|_| ContractError::AliasDoesNotExist { current_alias })?;
    CONTRACT_ALIAS_MAP.save(deps.storage, &new_alias, &address)?;
    Ok(Response::new().add_attribute("method", "change_contract_alias"))
}

// Remove an existing alias->address map in the registry
pub fn remove_contract_alias(
    deps: DepsMut,
    current_alias: &str,
) -> Result<Response, ContractError> {
    CONTRACT_ALIAS_MAP
        .load(deps.storage, current_alias)
        .map_err(|_| ContractError::AliasDoesNotExist {
            current_alias: current_alias.to_string(),
        })?;
    CONTRACT_ALIAS_MAP.remove(deps.storage, &current_alias);
    Ok(Response::new().add_attribute("method", "remove_contract_alias"))
}

// Chain to Chain Channel Registry

// Set a chain pair->channel_id map in the registry
pub fn set_chain_to_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
    channel_id: String,
) -> Result<Response, ContractError> {
    if CHAIN_TO_CHAIN_CHANNEL_MAP.has(deps.storage, (&source_chain, &destination_chain)) {
        return Err(ContractError::ChainToChainChannelLinkAlreadyExists {
            source_chain,
            destination_chain,
        });
    }
    CHAIN_TO_CHAIN_CHANNEL_MAP.save(
        deps.storage,
        (&source_chain, &destination_chain),
        &channel_id,
    )?;
    Ok(Response::new().add_attribute("method", "set_chain_to_chain_channel_link"))
}

// Change an existing chain pair->channel_id map in the registry
pub fn change_chain_to_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
    new_channel_id: String,
) -> Result<Response, ContractError> {
    CHAIN_TO_CHAIN_CHANNEL_MAP
        .load(deps.storage, (&source_chain, &destination_chain))
        .map_err(|_| ContractError::ChainChannelLinkDoesNotExist {
            source_chain: source_chain.clone(),
            destination_chain: destination_chain.clone(),
        })?;
    CHAIN_TO_CHAIN_CHANNEL_MAP.save(
        deps.storage,
        (&source_chain, &destination_chain),
        &new_channel_id,
    )?;
    Ok(Response::new().add_attribute("method", "change_chain_channel_link"))
}

// Remove an existing chain pair->channel_id map in the registry
pub fn remove_chain_to_chain_channel_link(
    deps: DepsMut,
    source_chain: String,
    destination_chain: String,
) -> Result<Response, ContractError> {
    CHAIN_TO_CHAIN_CHANNEL_MAP
        .load(deps.storage, (&source_chain, &destination_chain))
        .map_err(|_| ContractError::ChainChannelLinkDoesNotExist {
            source_chain: source_chain.clone(),
            destination_chain: destination_chain.clone(),
        })?;
    CHAIN_TO_CHAIN_CHANNEL_MAP.remove(deps.storage, (&source_chain, &destination_chain));
    Ok(Response::new().add_attribute("method", "remove_chain_channel_link"))
}

// Channel to Chain Chain Registry

// Set a channel_id->chain pair map in the registry
pub fn set_channel_to_chain_chain_link(
    deps: DepsMut,
    channel_id: String,
    source_chain: String,
    destination_chain: String,
) -> Result<Response, ContractError> {
    if CHANNEL_TO_CHAIN_CHAIN_MAP.has(deps.storage, (&channel_id, &source_chain)) {
        return Err(ContractError::ChannelToChainChainLinkAlreadyExists {
            channel_id,
            source_chain,
        });
    }
    CHANNEL_TO_CHAIN_CHAIN_MAP.save(
        deps.storage,
        (&channel_id, &source_chain),
        &destination_chain,
    )?;
    Ok(Response::new().add_attribute("method", "set_channel_to_chain_chain_link"))
}

// Change an existing channel_id->chain pair map in the registry
pub fn change_channel_to_chain_chain_link(
    deps: DepsMut,
    channel_id: String,
    source_chain: String,
    new_destination_chain: String,
) -> Result<Response, ContractError> {
    CHANNEL_TO_CHAIN_CHAIN_MAP
        .load(deps.storage, (&channel_id, &source_chain))
        .map_err(|_| ContractError::ChannelToChainChainLinkDoesNotExist {
            channel_id: channel_id.clone(),
            source_chain: source_chain.clone(),
        })?;
    CHANNEL_TO_CHAIN_CHAIN_MAP.save(
        deps.storage,
        (&channel_id, &source_chain),
        &new_destination_chain,
    )?;
    Ok(Response::new().add_attribute("method", "change_channel_to_chain_chain_link"))
}

// Remove an existing channel_id->chain pair map in the registry
pub fn remove_channel_to_chain_chain_link(
    deps: DepsMut,
    channel_id: String,
    source_chain: String,
) -> Result<Response, ContractError> {
    CHANNEL_TO_CHAIN_CHAIN_MAP
        .load(deps.storage, (&channel_id, &source_chain))
        .map_err(|_| ContractError::ChannelToChainChainLinkDoesNotExist {
            channel_id: channel_id.clone(),
            source_chain: source_chain.clone(),
        })?;
    CHANNEL_TO_CHAIN_CHAIN_MAP.remove(deps.storage, (&channel_id, &source_chain));
    Ok(Response::new().add_attribute("method", "remove_channel_to_chain_chain_link"))
}

// Osmosis Denom Registry

// Set a native_denom->ibc_denom map in the registry
pub fn set_native_denom_to_ibc_denom_link(
    deps: DepsMut,
    native_denom: String,
    ibc_denom: String,
) -> Result<Response, ContractError> {
    if NATIVE_DENOM_TO_IBC_DENOM_MAP.has(deps.storage, &native_denom) {
        return Err(ContractError::NativeDenomLinkAlreadyExists { native_denom });
    }
    NATIVE_DENOM_TO_IBC_DENOM_MAP.save(deps.storage, &native_denom, &ibc_denom)?;
    Ok(Response::new().add_attribute("method", "set_native_denom_to_ibc_denom_link"))
}

// Change an existing native_denom->ibc_denom map in the registry
pub fn change_native_denom_to_ibc_denom_link(
    deps: DepsMut,
    native_denom: String,
    new_ibc_denom: String,
) -> Result<Response, ContractError> {
    let denom = native_denom.clone(); // create a clone of native_denom
    NATIVE_DENOM_TO_IBC_DENOM_MAP
        .load(deps.storage, &native_denom)
        .map_err(|_| ContractError::NativeDenomLinkDoesNotExist { native_denom })?;
    NATIVE_DENOM_TO_IBC_DENOM_MAP.save(deps.storage, &denom, &new_ibc_denom)?;
    Ok(Response::new().add_attribute("method", "change_native_denom_to_ibc_denom_link"))
}

// Remove an existing native_denom->ibc_denom map in the registry
pub fn remove_native_denom_to_ibc_denom_link(
    deps: DepsMut,
    native_denom: String,
) -> Result<Response, ContractError> {
    let denom = native_denom.clone(); // create a clone of native_denom
    NATIVE_DENOM_TO_IBC_DENOM_MAP
        .load(deps.storage, &native_denom)
        .map_err(|_| ContractError::NativeDenomLinkDoesNotExist { native_denom })?;
    NATIVE_DENOM_TO_IBC_DENOM_MAP.remove(deps.storage, &denom);
    Ok(Response::new().add_attribute("method", "remove_native_denom_to_ibc_denom_link"))
}

// Queries

pub fn query_denom_trace(deps: Deps, hash: String) -> Result<String, StdError> {
    let res = QueryDenomTraceRequest { hash }.query(&deps.querier)?;
    if let Some(denom_trace) = res.denom_trace {
        let base_denom = denom_trace.base_denom;
        Ok(base_denom)
    } else {
        Err(StdError::generic_err("No denom trace found"))
    }
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
            CONTRACT_ALIAS_MAP
                .load(&deps.storage, "swap_router")
                .unwrap(),
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
            CONTRACT_ALIAS_MAP
                .load(&deps.storage, "swap_router")
                .unwrap(),
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
            CONTRACT_ALIAS_MAP
                .load(&deps.storage, "new_swap_router")
                .unwrap(),
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
        assert!(!CONTRACT_ALIAS_MAP.has(&deps.storage, "swap_router"));
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
        let msg = ExecuteMsg::SetChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(
                    &deps.storage,
                    (&"osmosis".to_string(), &"cosmos".to_string())
                )
                .unwrap(),
            "channel-0"
        );
    }

    #[test]
    fn test_set_chain_channel_link_fail_existing_link() {
        let mut deps = mock_dependencies();

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set the canonical channel link between osmosis and cosmos to channel-150
        // This should fail because the link already exists
        let msg = ExecuteMsg::SetChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-150".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::ChainToChainChannelLinkAlreadyExists {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(
                    &deps.storage,
                    (&"osmosis".to_string(), &"cosmos".to_string())
                )
                .unwrap(),
            "channel-0"
        );
    }

    #[test]
    fn test_change_chain_channel_link_success() {
        let mut deps = mock_dependencies();

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change the canonical channel link between osmosis and cosmos to channel-150
        let msg = ExecuteMsg::ChangeChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            new_channel_id: "channel-150".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that the channel between osmosis and cosmos has changed from channel-0 to channel-150
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(
                    &deps.storage,
                    (&"osmosis".to_string(), &"cosmos".to_string())
                )
                .unwrap(),
            "channel-150"
        );
    }

    #[test]
    fn test_change_chain_channel_link_fail_non_existing_link() {
        let mut deps = mock_dependencies();

        // Attempt to change a channel link that does not exist
        let msg = ExecuteMsg::ChangeChainToChainChannelLink {
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

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::SetChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
            channel_id: "channel-0".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Remove the link
        let msg = ExecuteMsg::RemoveChainToChainChannelLink {
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Verify that the link no longer exists
        assert!(!CHAIN_TO_CHAIN_CHANNEL_MAP.has(
            &deps.storage,
            (&"osmosis".to_string(), &"cosmos".to_string())
        ));
    }

    #[test]
    fn test_remove_chain_channel_link_fail_nonexistent_link() {
        let mut deps = mock_dependencies();

        // Attempt to remove a link that does not exist
        let msg = ExecuteMsg::RemoveChainToChainChannelLink {
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
    fn test_set_channel_to_chain_link_success() {
        let mut deps = mock_dependencies();

        // Set channel-0 link from osmosis to cosmos
        let msg = ExecuteMsg::SetChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that channel-0 on osmosis is linked to cosmos
        assert_eq!(
            CHANNEL_TO_CHAIN_CHAIN_MAP
                .load(
                    &deps.storage,
                    (&"channel-0".to_string(), &"osmosis".to_string())
                )
                .unwrap(),
            "cosmos"
        );
    }

    #[test]
    fn test_set_channel_to_chain_link_fail_existing_link() {
        let mut deps = mock_dependencies();

        // Set channel-0 link from osmosis to cosmos
        let msg = ExecuteMsg::SetChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set channel-0 link from osmosis to regen
        let msg = ExecuteMsg::SetChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            destination_chain: "regen".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::ChannelToChainChainLinkAlreadyExists {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            CHANNEL_TO_CHAIN_CHAIN_MAP
                .load(
                    &deps.storage,
                    (&"channel-0".to_string(), &"osmosis".to_string())
                )
                .unwrap(),
            "cosmos"
        );
    }

    #[test]
    fn test_change_channel_to_chain_link_success() {
        let mut deps = mock_dependencies();

        // Set channel-0 link from osmosis to cosmos
        let msg = ExecuteMsg::SetChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change channel-0 link from osmosis to regen
        let msg = ExecuteMsg::ChangeChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            new_destination_chain: "regen".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that channel-0 on osmosis is linked to regen
        assert_eq!(
            CHANNEL_TO_CHAIN_CHAIN_MAP
                .load(
                    &deps.storage,
                    (&"channel-0".to_string(), &"osmosis".to_string())
                )
                .unwrap(),
            "regen"
        );
    }

    #[test]
    fn test_change_channel_to_chain_link_fail_nonexistent_link() {
        let mut deps = mock_dependencies();

        // Attempt to change a link that does not exist
        let msg = ExecuteMsg::ChangeChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            new_destination_chain: "regen".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::ChannelToChainChainLinkDoesNotExist {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_channel_to_chain_link_success() {
        let mut deps = mock_dependencies();

        // Set channel-0 link from osmosis to cosmos
        let msg = ExecuteMsg::SetChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
            destination_chain: "cosmos".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Remove channel-0 link from osmosis to cosmos
        let msg = ExecuteMsg::RemoveChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that the link no longer exists
        assert!(!CHANNEL_TO_CHAIN_CHAIN_MAP.has(
            &deps.storage,
            (&"channel-0".to_string(), &"osmosis".to_string())
        ));
    }

    #[test]
    fn test_remove_channel_to_chain_link_fail_nonexistent_link() {
        let mut deps = mock_dependencies();

        // Attempt to remove a link that does not exist
        let msg = ExecuteMsg::RemoveChannelToChainChainLink {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::ChannelToChainChainLinkDoesNotExist {
            channel_id: "channel-0".to_string(),
            source_chain: "osmosis".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_set_native_denom_to_ibc_denom_success() {
        let mut deps = mock_dependencies();

        // Set uatom to ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F
        let msg = ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            ibc_denom: "ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that uatom is linked to ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F
        assert_eq!(
            NATIVE_DENOM_TO_IBC_DENOM_MAP
                .load(&deps.storage, "uatom")
                .unwrap(),
            "ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F"
        );
    }

    #[test]
    fn test_set_native_denom_to_ibc_denom_fail_duplicate_denom() {
        let mut deps = mock_dependencies();

        // Set uatom to ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F
        let msg = ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            ibc_denom: "ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Attempt to set uatom to ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
        let msg = ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            ibc_denom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::NativeDenomLinkAlreadyExists {
            native_denom: "uatom".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_change_native_denom_to_ibc_denom_success() {
        let mut deps = mock_dependencies();

        // Set uatom to ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F
        let msg = ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            ibc_denom: "ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Change uatom to ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
        let msg = ExecuteMsg::ChangeNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            new_ibc_denom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that uatom is now linked to ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
        assert_eq!(
            NATIVE_DENOM_TO_IBC_DENOM_MAP
                .load(&deps.storage, "uatom")
                .unwrap(),
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
        );
    }

    #[test]
    fn test_change_native_denom_to_ibc_denom_fail_no_link() {
        let mut deps = mock_dependencies();

        // Attempt to change uatom to ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
        let msg = ExecuteMsg::ChangeNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            new_ibc_denom: "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::NativeDenomLinkDoesNotExist {
            native_denom: "uatom".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_native_denom_to_ibc_denom_success() {
        let mut deps = mock_dependencies();

        // Set uatom to ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F
        let msg = ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
            ibc_denom: "ibc/0A6F4F0F9EE437A9F8E9B53E79D7D197B3B8A606B3C3C9F7B99D8B2A476A3C1F"
                .to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Remove uatom
        let msg = ExecuteMsg::RemoveNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);
        assert!(result.is_ok());

        // Verify that uatom is no longer linked to anything
        assert_eq!(
            NATIVE_DENOM_TO_IBC_DENOM_MAP
                .may_load(&deps.storage, "uatom")
                .unwrap(),
            None
        );
    }

    #[test]
    fn test_remove_native_denom_to_ibc_denom_fail_no_link() {
        let mut deps = mock_dependencies();

        // Attempt to remove uatom
        let msg = ExecuteMsg::RemoveNativeDenomToIbcDenom {
            native_denom: "uatom".to_string(),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::NativeDenomLinkDoesNotExist {
            native_denom: "uatom".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
    }
}
