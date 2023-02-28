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
        execute::set_contract_alias(
            deps.as_mut(),
            "contract_one".to_string(),
            "osmo1dfaselkjh32hnkljw3nlklk2lknmes".to_string(),
        )?;
        execute::set_contract_alias(
            deps.as_mut(),
            "contract_two".to_string(),
            "osmo1dfg4k3jhlknlfkjdslkjkl43klnfdl".to_string(),
        )?;
        execute::set_contract_alias(
            deps.as_mut(),
            "contract_three".to_string(),
            "osmo1dfgjlk4lkfklkld32fsdajknjrrgfg".to_string(),
        )?;

        // Set up the chain channels
        execute::connection_operation(
            deps.as_mut(),
            execute::ConnectionOperation::Set,
            "osmo".to_string(),
            "juno".to_string(),
            Some("channel-42".to_string()),
            None,
            None,
        )?;

        execute::connection_operation(
            deps.as_mut(),
            execute::ConnectionOperation::Set,
            "osmo".to_string(),
            "stars".to_string(),
            Some("channel-75".to_string()),
            None,
            None,
        )?;

        execute::connection_operation(
            deps.as_mut(),
            execute::ConnectionOperation::Set,
            "stars".to_string(),
            "osmo".to_string(),
            Some("channel-0".to_string()),
            None,
            None,
        )?;
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
