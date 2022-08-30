#![cfg(test)]

use crate::{contract::*, ContractError};
use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
use cosmwasm_std::{from_binary, Addr, Attribute};

use crate::helpers::tests::verify_query_response;
use crate::msg::{InstantiateMsg, PathMsg, QueryMsg, QuotaMsg, SudoMsg};
use crate::state::tests::RESET_TIME_WEEKLY;
use crate::state::{RateLimit, GOVMODULE, IBCMODULE, RATE_LIMIT_TRACKERS};

const IBC_ADDR: &str = "IBC_MODULE";
const GOV_ADDR: &str = "GOV_MODULE";

#[test] // Tests we ccan instantiate the contract and that the owners are set correctly
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

#[test] // Tests that when a packet is transferred, the peropper allowance is consummed
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
    let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();

    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "300");

    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
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
            channel_id: format!("channel"),
            denom: format!("denom"),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &vec![]);
    let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

    let send_msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    let recv_msg = SudoMsg::RecvPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };

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
            channel_id: format!("channel"),
            denom: format!("denom"),
            quotas: vec![quota],
        }],
    };
    let info = mock_info(GOV_ADDR, &vec![]);
    let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

    // Sending 2%
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 60,
    };
    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "60");

    // Sending 2% more. Allowed, as sending has a 4% allowance
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 60,
    };

    let res = sudo(deps.as_mut(), mock_env(), msg).unwrap();
    println!("{res:?}");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "120");

    // Receiving 1% should still work. 4% *sent* through the path, but we can still receive.
    let recv_msg = SudoMsg::RecvPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 30,
    };
    let res = sudo(deps.as_mut(), mock_env(), recv_msg).unwrap();
    let Attribute { key, value } = &res.attributes[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.attributes[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "90");

    // Sending 2%. Should fail. In balance, we've sent 4% and received 1%, so only 1% left to send.
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 60,
    };
    let err = sudo(deps.as_mut(), mock_env(), msg.clone()).unwrap_err();
    assert!(matches!(err, ContractError::RateLimitExceded { .. }));

    // Sending 1%: Allowed; because sending has a 4% allowance. We've sent 4% already, but received 1%, so there's send cappacity again
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 30,
    };
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

    let send_msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    sudo(deps.as_mut(), mock_env(), send_msg.clone()).unwrap();

    let recv_msg = SudoMsg::RecvPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 30,
    };
    sudo(deps.as_mut(), mock_env(), recv_msg.clone()).unwrap();

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

#[test] // Tests quota percentages are between [0,100]
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

#[test] // Tests that undo reverts a packet send without affecting expiration or channel value
fn undo_send() {
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

    let send_msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    let undo_msg = SudoMsg::UndoSend {
        channel_id: format!("channel"),
        denom: format!("denom"),
        funds: 300,
    };

    sudo(deps.as_mut(), mock_env(), send_msg.clone()).unwrap();

    let trackers = RATE_LIMIT_TRACKERS
        .load(&deps.storage, ("channel".to_string(), "denom".to_string()))
        .unwrap();
    assert_eq!(trackers.first().unwrap().flow.outflow, 300);
    let period_end = trackers.first().unwrap().flow.period_end;
    let channel_value = trackers.first().unwrap().quota.channel_value;

    sudo(deps.as_mut(), mock_env(), undo_msg.clone()).unwrap();

    let trackers = RATE_LIMIT_TRACKERS
        .load(&deps.storage, ("channel".to_string(), "denom".to_string()))
        .unwrap();
    assert_eq!(trackers.first().unwrap().flow.outflow, 0);
    assert_eq!(trackers.first().unwrap().flow.period_end, period_end);
    assert_eq!(trackers.first().unwrap().quota.channel_value, channel_value);
}
