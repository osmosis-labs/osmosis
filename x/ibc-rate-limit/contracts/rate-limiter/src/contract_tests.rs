#![cfg(test)]

use crate::packet::Packet;
use crate::state::rbac::Roles;
use crate::{contract::*, test_msg_recv, test_msg_send, ContractError};
use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
use cosmwasm_std::{from_binary, Addr, Attribute, MessageInfo, Uint256};
use cw2::set_contract_version;

use crate::blocking::{check_restricted_denoms, effective_restriction, NO_CHANNEL_ALLOWED};
use crate::helpers::tests::verify_query_response;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, PathMsg, QueryMsg, QuotaMsg, SudoMsg};
use crate::packet::hash_denom;
use crate::state::flow::{tests::RESET_TIME_WEEKLY, FlowType};
use crate::state::rate_limit::RateLimit;
use crate::state::storage::{
    ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM, GOVMODULE, IBCMODULE, PENDING_SENDS,
    RATE_LIMIT_TRACKERS, RBAC_PERMISSIONS,
};
const IBC_ADDR: &str = "osmo1vz5e6tzdjlzy2f7pjvx0ecv96h8r4m2y92thdm";
const GOV_ADDR: &str = "osmo1tzz5zf2u68t00un2j4lrrnkt2ztd46kfzfp58r";

#[test] // Tests we ccan instantiate the contract and that the owners are set correctly
fn proper_instantiation() {
    let mut deps = mock_dependencies();

    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![],
    };
    let info = mock_info(IBC_ADDR, &[]);

    // we can just call .unwrap() to assert this was a success
    let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
    assert_eq!(0, res.messages.len());

    // The ibc and gov modules are properly stored
    assert_eq!(IBCMODULE.load(deps.as_ref().storage).unwrap(), IBC_ADDR);
    assert_eq!(GOVMODULE.load(deps.as_ref().storage).unwrap(), GOV_ADDR);

    let permissions = RBAC_PERMISSIONS
        .load(&mut deps.storage, GOV_ADDR.to_string())
        .unwrap();
    for permission in Roles::all_roles() {
        assert!(permissions.contains(&permission));
    }
}

#[test] // Tests that when a packet is transferred, the peropper allowance is consummed
fn consume_allowance() {
    let mut deps = mock_dependencies();

    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &[]);
    let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom") ,
        channel_value: 3_300_u32.into(),
        funds: 300_u32.into()
    );
    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();

    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "300");

    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_300_u32.into(),
        funds: 300_u32.into()
    );
    let err = sudo(deps.as_mut(), mock_env(), msg).unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));
}

#[test] // Tests that the balance of send and receive is maintained (i.e: recives are sustracted from the send allowance and sends from the receives)
fn symetric_flows_dont_consume_allowance() {
    let mut deps = mock_dependencies();

    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &[]);
    let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

    let send_msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_300_u32.into(),
        funds: 300_u32.into()
    );
    let recv_msg = test_msg_recv!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 300_u32.into()
    );

    let res = sudo(deps.as_mut(), mock_env(), send_msg.clone()).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "300");

    let res = sudo(deps.as_mut(), mock_env(), recv_msg.clone()).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "0");

    // We can still use the path. Even if we have sent more than the
    // allowance through the path (900 > 3000*.1), the current "balance"
    // of inflow vs outflow is still lower than the path's capacity/quota
    let res = sudo(deps.as_mut(), mock_env(), recv_msg.clone()).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "300");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "0");

    let err = sudo(deps.as_mut(), mock_env(), recv_msg.clone()).unwrap_err();

    assert!(matches!(err, ContractError::RateLimitExceded { .. }));
}

#[test] // Tests that we can have different quotas for send and receive. In this test we use 4% send and 1% receive
fn asymetric_quotas() {
    let mut deps = mock_dependencies();

    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 4, 1);
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &[]);
    let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

    // Sending 2%
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_060_u32.into(),
        funds: 60_u32.into()
    );
    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "60");

    // Sending 2% more. Allowed, as sending has a 4% allowance
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_060_u32.into(),
        funds: 60_u32.into()
    );

    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "120");

    // Receiving 1% should still work. 4% *sent* through the path, but we can still receive.
    let recv_msg = test_msg_recv!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 30_u32.into()
    );
    let res = sudo(deps.as_mut(), mock_env(), recv_msg).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "90");

    // Sending 2%. Should fail. In balance, we've sent 4% and received 1%, so only 1% left to send.
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_060_u32.into(),
        funds: 60_u32.into()
    );
    let err = sudo(deps.as_mut(), mock_env(), msg.clone()).unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));

    // Sending 1%: Allowed; because sending has a 4% allowance. We've sent 4% already, but received 1%, so there's send cappacity again
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_060_u32.into(),
        funds: 30_u32.into()
    );
    let res = sudo(deps.as_mut(), mock_env(), msg.clone()).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "120");
}

#[test] // Tests we can get the current state of the trackers
fn query_state() {
    let mut deps = mock_dependencies();

    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &[]);
    let env = mock_env();
    let _res = instantiate(deps.as_mut(), env.clone(), info, msg).unwrap();

    let query_msg = QueryMsg::GetQuotas {
        channel_id: "any".to_string(),
        denom: "denom".to_string(),
    };

    let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
    let value: Vec<RateLimit> = from_binary(&res).unwrap();
    assert_eq!(value[0].quota.name, "weekly");
    assert_eq!(value[0].quota.max_percentage_send, 10);
    assert_eq!(value[0].quota.max_percentage_recv, 10);
    assert_eq!(value[0].quota.duration, RESET_TIME_WEEKLY);
    assert_eq!(value[0].flow.inflow, Uint256::from(0_u32));
    assert_eq!(value[0].flow.outflow, Uint256::from(0_u32));
    assert_eq!(
        value[0].flow.period_end,
        env.block.time.plus_seconds(RESET_TIME_WEEKLY)
    );

    let send_msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_300_u32.into(),
        funds: 300_u32.into()
    );
    sudo(deps.as_mut(), mock_env(), send_msg.clone()).unwrap();

    let recv_msg = test_msg_recv!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 30_u32.into()
    );
    sudo(deps.as_mut(), mock_env(), recv_msg.clone()).unwrap();

    // Query
    let res = query(deps.as_ref(), mock_env(), query_msg.clone()).unwrap();
    let value: Vec<RateLimit> = from_binary(&res).unwrap();
    verify_query_response(
        &value[0],
        "weekly",
        (10, 10),
        RESET_TIME_WEEKLY,
        30_u32.into(),
        300_u32.into(),
        env.block.time.plus_seconds(RESET_TIME_WEEKLY),
    );
}

#[test] // Tests quota percentages are between [0,100]
fn bad_quotas() {
    let mut deps = mock_dependencies();

    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg {
                name: "bad_quota".to_string(),
                duration: 200,
                send_recv: (5000, 101),
            }],
        }],
    };
    let info = mock_info(IBC_ADDR, &[]);

    let env = mock_env();
    instantiate(deps.as_mut(), env.clone(), info, msg).unwrap();

    // If a quota is higher than 100%, we set it to 100%
    let query_msg = QueryMsg::GetQuotas {
        channel_id: "any".to_string(),
        denom: "denom".to_string(),
    };
    let res = query(deps.as_ref(), env.clone(), query_msg).unwrap();
    let value: Vec<RateLimit> = from_binary(&res).unwrap();
    verify_query_response(
        &value[0],
        "bad_quota",
        (100, 100),
        200,
        0_u32.into(),
        0_u32.into(),
        env.block.time.plus_seconds(200),
    );
}

fn instantiate_any_denom(
    deps: cosmwasm_std::DepsMut,
    quotas: Vec<QuotaMsg>,
) -> Result<cosmwasm_std::Response, ContractError> {
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas,
        }],
    };
    instantiate(deps, mock_env(), mock_info(GOV_ADDR, &[]), msg)
}

fn send_with_sequence(sequence: u64, funds: u32) -> SudoMsg {
    let mut packet = Packet::mock(
        "channel".to_string(),
        "channel".to_string(),
        "denom".to_string(),
        funds.into(),
    );
    packet.sequence = sequence;
    SudoMsg::SendPacket {
        packet,
        channel_value_mock: Some(3_300_u32.into()),
    }
}

fn record_with_sequence(sequence: u64, funds: u32) -> SudoMsg {
    let mut packet = Packet::mock(
        "channel".to_string(),
        "channel".to_string(),
        "denom".to_string(),
        funds.into(),
    );
    packet.sequence = sequence;
    SudoMsg::RecordSend { packet }
}

fn undo_with_sequence(sequence: u64, funds: u32) -> SudoMsg {
    let mut packet = Packet::mock(
        "channel".to_string(),
        "channel".to_string(),
        "denom".to_string(),
        funds.into(),
    );
    packet.sequence = sequence;
    SudoMsg::UndoSend { packet }
}

fn attr(res: &cosmwasm_std::Response, key: &str) -> String {
    res.attributes
        .iter()
        .find(|a| a.key == key)
        .map(|a| a.value.clone())
        .unwrap_or_default()
}

fn any_denom_outflow(storage: &dyn cosmwasm_std::Storage) -> Uint256 {
    RATE_LIMIT_TRACKERS
        .load(storage, ("any".to_string(), "denom".to_string()))
        .unwrap()
        .first()
        .unwrap()
        .flow
        .outflow
}

#[test] // A failed send is refunded while the quota is still in the window the send was counted in
fn undo_send_refunds_within_the_same_window() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    sudo(deps.as_mut(), mock_env(), send_with_sequence(7, 300)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(7, 300)).unwrap();
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(300_u32));
    assert!(PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));
    let before = RATE_LIMIT_TRACKERS
        .load(&deps.storage, ("any".to_string(), "denom".to_string()))
        .unwrap();

    let res = sudo(deps.as_mut(), mock_env(), undo_with_sequence(7, 300)).unwrap();
    assert_eq!(attr(&res, "refunded"), "1");
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(0_u32));
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));

    // The refund touches nothing but the flow
    let after = RATE_LIMIT_TRACKERS
        .load(&deps.storage, ("any".to_string(), "denom".to_string()))
        .unwrap();
    assert_eq!(after[0].flow.period_end, before[0].flow.period_end);
    assert_eq!(after[0].quota.channel_value, before[0].quota.channel_value);

    // and a second undo of the same packet is a no-op
    let res = sudo(deps.as_mut(), mock_env(), undo_with_sequence(7, 300)).unwrap();
    assert_eq!(attr(&res, "refunded"), "0");
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(0_u32));
}

#[test] // A send the chain did not record (sequence 0, as before the reordering upgrade) keeps its cost
fn undo_send_without_record_keeps_the_cost() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    sudo(deps.as_mut(), mock_env(), send_with_sequence(0, 300)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(0, 300)).unwrap();
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 0)));

    let res = sudo(deps.as_mut(), mock_env(), undo_with_sequence(0, 300)).unwrap();
    assert_eq!(attr(&res, "refunded"), "0");
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(300_u32));
    let err = sudo(deps.as_mut(), mock_env(), send_with_sequence(0, 300)).unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));
}

#[test] // The attack the refund used to allow: a send from the previous window must not be refunded into the next
fn undo_send_after_window_reset_does_not_refund() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    // Window N: the full 300 out
    sudo(deps.as_mut(), mock_env(), send_with_sequence(7, 300)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(7, 300)).unwrap();

    // Window N+1: any transfer resets the flow
    let mut later = mock_env();
    later.block.time = later.block.time.plus_seconds(RESET_TIME_WEEKLY + 1);
    sudo(deps.as_mut(), later.clone(), send_with_sequence(8, 1)).unwrap();
    sudo(deps.as_mut(), later.clone(), record_with_sequence(8, 1)).unwrap();
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(1_u32));

    // The error ack for the window-N send arrives now: no refund, record dropped
    let res = sudo(deps.as_mut(), later.clone(), undo_with_sequence(7, 300)).unwrap();
    assert_eq!(attr(&res, "refunded"), "0");
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(1_u32));
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));

    // Window N+1 still has exactly its own allowance: 329 more fits (330 total), 330 more does not
    sudo(deps.as_mut(), later.clone(), send_with_sequence(9, 329)).unwrap();
    sudo(deps.as_mut(), later.clone(), record_with_sequence(9, 329)).unwrap();
    let err = sudo(deps.as_mut(), later, send_with_sequence(10, 1)).unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));
}

#[test] // A successful acknowledgement settles the record and the send stays counted
fn confirm_send_settles_the_record() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    sudo(deps.as_mut(), mock_env(), send_with_sequence(7, 300)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(7, 300)).unwrap();
    let mut packet = Packet::mock(
        "channel".to_string(),
        "channel".to_string(),
        "denom".to_string(),
        300_u32.into(),
    );
    packet.sequence = 7;
    let res = sudo(deps.as_mut(), mock_env(), SudoMsg::ConfirmSend { packet }).unwrap();
    assert_eq!(attr(&res, "settled"), "true");
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));

    // A late undo for the same packet cannot refund it any more
    sudo(deps.as_mut(), mock_env(), undo_with_sequence(7, 300)).unwrap();
    assert_eq!(any_denom_outflow(&deps.storage), Uint256::from(300_u32));
}

#[test] // Stale records can be purged by anyone once their windows have ended, and not before
fn purge_stale_sends_is_bounded_and_permissionless() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();
    sudo(deps.as_mut(), mock_env(), send_with_sequence(7, 100)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(7, 100)).unwrap();
    sudo(deps.as_mut(), mock_env(), send_with_sequence(8, 100)).unwrap();
    sudo(deps.as_mut(), mock_env(), record_with_sequence(8, 100)).unwrap();

    let purge = ExecuteMsg::PurgeStaleSends {
        start_after: None,
        limit: 10,
    };

    // Windows still open: nothing to purge
    let res = execute(
        deps.as_mut(),
        mock_env(),
        mock_info("anyone", &[]),
        purge.clone(),
    )
    .unwrap();
    assert_eq!(attr(&res, "purged"), "0");
    assert!(PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));

    // Windows over: both records go, in bounded steps
    let mut later = mock_env();
    later.block.time = later.block.time.plus_seconds(RESET_TIME_WEEKLY + 1);
    let one = ExecuteMsg::PurgeStaleSends {
        start_after: None,
        limit: 1,
    };
    let res = execute(deps.as_mut(), later.clone(), mock_info("anyone", &[]), one).unwrap();
    assert_eq!(attr(&res, "purged"), "1");
    assert_eq!(attr(&res, "last_key"), "channel/7");
    let rest = ExecuteMsg::PurgeStaleSends {
        start_after: Some(("channel".to_string(), 7)),
        limit: 10,
    };
    let res = execute(deps.as_mut(), later, mock_info("anyone", &[]), rest).unwrap();
    assert_eq!(attr(&res, "purged"), "1");
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 7)));
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 8)));
}

#[test] // Residue from older versions is removed by anyone, in bounded batches, never by the migration
fn purge_empty_paths_is_bounded_and_permissionless() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    // Residue as written by versions before 0.2.0
    let residue = [
        ("any", "ibc/AAAA"),
        ("channel-0", "ibc/AAAA"),
        ("channel-1", "uosmo"),
    ];
    for (channel, denom) in residue {
        RATE_LIMIT_TRACKERS
            .save(
                deps.as_mut().storage,
                (channel.to_string(), denom.to_string()),
                &vec![],
            )
            .unwrap();
    }

    // Migration leaves it alone
    let res = migrate(deps.as_mut(), mock_env(), MigrateMsg {}).unwrap();
    assert_eq!(attr(&res, "purged_empty_paths"), "");
    for (channel, denom) in residue {
        let key = (channel.to_string(), denom.to_string());
        assert!(RATE_LIMIT_TRACKERS.has(&deps.storage, key));
    }

    // First batch of two scans any/denom (kept) and any/ibc/AAAA (purged)
    let first = ExecuteMsg::PurgeEmptyPaths {
        start_after: None,
        limit: 2,
    };
    let res = execute(deps.as_mut(), mock_env(), mock_info("anyone", &[]), first).unwrap();
    assert_eq!(attr(&res, "scanned"), "2");
    assert_eq!(attr(&res, "purged"), "1");
    assert_eq!(attr(&res, "last_key"), "any/ibc/AAAA");

    // Continue from the cursor: the remaining two go
    let rest = ExecuteMsg::PurgeEmptyPaths {
        start_after: Some(("any".to_string(), "ibc/AAAA".to_string())),
        limit: 10,
    };
    let res = execute(deps.as_mut(), mock_env(), mock_info("anyone", &[]), rest).unwrap();
    assert_eq!(attr(&res, "scanned"), "2");
    assert_eq!(attr(&res, "purged"), "2");

    for (channel, denom) in residue {
        let key = (channel.to_string(), denom.to_string());
        assert!(!RATE_LIMIT_TRACKERS.has(&deps.storage, key));
    }
    let kept = RATE_LIMIT_TRACKERS
        .load(&deps.storage, ("any".to_string(), "denom".to_string()))
        .unwrap();
    assert_eq!(kept.len(), 1);

    // A zero limit is rejected rather than silently doing nothing
    let zero = ExecuteMsg::PurgeEmptyPaths {
        start_after: None,
        limit: 0,
    };
    let err = execute(deps.as_mut(), mock_env(), mock_info("anyone", &[]), zero).unwrap_err();
    assert!(matches!(err, ContractError::InvalidParameters(_)));
}

#[test] // RecordSend writes nothing for an unquoted path or an unknown sequence
fn record_send_records_only_charged_sends() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    // Sequence 0 is what a chain that has not learned the sequence passes
    let res = sudo(deps.as_mut(), mock_env(), record_with_sequence(0, 100)).unwrap();
    assert_eq!(attr(&res, "recorded"), "false");
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 0)));

    // A denom with no quota anywhere charged nothing, so there is nothing to record
    let mut packet = Packet::mock(
        "channel".to_string(),
        "channel".to_string(),
        "other".to_string(),
        100_u32.into(),
    );
    packet.sequence = 3;
    let res = sudo(deps.as_mut(), mock_env(), SudoMsg::RecordSend { packet }).unwrap();
    assert_eq!(attr(&res, "recorded"), "false");
    assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 3)));

    // A charged send is recorded with the quota's current window
    sudo(deps.as_mut(), mock_env(), send_with_sequence(4, 100)).unwrap();
    let res = sudo(deps.as_mut(), mock_env(), record_with_sequence(4, 100)).unwrap();
    assert_eq!(attr(&res, "recorded"), "true");
    let record = PENDING_SENDS
        .load(&deps.storage, ("channel".to_string(), 4))
        .unwrap();
    assert_eq!(record.funds, Uint256::from(100_u32));
    assert_eq!(record.any_windows.len(), 1);
    assert_eq!(record.any_windows[0].0, "weekly");
    assert!(record.channel_windows.is_empty());
}

#[test] // Each new record evicts the oldest stale records on its channel, so the map stays bounded by itself
fn record_send_evicts_stale_records_on_the_channel() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    // Window N: two recorded sends whose acks never arrive
    for sequence in [7, 8] {
        sudo(deps.as_mut(), mock_env(), send_with_sequence(sequence, 10)).unwrap();
        sudo(
            deps.as_mut(),
            mock_env(),
            record_with_sequence(sequence, 10),
        )
        .unwrap();
    }
    // Still inside the window: a new record evicts nothing
    sudo(deps.as_mut(), mock_env(), send_with_sequence(9, 10)).unwrap();
    let res = sudo(deps.as_mut(), mock_env(), record_with_sequence(9, 10)).unwrap();
    assert_eq!(attr(&res, "evicted_stale"), "0");

    // Window N+1: the first recorded send clears the three stale records
    let mut later = mock_env();
    later.block.time = later.block.time.plus_seconds(RESET_TIME_WEEKLY + 1);
    sudo(deps.as_mut(), later.clone(), send_with_sequence(10, 10)).unwrap();
    let res = sudo(deps.as_mut(), later, record_with_sequence(10, 10)).unwrap();
    assert_eq!(attr(&res, "evicted_stale"), "3");
    for sequence in [7_u64, 8, 9] {
        assert!(!PENDING_SENDS.has(&deps.storage, ("channel".to_string(), sequence)));
    }
    assert!(PENDING_SENDS.has(&deps.storage, ("channel".to_string(), 10)));
}

#[test] // The housekeeping batch size is the contract's bound, not the caller's
fn purge_batch_limit_is_capped() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();
    let info = mock_info("anyone", &[]);

    let at_cap = ExecuteMsg::PurgeEmptyPaths {
        start_after: None,
        limit: crate::execute::MAX_PURGE_BATCH,
    };
    execute(deps.as_mut(), mock_env(), info.clone(), at_cap).unwrap();

    let over_cap = ExecuteMsg::PurgeEmptyPaths {
        start_after: None,
        limit: crate::execute::MAX_PURGE_BATCH + 1,
    };
    let err = execute(deps.as_mut(), mock_env(), info.clone(), over_cap).unwrap_err();
    assert!(matches!(err, ContractError::InvalidParameters(_)));

    let over_cap = ExecuteMsg::PurgeStaleSends {
        start_after: None,
        limit: crate::execute::MAX_PURGE_BATCH + 1,
    };
    let err = execute(deps.as_mut(), mock_env(), info, over_cap).unwrap_err();
    assert!(matches!(err, ContractError::InvalidParameters(_)));
}

#[test] // Migration moves packet-form restriction entries to their canonical key and drops empty ones
fn migrate_canonicalises_restrictions() {
    let mut deps = mock_dependencies();
    instantiate_any_denom(
        deps.as_mut(),
        vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
    )
    .unwrap();

    let moved_denom = "transfer/channel-6897/usat";
    let merged_denom = "transfer/channel-1/uatom";
    let conflicting_denom = "transfer/channel-2/ujuno";
    let entries: [(String, Vec<&str>); 6] = [
        // legacy entry with no canonical counterpart: moved
        (moved_denom.to_string(), vec!["channel-6897"]),
        // legacy empty entry: dropped
        ("transfer/channel-3/ustars".to_string(), vec![]),
        // legacy and canonical overlap: intersected under the canonical key
        (merged_denom.to_string(), vec!["channel-1", "channel-9"]),
        (hash_denom(merged_denom), vec!["channel-1", "channel-8"]),
        // legacy and canonical disjoint: collapsed into one entry allowing no channel
        (conflicting_denom.to_string(), vec!["channel-2"]),
        (hash_denom(conflicting_denom), vec!["channel-5"]),
    ];
    for (key, channels) in entries {
        let channels: Vec<String> = channels.into_iter().map(str::to_string).collect();
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .save(deps.as_mut().storage, key, &channels)
            .unwrap();
    }

    let res = migrate(deps.as_mut(), mock_env(), MigrateMsg {}).unwrap();
    assert_eq!(attr(&res, "restrictions_moved"), "2");
    assert_eq!(attr(&res, "restrictions_dropped"), "1");
    assert_eq!(attr(&res, "restrictions_conflicting"), "1");

    let storage = deps.as_ref().storage;
    assert!(!ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.has(storage, moved_denom.to_string()));
    assert_eq!(
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .load(storage, hash_denom(moved_denom))
            .unwrap(),
        vec!["channel-6897".to_string()]
    );
    let dropped = "transfer/channel-3/ustars".to_string();
    assert!(!ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.has(storage, dropped));
    assert!(!ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.has(storage, merged_denom.to_string()));
    assert_eq!(
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .load(storage, hash_denom(merged_denom))
            .unwrap(),
        vec!["channel-1".to_string()]
    );
    assert!(!ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.has(storage, conflicting_denom.to_string()));
    assert_eq!(
        effective_restriction(storage, conflicting_denom).unwrap(),
        Some(vec![NO_CHANNEL_ALLOWED.to_string()])
    );
    // which still blocks every channel either alias used to allow
    for channel in ["channel-2", "channel-5"] {
        let packet = Packet::mock(
            channel.to_string(),
            "dest".to_string(),
            conflicting_denom.to_string(),
            Uint256::from(1_u32),
        );
        let result = check_restricted_denoms(deps.as_ref(), &packet, &FlowType::Out);
        assert!(
            matches!(result, Err(ContractError::ChannelBlocked { .. })),
            "{channel}"
        );
    }

    // After the migration, management by hash acts on the whole restriction:
    // unsetting the collapsed conflict lifts it completely
    let unset = ExecuteMsg::UnsetDenomRestrictions {
        denom: hash_denom(conflicting_denom),
    };
    execute(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), unset).unwrap();
    assert_eq!(
        effective_restriction(deps.as_ref().storage, conflicting_denom).unwrap(),
        None
    );
    let packet = Packet::mock(
        "channel-2".to_string(),
        "dest".to_string(),
        conflicting_denom.to_string(),
        Uint256::from(1_u32),
    );
    assert!(check_restricted_denoms(deps.as_ref(), &packet, &FlowType::Out).is_ok());

    // and setting by hash replaces the moved entry for every spelling
    let set = ExecuteMsg::SetDenomRestrictions {
        denom: hash_denom(moved_denom),
        allowed_channels: vec!["channel-77".to_string()],
    };
    execute(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), set).unwrap();
    assert_eq!(
        effective_restriction(deps.as_ref().storage, moved_denom).unwrap(),
        Some(vec!["channel-77".to_string()])
    );
    let unset = ExecuteMsg::UnsetDenomRestrictions {
        denom: hash_denom(moved_denom),
    };
    execute(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), unset).unwrap();
    assert_eq!(
        effective_restriction(deps.as_ref().storage, moved_denom).unwrap(),
        None
    );
}

#[test] // Packets on paths without quotas must not leave empty tracker entries behind
fn unquoted_paths_leave_no_state() {
    let mut deps = mock_dependencies();
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: "denom".to_string(),
            quotas: vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10)],
        }],
    };
    instantiate(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), msg).unwrap();

    // Quota only on "any": the per-channel side must not be written
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_300_u32.into(),
        funds: 1_u32.into()
    );
    sudo(deps.as_mut(), mock_env(), msg).unwrap();
    let channel_denom = ("channel".to_string(), "denom".to_string());
    assert!(!RATE_LIMIT_TRACKERS.has(&deps.storage, channel_denom));

    // No quota anywhere: nothing is written for either side
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("other"),
        channel_value: 3_300_u32.into(),
        funds: 1_u32.into()
    );
    sudo(deps.as_mut(), mock_env(), msg).unwrap();
    let channel_other = ("channel".to_string(), "other".to_string());
    let any_other = ("any".to_string(), "other".to_string());
    assert!(!RATE_LIMIT_TRACKERS.has(&deps.storage, channel_other));
    assert!(!RATE_LIMIT_TRACKERS.has(&deps.storage, any_other));
}

#[test]
fn test_basic_message() {
    let json = r#"{"send_packet":{"packet":{"sequence":2,"source_port":"transfer","source_channel":"channel-0","destination_port":"transfer","destination_channel":"channel-0","data":{"denom":"stake","amount":"125000000000011250","sender":"osmo1dwtagd6xzl4eutwtyv6mewra627lkg3n3w26h6","receiver":"osmo1yvjkt8lnpxucjmspaj5ss4aa8562gx0a3rks8s"},"timeout_height":{"revision_height":100}}}}"#;
    let _parsed: SudoMsg = serde_json_wasm::from_str(json).unwrap();
    //println!("{parsed:?}");
}

#[test]
fn test_testnet_message() {
    let json = r#"{"send_packet":{"packet":{"sequence":4,"source_port":"transfer","source_channel":"channel-0","destination_port":"transfer","destination_channel":"channel-1491","data":{"denom":"uosmo","amount":"100","sender":"osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks","receiver":"osmo1c584m4lq25h83yp6ag8hh4htjr92d954vklzja"},"timeout_height":{},"timeout_timestamp":1668024637477293371}}}"#;
    let _parsed: SudoMsg = serde_json_wasm::from_str(json).unwrap();
    //println!("{parsed:?}");
}

#[test]
fn test_tokenfactory_message() {
    let json = r#"{"send_packet":{"packet":{"sequence":4,"source_port":"transfer","source_channel":"channel-0","destination_port":"transfer","destination_channel":"channel-1491","data":{"denom":"transfer/channel-0/factory/osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj/czar","amount":"100000000000000000","sender":"osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks","receiver":"osmo1c584m4lq25h83yp6ag8hh4htjr92d954vklzja"},"timeout_height":{},"timeout_timestamp":1668024476848430980}}}"#;
    let _parsed: SudoMsg = serde_json_wasm::from_str(json).unwrap();
    //println!("{parsed:?}");
}

#[test] // Tests we ccan instantiate the contract and that the owners are set correctly
fn proper_migrate_for_v0_1_0() {
    let mut deps = mock_dependencies();
    let env = mock_env();

    crate::contract::instantiate(
        deps.as_mut(),
        env,
        MessageInfo {
            sender: Addr::unchecked("osmo16tumts0kckpfp9fk7e3rnx9ahzn70dyyqfypgh"),
            funds: vec![],
        },
        InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            paths: vec![],
        },
    )
    .unwrap();

    // force set contract version to 0.1.0
    set_contract_version(deps.as_mut().storage, "crates.io:rate-limiter", "0.1.0").unwrap();

    // test that instantiate set the correct gov module address and RBAC permissions
    let permissions = RBAC_PERMISSIONS
        .load(&mut deps.storage, GOV_ADDR.to_string())
        .unwrap();
    for permission in Roles::all_roles() {
        assert!(permissions.contains(&permission));
    }
    assert_eq!(GOVMODULE.load(deps.as_ref().storage).unwrap(), GOV_ADDR);

    // revoke all roles from the gov contract, migration from 0.1.0 should re-asssign
    crate::rbac::revoke_role(&mut deps.as_mut(), GOV_ADDR.to_string(), Roles::all_roles()).unwrap();

    migrate(deps.as_mut(), mock_env(), MigrateMsg {}).unwrap();

    // ensure migration assigned all the roles
    let permissions = RBAC_PERMISSIONS
        .load(&mut deps.storage, GOV_ADDR.to_string())
        .unwrap();
    for permission in Roles::all_roles() {
        assert!(permissions.contains(&permission));
    }
}

// Regression tests for osmosis-labs/osmosis#9742.
//
// Injective's channel to Osmosis is channel-8, Osmosis' side is channel-122, and Stride's
// stATOM lives on Injective as transfer/channel-89/stuatom. "channel-8" is a string prefix
// of "channel-89", so a prefix check without the trailing slash treated the packet as an
// Osmosis-native token returning home, stripped nothing, and asked the chain for the supply
// of an empty denom. Every such packet was rejected as "rate limit exceeded".
const INJECTIVE_TO_OSMOSIS: &str = "channel-8";
const OSMOSIS_FROM_INJECTIVE: &str = "channel-122";
const STATOM_ON_INJECTIVE_TRACE: &str = "transfer/channel-89/stuatom";
// sha256("transfer/channel-122/transfer/channel-89/stuatom")
const STATOM_VIA_INJECTIVE_ON_OSMOSIS: &str =
    "ibc/F65724D2AE4A14F5BC149FC12C984D53D2307D95EC57BBA1CE52F94EB670EF60";

fn statom_via_injective_recv(channel_value_mock: Option<Uint256>, funds: u128) -> SudoMsg {
    SudoMsg::RecvPacket {
        packet: Packet::mock(
            INJECTIVE_TO_OSMOSIS.to_string(),
            OSMOSIS_FROM_INJECTIVE.to_string(),
            STATOM_ON_INJECTIVE_TRACE.to_string(),
            funds.into(),
        ),
        channel_value_mock,
    }
}

#[test] // A colliding-prefix packet on a path with no quota passes, and the contract never asks the chain for a supply
fn recv_colliding_channel_prefix_without_quota_is_allowed() {
    let mut deps = mock_dependencies();
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![],
    };
    instantiate(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), msg).unwrap();

    // No channel value mock: the mock querier cannot answer a SupplyOf query, so this only
    // succeeds if the contract checks for quotas before querying the channel value.
    let res = sudo(
        deps.as_mut(),
        mock_env(),
        statom_via_injective_recv(None, 1_000_000),
    )
    .unwrap();

    let denom = res.attributes.iter().find(|a| a.key == "denom").unwrap();
    assert_eq!(denom.value, STATOM_VIA_INJECTIVE_ON_OSMOSIS);
    assert!(res
        .attributes
        .iter()
        .any(|a| a.key == "quota" && a.value == "none"));
}

#[test] // A colliding-prefix packet is accounted against the quota of the denom it actually becomes on Osmosis
fn recv_colliding_channel_prefix_consumes_foreign_denom_quota() {
    let mut deps = mock_dependencies();
    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);
    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths: vec![PathMsg {
            channel_id: "any".to_string(),
            denom: STATOM_VIA_INJECTIVE_ON_OSMOSIS.to_string(),
            quotas: vec![quota],
        }],
    };
    instantiate(deps.as_mut(), mock_env(), mock_info(GOV_ADDR, &[]), msg).unwrap();

    // 10% of a 1000 supply is 100. The first 100 fits, the next 100 does not.
    let res = sudo(
        deps.as_mut(),
        mock_env(),
        statom_via_injective_recv(Some(1_000_u32.into()), 100),
    )
    .unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "100");

    let err = sudo(
        deps.as_mut(),
        mock_env(),
        statom_via_injective_recv(Some(1_000_u32.into()), 100),
    )
    .unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));
}
