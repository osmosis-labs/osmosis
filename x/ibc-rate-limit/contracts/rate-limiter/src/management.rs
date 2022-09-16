use crate::msg::{PathMsg, QuotaMsg};
use crate::state::{Flow, Path, RateLimit, GOVMODULE, IBCMODULE, RATE_LIMIT_TRACKERS};
use crate::ContractError;
use cosmwasm_std::{Addr, DepsMut, Response, Timestamp};

pub fn add_new_paths(
    deps: DepsMut,
    path_msgs: Vec<PathMsg>,
    now: Timestamp,
) -> Result<(), ContractError> {
    for path_msg in path_msgs {
        let path = Path::new(path_msg.channel_id, path_msg.denom);

        RATE_LIMIT_TRACKERS.save(
            deps.storage,
            path.into(),
            &path_msg
                .quotas
                .iter()
                .map(|q| RateLimit {
                    quota: q.into(),
                    flow: Flow::new(0_u128, 0_u128, now, q.duration),
                })
                .collect(),
        )?
    }
    Ok(())
}

pub fn try_add_path(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    denom: String,
    quotas: Vec<QuotaMsg>,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // codenit: should we make a function for checking this authorization?
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }
    add_new_paths(deps, vec![PathMsg::new(&channel_id, &denom, quotas)], now)?;

    Ok(Response::new()
        .add_attribute("method", "try_add_channel")
        .add_attribute("channel_id", channel_id)
        .add_attribute("denom", denom))
}

pub fn try_remove_path(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    denom: String,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }

    let path = Path::new(&channel_id, &denom);
    RATE_LIMIT_TRACKERS.remove(deps.storage, path.into());
    Ok(Response::new()
        .add_attribute("method", "try_remove_channel")
        .add_attribute("denom", denom)
        .add_attribute("channel_id", channel_id))
}

// Reset specified quote_id for the given channel_id
pub fn try_reset_path_quota(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    denom: String,
    quota_id: String,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }

    let path = Path::new(&channel_id, &denom);
    RATE_LIMIT_TRACKERS.update(deps.storage, path.into(), |maybe_rate_limit| {
        match maybe_rate_limit {
            None => Err(ContractError::QuotaNotFound {
                quota_id,
                channel_id: channel_id.clone(),
                denom: denom.clone(),
            }),
            Some(mut limits) => {
                // Q: What happens here if quote_id not found? seems like we return ok?
                limits.iter_mut().for_each(|limit| {
                    if limit.quota.name == quota_id.as_ref() {
                        limit.flow.expire(now, limit.quota.duration)
                    }
                });
                Ok(limits)
            }
        }
    })?;

    Ok(Response::new()
        .add_attribute("method", "try_reset_channel")
        .add_attribute("channel_id", channel_id))
}

#[cfg(test)]
mod tests {

    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{from_binary, Addr, StdError};

    use crate::contract::{execute, query};
    use crate::helpers::tests::verify_query_response;
    use crate::msg::{ExecuteMsg, QueryMsg, QuotaMsg};
    use crate::state::{RateLimit, GOVMODULE, IBCMODULE};

    const IBC_ADDR: &str = "IBC_MODULE";
    const GOV_ADDR: &str = "GOV_MODULE";

    #[test] // Tests AddPath and RemovePath messages
    fn management_add_and_remove_path() {
        let mut deps = mock_dependencies();
        IBCMODULE
            .save(deps.as_mut().storage, &Addr::unchecked(IBC_ADDR))
            .unwrap();
        GOVMODULE
            .save(deps.as_mut().storage, &Addr::unchecked(GOV_ADDR))
            .unwrap();

        let msg = ExecuteMsg::AddPath {
            channel_id: format!("channel"),
            denom: format!("denom"),
            quotas: vec![QuotaMsg {
                name: "daily".to_string(),
                duration: 1600,
                send_recv: (3, 5),
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        let env = mock_env();
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        let query_msg = QueryMsg::GetQuotas {
            channel_id: format!("channel"),
            denom: format!("denom"),
        };

        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();

        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        verify_query_response(
            &value[0],
            "daily",
            (3, 5),
            1600,
            0,
            0,
            env.block.time.plus_seconds(1600),
        );

        assert_eq!(value.len(), 1);

        // Add another path
        let msg = ExecuteMsg::AddPath {
            channel_id: format!("channel2"),
            denom: format!("denom"),
            quotas: vec![QuotaMsg {
                name: "daily".to_string(),
                duration: 1600,
                send_recv: (3, 5),
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // remove the first one
        let msg = ExecuteMsg::RemovePath {
            channel_id: format!("channel"),
            denom: format!("denom"),
        };

        let info = mock_info(IBC_ADDR, &vec![]);
        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // The channel is not there anymore
        let err = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap_err();
        assert!(matches!(err, StdError::NotFound { .. }));

        // The second channel is still there
        let query_msg = QueryMsg::GetQuotas {
            channel_id: format!("channel2"),
            denom: format!("denom"),
        };
        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        assert_eq!(value.len(), 1);
        verify_query_response(
            &value[0],
            "daily",
            (3, 5),
            1600,
            0,
            0,
            env.block.time.plus_seconds(1600),
        );

        // Paths are overriden if they share a name and denom
        let msg = ExecuteMsg::AddPath {
            channel_id: format!("channel2"),
            denom: format!("denom"),
            quotas: vec![QuotaMsg {
                name: "different".to_string(),
                duration: 5000,
                send_recv: (50, 30),
            }],
        };
        let info = mock_info(IBC_ADDR, &vec![]);

        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        let query_msg = QueryMsg::GetQuotas {
            channel_id: format!("channel2"),
            denom: format!("denom"),
        };
        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        assert_eq!(value.len(), 1);

        verify_query_response(
            &value[0],
            "different",
            (50, 30),
            5000,
            0,
            0,
            env.block.time.plus_seconds(5000),
        );
    }
}
