#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Timestamp,
};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::management::{add_new_paths, try_add_path, try_remove_path, try_reset_path_quota};
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};
use crate::state::{FlowType, Path, RateLimit, GOVMODULE, IBCMODULE, RATE_LIMIT_TRACKERS};

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

    add_new_paths(deps, msg.paths, env.block.time)?;

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
            denom,
        } => {
            let path = Path::new(&channel_id, &denom);
            try_transfer(
                deps,
                info.sender,
                &path,
                channel_value,
                funds,
                FlowType::Out,
                env.block.time,
            )
        }
        ExecuteMsg::RecvPacket {
            channel_id,
            channel_value,
            funds,
            denom,
        } => {
            let path = Path::new(&channel_id, &denom);
            try_transfer(
                deps,
                info.sender,
                &path,
                channel_value,
                funds,
                FlowType::In,
                env.block.time,
            )
        }
        ExecuteMsg::AddPath {
            channel_id,
            denom,
            quotas,
        } => try_add_path(deps, info.sender, channel_id, denom, quotas, env.block.time),
        ExecuteMsg::RemovePath { channel_id, denom } => {
            try_remove_path(deps, info.sender, channel_id, denom)
        }
        ExecuteMsg::ResetPathQuota {
            channel_id,
            denom,
            quota_id,
        } => try_reset_path_quota(
            deps,
            info.sender,
            channel_id,
            denom,
            quota_id,
            env.block.time,
        ),
    }
}

pub struct RateLimitResponse {
    pub rate_limit: RateLimit,
    pub used: u128,
    pub max: u128,
}

// Q: Is an ICS 20 transfer only 1 denom at a time, or does the caller have to split into several
// calls if its a multi-denom ICS-20 transfer

/// This function checks the rate limit and, if successful, stores the updated data about the value
/// that has been transfered through the channel for a specific denom.
/// If the period for a RateLimit has ended, the Flow information is reset.
///
/// The channel_value is the current value of the denom for the the channel as
/// calculated by the caller. This should be the total supply of a denom
pub fn try_transfer(
    deps: DepsMut,
    sender: Addr,
    path: &Path,
    channel_value: u128,
    funds: u128,
    direction: FlowType,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Only the IBCMODULE can execute transfers
    // TODO: Should we move this to a helper method?
    //       This may not be needed once we move this function to be under sudo.
    //       Though it might still be worth checking that only the transfer module is calling it
    let ibc_module = IBCMODULE.load(deps.storage)?;
    if sender != ibc_module {
        return Err(ContractError::Unauthorized {});
    }

    let trackers = RATE_LIMIT_TRACKERS.may_load(deps.storage, path.into())?;

    let configured = match trackers {
        None => false,
        Some(ref x) if x.is_empty() => false,
        _ => true,
    };

    if !configured {
        // No Quota configured for the current path. Allowing all messages.
        // TODO: Should there be an attribute for it being allowed vs denied?
        return Ok(Response::new()
            .add_attribute("method", "try_transfer")
            .add_attribute("channel_id", path.channel.to_string())
            .add_attribute("denom", path.denom.to_string())
            .add_attribute("quota", "none"));
    }

    let mut rate_limits = trackers.unwrap();

    let results: Result<Vec<RateLimitResponse>, _> = rate_limits
        .iter_mut()
        .map(|limit| {
            // TODO: Should we move this to more methods on ChannelFlow?
            // e.g. new pseudocode
            // channel.updateIfExpired(now)
            // channel.trackSend(&direction, funds)
            // channel.CheckRateLimit(direction.clone())?;
            // (Or at least update for time + rename for track send. the last one is a bit of a code style nit)
            let max = limit.quota.capacity_at(&channel_value, &direction);
            if limit.flow.is_expired(now) {
                limit.flow.expire(now, limit.quota.duration)
            }
            limit.flow.add_flow(direction.clone(), funds);

            let balance = limit.flow.balance();
            if balance > max {
                return Err(ContractError::RateLimitExceded {
                    channel: path.channel.to_string(),
                    denom: path.denom.to_string(),
                    reset: limit.flow.period_end,
                });
            };
            Ok(RateLimitResponse {
                // TODO: nit, can we derive channel.Clone()?
                rate_limit: RateLimit {
                    quota: limit.quota.clone(),
                    flow: limit.flow,
                },
                used: balance,
                max,
            })
        })
        .collect();
    let results = results?;

    RATE_LIMIT_TRACKERS.save(
        deps.storage,
        path.into(),
        &results.iter().map(|r| r.rate_limit.clone()).collect(),
    )?;

    let response = Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string());

    // Adding the attributes from each quota to the response
    // Code style Q: Should we move attribute setting to a function on response?
    // Rust question: How does this work with one response being an error, I'm not sure how the flow works here
    results.iter().fold(Ok(response), |acc, result| {
        Ok(acc?
            .add_attribute(
                format!("{}_used", result.rate_limit.quota.name),
                result.used.to_string(),
            )
            .add_attribute(
                format!("{}_max", result.rate_limit.quota.name),
                result.max.to_string(),
            )
            .add_attribute(
                format!("{}_period_end", result.rate_limit.quota.name),
                result.rate_limit.flow.period_end.to_string(),
            ))
    })
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetQuotas { channel_id, denom } => get_quotas(deps, channel_id, denom),
    }
}

fn get_quotas(
    deps: Deps,
    channel_id: impl Into<String>,
    denom: impl Into<String>,
) -> StdResult<Binary> {
    let path = Path::new(channel_id, denom);
    to_binary(&RATE_LIMIT_TRACKERS.load(deps.storage, path.into())?)
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(_deps: DepsMut, _env: Env, _msg: MigrateMsg) -> Result<Response, ContractError> {
    unimplemented!()
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{from_binary, Addr, Attribute};

    use crate::helpers::tests::verify_query_response;
    use crate::msg::{PathMsg, QuotaMsg};
    use crate::state::RESET_TIME_WEEKLY;

    const IBC_ADDR: &str = "IBC_MODULE";
    const GOV_ADDR: &str = "GOV_MODULE";

    #[test]
    fn proper_instantiation() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // The ibc and gov modules are properly stored
        assert_eq!(IBCMODULE.load(deps.as_ref().storage).unwrap(), IBC_ADDR);
        assert_eq!(GOVMODULE.load(deps.as_ref().storage).unwrap(), GOV_ADDR);
    }

    #[test]
    fn permissions() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("Weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };

        // This succeeds
        execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let info = mock_info("SomeoneElse", &vec![]);

        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };
        let err = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized { .. }));
    }

    #[test]
    fn consume_allowance() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
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
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let info = mock_info(IBC_ADDR, &vec![]);
        let send_msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };
        let recv_msg = ExecuteMsg::RecvPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };

        let res = execute(deps.as_mut(), mock_env(), info.clone(), send_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let res = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "0");

        // We can still use the path. Even if we have sent more than the
        // allowance through the path (900 > 3000*.1), the current "balance"
        // of inflow vs outflow is still lower than the path's capacity/quota
        let res = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap();
        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "300");

        let err = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap_err();

        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
    }

    #[test]
    fn asymetric_quotas() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 1);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        // Sending 2%
        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 60,
        };
        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "60");

        // Sending 1% more. Allowed, as sending has a 10% allowance
        let msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 30,
        };

        let info = mock_info(IBC_ADDR, &vec![]);
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[3];
        assert_eq!(key, "weekly_used");
        assert_eq!(value, "90");

        // Receiving 1% should fail. 3% already executed through the path
        let recv_msg = ExecuteMsg::RecvPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 30,
        };

        let err = execute(deps.as_mut(), mock_env(), info.clone(), recv_msg.clone()).unwrap_err();
        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
    }

    #[test]
    fn query_state() {
        let mut deps = mock_dependencies();

        let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![quota],
            }],
        };
        let info = mock_info(GOV_ADDR, &vec![]);
        let env = mock_env();
        let _res = instantiate(deps.as_mut(), env.clone(), info, msg).unwrap();

        let query_msg = QueryMsg::GetQuotas {
            channel_id: format!("channel"),
            denom: format!("denom"),
        };

        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        assert_eq!(value[0].quota.name, "weekly");
        assert_eq!(value[0].quota.max_percentage_send, 10);
        assert_eq!(value[0].quota.max_percentage_recv, 10);
        assert_eq!(value[0].quota.duration, RESET_TIME_WEEKLY);
        assert_eq!(value[0].flow.inflow, 0);
        assert_eq!(value[0].flow.outflow, 0);
        assert_eq!(
            value[0].flow.period_end,
            env.block.time.plus_seconds(RESET_TIME_WEEKLY)
        );

        let info = mock_info(IBC_ADDR, &vec![]);
        let send_msg = ExecuteMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 300,
        };
        execute(deps.as_mut(), mock_env(), info.clone(), send_msg.clone()).unwrap();

        let recv_msg = ExecuteMsg::RecvPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 3_000,
            funds: 30,
        };
        execute(deps.as_mut(), mock_env(), info, recv_msg.clone()).unwrap();

        // Query
        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        verify_query_response(
            &value[0],
            "weekly",
            (10, 10),
            RESET_TIME_WEEKLY,
            30,
            300,
            env.block.time.plus_seconds(RESET_TIME_WEEKLY),
        );
    }

    #[test]
    fn bad_quotas() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![PathMsg {
                channel_id: format!("channel"),
                denom: format!("denom"),
                quotas: vec![QuotaMsg {
                    name: "bad_quota".to_string(),
                    duration: 200,
                    send_recv: (5000, 101),
                }],
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        // we can just call .unwrap() to assert this was a success
        let env = mock_env();
        instantiate(deps.as_mut(), env.clone(), info, msg).unwrap();

        // If a quota is higher than 100%, we set it to 100%
        let query_msg = QueryMsg::GetQuotas {
            channel_id: format!("channel"),
            denom: format!("denom"),
        };
        let res = query(deps.as_ref(), env.clone(), query_msg).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        verify_query_response(
            &value[0],
            "bad_quota",
            (100, 100),
            200,
            0,
            0,
            env.block.time.plus_seconds(200),
        );
    }
}
