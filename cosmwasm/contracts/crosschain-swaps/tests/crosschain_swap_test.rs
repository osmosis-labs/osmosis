#![allow(unused_imports)]
#![allow(dead_code)]

mod test_env;
use std::str::FromStr;

use cosmwasm_std::{Addr, Coin, Decimal};
use osmosis_std::types::osmosis::gamm::v1beta1::SwapAmountInRoute;
use osmosis_testing::cosmrs::proto::cosmos::bank::v1beta1::QueryAllBalancesRequest;

use crosschain_swaps::msg::{ExecuteMsg as CrossChainExecute, FailedDeliveryAction};
use osmosis_testing::{Account, Bank, Module, Wasm};
use swaprouter::msg::{ExecuteMsg as SwapRouterExecute, Slippage};
use test_env::*;

const INITIAL_AMOUNT: u128 = 1_000_000_000_000;

#[test]
fn crosschain_swap() {
    let TestEnv {
        app,
        swaprouter_address,
        crosschain_address,
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
    let set_route_msg = SwapRouterExecute::SetRoute {
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

    println!("{:?}", serde_json_wasm::to_string(&set_route_msg).unwrap());

    // setup route by swaprouter's owner
    wasm.execute(&swaprouter_address, &set_route_msg, &[], &owner)
        .expect("Setup route fixture must always succeed");

    // execute swap
    let output_denom = "uion".to_string();
    let msg = CrossChainExecute::OsmosisSwap {
        output_denom,
        slippage: Slippage::Twap {
            window_seconds: Some(1),
            slippage_percentage: Decimal::from_str("5").unwrap(),
        },
        receiver: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89".to_string(),
        on_failed_delivery: FailedDeliveryAction::DoNothing,
        next_memo: None,
    };
    let funds: &[Coin] = &[Coin::new(10000, "uosmo")];
    println!("{}", serde_json_wasm::to_string(&msg).unwrap());
    let _res = wasm.execute(&crosschain_address, &msg, funds, &sender);
    //dbg!(&res);

    // This test cannot be completed until we have ibc tests on osmosis testing.

    // let bank = Bank::new(&app);
    // let balances = bank
    //     .query_all_balances(&QueryAllBalancesRequest {
    //         address: sender.address().to_string(),
    //         pagination: None,
    //     })
    //     .unwrap()
    //     .balances;
    // let input_amount = get_amount(&balances, &input_coin.denom);
    // let output_amount = get_amount(&balances, &output_denom);

    // assert!(
    //     input_amount < INITIAL_AMOUNT,
    //     "Input must be decreased after swap"
    // );
    // assert!(
    //     output_amount > INITIAL_AMOUNT,
    //     "Output must be increased after swap"
    // );
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
