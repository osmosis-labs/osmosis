mod test_env;
use cosmwasm_std::Coin;
use osmosis_std::types::osmosis::gamm::v1beta1::SwapAmountInRoute;
use osmosis_testing::{Module, RunnerError, Wasm};
use swaprouter::msg::{ExecuteMsg, GetRouteResponse, QueryMsg};
use test_env::*;

test_set_route!(
    set_initial_route_by_non_owner
    should failed_with "Unauthorized: execute wasm contract failed",

    sender = NonOwner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![SwapAmountInRoute {
            pool_id: 1,
            token_out_denom: "uion".to_string(),
        }],
    }
);

test_set_route!(
    set_initial_route_by_owner
    should succeed,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![SwapAmountInRoute {
            pool_id: 1,
            token_out_denom: "uion".to_string(),
        }],
    }
);

test_set_route!(
    override_route_with_multi_hop
    should succeed,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 2, // uatom/uosmo
                token_out_denom: "uatom".to_string(),
            },
            SwapAmountInRoute {
                pool_id: 3, // uatom/uion
                token_out_denom: "uion".to_string(),
            }
        ],
    }
);

test_set_route!(
    output_denom_that_does_not_ending_pool_route
    should failed_with
    r#"Invalid Pool Route: "last denom doesn't match": execute wasm contract failed"#,
    // r#"Invalid Pool Route: "denom uosmo is not in pool id 1": execute wasm contract failed"#,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 1, // uosmo/uion
                token_out_denom: "uosmo".to_string(),
            },
        ],
    }
);

test_set_route!(
    pool_does_not_have_input_asset
    should failed_with
    r#"Invalid Pool Route: "denom uatom is not in pool id 1": execute wasm contract failed"#,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uatom".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 1, // uosmo/uion
                token_out_denom: "uion".to_string(),
            },
        ],
    }
);

test_set_route!(
    pool_does_not_have_output_asset
    should failed_with
    r#"Invalid Pool Route: "denom uosmo is not in pool id 1": execute wasm contract failed"#,
    // confusing error message from chain, should state that:
    // > `denom uatom is not in pool id 1": execute wasm contract failed`
    // instead.

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uatom".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 1, // uosmo/uion
                token_out_denom: "uatom".to_string(),
            },
        ],
    }
);

test_set_route!(
    intermediary_pool_does_not_have_output_asset
    should failed_with
    r#"Invalid Pool Route: "denom uosmo is not in pool id 1": execute wasm contract failed"#,
    // confusing error message from chain, should state that:
    // > `denom foocoin is not in pool id 1": execute wasm contract failed`
    // instead.

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uatom".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 1, // uosmo/uion
                token_out_denom: "foocoin".to_string(),
            },
            SwapAmountInRoute {
                pool_id: 2, // uatom/uosmo
                token_out_denom: "uatom".to_string(),
            },
        ],
    }
);

test_set_route!(
    intermediary_pool_does_not_have_input_asset
    should failed_with
    r#"Invalid Pool Route: "denom uion is not in pool id 2": execute wasm contract failed"#,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uatom".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 1, // uosmo/uion
                token_out_denom: "uion".to_string(),
            },
            SwapAmountInRoute {
                pool_id: 2, // uatom/uosmo
                token_out_denom: "uatom".to_string(),
            },
        ],
    }
);

test_set_route!(
    non_existant_pool
    should failed_with
    r#"Invalid Pool Route: "denom uosmo is not in pool id 3": execute wasm contract failed"#,

    sender = Owner,
    msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uatom".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 3, // uatom/uion
                token_out_denom: "uion".to_string(),
            },
        ],
    }
);

// ======= helpers ========

#[macro_export]
macro_rules! test_set_route {
    ($test_name:ident should succeed, sender = Owner, msg = $msg:expr) => {
        #[test]
        fn $test_name() {
            test_set_route_success_case($msg)
        }
    };

    ($test_name:ident should failed_with $err:expr, sender = $sender:ident, msg = $msg:expr) => {
        #[test]
        fn $test_name() {
            test_set_route_failed_case(Sender::$sender, $msg, $err)
        }
    };
}

enum Sender {
    Owner,
    NonOwner,
}

fn test_set_route_success_case(msg: ExecuteMsg) {
    let TestEnv {
        app,
        contract_address,
        owner,
    } = TestEnv::new();
    let wasm = Wasm::new(&app);
    let res = wasm.execute(&contract_address, &msg, &[], &owner);

    // check if execution succeeded
    assert!(res.is_ok(), "{}", res.unwrap_err());

    // check if previously set route can be queried correctly
    match msg {
        ExecuteMsg::SetRoute {
            input_denom,
            output_denom,
            ..
        } => {
            let query = QueryMsg::GetRoute {
                input_denom,
                output_denom,
            };

            // expect route to always be found in this case`
            let res = wasm.query::<QueryMsg, GetRouteResponse>(&contract_address, &query);
            assert!(res.is_ok(), "{:?}", res.unwrap_err());
        }
        _ => {
            panic!("ExecuteMsg must be `SetRoute`");
        }
    }
}

fn test_set_route_failed_case(sender: Sender, msg: ExecuteMsg, expected_error: &str) {
    let TestEnv {
        app,
        contract_address,
        owner,
    } = TestEnv::new();
    let wasm = Wasm::new(&app);

    let sender = match sender {
        Sender::Owner => owner,
        Sender::NonOwner => {
            let initial_balance = [
                Coin::new(1_000_000_000_000, "uosmo"),
                Coin::new(1_000_000_000_000, "uion"),
                Coin::new(1_000_000_000_000, "uatom"),
            ];
            app.init_account(&initial_balance).unwrap()
        }
    };

    let res = wasm.execute::<ExecuteMsg>(&contract_address, &msg, &[], &sender);
    let err = res.unwrap_err();

    // assert on error message
    if let RunnerError::ExecuteError { msg } = &err {
        let expected_err = &format!(
            "failed to execute message; message index: 0: {}",
            expected_error
        );
        assert_eq!(msg, expected_err);
    } else {
        panic!("unexpected error: {:?}", err);
    }
}
