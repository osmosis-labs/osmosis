use std::collections::HashSet;

use cosmwasm_std::{Addr, DepsMut, MessageInfo};

use crate::{msg::ExecuteMsg, state::{rbac::Roles, storage::{RBAC_PERMISSIONS, TIMELOCK_DELAY}}, ContractError};

/// Check to see if the sender of the message can invoke the message by holding the required rbac role
/// 
/// # Errors
/// 
/// ContractError::Unauthorized if the sender does not have the required permission
/// 
/// StdErr::NotFound if the RBAC_PERMISSIONS storage variable does not have an entry for the sender
pub fn can_invoke_message(
    deps: &DepsMut,
    info: &MessageInfo,
    msg: &ExecuteMsg,
) -> Result<(), ContractError> {
    // get the required permission to execute the message
    let Some(required_permission) = msg.required_permission() else {
        // no permission required so return ok
        return Ok(());
    };
    let permissions = RBAC_PERMISSIONS.load(deps.storage, info.sender.to_string())?;
    if permissions.contains(&required_permission) {
        return Ok(())
    }
    Err(ContractError::Unauthorized {  })
}

/// Sets a timelock delay for `signer` of `hours`
pub fn set_timelock_delay(
    deps: &mut DepsMut,
    signer: String,
    hours: u64
) -> Result<(), ContractError> {
    let signer = deps.api.addr_validate(&signer)?;
    Ok(TIMELOCK_DELAY.save(deps.storage, signer.to_string(), &hours)?)
}

/// Grants `roles` to `signer`
pub fn grant_role(
    deps: &mut DepsMut,
    signer: String,
    roles: Vec<Roles>
) -> Result<(), ContractError> {
    let signer = deps.api.addr_validate(&signer)?;
    // get the current roles, if no current roles will be an empty vec
    let mut current_roles = RBAC_PERMISSIONS.load(deps.storage, signer.to_string()).unwrap_or_default();
    for role in roles {
        current_roles.insert(role);
    }

    // persist new roles
    Ok(RBAC_PERMISSIONS.save(deps.storage, signer.to_string(), &current_roles)?)
}

// Revokes `roles` from `signer`, if this results in an empty set of roles remove the storage variable
pub fn revoke_role(
    deps: &mut DepsMut,
    signer: String,
    roles: Vec<Roles>
) -> Result<(), ContractError> {
    let signer = deps.api.addr_validate(&signer)?;

    let mut current_roles = RBAC_PERMISSIONS.load(deps.storage, signer.to_string())?;
    for role in roles {
        current_roles.remove(&role);
    }
    if current_roles.is_empty() {
        // no more roles, remove storage variable to save resources
        RBAC_PERMISSIONS.remove(deps.storage, signer.to_string());
        Ok(())
    } else {
        Ok(RBAC_PERMISSIONS.save(deps.storage, signer.to_string(), &current_roles)?)
    }
    
}

#[cfg(test)]
mod test {
    use std::collections::BTreeSet;

    use cosmwasm_std::{testing::mock_dependencies, Addr};
    use itertools::Itertools;
    use crate::{msg::QuotaMsg, state::rbac::Roles};

    use super::*;
    #[test]
    fn test_set_timelock_delay() {
        let mut deps = mock_dependencies();
        assert!(TIMELOCK_DELAY.load(&deps.storage, "foobar".to_string()).is_err());
        set_timelock_delay(&mut deps.as_mut(), "foobar".to_string(), 6).unwrap();
        assert_eq!(TIMELOCK_DELAY.load(&deps.storage, "foobar".to_string()).unwrap(), 6);
    }
    #[test]
    fn test_can_invoke_add_path() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };
        let msg = ExecuteMsg::AddPath { 
            channel_id: "channelid".into(), 
            denom: "denom".into(), 
            quotas: vec![]
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::AddRateLimit].into_iter().collect()).unwrap();

        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());

    }

    #[test]
    fn test_can_invoke_remove_path() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };
        let msg = ExecuteMsg::RemovePath { 
            channel_id: "channelid".into(), 
            denom: "denom".into(), 
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::RemoveRateLimit].into_iter().collect()).unwrap();

        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }

    #[test]
    fn test_can_invoke_reset_path_quota() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::ResetPathQuota { 
            channel_id: "channelid".into(), 
            denom: "denom".into(),
            quota_id: "quota".into()
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::ResetPathQuota].into_iter().collect()).unwrap();

        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }

    #[test]
    fn test_can_invoke_grant_role() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::GrantRole { 
            signer: "signer".into(),
            roles: vec![Roles::GrantRole]
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::GrantRole].into_iter().collect()).unwrap();

        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }

    #[test]
    fn test_can_invoke_revoke_role() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::RevokeRole { 
            signer: "signer".into(),
            roles: vec![Roles::GrantRole]
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::RevokeRole].into_iter().collect()).unwrap();
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }

    #[test]
    fn test_can_invoke_edit_path_quota() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::EditPathQuota { 
            quota: QuotaMsg {
                name: "name".into(),
                duration: 0,
                send_recv: (1, 2),
            },
            channel_id: "channel_id".into(),
            denom: "denom".into()
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::EditPathQuota].into_iter().collect()).unwrap();
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }

    #[test]
    fn test_can_invoke_remove_message() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::RemoveMessage { 
            message_id: "message".into()
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::RemoveMessage].into_iter().collect()).unwrap();
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());
    }


    #[test]
    fn test_can_invoke_set_timelock_delay() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::SetTimelockDelay { 
            signer: "signer".into(),
            hours: 5,
        };
        RBAC_PERMISSIONS.save(&mut deps.storage, "foobar".to_string(), &vec![Roles::SetTimelockDelay].into_iter().collect()).unwrap();
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobar,
            &msg            
        ).is_ok());
        assert!(can_invoke_message(
            &deps.as_mut(),
            &info_foobarbaz,
            &msg            
        ).is_err());

    }

    #[test]
    fn test_can_invoke_process_messages() {
        let mut deps = mock_dependencies();


        let info_foobar = MessageInfo {
            sender: Addr::unchecked("foobar".to_string()),
            funds: vec![]
        };
        let info_foobarbaz = MessageInfo {
            sender: Addr::unchecked("foobarbaz".to_string()),
            funds: vec![]
        };

        let msg = ExecuteMsg::ProcessMessages { count: Some(1),message_ids: None };

        // all addresses should be able to invoke this
        assert!(
            can_invoke_message(
                &deps.as_mut(),
                &info_foobar,
                &msg
            ).is_ok()
        );
        assert!(
            can_invoke_message(
                &deps.as_mut(),
                &info_foobarbaz,
                &msg
            ).is_ok()
        );

        // try again with message_ids Some


        let msg = ExecuteMsg::ProcessMessages { count: None, message_ids: Some(vec!["foobar".to_string()]) };

        // all addresses should be able to invoke this
        assert!(
            can_invoke_message(
                &deps.as_mut(),
                &info_foobar,
                &msg
            ).is_ok()
        );
        assert!(
            can_invoke_message(
                &deps.as_mut(),
                &info_foobarbaz,
                &msg
            ).is_ok()
        );

    }

    #[test]
    fn test_grant_role() {
        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();
        
        let all_roles = Roles::all_roles().into_iter().chunks(2);

        // no roles, should fail
        assert!(RBAC_PERMISSIONS.load(deps.storage, "signer".to_string()).is_err());

        let mut granted_roles = BTreeSet::new();

        for roles in &all_roles {
            let roles = roles.collect::<Vec<_>>();

            grant_role(
                &mut deps,
                "signer".to_string(),
                roles.clone()
            ).unwrap();
            roles.iter().for_each(|role| { granted_roles.insert(*role); } );

            let assigned_roles = RBAC_PERMISSIONS.load(deps.storage, "signer".to_string()).unwrap();

            assert_eq!(granted_roles, assigned_roles);
        }

    }

    #[test]
    fn test_revoke_role() {
        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();

        let all_roles = Roles::all_roles();
        // no roles, should fail
        assert!(RBAC_PERMISSIONS.load(deps.storage, "signer".to_string()).is_err());

        // grant all roles
        RBAC_PERMISSIONS.save(deps.storage, "signer".to_string(), &all_roles.iter().copied().collect::<BTreeSet<_>>()).unwrap();

        let mut granted_roles: BTreeSet<_> = all_roles.iter().copied().collect();

        for roles in &all_roles.iter().chunks(2) {
            let roles = roles.map(|role| *role).collect::<Vec<_>>();

            revoke_role(
                &mut deps,
                "signer".to_string(),
                roles.clone()
            ).unwrap();

            roles.iter().for_each(|role| { granted_roles.remove(role); });

            if granted_roles.is_empty() {
                // no roles, should fail
                assert!(RBAC_PERMISSIONS.load(deps.storage, "signer".to_string()).is_err());
            } else {
                let assigned_roles = RBAC_PERMISSIONS.load(deps.storage, "signer".to_string()).unwrap();

                assert_eq!(assigned_roles, granted_roles);
            }
        }



    }
}