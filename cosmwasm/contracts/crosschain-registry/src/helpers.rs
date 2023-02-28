use osmosis_std_derive::CosmwasmExt;
use sha2::{Digest, Sha256};

#[cfg(test)]
pub mod test {
    use crate::execute;
    use crate::ContractError;
    use cosmwasm_std::testing::{mock_dependencies, MockApi, MockQuerier, MockStorage};
    use cosmwasm_std::OwnedDeps;

    pub fn setup() -> Result<OwnedDeps<MockStorage, MockApi, MockQuerier>, ContractError> {
        let mut deps = mock_dependencies();

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

        execute::contract_alias_operations(deps.as_mut(), operation)?;

        // Set up the chain channels
        let operations = vec![
            execute::ConnectionInput {
                operation: execute::Operation::Set,
                source_chain: "osmo".to_string(),
                destination_chain: "juno".to_string(),
                channel_id: Some("channel-42".to_string()),
                new_channel_id: None,
                new_destination_chain: None,
            },
            execute::ConnectionInput {
                operation: execute::Operation::Set,
                source_chain: "osmo".to_string(),
                destination_chain: "stars".to_string(),
                channel_id: Some("channel-75".to_string()),
                new_channel_id: None,
                new_destination_chain: None,
            },
            execute::ConnectionInput {
                operation: execute::Operation::Set,
                source_chain: "stars".to_string(),
                destination_chain: "osmo".to_string(),
                channel_id: Some("channel-0".to_string()),
                new_channel_id: None,
                new_destination_chain: None,
            },
        ];
        execute::connection_operations(deps.as_mut(), operations)?;

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
