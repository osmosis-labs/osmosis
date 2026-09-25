use std::collections::BTreeSet;

use crate::blocking::restriction_key;
use crate::msg::{PathMsg, QuotaMsg};

use crate::state::storage::{ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM, PENDING_SENDS};
use crate::state::{
    flow::Flow, path::Path, quota::Quota, rate_limit::RateLimit, storage::RATE_LIMIT_TRACKERS,
};
use crate::ContractError;
use cosmwasm_std::{DepsMut, Order, Response, StdResult, Timestamp};
use cw_storage_plus::Bound;

pub fn add_new_paths(
    deps: &mut DepsMut,
    path_msgs: Vec<PathMsg>,
    now: Timestamp,
) -> Result<(), ContractError> {
    for path_msg in path_msgs {
        validate_path_quotas(&path_msg)?;
        let path = Path::new(path_msg.channel_id, path_msg.denom);

        let limits = path_msg
            .quotas
            .iter()
            .map(|q| -> Result<RateLimit, ContractError> {
                Ok(RateLimit {
                    quota: Quota::try_from(q)?,
                    flow: Flow::new(0_u128, 0_u128, now, q.duration),
                })
            })
            .collect::<Result<Vec<_>, _>>()?;
        RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &limits)?
    }
    Ok(())
}

/// A path needs at least one quota, and ResetPathQuota and EditPathQuota
/// address quotas by name, so names must be unique within the path.
fn validate_path_quotas(path_msg: &PathMsg) -> Result<(), ContractError> {
    if path_msg.quotas.is_empty() {
        let reason = format!(
            "path {}/{} has no quotas; use RemovePath to lift a rate limit",
            path_msg.channel_id, path_msg.denom
        );
        return Err(ContractError::InvalidParameters(reason));
    }
    let mut names = BTreeSet::new();
    for quota in &path_msg.quotas {
        if !names.insert(quota.name.as_str()) {
            let reason = format!(
                "quota name {} is used more than once on {}/{}",
                quota.name, path_msg.channel_id, path_msg.denom
            );
            return Err(ContractError::InvalidParameters(reason));
        }
    }
    Ok(())
}

pub fn try_add_path(
    deps: &mut DepsMut,
    channel_id: String,
    denom: String,
    quotas: Vec<QuotaMsg>,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Adding a path that already exists replaces every quota on it and resets
    // the flows. That is allowed, but the execution event must say so, because a
    // proposal that meant to add a second quota would otherwise silently wipe
    // the first.
    let replaced = RATE_LIMIT_TRACKERS.has(deps.storage, Path::new(&channel_id, &denom).into());

    add_new_paths(deps, vec![PathMsg::new(&channel_id, &denom, quotas)], now)?;

    Ok(Response::new()
        .add_attribute("method", "try_add_channel")
        .add_attribute("channel_id", channel_id)
        .add_attribute("denom", denom)
        .add_attribute("replaced", replaced.to_string()))
}

pub fn try_remove_path(
    deps: &mut DepsMut,
    channel_id: String,
    denom: String,
) -> Result<Response, ContractError> {
    let path = Path::new(&channel_id, &denom);
    RATE_LIMIT_TRACKERS.remove(deps.storage, path.into());
    Ok(Response::new()
        .add_attribute("method", "try_remove_channel")
        .add_attribute("denom", denom)
        .add_attribute("channel_id", channel_id))
}

// Reset specified quote_id for the given channel_id
pub fn try_reset_path_quota(
    deps: &mut DepsMut,
    channel_id: String,
    denom: String,
    quota_id: String,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let path = Path::new(&channel_id, &denom);
    RATE_LIMIT_TRACKERS.update(deps.storage, path.into(), |maybe_rate_limit| {
        match maybe_rate_limit {
            None => Err(ContractError::QuotaNotFound {
                quota_id,
                channel_id: channel_id.clone(),
                denom: denom.clone(),
            }),
            Some(mut limits) => {
                let mut matched = false;
                limits.iter_mut().for_each(|limit| {
                    if limit.quota.name == quota_id.as_ref() {
                        limit.flow.expire(now, limit.quota.duration);
                        matched = true;
                    }
                });
                // A reset that matches nothing must fail rather than pass silently,
                // otherwise a typo in a proposal looks like a successful reset.
                if !matched {
                    return Err(ContractError::QuotaNotFound {
                        quota_id,
                        channel_id: channel_id.clone(),
                        denom: denom.clone(),
                    });
                }
                Ok(limits)
            }
        }
    })?;

    Ok(Response::new()
        .add_attribute("method", "try_reset_channel")
        .add_attribute("channel_id", channel_id))
}

pub fn edit_path_quota(
    deps: &mut DepsMut,
    channel_id: String,
    denom: String,
    quota: QuotaMsg,
) -> Result<(), ContractError> {
    let path = Path::new(&channel_id, &denom);
    RATE_LIMIT_TRACKERS.update(deps.storage, path.into(), |maybe_rate_limit| {
        match maybe_rate_limit {
            None => Err(ContractError::QuotaNotFound {
                quota_id: quota.name,
                channel_id: channel_id.clone(),
                denom: denom.clone(),
            }),
            Some(mut limits) => {
                let mut matched = false;
                for limit in limits.iter_mut() {
                    if limit.quota.name.eq(&quota.name) {
                        // Keep the channel value cached for the current window;
                        // the new bounds apply to it straight away.
                        let channel_value = limit.quota.channel_value;
                        limit.quota = Quota::try_from(&quota)?;
                        limit.quota.channel_value = channel_value;
                        matched = true;
                    }
                }
                // An edit that matches nothing must fail rather than pass silently
                if !matched {
                    return Err(ContractError::QuotaNotFound {
                        quota_id: quota.name,
                        channel_id: channel_id.clone(),
                        denom: denom.clone(),
                    });
                }
                Ok(limits)
            }
        }
    })?;
    Ok(())
}

pub fn set_denom_restrictions(
    deps: &mut DepsMut,
    denom: String,
    allowed_channels: Vec<String>,
) -> Result<Response, ContractError> {
    // An empty list is not "block everywhere": check_restricted_denoms treats it
    // as no restriction at all, so storing it would be a no-op that looks like a
    // block. Lifting a restriction goes through UnsetDenomRestrictions.
    if allowed_channels.is_empty() {
        let reason = "allowed_channels is empty; use UnsetDenomRestrictions to lift a restriction";
        return Err(ContractError::InvalidParameters(reason.to_string()));
    }
    // Store under the canonical key so the same token can never carry two
    // entries. A legacy packet-form entry for this denom is removed with it,
    // otherwise it would keep shadowing the new one.
    let key = restriction_key(&denom);
    if key != denom {
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(deps.storage, denom);
    }
    ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.save(deps.storage, key.clone(), &allowed_channels)?;
    Ok(Response::new()
        .add_attribute("method", "set_denom_restrictions")
        .add_attribute("key", key))
}

pub fn unset_denom_restrictions(
    deps: &mut DepsMut,
    denom: String,
) -> Result<Response, ContractError> {
    // Remove both spellings so no legacy packet-form entry survives
    ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(deps.storage, restriction_key(&denom));
    ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(deps.storage, denom);
    Ok(Response::new().add_attribute("method", "unset_denom_restrictions"))
}

/// Upper bound on the entries one housekeeping call may scan and decode. The
/// purge messages are permissionless, so the bound is the contract's, not the
/// caller's; a limit above it is rejected rather than silently clamped.
pub const MAX_PURGE_BATCH: u32 = 100;

fn checked_batch_limit(limit: u32) -> Result<usize, ContractError> {
    if limit == 0 || limit > MAX_PURGE_BATCH {
        let reason = format!("limit must be between 1 and {MAX_PURGE_BATCH}");
        return Err(ContractError::InvalidParameters(reason));
    }
    Ok(limit as usize)
}

/// Permissionless. Removes up to `limit` tracker entries, starting after
/// `start_after`, whose quota vector is empty. Versions before 0.2.0 wrote one
/// of those for every (channel, denom) pair they saw, and anyone could add
/// more by sending a new denom, so the cleanup is bounded per call and never
/// part of a migration. Reports the last key scanned for the next call.
pub fn purge_empty_paths(
    deps: &mut DepsMut,
    start_after: Option<(String, String)>,
    limit: u32,
) -> Result<Response, ContractError> {
    let limit = checked_batch_limit(limit)?;
    let scanned: Vec<((String, String), bool)> = RATE_LIMIT_TRACKERS
        .range(
            deps.storage,
            start_after.map(Bound::exclusive),
            None,
            Order::Ascending,
        )
        .take(limit)
        .map(|item| item.map(|(key, limits)| (key, limits.is_empty())))
        .collect::<StdResult<_>>()?;

    let mut purged = 0_u32;
    let mut last_key = "none".to_string();
    for (key, empty) in scanned.iter() {
        last_key = format!("{}/{}", key.0, key.1);
        if *empty {
            RATE_LIMIT_TRACKERS.remove(deps.storage, key.clone());
            purged += 1;
        }
    }
    Ok(Response::new()
        .add_attribute("method", "purge_empty_paths")
        .add_attribute("scanned", scanned.len().to_string())
        .add_attribute("purged", purged.to_string())
        .add_attribute("last_key", last_key))
}

/// Permissionless. Removes up to `limit` pending-send records, starting after
/// `start_after`, whose windows have all ended. A refund for those could no
/// longer change an active window, so the record only takes up space.
pub fn purge_stale_sends(
    deps: &mut DepsMut,
    now: Timestamp,
    start_after: Option<(String, u64)>,
    limit: u32,
) -> Result<Response, ContractError> {
    let limit = checked_batch_limit(limit)?;
    let scanned: Vec<((String, u64), bool)> = PENDING_SENDS
        .range(
            deps.storage,
            start_after.map(Bound::exclusive),
            None,
            Order::Ascending,
        )
        .take(limit)
        .map(|item| item.map(|(key, pending)| (key, pending.latest_period_end() < now)))
        .collect::<StdResult<_>>()?;

    let mut purged = 0_u32;
    let mut last_key = "none".to_string();
    for (key, stale) in scanned.iter() {
        last_key = format!("{}/{}", key.0, key.1);
        if *stale {
            PENDING_SENDS.remove(deps.storage, key.clone());
            purged += 1;
        }
    }
    Ok(Response::new()
        .add_attribute("method", "purge_stale_sends")
        .add_attribute("scanned", scanned.len().to_string())
        .add_attribute("purged", purged.to_string())
        .add_attribute("last_key", last_key))
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{from_binary, Addr, StdError};

    use crate::contract::{execute, query};
    use crate::helpers::tests::verify_query_response;
    use crate::msg::{ExecuteMsg, QueryMsg, QuotaMsg};
    use crate::state::rbac::Roles;
    use crate::state::{
        rate_limit::RateLimit,
        storage::{ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM, GOVMODULE, IBCMODULE},
    };
    use crate::ContractError;

    const IBC_ADDR: &str = "osmo1vz5e6tzdjlzy2f7pjvx0ecv96h8r4m2y92thdm";
    const GOV_ADDR: &str = "osmo1tzz5zf2u68t00un2j4lrrnkt2ztd46kfzfp58r";

    #[test] // Tests AddPath and RemovePath messages
    fn management_add_and_remove_path() {
        let mut deps = mock_dependencies();
        IBCMODULE
            .save(deps.as_mut().storage, &Addr::unchecked(IBC_ADDR))
            .unwrap();
        GOVMODULE
            .save(deps.as_mut().storage, &Addr::unchecked(GOV_ADDR))
            .unwrap();

        // grant role to IBC_ADDR
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            IBC_ADDR.to_string(),
            vec![Roles::AddRateLimit, Roles::RemoveRateLimit],
        )
        .unwrap();

        let msg = ExecuteMsg::AddPath {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("daily", 1600, 3, 5)],
        };
        let info = mock_info(IBC_ADDR, &[]);

        let env = mock_env();
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        let query_msg = QueryMsg::GetQuotas {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
        };

        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();

        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        verify_query_response(
            &value[0],
            "daily",
            (3, 5),
            1600,
            0_u32.into(),
            0_u32.into(),
            env.block.time.plus_seconds(1600),
        );

        assert_eq!(value.len(), 1);

        // Add another path
        let msg = ExecuteMsg::AddPath {
            channel_id: "channel2".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("daily", 1600, 3, 5)],
        };
        let info = mock_info(IBC_ADDR, &[]);

        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // remove the first one
        let msg = ExecuteMsg::RemovePath {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
        };

        let info = mock_info(IBC_ADDR, &[]);
        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // The channel is not there anymore
        let err = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap_err();
        assert!(matches!(err, StdError::NotFound { .. }));

        // The second channel is still there
        let query_msg = QueryMsg::GetQuotas {
            channel_id: "channel2".to_string(),
            denom: "denom".to_string(),
        };
        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        assert_eq!(value.len(), 1);
        verify_query_response(
            &value[0],
            "daily",
            (3, 5),
            1600,
            0_u32.into(),
            0_u32.into(),
            env.block.time.plus_seconds(1600),
        );

        // Paths are overriden if they share a name and denom
        let msg = ExecuteMsg::AddPath {
            channel_id: "channel2".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("different", 5000, 50, 30)],
        };
        let info = mock_info(IBC_ADDR, &[]);

        let env = mock_env();
        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        let query_msg = QueryMsg::GetQuotas {
            channel_id: "channel2".to_string(),
            denom: "denom".to_string(),
        };
        let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
        let value: Vec<RateLimit> = from_binary(&res).unwrap();
        assert_eq!(value.len(), 1);

        verify_query_response(
            &value[0],
            "different",
            (50, 30),
            5000,
            0_u32.into(),
            0_u32.into(),
            env.block.time.plus_seconds(5000),
        );
    }

    #[test]
    fn test_execute_set_denom_restrictions() {
        let mut deps = mock_dependencies();

        // Set up the message and the environment
        let denom = "denom1".to_string();
        let allowed_channels = vec!["channel1".to_string(), "channel2".to_string()];
        let msg = ExecuteMsg::SetDenomRestrictions {
            denom: denom.clone(),
            allowed_channels: allowed_channels.clone(),
        };
        let info = mock_info("executor", &[]);

        // Grant the necessary role
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            "executor".to_string(),
            vec![Roles::ManageDenomRestrictions],
        )
        .unwrap();

        // Execute the message
        let res = execute(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "set_denom_restrictions");

        // Verify the restriction was set
        let stored_channels = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .load(deps.as_ref().storage, denom)
            .unwrap();
        assert_eq!(stored_channels, allowed_channels);
    }

    #[test]
    fn test_execute_unset_denom_restrictions() {
        let mut deps = mock_dependencies();

        // First, set a restriction
        let denom = "denom1".to_string();
        let allowed_channels = vec!["channel1".to_string()];
        let set_msg = ExecuteMsg::SetDenomRestrictions {
            denom: denom.clone(),
            allowed_channels: allowed_channels.clone(),
        };
        let info = mock_info("executor", &[]);

        // Grant the necessary role
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            "executor".to_string(),
            vec![Roles::ManageDenomRestrictions],
        )
        .unwrap();

        // Execute the set message
        execute(deps.as_mut(), mock_env(), info.clone(), set_msg).unwrap();

        // Verify the restriction was set
        let stored_channels = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .load(deps.as_ref().storage, denom.clone())
            .unwrap();
        assert_eq!(stored_channels, allowed_channels);

        // Now unset the restriction
        let unset_msg = ExecuteMsg::UnsetDenomRestrictions {
            denom: denom.clone(),
        };
        let res = execute(deps.as_mut(), mock_env(), info, unset_msg).unwrap();
        assert_eq!(res.attributes[0].value, "unset_denom_restrictions");

        // Verify the restriction was removed
        let stored_channels = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .may_load(deps.as_ref().storage, denom)
            .unwrap();
        assert!(stored_channels.is_none());
    }

    #[test]
    fn test_query_denom_restrictions() {
        let mut deps = mock_dependencies();

        // Set up initial restrictions
        let denom = "denom1".to_string();
        let allowed_channels = vec!["channel1".to_string(), "channel2".to_string()];
        let set_msg = ExecuteMsg::SetDenomRestrictions {
            denom: denom.clone(),
            allowed_channels: allowed_channels.clone(),
        };
        let info = mock_info("executor", &[]);

        // Grant the necessary role
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            "executor".to_string(),
            vec![Roles::ManageDenomRestrictions],
        )
        .unwrap();

        // Execute the set message
        execute(deps.as_mut(), mock_env(), info, set_msg).unwrap();

        // Query the restrictions
        let query_msg = QueryMsg::GetDenomRestrictions {
            denom: denom.clone(),
        };
        let res = query(deps.as_ref(), mock_env(), query_msg).unwrap();
        let returned_channels: Vec<String> = from_binary(&res).unwrap();
        assert_eq!(returned_channels, allowed_channels);
    }

    #[test]
    fn test_query_unset_denom_restrictions() {
        let deps = mock_dependencies();

        // Attempt to query restrictions on a denom with no restrictions
        let denom = "denom1".to_string();
        let query_msg = QueryMsg::GetDenomRestrictions {
            denom: denom.clone(),
        };
        query(deps.as_ref(), mock_env(), query_msg).unwrap_err();
    }

    #[test]
    fn test_set_denom_restrictions_rejects_empty_channel_list() {
        let mut deps = mock_dependencies();
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            "executor".to_string(),
            vec![Roles::ManageDenomRestrictions],
        )
        .unwrap();

        let msg = ExecuteMsg::SetDenomRestrictions {
            denom: "denom1".to_string(),
            allowed_channels: vec![],
        };
        let err = execute(deps.as_mut(), mock_env(), mock_info("executor", &[]), msg).unwrap_err();
        assert!(matches!(err, ContractError::InvalidParameters(_)));
        assert!(ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .may_load(deps.as_ref().storage, "denom1".to_string())
            .unwrap()
            .is_none());
    }

    #[test]
    fn test_reset_and_edit_unknown_quota_fail() {
        let mut deps = mock_dependencies();
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            GOV_ADDR.to_string(),
            vec![
                Roles::AddRateLimit,
                Roles::ResetPathQuota,
                Roles::EditPathQuota,
            ],
        )
        .unwrap();
        let info = mock_info(GOV_ADDR, &[]);

        let add = ExecuteMsg::AddPath {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("daily", 1600, 3, 5)],
        };
        execute(deps.as_mut(), mock_env(), info.clone(), add).unwrap();

        // The path exists but no quota is called "weekly": both must fail, not pass silently
        let reset = ExecuteMsg::ResetPathQuota {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quota_id: "weekly".to_string(),
        };
        let err = execute(deps.as_mut(), mock_env(), info.clone(), reset).unwrap_err();
        assert!(matches!(err, ContractError::QuotaNotFound { .. }));

        let edit = ExecuteMsg::EditPathQuota {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quota: QuotaMsg::new("weekly", 1600, 3, 5),
        };
        let err = execute(deps.as_mut(), mock_env(), info.clone(), edit).unwrap_err();
        assert!(matches!(err, ContractError::QuotaNotFound { .. }));

        // The matching name still works
        let reset = ExecuteMsg::ResetPathQuota {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quota_id: "daily".to_string(),
        };
        execute(deps.as_mut(), mock_env(), info, reset).unwrap();
    }

    #[test]
    fn test_add_path_reports_replacement() {
        let mut deps = mock_dependencies();
        crate::rbac::grant_role(
            &mut deps.as_mut(),
            GOV_ADDR.to_string(),
            vec![Roles::AddRateLimit],
        )
        .unwrap();
        let info = mock_info(GOV_ADDR, &[]);
        let add = ExecuteMsg::AddPath {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("daily", 1600, 3, 5)],
        };

        let res = execute(deps.as_mut(), mock_env(), info.clone(), add.clone()).unwrap();
        let replaced = res.attributes.iter().find(|a| a.key == "replaced").unwrap();
        assert_eq!(replaced.value, "false");

        let res = execute(deps.as_mut(), mock_env(), info, add).unwrap();
        let replaced = res.attributes.iter().find(|a| a.key == "replaced").unwrap();
        assert_eq!(replaced.value, "true");
    }

    #[test]
    fn test_permissions_enforced() {
        let mut deps = mock_dependencies();

        // Set up the message and the environment
        let denom = "denom1".to_string();
        let allowed_channels = vec!["channel1".to_string(), "channel2".to_string()];
        let msg = ExecuteMsg::SetDenomRestrictions {
            denom: denom.clone(),
            allowed_channels: allowed_channels.clone(),
        };
        let info = mock_info("unauthorized_user", &[]);

        // Attempt to execute the message without the necessary role
        let err = execute(deps.as_mut(), mock_env(), info, msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized { .. }));

        // Verify no restrictions were set
        let stored_channels = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .may_load(deps.as_ref().storage, denom)
            .unwrap();
        assert!(stored_channels.is_none());
    }
}
