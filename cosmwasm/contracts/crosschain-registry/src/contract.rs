#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{
    ExecuteMsg, GetAddressFromAliasResponse, GetChannelFromChainPairResponse,
    GetDestinationChainFromSourceChainViaChannelResponse, InstantiateMsg, QueryMsg,
};
use crate::state::{
    Config, CHAIN_TO_CHAIN_CHANNEL_MAP, CHANNEL_ON_CHAIN_CHAIN_MAP, CONFIG, CONTRACT_ALIAS_MAP,
};
use crate::{execute, Registries};

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
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        // Contract aliases
        ExecuteMsg::ModifyContractAlias { operations } => {
            execute::contract_alias_operations(deps, operations)
        }

        // Chain channel links
        ExecuteMsg::ModifyChainChannelLinks { operations } => {
            execute::connection_operations(deps, operations)
        }
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
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
        } => {
            let destination_chain =
                CHANNEL_ON_CHAIN_CHAIN_MAP.load(deps.storage, (&via_channel, &on_chain))?;
            let response =
                GetDestinationChainFromSourceChainViaChannelResponse { destination_chain };
            to_binary(&response)
        }

        QueryMsg::GetChannelFromChainPair {
            source_chain,
            destination_chain,
        } => {
            let channel_id = CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(deps.storage, (&source_chain, &destination_chain))?;
            let response = GetChannelFromChainPairResponse { channel_id };
            to_binary(&response)
        }

        QueryMsg::GetDenomTrace { ibc_denom } => {
            to_binary(&execute::query_denom_trace_from_ibc_denom(deps, ibc_denom)?)
        }

        QueryMsg::UnwrapDenom { ibc_denom } => {
            let registries = Registries::new(deps, env.contract.address.to_string())?;
            to_binary(&registries.unwrap_denom(&ibc_denom)?)
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::helpers::test::setup;

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
    fn setup_and_query_chain_and_channel() {
        // Store three chain<>channel mappings
        let deps = setup().unwrap();

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

        // Retrieve osmo<>juno link and check the channel is what we expect
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
    }
}
