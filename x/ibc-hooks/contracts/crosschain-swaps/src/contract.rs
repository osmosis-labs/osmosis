#[cfg(not(feature = "imported"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Reply, Response, StdResult,
};
use cw2::set_contract_version;

use crate::consts::MsgReplyID;
use crate::error::ContractError;
use crate::msg::{ExecuteMsg, IBCLifecycleComplete, InstantiateMsg, MigrateMsg, QueryMsg, SudoMsg};
use crate::state::{Config, CHANNEL_MAP, CONFIG, RECOVERY_STATES};
use crate::{execute, ibc_lifecycle};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:crosschain-swaps";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Handling contract instantiation
#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    // validate contract addresses and save to config
    let swap_contract = deps.api.addr_validate(&msg.swap_contract)?;
    let state = Config { swap_contract };
    CONFIG.save(deps.storage, &state)?;
    for (prefix, channel) in msg.channels.into_iter() {
        CHANNEL_MAP.save(deps.storage, &prefix, &channel)?;
    }

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
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
        ExecuteMsg::OsmosisSwap {
            input_coin,
            output_denom,
            receiver,
            slippage,
            next_memo,
            on_failed_delivery,
        } => execute::swap_and_forward(
            deps,
            env.block.time,
            env.contract.address,
            input_coin,
            output_denom,
            slippage,
            receiver,
            next_memo,
            on_failed_delivery,
        ),
        ExecuteMsg::Recover {} => execute::recover(deps, info.sender),
    }
}

/// Handling contract queries
#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Recoverable { addr } => to_binary(
            &RECOVERY_STATES
                .may_load(deps.storage, &addr)?
                .or(Some(vec![])),
        ),
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::IBCLifecycleComplete(IBCLifecycleComplete::IBCAck {
            channel,
            sequence,
            ack,
            success,
        }) => ibc_lifecycle::receive_ack(deps, channel, sequence, ack, success),
        SudoMsg::IBCLifecycleComplete(IBCLifecycleComplete::IBCTimeout { channel, sequence }) => {
            ibc_lifecycle::receive_timeout(deps, channel, sequence)
        }
    }
}

#[cfg_attr(not(feature = "imported"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, reply: Reply) -> Result<Response, ContractError> {
    deps.api
        .debug(&format!("executing crosschain reply: {reply:?}"));
    match MsgReplyID::from_repr(reply.id) {
        Some(MsgReplyID::Swap) => execute::handle_swap_reply(deps, reply),
        Some(MsgReplyID::Forward) => execute::handle_forward_reply(deps, reply),
        None => Err(ContractError::InvalidReplyID { id: reply.id }),
    }
}
