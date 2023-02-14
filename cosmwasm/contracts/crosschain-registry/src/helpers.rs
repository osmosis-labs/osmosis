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
    execute::set_chain_channel_link(
        deps.as_mut(),
        "osmo".to_string(),
        "juno".to_string(),
        "channel-42".to_string(),
    )?;
    execute::set_chain_channel_link(
        deps.as_mut(),
        "osmo".to_string(),
        "stars".to_string(),
        "channel-75".to_string(),
    )?;
    execute::set_chain_channel_link(
        deps.as_mut(),
        "stars".to_string(),
        "osmo".to_string(),
        "channel-0".to_string(),
    )?;

    // Set up the asset mappings
    execute::set_asset_map(
        deps.as_mut(),
        "uosmo".to_string(),
        "osmo".to_string(),
        "uosmo".to_string(),
    )?;
    execute::set_asset_map(
        deps.as_mut(),
        "uatom".to_string(),
        "osmo".to_string(),
        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2".to_string(),
    )?;
    execute::set_asset_map(
        deps.as_mut(),
        "ustars".to_string(),
        "osmo".to_string(),
        "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4".to_string(),
    )?;

    Ok(deps)
}

pub fn make_chain_channel_key(source_chain: &str, destination_chain: &str) -> String {
    format!("{}|{}", source_chain, destination_chain)
}

pub fn make_asset_key(native_denom: &str, destination_chain: &str) -> String {
    format!("{}|{}", native_denom, destination_chain)
}
