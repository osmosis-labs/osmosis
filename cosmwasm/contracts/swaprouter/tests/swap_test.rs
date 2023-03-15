mod test_env;
use std::str::FromStr;

use cosmwasm_std::{Coin, Decimal};
use osmosis_std::types::osmosis::gamm::v1beta1::SwapAmountInRoute;
use osmosis_testing::cosmrs::proto::cosmos::bank::v1beta1::QueryAllBalancesRequest;
use osmosis_testing::cosmrs::proto::cosmwasm::wasm::v1::MsgExecuteContractResponse;
use osmosis_testing::{
    Account, Bank, Module, OsmosisTestApp, RunnerError, RunnerExecuteResult, SigningAccount, Wasm,
};
use swaprouter::msg::{ExecuteMsg, Slippage};
use test_env::*;

test_swap!(
    try_swap_for_correct_route
    should succeed,

    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uosmo"),
        output_denom: "uion".to_string(),
        slippage: Slippage::MinOutputAmount(1u128.into()),
    },
    funds: [
        Coin::new(1000, "uosmo")
    ]
);

test_swap!(
    not_enough_attached_funds_to_swap should failed_with
    "Insufficient Funds: execute wasm contract failed",

    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uosmo"),
        output_denom: "uion".to_string(),
        slippage: Slippage::MinOutputAmount(1u128.into()),
    },
    funds: [
        Coin::new(10, "uosmo")
    ]
);

test_swap!(
    wrong_denom_attached_funds should failed_with
    "Insufficient Funds: execute wasm contract failed",

    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uosmo"),
        output_denom: "uion".to_string(),
        slippage: Slippage::MinOutputAmount(1u128.into()),
    },
    funds: [
        Coin::new(10, "uion")
    ]
);

test_swap!(
    minimum_output_amount_too_high should failed_with
    "dispatch: submessages: uion token is lesser than min amount: calculated amount is lesser than min amount",

    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uosmo"),
        output_denom: "uion".to_string(),
        slippage: Slippage::MinOutputAmount(1000000000000000000000000u128.into()),
    },
    funds: [
        Coin::new(1000, "uosmo")
    ]
);

test_swap!(
    non_existant_route should failed_with
    "alloc::vec::Vec<osmosis_std::types::osmosis::gamm::v1beta1::SwapAmountInRoute> not found: execute wasm contract failed",

    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uion"),
        output_denom: "uosmo".to_string(),
        slippage: Slippage::MinOutputAmount(1000000000000000000000000u128.into()),
    },
    funds: [
        Coin::new(1000, "uion")
    ]
);

test_swap!(
    twap_based_swap
    should succeed,
    msg = ExecuteMsg::Swap {
        input_coin: Coin::new(1000, "uosmo"),
        output_denom: "uion".to_string(),
        slippage: Slippage::Twap{ window_seconds: Some(1), slippage_percentage: Decimal::from_str("5").unwrap() },
    },
    funds: [
        Coin::new(10000, "uosmo")
    ]
);

// ======= helpers ========

#[macro_export]
macro_rules! test_swap {
    ($test_name:ident should succeed, msg = $msg:expr, funds: $funds:expr) => {
        #[test]
        fn $test_name() {
            test_swap_success_case($msg, &$funds);
        }
    };
    ($test_name:ident should failed_with $err:expr, msg = $msg:expr, funds: $funds:expr) => {
        #[test]
        fn $test_name() {
            test_swap_failed_case($msg, &$funds, $err);
        }
    };
}

const INITIAL_AMOUNT: u128 = 1_000_000_000_000;

fn test_swap_success_case(msg: ExecuteMsg, funds: &[Coin]) {
    let (app, sender, _res) = setup_route_and_execute_swap(&msg, funds);
    // dbg!(res);
    // println!("{:?}", String::from_utf8(to_vec(&msg).unwrap()).unwrap());
    assert_input_decreased_and_output_increased(&app, &sender.address(), &msg);
}

fn test_swap_failed_case(msg: ExecuteMsg, funds: &[Coin], expected_error: &str) {
    let (_app, _sender, res) = setup_route_and_execute_swap(&msg, funds);
    let err = res.unwrap_err();
    assert_eq!(
        err,
        RunnerError::ExecuteError {
            msg: format!("failed to execute message; message index: 0: {expected_error}")
        }
    );
}

fn setup_route_and_execute_swap(
    msg: &ExecuteMsg,
    funds: &[Coin],
) -> (
    OsmosisTestApp,
    SigningAccount,
    RunnerExecuteResult<MsgExecuteContractResponse>,
) {
    let TestEnv {
        app,
        contract_address,
        owner,
    } = TestEnv::new();
    let wasm = Wasm::new(&app);

    let initial_balance = [
        Coin::new(INITIAL_AMOUNT, "uosmo"),
        Coin::new(INITIAL_AMOUNT, "uion"),
        Coin::new(INITIAL_AMOUNT, "uatom"),
    ];

    let sender = app.init_account(&initial_balance).unwrap();

    // setup route
    // uosmo/uion = pool(2): uosmo/uatom -> pool(3): uatom/uion
    let set_route_msg = ExecuteMsg::SetRoute {
        input_denom: "uosmo".to_string(),
        output_denom: "uion".to_string(),
        pool_route: vec![
            SwapAmountInRoute {
                pool_id: 2,
                token_out_denom: "uatom".to_string(),
            },
            SwapAmountInRoute {
                pool_id: 3,
                token_out_denom: "uion".to_string(),
            },
        ],
    };

    // setup route by swaprouter's owner
    wasm.execute(&contract_address, &set_route_msg, &[], &owner)
        .expect("Setup route fixture must always succeed");

    // execute swap
    assert!(
        matches!(msg, ExecuteMsg::Swap { .. }),
        "only allow `ExecuteMsg::Swap` msg for this test"
    );
    let res = wasm.execute(&contract_address, &msg, funds, &sender);
    (app, sender, res)
}

fn assert_input_decreased_and_output_increased(
    app: &OsmosisTestApp,
    sender: &str,
    msg: &ExecuteMsg,
) {
    let bank = Bank::new(app);
    let balances = bank
        .query_all_balances(&QueryAllBalancesRequest {
            address: sender.to_string(),
            pagination: None,
        })
        .unwrap()
        .balances;
    match msg {
        ExecuteMsg::Swap {
            input_coin,
            output_denom,
            ..
        } => {
            let input_amount = get_amount(&balances, &input_coin.denom);
            let output_amount = get_amount(&balances, output_denom);

            assert!(
                input_amount < INITIAL_AMOUNT,
                "Input must be decreased after swap"
            );
            assert!(
                output_amount > INITIAL_AMOUNT,
                "Output must be increased after swap"
            );
        }
        _ => {
            panic!("Wrong message type: Must be `ExecuteMsg::Swap`");
        }
    }
}

fn get_amount(
    balances: &Vec<osmosis_testing::cosmrs::proto::cosmos::base::v1beta1::Coin>,
    denom: &str,
) -> u128 {
    balances
        .iter()
        .find(|b| b.denom == denom)
        .unwrap()
        .amount
        .parse::<u128>()
        .unwrap()
}
