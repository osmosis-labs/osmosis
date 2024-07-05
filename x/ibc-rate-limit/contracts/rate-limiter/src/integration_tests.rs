#![cfg(test)]
use std::any::Any;

use crate::{helpers::RateLimitingContract, msg::{ExecuteMsg, QueryMsg}, state::{rate_limit::RateLimit, rbac::Roles}, test_msg_send, ContractError};
use cosmwasm_std::{to_binary, Addr, Coin, Empty, Timestamp, Uint128, Uint256};
use cw_multi_test::{App, AppBuilder, Contract, ContractWrapper, Executor};
use cosmwasm_std::Querier;
use crate::{
    msg::{InstantiateMsg, PathMsg, QuotaMsg},
    state::flow::tests::{RESET_TIME_DAILY, RESET_TIME_MONTHLY, RESET_TIME_WEEKLY},
};

pub fn contract_template() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        crate::contract::execute,
        crate::contract::instantiate,
        crate::contract::query,
    )
    .with_sudo(crate::contract::sudo);
    Box::new(contract)
}

const USER: &str = "USER";
const IBC_ADDR: &str = "osmo1vz5e6tzdjlzy2f7pjvx0ecv96h8r4m2y92thdm";
const GOV_ADDR: &str = "osmo1tzz5zf2u68t00un2j4lrrnkt2ztd46kfzfp58r";
const NATIVE_DENOM: &str = "nosmo";

fn mock_app() -> App {
    AppBuilder::new().build(|router, _, storage| {
        router
            .bank
            .init_balance(
                storage,
                &Addr::unchecked(USER),
                vec![Coin {
                    denom: NATIVE_DENOM.to_string(),
                    amount: Uint128::new(1_000),
                }],
            )
            .unwrap();
    })
}

// Instantiate the contract
fn proper_instantiate(paths: Vec<PathMsg>) -> (App, RateLimitingContract) {
    let mut app = mock_app();
    let cw_code_id = app.store_code(contract_template());

    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths,
    };

    let cw_rate_limit_contract_addr = app
        .instantiate_contract(
            cw_code_id,
            Addr::unchecked(GOV_ADDR),
            &msg,
            &[],
            "test",
            None,
        )
        .unwrap();

    let cw_rate_limit_contract = RateLimitingContract(cw_rate_limit_contract_addr);

    (app, cw_rate_limit_contract)
}

use cosmwasm_std::Attribute;

#[test] // Checks that the RateLimit flows are expired properly when time passes
fn expiration() {
    let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas: vec![quota],
    }]);

    // Using all the allowance
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 300_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    let res = app.sudo(cosmos_msg).unwrap();

    let Attribute { key, value } = &res.custom_attrs(1)[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.custom_attrs(1)[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "300");
    let Attribute { key, value } = &res.custom_attrs(1)[5];
    assert_eq!(key, "weekly_max_in");
    assert_eq!(value, "300");
    let Attribute { key, value } = &res.custom_attrs(1)[6];
    assert_eq!(key, "weekly_max_out");
    assert_eq!(value, "300");

    // Another packet is rate limited
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 300_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    let err = app.sudo(cosmos_msg).unwrap_err();

    assert_eq!(
        err.downcast_ref::<ContractError>().unwrap(),
        &ContractError::RateLimitExceded {
            channel: "channel".to_string(),
            denom: "denom".to_string(),
            amount: Uint256::from_u128(300),
            quota_name: "weekly".to_string(),
            used: Uint256::from_u128(300),
            max: Uint256::from_u128(300),
            reset: Timestamp::from_nanos(1572402219879305533),
        }
    );

    // ... Time passes
    app.update_block(|b| {
        b.height += 1000;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // Sending the packet should work now
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 300_u32.into()
    );

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    let res = app.sudo(cosmos_msg).unwrap();

    let Attribute { key, value } = &res.custom_attrs(1)[3];
    assert_eq!(key, "weekly_used_in");
    assert_eq!(value, "0");
    let Attribute { key, value } = &res.custom_attrs(1)[4];
    assert_eq!(key, "weekly_used_out");
    assert_eq!(value, "300");
    let Attribute { key, value } = &res.custom_attrs(1)[5];
    assert_eq!(key, "weekly_max_in");
    assert_eq!(value, "300");
    let Attribute { key, value } = &res.custom_attrs(1)[6];
    assert_eq!(key, "weekly_max_out");
    assert_eq!(value, "300");
}

#[test] // Tests we can have different maximums for different quotaas (daily, weekly, etc) and that they all are active at the same time
fn multiple_quotas() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% to use the daily allowance
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // Another packet is rate limited
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One day passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
    });

    // Sending the packet should work now
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // Do that for 4 more days
    for _ in 1..4 {
        // ... One day passes
        app.update_block(|b| {
            b.height += 10;
            b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
        });

        // Sending the packet should work now
        let msg = test_msg_send!(
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 101_u32.into(),
            funds: 1_u32.into()
        );
        let cosmos_msg = cw_rate_limit_contract.sudo(msg);
        app.sudo(cosmos_msg).unwrap();
    }

    // ... One day passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
    });

    // We now have exceeded the weekly limit!  Even if the daily limit allows us, the weekly doesn't
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // We can still can't send because the weekly and monthly limits are the same
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Waiting a week again, doesn't help!!
    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // We can still can't send because the  monthly limit hasn't passed
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Only after two more weeks we can send again
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds((RESET_TIME_WEEKLY * 2) + 1) // Two weeks
    });

    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();
}

#[test] // Tests that the channel value is based on the value at the beginning of the period
fn channel_value_cached() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 2, 2),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% (half of the daily allowance)
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // Sending 3% is now rate limited
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100_u32.into(),
        funds: 3_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Even if the channel value increases, the percentage is calculated based on the value at period start
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100000_u32.into(),
        funds: 3_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One day passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
    });

    // New Channel Value world!

    // Sending 1% of a new value (10_000) passes the daily check, cause it
    // has expired, but not the weekly check (The value for last week is
    // sitll 100, as only 1 day has passed)
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 10_000_u32.into(),
        funds: 100_u32.into()
    );

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // Sending 1% of a new value should work and set the value for the day at 10_000
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 10_000_u32.into(),
        funds: 100_u32.into()
    );

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // If the value magically decreasses. We can still send up to 100 more (1% of 10k)
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 1_u32.into(),
        funds: 75_u32.into()
    );

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();
}

#[test] // Checks that RateLimits added after instantiation are respected
fn add_paths_later() {
    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![]);

    // All sends are allowed
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000_u32.into(),
        funds: 300_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg.clone());
    let res = app.sudo(cosmos_msg).unwrap();

    let Attribute { key, value } = &res.custom_attrs(1)[3];
    assert_eq!(key, "quota");
    assert_eq!(value, "none");

    // Add a weekly limit of 1%
    let management_msg = ExecuteMsg::AddPath {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas: vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 1, 1)],
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg).unwrap();

    // Executing the same message again should fail, as it is now rate limited
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();
}


#[test]
fn test_execute_add_path() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    let management_msg = ExecuteMsg::AddPath {
        channel_id: format!("new_channel_id"),
        denom: format!("new_denom"),
        quotas: vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 1, 1)],
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();


    // Sending 1% to use the daily allowance
    let msg = test_msg_send!(
        channel_id: format!("new_channel_id"),
        denom: format!("new_denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg.clone());
    app.sudo(cosmos_msg).unwrap();

    let response: Vec<crate::state::rate_limit::RateLimit> = app.wrap().query_wasm_smart(cw_rate_limit_contract.addr(), &QueryMsg::GetQuotas {
        channel_id: "new_channel_id".to_string(),
        denom: "new_denom".to_string()
    }).unwrap();
    assert_eq!(response.len(), 1);
    assert_eq!(response[0].flow.outflow, Uint256::one());
    assert_eq!(response[0].quota.channel_value, Some(Uint256::from_u128(101)));

}
#[test]
fn test_execute_remove_path() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    let management_msg = ExecuteMsg::RemovePath {
        channel_id: "any".to_string(),
        denom: "denom".to_string(),
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // rate limits should be removed
    assert!(app.wrap().query_wasm_smart::<crate::state::rate_limit::RateLimit>(cw_rate_limit_contract.addr(), &QueryMsg::GetQuotas {
        channel_id: "any".to_string(),
        denom: "denom".to_string()
    }).is_err());

}

#[test]
fn test_execute_reset_path_quota() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% to use the daily allowance
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg.clone());
    app.sudo(cosmos_msg).unwrap();

    let management_msg = ExecuteMsg::ResetPathQuota {
        channel_id: "any".to_string(),
        denom: "denom".to_string(),
        quota_id: "daily".to_string()
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();
    
    let response =  app.wrap().query_wasm_smart::<Vec<crate::state::rate_limit::RateLimit>>(cw_rate_limit_contract.addr(), &QueryMsg::GetQuotas {
        channel_id: "any".to_string(),
        denom: "denom".to_string()
    }).unwrap();

    // daily quota should be reset
    let daily_quota = response.iter().find(|rate_limit| rate_limit.quota.name.eq("daily")).unwrap();
    assert_eq!(daily_quota.flow.inflow, Uint256::zero());
    assert_eq!(daily_quota.flow.outflow, Uint256::zero());

    // weekly and monthly should not be reset
    let weekly_quota = response.iter().find(|rate_limit| rate_limit.quota.name.eq("weekly")).unwrap();
    assert_eq!(weekly_quota.flow.inflow, Uint256::zero());
    assert_eq!(weekly_quota.flow.outflow, Uint256::one());

    let  monthly_quota = response.iter().find(|rate_limit| rate_limit.quota.name.eq("monthly")).unwrap();
    assert_eq!(monthly_quota.flow.inflow, Uint256::zero());
    assert_eq!(monthly_quota.flow.outflow, Uint256::one());
}

#[test]
fn test_execute_grant_and_revoke_role() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);


    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobar".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    let response = app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles {
            owner: "foobar".to_string()
        }
    ).unwrap();
    assert_eq!(response.len(), 1);
    assert_eq!(response[0], Roles::GrantRole);

    // test foobar can grant a role
    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobarbaz".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();

    let response = app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles {
            owner: "foobarbaz".to_string()
        }
    ).unwrap();
    assert_eq!(response.len(), 1);
    assert_eq!(response[0], Roles::GrantRole);


    // test role revocation

    let management_msg = ExecuteMsg::RevokeRole {
        signer: "foobar".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // foobar should no longer have roles
    assert!(app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles { owner: "foobar".to_string() }
    ).is_err());

}

#[test]
fn test_execute_edit_path_quota() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% to use the daily allowance
    let msg = test_msg_send!(
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 101_u32.into(),
        funds: 1_u32.into()
    );
    let cosmos_msg = cw_rate_limit_contract.sudo(msg.clone());
    app.sudo(cosmos_msg).unwrap();



    let management_msg = ExecuteMsg::EditPathQuota {
        channel_id: "any".to_string(),
        denom: "denom".to_string(),
        quota: QuotaMsg {
            send_recv: (81, 58),
            name: "monthly".to_string(),
            duration: RESET_TIME_MONTHLY
        }
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    let response = app.wrap().query_wasm_smart::<Vec<RateLimit>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetQuotas {
            channel_id: "any".to_string(),
            denom: "denom".to_string()
        }
    ).unwrap();
    let monthly_quota = response.iter().find(|rate_limit| rate_limit.quota.name.eq("monthly")).unwrap();
    assert_eq!(monthly_quota.quota.max_percentage_send, 81);
    assert_eq!(monthly_quota.quota.max_percentage_recv, 58);
}
#[test]
fn test_execute_remove_message() {
    
    // this test case also covers timelock delay set, as a non zero timelock
    // will force the message to be queued, thus allowing queue removal

    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);



    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobar".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // set a timelock delay for foobar
    let management_msg = ExecuteMsg::SetTimelockDelay {
        signer: "foobar".to_string(),
        hours: 1
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke as insufficient permissions
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // message submitter by foobar should not be queued
    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobarbaz".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();
    let response = app.wrap().query_wasm_smart::<Vec<String>>(
            cw_rate_limit_contract.addr(),
            &QueryMsg::GetMessageIds
        ).unwrap();
    assert_eq!(
        response.len(),
        1
    );

    // remove the message
    let management_msg = ExecuteMsg::RemoveMessage {
        message_id: response[0].clone(),
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // no messges should be present
    assert_eq!(app.wrap().query_wasm_smart::<Vec<String>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetMessageIds
    ).unwrap().len(), 0);
}

#[test]
fn test_execute_process_messages() {
    let quotas = vec![
        QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
        QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
        QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
    ];

    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![PathMsg {
        channel_id: format!("any"),
        denom: format!("denom"),
        quotas,
    }]);


    // allocate GrantRole and RevokeRole to `foobar`
    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobar".to_string(),
        roles: vec![Roles::GrantRole, Roles::RevokeRole]
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // set a timelock delay for foobar
    let management_msg = ExecuteMsg::SetTimelockDelay {
        signer: "foobar".to_string(),
        hours: 1
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    // non gov cant invoke as insufficient permissions
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
    // gov addr can invoke
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg.clone()).unwrap();

    // message submitted by foobar should be queued
    // allocate GrantRole to foobarbaz
    let management_msg = ExecuteMsg::GrantRole {
        signer: "foobarbaz".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();
    let response = app.wrap().query_wasm_smart::<Vec<String>>(
            cw_rate_limit_contract.addr(),
            &QueryMsg::GetMessageIds
        ).unwrap();
    assert_eq!(
        response.len(),
        1
    );

    // any address should be able to trigger queue message processing
    let management_msg = ExecuteMsg::ProcessMessages {
        count: Some(1),
        message_ids: None
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("veryrandomaddress"), cosmos_msg).unwrap();

    // insufficient time has passed so queue should still be 1
    assert_eq!(app.wrap().query_wasm_smart::<Vec<String>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetMessageIds
    ).unwrap().len(), 1);

    // advance time
    app.update_block(|block| {
        block.height += 100;
        block.time = block.time.plus_seconds(3601)
    });

    // any address should be able to trigger queue message processing
    let management_msg = ExecuteMsg::ProcessMessages {
        count: Some(1),
       message_ids: None
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("veryrandomaddress"), cosmos_msg).unwrap();

    // no messges should be present as time passed and message was executed
    assert_eq!(app.wrap().query_wasm_smart::<Vec<String>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetMessageIds
    ).unwrap().len(), 0);

    // foobarbaz should have the GrantRole permission
    let response = app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles {
            owner: "foobarbaz".to_string()
        }
    ).unwrap();
    assert_eq!(response.len(), 1);
    assert_eq!(response[0], Roles::GrantRole);

    app.update_block(|block| {
        block.height += 1;
        block.time = block.time.plus_seconds(3600);
    });

    let management_msg = ExecuteMsg::RevokeRole {
        signer: "foobarbaz".to_string(),
        roles: vec![Roles::GrantRole]
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();

    let message_ids = app.wrap().query_wasm_smart::<Vec<String>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetMessageIds
    ).unwrap();
    assert_eq!(message_ids.len(), 1);

    app.update_block(|block| {
        block.height += 1;
    });

    let management_msg = ExecuteMsg::ProcessMessages {
        count: Some(1),
        message_ids: None
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();

    // insufficient time has passed so queue length is still 1
    let response = app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles {
            owner: "foobarbaz".to_string()
        }
    ).unwrap();
    assert_eq!(response.len(), 1);

    // advance time
    app.update_block(|block| {
        block.height += 100;
        block.time = block.time.plus_seconds(3601);
    });

    let management_msg = ExecuteMsg::ProcessMessages {
        count: None,
        message_ids: Some(message_ids.clone())
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).unwrap();
    
    // sufficient time has passed, empty queue
    let message_ids = app.wrap().query_wasm_smart::<Vec<String>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetMessageIds
    ).unwrap();
    assert_eq!(message_ids.len(), 0);

    // no rolles allocated, storage key should be removed
    assert!(app.wrap().query_wasm_smart::<Vec<Roles>>(
        cw_rate_limit_contract.addr(),
        &QueryMsg::GetRoles {
            owner: "foobarbaz".to_string()
        }
    ).is_err());

    // error should be returned when all params are None
    let management_msg = ExecuteMsg::ProcessMessages {
        count: None,
       message_ids: None
    };
    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    assert!(app.execute(Addr::unchecked("foobar"), cosmos_msg.clone()).is_err());
}