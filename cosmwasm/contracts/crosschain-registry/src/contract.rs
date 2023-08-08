#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, GetAddressFromAliasResponse, InstantiateMsg, QueryMsg};
use crate::state::{Config, CONFIG, CONTRACT_ALIAS_MAP};
use crate::{execute, query};
use registry::Registry;

// version info for migration
const CONTRACT_NAME: &str = "crates.io:crosschain-registry";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    // validate owner address and save to state
    let owner = deps.api.addr_validate(&msg.owner)?;
    let state = Config { owner };
    CONFIG.save(deps.storage, &state)?;

    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        // Contract aliases
        ExecuteMsg::ModifyContractAlias { operations } => {
            execute::contract_alias_operations(deps, info.sender, operations)
        }

        // Chain channel links
        ExecuteMsg::ModifyChainChannelLinks { operations } => {
            execute::connection_operations(deps, info.sender, operations)
        }

        // Bech32 prefixes
        ExecuteMsg::ModifyBech32Prefixes { operations } => {
            execute::chain_to_prefix_operations(deps, info.sender, operations)
        }

        // Authorized addresses
        ExecuteMsg::ModifyAuthorizedAddresses { operations } => {
            execute::authorized_address_operations(deps, info.sender, operations)
        }

        ExecuteMsg::UnwrapCoin {
            receiver,
            into_chain,
            with_memo,
        } => {
            let registries = Registry::new(deps.as_ref(), env.contract.address.to_string())?;
            let coin = cw_utils::one_coin(&info)?;
            let transfer_msg = registries.unwrap_coin_into(
                coin,
                receiver,
                into_chain.as_deref(),
                env.contract.address.to_string(),
                env.block.time,
                with_memo,
                None,
            )?;
            deps.api.debug(&format!("transfer_msg: {transfer_msg:?}"));
            Ok(Response::new()
                .add_message(transfer_msg)
                .add_attribute("method", "unwrap_coin"))
        }
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    deps.api.debug(&format!("executing query: {msg:?}"));
    match msg {
        QueryMsg::GetAddressFromAlias { contract_alias } => {
            let address = CONTRACT_ALIAS_MAP.load(deps.storage, &contract_alias)?;
            let response = GetAddressFromAliasResponse { address };
            to_binary(&response)
        }

        QueryMsg::GetDestinationChainFromSourceChainViaChannel {
            on_chain,
            via_channel,
        } => to_binary(&query::query_chain_from_channel_chain_pair(
            deps,
            on_chain,
            via_channel,
        )?),

        QueryMsg::GetChannelFromChainPair {
            source_chain,
            destination_chain,
        } => to_binary(&query::query_channel_from_chain_pair(
            deps,
            source_chain,
            destination_chain,
        )?),

        QueryMsg::GetBech32PrefixFromChainName { chain_name } => to_binary(
            &query::query_bech32_prefix_from_chain_name(deps, chain_name)?,
        ),

        QueryMsg::GetDenomTrace { ibc_denom } => {
            to_binary(&query::query_denom_trace_from_ibc_denom(deps, ibc_denom)?)
        }
        QueryMsg::GetChainNameFromBech32Prefix { prefix } => {
            to_binary(&query::query_chain_name_from_bech32_prefix(deps, prefix)?)
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::execute::ConnectionInput;
    use crate::helpers::test::setup;

    use cosmwasm_std::from_binary;
    use cosmwasm_std::testing::{mock_env, mock_info};

    static CREATOR_ADDRESS: &str = "creator";

    #[test]
    fn query_aliases() {
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
        let address: GetAddressFromAliasResponse = from_binary(&address_binary).unwrap();
        assert_eq!(
            GetAddressFromAliasResponse {
                address: "osmo1dfaselkjh32hnkljw3nlklk2lknmes".to_string(),
            },
            address
        );

        // Retrieve alias two and check the contract address is what we expect
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "contract_two".to_string(),
            },
        )
        .unwrap();
        let address: GetAddressFromAliasResponse = from_binary(&address_binary).unwrap();
        assert_eq!(
            GetAddressFromAliasResponse {
                address: "osmo1dfg4k3jhlknlfkjdslkjkl43klnfdl".to_string(),
            },
            address
        );

        // Retrieve alias three and check the contract address is what we expect
        let address_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetAddressFromAlias {
                contract_alias: "contract_three".to_string(),
            },
        )
        .unwrap();
        let address: GetAddressFromAliasResponse = from_binary(&address_binary).unwrap();
        assert_eq!(
            GetAddressFromAliasResponse {
                address: "osmo1dfgjlk4lkfklkld32fsdajknjrrgfg".to_string(),
            },
            address
        );

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
    fn query_chain_and_channel() {
        // Store three chain<>channel mappings
        let mut deps = setup().unwrap();

        // Retrieve osmo<>juno link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "osmosis".to_string(),
                destination_chain: "juno".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-42", channel);

        // Check that osmosis' channel-42 is connected to juno
        let destination_chain = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetDestinationChainFromSourceChainViaChannel {
                on_chain: "osmosis".to_string(),
                via_channel: "channel-42".to_string(),
            },
        )
        .unwrap();
        let destination_chain: String = from_binary(&destination_chain).unwrap();
        assert_eq!("juno", destination_chain);

        // Retrieve osmo<>stars link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "osmosis".to_string(),
                destination_chain: "stargaze".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-75", channel);

        // Check that osmosis' channel-75 is connected to stars
        let destination_chain = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetDestinationChainFromSourceChainViaChannel {
                on_chain: "osmosis".to_string(),
                via_channel: "channel-75".to_string(),
            },
        )
        .unwrap();
        let destination_chain: String = from_binary(&destination_chain).unwrap();
        assert_eq!("stargaze", destination_chain);

        // Retrieve stargaze<>osmosis link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "stargaze".to_string(),
                destination_chain: "osmosis".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-0", channel);

        // Check that stars' channel-0 is connected to osmo
        let destination_chain = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetDestinationChainFromSourceChainViaChannel {
                on_chain: "stargaze".to_string(),
                via_channel: "channel-0".to_string(),
            },
        )
        .unwrap();
        let destination_chain: String = from_binary(&destination_chain).unwrap();
        assert_eq!("osmosis", destination_chain);

        // Attempt to retrieve a link that doesn't exist and check that we get an error
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "osmosis".to_string(),
                destination_chain: "cerberus".to_string(),
            },
        );
        assert!(channel_binary.is_err());

        // Disable the osmo<>juno link with the global admin
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: execute::FullOperation::Disable,
                source_chain: "OSMOSIS".to_string(),
                destination_chain: "JUNO".to_string(),
                channel_id: Some("CHANNEL-42".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info_creator = mock_info(CREATOR_ADDRESS, &[]);
        let result = execute(deps.as_mut(), mock_env(), info_creator.clone(), msg);
        assert!(result.is_ok());

        // Retrieve osmo<>juno link again, but this time it should be disabled
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "osmosis".to_string(),
                destination_chain: "juno".to_string(),
            },
        );
        assert!(res.is_err());

        // Enable the osmo<>juno link with the global admin
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: execute::FullOperation::Enable,
                source_chain: "OSMOSIS".to_string(),
                destination_chain: "JUNO".to_string(),
                channel_id: Some("CHANNEL-42".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let result = execute(deps.as_mut(), mock_env(), info_creator, msg);
        assert!(result.is_ok());

        // Retrieve osmo<>juno link again, but this time it should be enabled
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChannelFromChainPair {
                source_chain: "osmosis".to_string(),
                destination_chain: "juno".to_string(),
            },
        )
        .unwrap();
        let channel: String = from_binary(&channel_binary).unwrap();
        assert_eq!("channel-42", channel);
    }
}
