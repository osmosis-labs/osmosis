use cosmwasm_std::{Addr, Deps};
use osmosis_std_derive::CosmwasmExt;
use sha2::{Digest, Sha256};

use crate::execute::{FullOperation, Permission};
use crate::state::{CHAIN_ADMIN_MAP, CHAIN_MAINTAINER_MAP, CONFIG, GLOBAL_ADMIN_MAP};
use crate::ContractError;

// check_is_contract_governor is used for functions that can only be called by the contract governor
pub fn check_is_contract_governor(deps: Deps, sender: Addr) -> Result<(), ContractError> {
    let config = CONFIG.load(deps.storage).unwrap();
    if config.owner != sender {
        Err(ContractError::Unauthorized {})
    } else {
        Ok(())
    }
}

// check_permission checks if an account with the provided permission is authorized to perform the action requested
// GlobalAdmin can add addresses to any address map, ChainAdmin can only add addresses to their
// own chain map, and ChainMaintainer cant add addresses to any address map
pub fn check_permission(
    provided_permission: Permission,
    max_permission: Permission,
) -> Result<(), ContractError> {
    if max_permission == Permission::GlobalAdmin {
        return Ok(());
    }
    if max_permission == Permission::ChainAdmin
        && (provided_permission == Permission::ChainAdmin
            || provided_permission == Permission::ChainMaintainer)
    {
        return Ok(());
    }
    if max_permission == Permission::ChainMaintainer
        && provided_permission == Permission::ChainMaintainer
    {
        return Ok(());
    }
    Err(ContractError::Unauthorized {})
}

pub fn check_action_permission(
    provided_action: FullOperation,
    max_permission: Permission,
) -> Result<(), ContractError> {
    if max_permission == Permission::GlobalAdmin || max_permission == Permission::ChainAdmin {
        return Ok(());
    }
    if max_permission == Permission::ChainMaintainer
        && (provided_action == FullOperation::Set
            || provided_action == FullOperation::Disable
            || provided_action == FullOperation::Enable)
    {
        return Ok(());
    }
    Err(ContractError::Unauthorized {})
}

pub fn check_is_authorized(
    deps: Deps,
    sender: Addr,
    source_chain: Option<String>,
) -> Result<Permission, ContractError> {
    if check_is_global_admin(deps, sender.clone()).is_ok() {
        return Ok(Permission::GlobalAdmin);
    }
    if check_is_chain_admin(deps, sender.clone(), source_chain.clone()).is_ok() {
        return Ok(Permission::ChainAdmin);
    }
    check_is_chain_maintainer(deps, sender, source_chain)?;
    Ok(Permission::ChainMaintainer)
}

// check_is_global_admin is used for functions that can only be called by the contract governor
pub fn check_is_global_admin(deps: Deps, sender: Addr) -> Result<(), ContractError> {
    let config = CONFIG.load(deps.storage).unwrap();
    // If the sender is the contract governor, they are authorized to make changes
    if config.owner == sender {
        return Ok(());
    }

    // If the sender an authorized address, they are authorized to make changes
    let authorized_addr = GLOBAL_ADMIN_MAP
        .may_load(deps.storage, "osmosis")
        .unwrap_or_default();
    if authorized_addr.eq(&Some(sender)) {
        return Ok(());
    }

    Err(ContractError::Unauthorized {})
}

// check_is_chain_admin_address checks if the sender is the contract governor or if the sender is
// authorized to make changes to the provided source chain
pub fn check_is_chain_admin(
    deps: Deps,
    sender: Addr,
    source_chain: Option<String>,
) -> Result<(), ContractError> {
    // If the sender is the authorized address for the source chain, they are authorized to make changes
    if let Some(source_chain) = source_chain {
        let authorized_addr = CHAIN_ADMIN_MAP
            .may_load(deps.storage, &source_chain.to_lowercase())
            .unwrap_or_default();
        if authorized_addr.eq(&Some(sender)) {
            return Ok(());
        }
    }
    Err(ContractError::Unauthorized {})
}

// check_is_chain_maintainer_address checks if the sender is the contract governor or if the sender is
// authorized to make changes to the provided source chain
pub fn check_is_chain_maintainer(
    deps: Deps,
    sender: Addr,
    source_chain: Option<String>,
) -> Result<(), ContractError> {
    // If the sender is the authorized address for the source chain, they are authorized to make changes
    if let Some(source_chain) = source_chain {
        let authorized_addr = CHAIN_MAINTAINER_MAP
            .may_load(deps.storage, &source_chain.to_lowercase())
            .unwrap_or_default();
        if authorized_addr.eq(&Some(sender)) {
            return Ok(());
        }
    }
    Err(ContractError::Unauthorized {})
}

#[cfg(test)]
pub mod test {
    use crate::execute;
    use crate::execute::AuthorizedAddressInput;
    use crate::msg;
    use crate::ContractError;
    use crate::{contract, msg::InstantiateMsg};
    use cosmwasm_std::testing::{
        mock_dependencies, mock_env, mock_info, MockApi, MockQuerier, MockStorage,
    };
    use cosmwasm_std::{Addr, DepsMut, OwnedDeps};

    static CREATOR_ADDRESS: &str = "creator";
    static CHAIN_ADMIN: &str = "chain_admin";
    static CHAIN_MAINTAINER: &str = "chain_maintainer";

    pub fn initialize_contract(mut deps: DepsMut) -> Addr {
        let msg = InstantiateMsg {
            owner: String::from(CREATOR_ADDRESS),
        };
        let creator_info = mock_info(CREATOR_ADDRESS, &[]);

        contract::instantiate(deps.branch(), mock_env(), creator_info.clone(), msg).unwrap();

        // Set the CHAIN_ADMIN address as the osmosis and mars chain admin
        let msg = msg::ExecuteMsg::ModifyAuthorizedAddresses {
            operations: vec![
                AuthorizedAddressInput {
                    operation: execute::Operation::Set,
                    source_chain: "osmosis".to_string(),
                    permission: Some(execute::Permission::ChainAdmin),
                    addr: Addr::unchecked(CHAIN_ADMIN.to_string()),
                    new_addr: None,
                },
                AuthorizedAddressInput {
                    operation: execute::Operation::Set,
                    source_chain: "mars".to_string(),
                    permission: Some(execute::Permission::ChainAdmin),
                    addr: Addr::unchecked(CHAIN_ADMIN.to_string()),
                    new_addr: None,
                },
            ],
        };
        contract::execute(deps.branch(), mock_env(), creator_info.clone(), msg).unwrap();

        // Set the CHAIN_MAINTAINER address as the osmosis and mars chain maintainer with the chain admin
        let msg = msg::ExecuteMsg::ModifyAuthorizedAddresses {
            operations: vec![
                AuthorizedAddressInput {
                    operation: execute::Operation::Set,
                    source_chain: "osmosis".to_string(),
                    permission: Some(execute::Permission::ChainMaintainer),
                    addr: Addr::unchecked(CHAIN_MAINTAINER.to_string()),
                    new_addr: None,
                },
                AuthorizedAddressInput {
                    operation: execute::Operation::Set,
                    source_chain: "mars".to_string(),
                    permission: Some(execute::Permission::ChainMaintainer),
                    addr: Addr::unchecked(CHAIN_MAINTAINER.to_string()),
                    new_addr: None,
                },
            ],
        };
        let chain_admin_info = mock_info(CHAIN_ADMIN, &[]);
        contract::execute(deps.branch(), mock_env(), chain_admin_info.clone(), msg).unwrap();

        // Set the CHAIN_ADMIN address as the juno chain maintainer
        // This is used to ensure that permissions don't bleed over from one chain to another
        let msg = msg::ExecuteMsg::ModifyAuthorizedAddresses {
            operations: vec![AuthorizedAddressInput {
                operation: execute::Operation::Set,
                source_chain: "juno".to_string(),
                permission: Some(execute::Permission::ChainMaintainer),
                addr: Addr::unchecked(CHAIN_ADMIN.to_string()),
                new_addr: None,
            }],
        };
        contract::execute(deps, mock_env(), creator_info.clone(), msg).unwrap();

        creator_info.sender
    }

    pub fn setup() -> Result<OwnedDeps<MockStorage, MockApi, MockQuerier>, ContractError> {
        let mut deps = mock_dependencies();
        let governor = initialize_contract(deps.as_mut());
        let info = mock_info(governor.as_str(), &[]);

        // Set up the contract aliases
        let operation = vec![
            execute::ContractAliasInput {
                operation: execute::Operation::Set,
                alias: "contract_one".to_string(),
                address: Some("osmo1dfaselkjh32hnkljw3nlklk2lknmes".to_string()),
                new_alias: None,
            },
            execute::ContractAliasInput {
                operation: execute::Operation::Set,
                alias: "contract_two".to_string(),
                address: Some("osmo1dfg4k3jhlknlfkjdslkjkl43klnfdl".to_string()),
                new_alias: None,
            },
            execute::ContractAliasInput {
                operation: execute::Operation::Set,
                alias: "contract_three".to_string(),
                address: Some("osmo1dfgjlk4lkfklkld32fsdajknjrrgfg".to_string()),
                new_alias: None,
            },
        ];

        execute::contract_alias_operations(deps.as_mut(), info.sender.clone(), operation)?;

        // Set up the chain channels
        let operations = vec![
            execute::ConnectionInput {
                operation: execute::FullOperation::Set,
                source_chain: "osmosis".to_string(),
                destination_chain: "juno".to_string(),
                channel_id: Some("channel-42".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            },
            execute::ConnectionInput {
                operation: execute::FullOperation::Set,
                source_chain: "osmosis".to_string(),
                destination_chain: "stargaze".to_string(),
                channel_id: Some("channel-75".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            },
            execute::ConnectionInput {
                operation: execute::FullOperation::Set,
                source_chain: "stargaze".to_string(),
                destination_chain: "osmosis".to_string(),
                channel_id: Some("channel-0".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            },
        ];
        execute::connection_operations(deps.as_mut(), info.sender.clone(), operations)?;

        Ok(deps)
    }
}

// takes a transfer message and returns ibc/<hash of denom>
pub fn hash_denom_trace(unwrapped: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(unwrapped.as_bytes());
    let result = hasher.finalize();
    let hash = hex::encode(result);
    format!("ibc/{}", hash.to_uppercase())
}

// DenomTrace query message definition.
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/ibc.applications.transfer.v1.QueryDenomTraceRequest")]
#[proto_query(
    path = "/ibc.applications.transfer.v1.Query/DenomTrace",
    response_type = QueryDenomTraceResponse
)]
pub struct QueryDenomTraceRequest {
    #[prost(string, tag = "1")]
    pub hash: ::prost::alloc::string::String,
}

#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/ibc.applications.transfer.v1.QueryDenomTraceResponse")]
pub struct QueryDenomTraceResponse {
    #[prost(message, optional, tag = "1")]
    pub denom_trace: Option<DenomTrace>,
}

#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
)]
pub struct DenomTrace {
    #[prost(string, tag = "1")]
    pub path: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub base_denom: ::prost::alloc::string::String,
}
