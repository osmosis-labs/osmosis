#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::execute::execute_swap;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg};
use crate::state::{Config, CONFIG};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:outpost";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Handling contract instantiation
#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    // Validate the XC swap contract addr.
    // This needs to be done with bech32 because the prefix may be different than the current chain
    let Ok((prefix, _, _)) = bech32::decode(msg.crosschain_swaps_contract.as_str()) else {
        return Err(ContractError::InvalidCrosschainSwapsContract {
            contract: msg.crosschain_swaps_contract,
        })
    };
    if prefix != "osmo" {
        return Err(ContractError::InvalidCrosschainSwapsContract {
            contract: format!("invalid prefix: {}", msg.crosschain_swaps_contract),
        });
    }

    // Store the contract addr and the osmosis channel
    let state = Config {
        osmosis_channel: msg.osmosis_channel,
        crosschain_swaps_contract: msg.crosschain_swaps_contract,
    };
    CONFIG.save(deps.storage, &state)?;
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn migrate(_deps: DepsMut, _env: Env, msg: MigrateMsg) -> Result<Response, ContractError> {
    match msg {}
}

/// Handling contract execution
#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::OsmosisSwap { .. } => {
            // IBC transfers support only one token at a time
            let coin = cw_utils::one_coin(&info)?;
            execute_swap(deps, env.contract.address, env.block.time, coin, msg)
        }
    }
}
