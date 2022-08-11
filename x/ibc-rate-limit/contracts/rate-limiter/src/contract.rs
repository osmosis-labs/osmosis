#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Timestamp};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{Channel, ExecuteMsg, InstantiateMsg, QuotaMsg};
use crate::state::{ChannelFlow, Flow, FlowType, CHANNEL_FLOWS, GOVMODULE, IBCMODULE};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:rate-limiter";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    IBCMODULE.save(deps.storage, &msg.ibc_module)?;
    GOVMODULE.save(deps.storage, &msg.gov_module)?;

    add_new_channels(deps, msg.channels, env.block.time)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("ibc_module", msg.ibc_module.to_string())
        .add_attribute("gov_module", msg.gov_module.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SendPacket {
            channel_id,
            channel_value,
            funds,
        } => try_transfer(
            deps,
            info.sender,
            channel_id,
            channel_value,
            funds,
            FlowType::Out,
            env.block.time,
        ),
        ExecuteMsg::RecvPacket {
            channel_id,
            channel_value,
            funds,
        } => try_transfer(
            deps,
            info.sender,
            channel_id,
            channel_value,
            funds,
            FlowType::In,
            env.block.time,
        ),
        ExecuteMsg::AddChannel { channel_id, quotas } => {
            try_add_channel(deps, info.sender, channel_id, quotas, env.block.time)
        }
        ExecuteMsg::RemoveChannel { channel_id } => {
            try_remove_channel(deps, info.sender, channel_id)
        }
        ExecuteMsg::ResetChannelQuota {
            channel_id,
            quota_id,
        } => try_reset_channel_quota(deps, info.sender, channel_id, quota_id, env.block.time),
    }
}

pub fn add_new_channels(
    deps: DepsMut,
    channels: Vec<Channel>,
    now: Timestamp,
) -> Result<(), ContractError> {
    for channel in channels {
        CHANNEL_FLOWS.save(
            deps.storage,
            &channel.name,
            &channel
                .quotas
                .iter()
                .map(|q| ChannelFlow {
                    quota: q.into(),
                    flow: Flow::new(0_u128, 0_u128, now, q.duration),
                })
                .collect(),
        )?
    }
    Ok(())
}

pub fn try_add_channel(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    quotas: Vec<QuotaMsg>,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }
    add_new_channels(
        deps,
        vec![Channel {
            name: channel_id.to_string(),
            quotas,
        }],
        now,
    )?;

    Ok(Response::new()
        .add_attribute("method", "try_add_channel")
        .add_attribute("channel_id", channel_id))
}

pub fn try_remove_channel(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }
    CHANNEL_FLOWS.remove(deps.storage, &channel_id);
    Ok(Response::new()
        .add_attribute("method", "try_remove_channel")
        .add_attribute("channel_id", channel_id))
}

pub fn try_reset_channel_quota(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    quota_id: String,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }

    CHANNEL_FLOWS.update(
        deps.storage,
        &channel_id.clone(),
        |maybe_flows| match maybe_flows {
            None => Err(ContractError::QuotaNotFound {
                quota_id,
                channel_id: channel_id.clone(),
            }),
            Some(mut flows) => {
                flows.iter_mut().for_each(|channel| {
                    if channel.quota.name == channel_id.as_ref() {
                        channel.flow.expire(now, channel.quota.duration)
                    }
                });
                Ok(flows)
            }
        },
    )?;

    Ok(Response::new()
        .add_attribute("method", "try_reset_channel")
        .add_attribute("channel_id", channel_id))
}

pub struct ChannelFlowResponse {
    pub channel_flow: ChannelFlow,
    pub used: u128,
    pub max: u128,
}

pub fn try_transfer(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    channel_value: u128,
    funds: u128,
    direction: FlowType,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Only the IBCMODULE can execute transfers
    let ibc_module = IBCMODULE.load(deps.storage)?;
    if sender != ibc_module {
        return Err(ContractError::Unauthorized {});
    }

    let channels = CHANNEL_FLOWS.may_load(deps.storage, &channel_id)?;

    let configured = match channels {
        None => false,
        Some(ref x) if x.len() == 0 => false,
        _ => true,
    };

    if !configured {
        // No Quota configured for the current channel. Allowing all messages.
        return Ok(Response::new()
            .add_attribute("method", "try_transfer")
            .add_attribute("channel_id", channel_id)
            .add_attribute("quota", "none"));
    }

    let mut channels = channels.unwrap();

    let results: Result<Vec<ChannelFlowResponse>, _> = channels
        .iter_mut()
        .map(|channel| {
            let max = channel.quota.capacity_at(&channel_value, &direction);
            if channel.flow.is_expired(now) {
                channel.flow.expire(now, channel.quota.duration)
            }
            channel.flow.add_flow(direction.clone(), funds);

            let balance = channel.flow.balance();
            if balance > max {
                return Err(ContractError::RateLimitExceded {
                    channel: channel_id.to_string(),
                    reset: channel.flow.period_end,
                });
            };
            Ok(ChannelFlowResponse {
                channel_flow: ChannelFlow {
                    quota: channel.quota.clone(),
                    flow: channel.flow.clone(),
                },
                used: balance,
                max,
            })
        })
        .collect();
    let results = results?;

    CHANNEL_FLOWS.save(
        deps.storage,
        &channel_id,
        &results.iter().map(|r| r.channel_flow.clone()).collect(),
    )?;

    let response = Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", channel_id);

    // Adding the attributes from each quota to the response
    results.iter().fold(Ok(response), |acc, result| {
        Ok(acc?
            .add_attribute(
                format!("{}_used", result.channel_flow.quota.name),
                result.used.to_string(),
            )
            .add_attribute(
                format!("{}_max", result.channel_flow.quota.name),
                result.max.to_string(),
            )
            .add_attribute(
                format!("{}_period_end", result.channel_flow.quota.name),
                result.channel_flow.flow.period_end.to_string(),
            ))
    })
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: ExecuteMsg) -> StdResult<Binary> {
    todo!()
}

#[cfg(test)]
mod tests {
    use crate::msg::{Channel, QuotaMsg};
    use crate::state::RESET_TIME_WEEKLY;

    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{Addr, Attribute};

    const IBC_ADDR: &str = "IBC_MODULE";
    const GOV_ADDR: &str = "GOV_MODULE";

    #[test]
    fn proper_instantiation() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels: vec![],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // TODO: Check initialization values are correct
    }

    #[test]
    fn permissions() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("Weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels: vec![Channel {
                name: "channel".to_string(),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };

        // This succeeds
        execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let info = mock_info("SomeoneElse", &vec![]);

        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let err = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized { .. }));
    }

    #[test]
    fn consume_allowance1() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels: vec![Channel {
                name: "channel".to_string(),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let err = execute(deps.as_mut(), mock_env(), info, msg).unwrap_err();
        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
    }

    #[test]
    fn symetric_flows_dont_consume_allowance() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels: vec![Channel {
                name: "channel".to_string(),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let info = mock_info(IBC_ADDR, &vec![]);
        let send_msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let recv_msg = ExecuteMsg::RecvPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };

        let res = execute(deps.as_mut(), mock_env(), info.clone(), send_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let res = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "0");

        // We can still use the channel. Even if we have sent more than the
        // allowance through the channel (900 > 3000*.1), the current "balance"
        // of inflow vs outflow is still lower than the channel's capacity/quota
        let res = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let err = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap_err();

        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
        //assert_eq!(18, value.count);
    }

    #[test]
    fn asymetric_quotas() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 1);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels: vec![Channel {
                name: "channel".to_string(),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        // Sending 2%
        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 60,
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "60");

        // Sending 1% more. Allowed, as sending has a 10% allowance
        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 30,
        };

        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "90");

        // Receiving 1% should fail. 3% already executed through the channel
        let recv_msg = ExecuteMsg::RecvPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 30,
        };

        let err = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap_err();
        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
    }
}
