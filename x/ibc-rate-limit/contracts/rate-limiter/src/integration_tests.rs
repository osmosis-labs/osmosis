#![cfg(test)]
use crate::{helpers::RateLimitingContract, msg::ExecuteMsg};
use cosmwasm_std::{Addr, Coin, Empty, Uint128};
use cw_multi_test::{App, AppBuilder, Contract, ContractWrapper, Executor};

use crate::{
    msg::{InstantiateMsg, PathMsg, QuotaMsg, SudoMsg},
    state::tests::{RESET_TIME_DAILY, RESET_TIME_MONTHLY, RESET_TIME_WEEKLY},
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
const IBC_ADDR: &str = "IBC_MODULE";
const GOV_ADDR: &str = "GOV_MODULE";
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
    let cw_template_id = app.store_code(contract_template());

    let msg = InstantiateMsg {
        gov_module: Addr::unchecked(GOV_ADDR),
        ibc_module: Addr::unchecked(IBC_ADDR),
        paths,
    };

    let cw_rate_limit_contract_addr = app
        .instantiate_contract(
            cw_template_id,
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
        channel_id: format!("channel"),
        denom: format!("denom"),
        quotas: vec![quota],
    }]);

    // Using all the allowance
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
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
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    let _err = app.sudo(cosmos_msg).unwrap_err();

    // TODO: how do we check the error type here?

    // ... Time passes
    app.update_block(|b| {
        b.height += 1000;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // Sending the packet should work now
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };

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
        channel_id: format!("channel"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% to use the daily allowance
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // Another packet is rate limited
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One day passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
    });

    // Sending the packet should work now
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };

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
        let msg = SudoMsg::SendPacket {
            channel_id: format!("channel"),
            denom: format!("denom"),
            channel_value: 100,
            funds: 1,
        };
        let cosmos_msg = cw_rate_limit_contract.sudo(msg);
        app.sudo(cosmos_msg).unwrap();
    }

    // ... One day passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
    });

    // We now have exceeded the weekly limit!  Even if the daily limit allows us, the weekly doesn't
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // We can still can't send because the weekly and monthly limits are the same
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Waiting a week again, doesn't help!!
    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // We can still can't send because the  monthly limit hasn't passed
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Only after two more weeks we can send again
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds((RESET_TIME_WEEKLY * 2) + 1) // Two weeks
    });

    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
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
        channel_id: format!("channel"),
        denom: format!("denom"),
        quotas,
    }]);

    // Sending 1% (half of the daily allowance)
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 1,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // Sending 3% is now rate limited
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100,
        funds: 3,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // Even if the channel value increases, the percentage is calculated based on the value at period start
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 100000,
        funds: 3,
    };
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
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 10_000,
        funds: 100,
    };

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();

    // ... One week passes
    app.update_block(|b| {
        b.height += 10;
        b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
    });

    // Sending 1% of a new value should work and set the value for the day at 10_000
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 10_000,
        funds: 100,
    };

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();

    // If the value magically decreasses. We can still send up to 100 more (1% of 10k)
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 1,
        funds: 75,
    };

    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap();
}

#[test] // Checks that RateLimits added after instantiation are respected
fn add_paths_later() {
    let (mut app, cw_rate_limit_contract) = proper_instantiate(vec![]);

    // All sends are allowed
    let msg = SudoMsg::SendPacket {
        channel_id: format!("channel"),
        denom: format!("denom"),
        channel_value: 3_000,
        funds: 300,
    };
    let cosmos_msg = cw_rate_limit_contract.sudo(msg.clone());
    let res = app.sudo(cosmos_msg).unwrap();
    let Attribute { key, value } = &res.custom_attrs(1)[3];
    assert_eq!(key, "quota");
    assert_eq!(value, "none");

    // Add a weekly limit of 1%
    let management_msg = ExecuteMsg::AddPath {
        channel_id: format!("channel"),
        denom: format!("denom"),
        quotas: vec![QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 1, 1)],
    };

    let cosmos_msg = cw_rate_limit_contract.call(management_msg).unwrap();
    app.execute(Addr::unchecked(GOV_ADDR), cosmos_msg).unwrap();

    // Executing the same message again should fail, as it is now rate limited
    let cosmos_msg = cw_rate_limit_contract.sudo(msg);
    app.sudo(cosmos_msg).unwrap_err();
}
