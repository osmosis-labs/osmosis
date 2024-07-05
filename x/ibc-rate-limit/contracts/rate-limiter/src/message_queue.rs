use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, Storage};

use crate::{
    error::ContractError,
    msg::ExecuteMsg,
    state::{
        rbac::QueuedMessage,
        storage::{MESSAGE_QUEUE, TIMELOCK_DELAY},
    },
};

/// Used to iterate over the message queue and process any messages that have passed the time lock delay.
///
/// If count is a non-zero value, we process no more than `count` message. This can be used to limit the number
/// of message processed in a single transaction to avoid running into OOG (out of gas) errors.
///
/// Because we iterate over the queue by popping items from the front, multiple transactions can be issued
/// in sequence to iterate over the queue
pub fn process_message_queue(
    deps: &mut DepsMut,
    env: &Env,
    count: Option<u64>,
    message_ids: Option<Vec<String>>,
) -> Result<Response, ContractError> {
    let mut response = Response::new();

    if let Some(message_ids) = message_ids {
        let messages = pop_messages(deps.storage, &message_ids)?;

        for message in messages {
            let message_id = message.message_id.clone();
            if let Err(err) = try_process_message(deps, env, message) {
                response = response.add_attribute(message_id, err.to_string());
            } else {
                response = response.add_attribute(message_id, "ok");
            }
        }
        return Ok(response);
    }

    let Some(count) = count else {
        return Err(ContractError::InvalidParameters(
            "both count and message_ids are None".to_string(),
        ));
    };

    let queue_len = MESSAGE_QUEUE.len(deps.storage)? as usize;

    for idx in 0..queue_len {
        if idx + 1 > count as usize {
            break;
        }
        if let Some(message) = MESSAGE_QUEUE.pop_front(deps.storage)? {
            let message_id = message.message_id.clone();
            if let Err(err) = try_process_message(deps, env, message) {
                response = response.add_attribute(message_id, err.to_string());
            } else {
                response = response.add_attribute(message_id, "ok");
            }
        }
    }
    Ok(response)
}

/// Given a message to execute, insert into the message queued with execution delayed by the timelock that is applied to the sender of the message
///
/// Returns the id of the queued message
pub fn queue_message(
    deps: &mut DepsMut,
    env: Env,
    msg: ExecuteMsg,
    info: MessageInfo,
) -> Result<String, ContractError> {
    let timelock_delay = TIMELOCK_DELAY.load(deps.storage, info.sender.to_string())?;
    let message_id = format!("{}_{}", env.block.height, env.transaction.unwrap().index);
    MESSAGE_QUEUE.push_back(
        deps.storage,
        &QueuedMessage {
            message_id: message_id.clone(),
            message: msg,
            timelock_delay,
            submitted_at: env.block.time,
        },
    )?;
    Ok(message_id)
}

/// Check to see if the message sender has a non-zero timelock delay configured
pub fn must_queue_message(deps: &mut DepsMut, info: &MessageInfo) -> bool {
    // if a non zero value is set, then it means a timelock delay is required
    TIMELOCK_DELAY
        .load(deps.storage, info.sender.to_string())
        .unwrap_or(0)
        > 0
}

/// Removes a message from the message queue if it matches message_id
pub fn remove_message(deps: &mut DepsMut, message_id: String) -> Result<(), ContractError> {
    pop_messages(deps.storage, &[message_id])?;
    Ok(())
}

fn pop_messages(
    storage: &mut dyn Storage,
    message_ids: &[String],
) -> Result<Vec<QueuedMessage>, ContractError> {
    let queue_len = MESSAGE_QUEUE.len(storage)? as usize;
    let mut messages = Vec::with_capacity(message_ids.len());

    for _ in 0..queue_len {
        if let Some(message) = MESSAGE_QUEUE.pop_front(storage)? {
            if message_ids.contains(&message.message_id) {
                messages.push(message);
            } else {
                // reinsert
                MESSAGE_QUEUE.push_back(storage, &message)?;
            }
        }
    }
    Ok(messages)
}

// attempts to process the message if the timelock has passed, otherwise reinsert into queue
//
// returns the id of the message
fn try_process_message(
    deps: &mut DepsMut,
    env: &Env,
    message: QueuedMessage,
) -> Result<(), ContractError> {
    // compute the minimum time at which the message is unlocked
    let min_unlock = message
        .submitted_at
        .plus_seconds(message.timelock_delay * 60 * 60);

    // check to see if the timelock delay has passed, which we need to first convert from hours into seconds
    if env.block.time.ge(&min_unlock) {
        crate::contract::match_execute(deps, &env, message.message)?;
    } else {
        MESSAGE_QUEUE.push_back(deps.storage, &message)?;
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{
        from_binary,
        testing::{mock_dependencies, mock_env},
        Addr, MemoryStorage, Timestamp, TransactionInfo,
    };

    use crate::{msg::QuotaMsg, query::get_queued_message, rbac::set_timelock_delay};

    use super::*;

    #[test]
    fn test_must_queue_message() {
        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();
        let foobar_info = MessageInfo {
            sender: Addr::unchecked("foobar"),
            funds: vec![],
        };
        let foobarbaz_info = MessageInfo {
            sender: Addr::unchecked("foobarbaz"),
            funds: vec![],
        };

        TIMELOCK_DELAY
            .save(deps.storage, "foobar".to_string(), &1)
            .unwrap();

        assert!(must_queue_message(&mut deps, &foobar_info));
        assert!(!must_queue_message(&mut deps, &foobarbaz_info));
    }

    #[test]
    fn test_queue_message() {
        let mut env = mock_env();
        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();

        let foobar_info = MessageInfo {
            sender: Addr::unchecked("foobar"),
            funds: vec![],
        };
        let foobarbaz_info = MessageInfo {
            sender: Addr::unchecked("foobarbaz"),
            funds: vec![],
        };
        let foobar_test_msg = ExecuteMsg::AddPath {
            channel_id: "channel".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg {
                name: "quota".to_string(),
                duration: 5,
                send_recv: (10, 10),
            }],
        };
        let foobarbaz_test_msg = ExecuteMsg::SetTimelockDelay {
            signer: "gov".to_string(),
            hours: 5,
        };
        set_timelock_delay(&mut deps, "foobar".to_string(), 10).unwrap();
        set_timelock_delay(&mut deps, "foobarbaz".to_string(), 1).unwrap();
        let foobar_message_id = {
            let mut env = env.clone();
            env.transaction = Some(TransactionInfo { index: 1 });
            queue_message(
                &mut deps,
                env.clone(),
                foobar_test_msg.clone(),
                foobar_info.clone(),
            )
            .unwrap()
        };
        let foobarbaz_message_id = {
            let mut env = env.clone();
            env.transaction = Some(TransactionInfo { index: 2 });
            queue_message(
                &mut deps,
                env.clone(),
                foobarbaz_test_msg.clone(),
                foobarbaz_info.clone(),
            )
            .unwrap()
        };
        // get foobar's queued message, and validate the type is as expected + timelock delays
        let msg = get_queued_message(deps.storage, foobar_message_id.clone()).unwrap();
        let msg: QueuedMessage = from_binary(&msg).unwrap();
        assert_eq!(msg.timelock_delay, 10);
        assert_eq!(msg.message, foobar_test_msg);

        // get foobarbaz's queued message, and validate the type is as expected + timelock delays
        let msg = get_queued_message(deps.storage, foobarbaz_message_id.clone()).unwrap();
        let msg: QueuedMessage = from_binary(&msg).unwrap();
        assert_eq!(msg.timelock_delay, 1);
        assert_eq!(msg.message, foobarbaz_test_msg);
    }

    #[test]
    fn test_process_message_queue_basic() {
        // basic test which simply iterates over the message queues
        // does include tests with some unlocked items vs some locked items

        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();
        let mut env = mock_env();
        create_n_messages(&mut deps, 10, &mut |_i: u64| Timestamp::default());
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 10);

        process_message_queue(&mut deps, &env.clone(), Some(1), None).unwrap();
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 9);

        process_message_queue(&mut deps, &env.clone(), Some(0), None).unwrap();
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 9);

        process_message_queue(&mut deps, &env.clone(), Some(5), None).unwrap();
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 4);

        process_message_queue(&mut deps, &env.clone(), Some(10), None).unwrap();
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 0);

        create_n_messages(&mut deps, 10, &mut |_i: u64| Timestamp::default());

        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 10);

        let message_ids = MESSAGE_QUEUE
            .iter(deps.storage)
            .unwrap()
            .filter_map(|msg| Some(msg.ok()?.message_id))
            .collect::<Vec<_>>();

        // get the first 4 message ids
        let msg_ids = message_ids[0..4].to_vec();
        process_message_queue(&mut deps, &env.clone(), None, Some(msg_ids)).unwrap();
        // should be 6 messages left
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 6);

        // get the remaining messages
        let msg_ids = message_ids[4..].to_vec();
        process_message_queue(&mut deps, &env.clone(), None, Some(msg_ids)).unwrap();
        // should be 0 messages left
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 0);
    }

    #[test]
    fn test_process_message_queue_complete() {
        // complete message queues testing, including some locked vs unlocked
        // as well as validating execution

        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();
        let mut env = mock_env();

        // starting time for tests, may 20th 12:32am PST
        let time = Timestamp::from_seconds(1716190293);
        env.block.time = time;

        create_n_messages(&mut deps, 10, &mut |i: u64| {
            // increment time by 1 hour * i
            time.plus_seconds(3600 * i)
        });

        // no messages should be processed as not enough time has passed
        process_message_queue(&mut deps, &env.clone(), Some(10), None).unwrap();

        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 10);

        // increase time by 24 hours
        env.block.time = env.block.time.plus_seconds(3600 * 24);

        // one message should be processed
        process_message_queue(&mut deps, &env.clone(), Some(10), None).unwrap();

        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 9);

        // signer should have a timelock delay of 1 hour
        assert_eq!(
            TIMELOCK_DELAY
                .load(deps.storage, "signer".to_string())
                .unwrap(),
            1
        );

        // advance time by 2 hours
        env.block.time = env.block.time.plus_seconds(3600 * 2);

        // 2 messages should be processed,
        process_message_queue(&mut deps, &env.clone(), Some(10), None).unwrap();

        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 7);

        //signer should have a timelock delay of 3 hours
        assert_eq!(
            TIMELOCK_DELAY
                .load(deps.storage, "signer".to_string())
                .unwrap(),
            3
        );

        // advance time by 24 hours
        env.block.time = env.block.time.plus_seconds(3600 * 24);

        // all messages should be processed
        process_message_queue(&mut deps, &env.clone(), Some(10), None).unwrap();

        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 0);

        // signer should have a delay of 10 hours
        assert_eq!(
            TIMELOCK_DELAY
                .load(deps.storage, "signer".to_string())
                .unwrap(),
            10
        );

        create_n_messages(&mut deps, 1, &mut |i: u64| {
            // increment time by 1 hour * i
            time.plus_seconds(3600 * i)
        });

        let message_ids = MESSAGE_QUEUE
            .iter(deps.storage)
            .unwrap()
            .filter_map(|msg| Some(msg.ok()?.message_id))
            .collect::<Vec<_>>();

        process_message_queue(&mut deps, &env.clone(), None, Some(message_ids)).unwrap();
        // signer should have a delay of 1 hours
        assert_eq!(
            TIMELOCK_DELAY
                .load(deps.storage, "signer".to_string())
                .unwrap(),
            1
        );
    }

    #[test]
    #[should_panic(expected = "both count and message_ids are None")]
    fn test_process_message_queue_invalid_parameters() {
        let mut deps = mock_dependencies();
        let mut deps = deps.as_mut();
        let mut env = mock_env();
        create_n_messages(&mut deps, 10, &mut |_i: u64| Timestamp::default());
        assert_eq!(MESSAGE_QUEUE.len(deps.storage).unwrap(), 10);

        process_message_queue(&mut deps, &env.clone(), None, None).unwrap();
    }

    // helper function which inserts N messages into the message queue
    // message types inserted are of ExecuteMsg::SetTimelockDelay
    fn create_n_messages(deps: &mut DepsMut, n: usize, ts: &mut dyn FnMut(u64) -> Timestamp) {
        for i in 0..n {
            MESSAGE_QUEUE
                .push_back(
                    deps.storage,
                    &QueuedMessage {
                        message: ExecuteMsg::SetTimelockDelay {
                            signer: "signer".to_string(),
                            hours: i as u64 + 1,
                        },
                        submitted_at: ts(i as u64),
                        timelock_delay: 24,
                        message_id: format!("prop-{i}"),
                    },
                )
                .unwrap();
        }
    }
}
