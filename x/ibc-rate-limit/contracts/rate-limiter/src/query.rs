use crate::state::{path::Path, storage::{MESSAGE_QUEUE, RATE_LIMIT_TRACKERS, RBAC_PERMISSIONS}};
use cosmwasm_std::{Order::Ascending, StdError, Storage};
use cosmwasm_std::{to_binary, Binary, Deps, StdResult};

pub fn get_quotas(
    storage: &dyn Storage,
    channel_id: impl Into<String>,
    denom: impl Into<String>,
) -> StdResult<Binary> {
    let path = Path::new(channel_id, denom);
    to_binary(&RATE_LIMIT_TRACKERS.load(storage, path.into())?)
}

/// Returns all addresses which have been assigned one or more roles
pub fn get_role_owners(storage: &dyn Storage) -> StdResult<Binary> {
    to_binary(
        &RBAC_PERMISSIONS
            .keys(storage, None, None, Ascending)
            .filter_map(|key| key.ok())
            .collect::<Vec<_>>(),
    )
}

/// Returns all the roles that have been granted to `owner` (if any)
pub fn get_roles(storage: &dyn Storage, owner: String) -> StdResult<Binary> {
    to_binary(&RBAC_PERMISSIONS.load(storage, owner)?)
}

/// Returns the id's of all queued messages
pub fn get_message_ids(storage: &dyn Storage) -> StdResult<Binary> {
    to_binary(
        &MESSAGE_QUEUE
            .iter(storage)?
            .filter_map(|message| Some(message.ok()?.message_id))
            .collect::<Vec<_>>(),
    )
}

/// Searches MESSAGE_QUEUE for a message_id matching `id`
pub fn get_queued_message(storage: &dyn Storage, id: String) -> StdResult<Binary> {
    to_binary(&MESSAGE_QUEUE.iter(storage)?.find(|message| {
        let Ok(message) = message else {
            return false
        };
        message.message_id.eq(&id)
    }).ok_or_else(|| StdError::not_found(id))??)
}

#[cfg(test)]
mod test {
    use cosmwasm_std::{from_binary, testing::mock_dependencies, Timestamp};

    use crate::{
        msg::ExecuteMsg,
        state::rbac::{QueuedMessage, Roles},
    };

    use super::*;
    #[test]
    fn test_get_role_owners() {
        let mut deps = mock_dependencies();

        // test getting role owners when no owners exist
        let response = get_role_owners(deps.as_ref().storage).unwrap();
        let decoded: Vec<String> = from_binary(&response).unwrap();
        assert!(decoded.is_empty());

        // insert 1 role owner, and test getting role owners
        RBAC_PERMISSIONS
            .save(
                &mut deps.storage,
                "foobar".to_string(),
                &vec![Roles::SetTimelockDelay].into_iter().collect(),
            )
            .unwrap();
        let response = get_role_owners(deps.as_ref().storage).unwrap();
        let decoded: Vec<String> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 1);
        assert_eq!(decoded[0], "foobar");

        // insert another role owner and test getting role owners
        RBAC_PERMISSIONS
            .save(
                &mut deps.storage,
                "foobarbaz".to_string(),
                &vec![Roles::SetTimelockDelay].into_iter().collect(),
            )
            .unwrap();
        let response = get_role_owners(deps.as_ref().storage).unwrap();
        let decoded: Vec<String> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 2);
        assert_eq!(decoded[0], "foobar");
        assert_eq!(decoded[1], "foobarbaz");
    }

    #[test]
    fn test_get_roles() {
        let mut deps = mock_dependencies();

        // test retrieving roles for a missing role owner
        assert!(get_roles(deps.as_ref().storage, "foobar".to_string()).is_err());

        // assign roles and test retrieving roles owned by address
        RBAC_PERMISSIONS
            .save(
                &mut deps.storage,
                "foobar".to_string(),
                &vec![Roles::SetTimelockDelay].into_iter().collect(),
            )
            .unwrap();
        let response = get_roles(deps.as_ref().storage, "foobar".to_string()).unwrap();
        let decoded: Vec<Roles> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 1);
        assert_eq!(decoded[0], Roles::SetTimelockDelay);

        // add additional roles foobar, and test retrierval
        RBAC_PERMISSIONS
            .save(
                &mut deps.storage,
                "foobar".to_string(),
                &vec![Roles::SetTimelockDelay, Roles::EditPathQuota].into_iter().collect(),
            )
            .unwrap();
        let response = get_roles(deps.as_ref().storage, "foobar".to_string()).unwrap();
        let decoded: Vec<Roles> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 2);
        assert!(decoded.contains(&Roles::SetTimelockDelay));
        assert!(decoded.contains(&Roles::EditPathQuota));
    }

    #[test]
    fn test_get_messageids() {
        let mut deps = mock_dependencies();
        let response = get_message_ids(deps.as_ref().storage).unwrap();
        let decoded: Vec<String> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 0);
        
        MESSAGE_QUEUE
            .push_back(
                &mut deps.storage,
                &QueuedMessage {
                    message_id: "prop-1".to_string(),
                    message: ExecuteMsg::ProcessMessages { count: Some(1),message_ids: None },
                    submitted_at: Timestamp::default(),
                    timelock_delay: 0,
                },
            )
            .unwrap();
        MESSAGE_QUEUE
            .push_back(
                &mut deps.storage,
                &QueuedMessage {
                    message_id: "prop-2".to_string(),
                    message: ExecuteMsg::ProcessMessages { count: Some(1),message_ids: None },
                    submitted_at: Timestamp::default(),
                    timelock_delay: 0,
                },
            )
            .unwrap();
        let response = get_message_ids(deps.as_ref().storage).unwrap();
        let decoded: Vec<String> = from_binary(&response).unwrap();
        assert_eq!(decoded.len(), 2);
        assert_eq!(decoded[0], "prop-1");
        assert_eq!(decoded[1], "prop-2");
    }
}
