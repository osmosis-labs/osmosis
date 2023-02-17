#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::execute;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{State, CHAIN_TO_CHAIN_CHANNEL_MAP, CONTRACT_ALIAS_MAP, STATE};

// version info for migration info
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
    let state = State { owner };
    STATE.save(deps.storage, &state)?;

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

        // Chain to chain channel links
        ExecuteMsg::SetChainToChainChannelLink {
            source_chain,
            destination_chain,
            channel_id,
        } => execute::set_chain_to_chain_channel_link(
            deps,
            source_chain,
            destination_chain,
            channel_id,
        ),
        ExecuteMsg::ChangeChainToChainChannelLink {
            source_chain,
            destination_chain,
            new_channel_id,
        } => execute::change_chain_to_chain_channel_link(
            deps,
            source_chain,
            destination_chain,
            new_channel_id,
        ),
        ExecuteMsg::RemoveChainToChainChannelLink {
            source_chain,
            destination_chain,
        } => execute::remove_chain_to_chain_channel_link(deps, source_chain, destination_chain),

        // Channel to chain chain links
        ExecuteMsg::SetChannelToChainChainLink {
            channel_id,
            source_chain,
            destination_chain,
        } => execute::set_channel_to_chain_chain_link(
            deps,
            channel_id,
            source_chain,
            destination_chain,
        ),
        ExecuteMsg::ChangeChannelToChainChainLink {
            channel_id,
            source_chain,
            new_destination_chain,
        } => execute::change_channel_to_chain_chain_link(
            deps,
            channel_id,
            source_chain,
            new_destination_chain,
        ),
        ExecuteMsg::RemoveChannelToChainChainLink {
            channel_id,
            source_chain,
        } => execute::remove_channel_to_chain_chain_link(deps, channel_id, source_chain),

        // Osmosis denom links
        ExecuteMsg::SetNativeDenomToIbcDenom {
            native_denom,
            ibc_denom,
        } => execute::set_native_denom_to_ibc_denom_link(deps, native_denom, ibc_denom),
        ExecuteMsg::ChangeNativeDenomToIbcDenom {
            native_denom,
            new_ibc_denom,
        } => execute::change_native_denom_to_ibc_denom_link(deps, native_denom, new_ibc_denom),
        ExecuteMsg::RemoveNativeDenomToIbcDenom { native_denom } => {
            execute::remove_native_denom_to_ibc_denom_link(deps, native_denom)
        }
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    deps.api.debug(&format!("executing  query: {msg:?}"));
    match msg {
        QueryMsg::GetAddressFromAlias { contract_alias } => {
            to_binary(&CONTRACT_ALIAS_MAP.load(deps.storage, &contract_alias)?)
        }
        QueryMsg::GetChainToChainChannelLink {
            source_chain,
            destination_chain,
        } => to_binary(
            &CHAIN_TO_CHAIN_CHANNEL_MAP.load(deps.storage, (&source_chain, &destination_chain))?,
        ),
        QueryMsg::GetDenomTrace { hash } => to_binary(&execute::query_denom_trace(deps, hash)?),
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
    fn setup_and_query_chain_to_chain_channel() {
        // Store three chain<>channel mappings
        let deps = setup().unwrap();

        // Retrieve osmo<>juno link and check the channel is what we expect
        let channel_binary = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetChainToChainChannelLink {
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
            QueryMsg::GetChainToChainChannelLink {
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
            QueryMsg::GetChainToChainChannelLink {
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
            QueryMsg::GetChainToChainChannelLink {
                source_chain: "osmo".to_string(),
                destination_chain: "cerberus".to_string(),
            },
        );
        assert!(channel_binary.is_err());
    }
}
