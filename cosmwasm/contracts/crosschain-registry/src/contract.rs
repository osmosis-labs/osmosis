use crate::helpers::{make_asset_key, make_chain_channel_key};
#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::execute;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{ASSET_MAP, CHAIN_CHANNEL_MAP, CONTRACT_MAP};

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
        // Contract aliases
        ExecuteMsg::SetContractAlias {
            contract_alias,
            contract_address,
        } => execute::set_contract_alias(deps, contract_alias, contract_address),
        ExecuteMsg::ChangeContractAlias {
            current_contract_alias,
            new_contract_alias,
        } => execute::change_contract_alias(deps, current_contract_alias, new_contract_alias),
        ExecuteMsg::RemoveContractAlias { contract_alias } => {
            execute::remove_contract_alias(deps, &contract_alias)
        }

        // Chain channel links
        ExecuteMsg::SetChainChannelLink {
            source_chain,
            destination_chain,
            channel_id,
        } => execute::set_chain_channel_link(deps, source_chain, destination_chain, channel_id),
        ExecuteMsg::ChangeChainChannelLink {
            source_chain,
            destination_chain,
            new_channel_id,
        } => execute::change_chain_channel_link(
            deps,
            source_chain,
            destination_chain,
            new_channel_id,
        ),
        ExecuteMsg::RemoveChainChannelLink {
            source_chain,
            destination_chain,
        } => execute::remove_chain_channel_link(deps, source_chain, destination_chain),

        // Asset mappings
        ExecuteMsg::SetAssetMapping {
            native_denom,
            destination_chain,
            destination_chain_denom,
        } => execute::set_asset_map(
            deps,
            native_denom,
            destination_chain,
            destination_chain_denom,
        ),
        ExecuteMsg::ChangeAssetMapping {
            native_denom,
            destination_chain,
            new_destination_chain_denom,
        } => execute::change_asset_map(
            deps,
            native_denom,
            destination_chain,
            new_destination_chain_denom,
        ),
        ExecuteMsg::RemoveAssetMapping {
            native_denom,
            destination_chain,
        } => execute::remove_asset_map(deps, native_denom, destination_chain),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetAddressFromAlias { contract_alias } => {
            to_binary(&CONTRACT_MAP.load(deps.storage, &contract_alias)?)
        }
        QueryMsg::GetChainChannelLink {
            source_chain,
            destination_chain,
        } => to_binary(&CHAIN_CHANNEL_MAP.load(
            deps.storage,
            &make_chain_channel_key(&source_chain, &destination_chain),
        )?),
        QueryMsg::GetAssetMapping {
            native_denom,
            destination_chain,
        } => to_binary(&ASSET_MAP.load(
            deps.storage,
            &make_asset_key(&native_denom, &destination_chain),
        )?),
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::helpers::setup;

    use cosmwasm_std::from_binary;
    use cosmwasm_std::testing::mock_env;

    #[test]
    fn setup_and_query_aliases() {
        // Store three alias<>address mappings
        let deps = setup().unwrap();

        // Retrieve alias one and check the contract address is what we expect
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "contract_one".to_string(),
            },
        )
        .unwrap();
        let address: String = from_binary(&address_binary).unwrap();
        assert_eq!("osmo1dfaselkjh32hnkljw3nlklk2lknmes", address);

        // Retrieve alias two and check the contract address is what we expect
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "contract_two".to_string(),
            },
        )
        .unwrap();
        let address: String = from_binary(&address_binary).unwrap();
        assert_eq!("osmo1dfg4k3jhlknlfkjdslkjkl43klnfdl", address);

        // Retrieve alias three and check the contract address is what we expect
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "contract_three".to_string(),
            },
        )
        .unwrap();
        let address: String = from_binary(&address_binary).unwrap();
        assert_eq!("osmo1dfgjlk4lkfklkld32fsdajknjrrgfg", address);

        // Attempt to retrieve an alias that doesn't exist and check that we get an error
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "invalid_contract_alias".to_string(),
            },
        );
        assert!(address_binary.is_err());
    }

    #[test]
    fn setup_and_query_channels() {
        // Store three chain<>channel mappings
        let deps = setup().unwrap();

        // Retrieve osmo<>juno link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChainChannelLink {
                source_chain: "osmo".to_string(),
                destination_chain: "juno".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-42", channel);

        // Retrieve osmo<>stars link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChainChannelLink {
                source_chain: "osmo".to_string(),
                destination_chain: "stars".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-75", channel);

        // Retrieve osmo<>juno link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChainChannelLink {
                source_chain: "stars".to_string(),
                destination_chain: "osmo".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-0", channel);

        // Attempt to retrieve a link that doesn't exist and check that we get an error
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChainChannelLink {
                source_chain: "osmo".to_string(),
                destination_chain: "cerberus".to_string(),
            },
        );
        assert!(channel_binary.is_err());
    }

    #[test]
    fn setup_and_query_denoms() {
        // Store three denom mappings
        let deps = setup().unwrap();

        // Retrieve uosmo on osmosis mapping and check the denom is what we expect
        let denom_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAssetMapping {
                native_denom: "uosmo".to_string(),
                destination_chain: "osmo".to_string(),
            },
        )
        .unwrap();
        let denom: String = from_binary(&denom_binary).unwrap();
        assert_eq!("uosmo", denom);

        // Retrieve uatom on osmosis mapping and check the denom is what we expect
        let denom_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAssetMapping {
                native_denom: "uatom".to_string(),
                destination_chain: "osmo".to_string(),
            },
        )
        .unwrap();
        let denom: String = from_binary(&denom_binary).unwrap();
        assert_eq!(
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
            denom
        );

        // Retrieve ustars on osmosis mapping and check the denom is what we expect
        let denom_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAssetMapping {
                native_denom: "ustars".to_string(),
                destination_chain: "osmo".to_string(),
            },
        )
        .unwrap();
        let denom: String = from_binary(&denom_binary).unwrap();
        assert_eq!(
            "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4",
            denom
        );

        // Attempt to retrieve a denom that doesn't exist on osmosis and check that we get an error
        let denom_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAssetMapping {
                native_denom: "uczar".to_string(),
                destination_chain: "osmo".to_string(),
            },
        );
        assert!(denom_binary.is_err());
    }
}
