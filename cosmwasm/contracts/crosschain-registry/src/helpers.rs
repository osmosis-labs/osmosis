use cosmwasm_std::{Addr, Deps};
use osmosis_std_derive::CosmwasmExt;
use sha2::{Digest, Sha256};

use crate::state::{AUTHORIZED_ADDRESSES, CONFIG};
use crate::ContractError;

pub fn check_is_contract_governor(deps: Deps, sender: Addr) -> Result<(), ContractError> {
    let config = CONFIG.load(deps.storage).unwrap();
    if config.owner != sender {
        Err(ContractError::Unauthorized {})
    } else {
        Ok(())
    }
}

// check_is_authorized_address checks if the sender is the contract governor or if the sender is
// authorized to make changes to the provided source chain
pub fn check_is_authorized_address(
    deps: Deps,
    sender: Addr,
    source_chain: Option<String>,
) -> Result<(), ContractError> {
    let config = CONFIG.load(deps.storage).unwrap();
    if config.owner == sender {
        return Ok(());
    }
    if let Some(source_chain) = source_chain {
        let authorized_addr = AUTHORIZED_ADDRESSES
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
    use crate::ContractError;
    use crate::{contract, msg::InstantiateMsg};
    use cosmwasm_std::testing::{
        mock_dependencies, mock_env, mock_info, MockApi, MockQuerier, MockStorage,
    };
    use cosmwasm_std::{Addr, DepsMut, OwnedDeps};

    static CREATOR_ADDRESS: &str = "creator";

    pub fn initialize_contract(deps: DepsMut) -> Addr {
        let msg = InstantiateMsg {
            owner: String::from(CREATOR_ADDRESS),
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);

        contract::instantiate(deps, mock_env(), info.clone(), msg).unwrap();

        info.sender
    }

    pub fn setup() -> Result<OwnedDeps<MockStorage, MockApi, MockQuerier>, ContractError> {
        let mut deps = mock_dependencies();
        let governor = initialize_contract(deps.as_mut());
        //let governor_info = mock_info(governor.as_str(), &vec![] as &Vec<Coin>);
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
                operation: execute::Operation::Set,
                source_chain: "osmosis".to_string(),
                destination_chain: "juno".to_string(),
                channel_id: Some("channel-42".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            },
            execute::ConnectionInput {
                operation: execute::Operation::Set,
                source_chain: "osmosis".to_string(),
                destination_chain: "stargaze".to_string(),
                channel_id: Some("channel-75".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            },
            execute::ConnectionInput {
                operation: execute::Operation::Set,
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
