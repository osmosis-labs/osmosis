use std::path::PathBuf;

use cosmwasm_std::Coin;
use osmosis_testing::{Account, OsmosisTestApp, SigningAccount};
use osmosis_testing::{Gamm, Module, Wasm};
use swaprouter::msg::InstantiateMsg;

pub struct TestEnv {
    pub app: OsmosisTestApp,
    pub contract_address: String,
    pub owner: SigningAccount,
}
impl TestEnv {
    pub fn new() -> Self {
        let app = OsmosisTestApp::new();
        let gamm = Gamm::new(&app);
        let wasm = Wasm::new(&app);

        // setup owner account
        let initial_balance = [
            Coin::new(1_000_000_000_000, "uosmo"),
            Coin::new(1_000_000_000_000, "uion"),
            Coin::new(1_000_000_000_000, "uatom"),
        ];
        let owner = app.init_account(&initial_balance).unwrap();

        // create pools
        gamm.create_basic_pool(
            &[
                Coin::new(100_000_000, "uion"),
                Coin::new(100_000_000, "uosmo"),
            ],
            &owner,
        )
        .unwrap();
        gamm.create_basic_pool(
            &[
                Coin::new(100_000_000, "uatom"),
                Coin::new(200_000_000, "uosmo"),
            ],
            &owner,
        )
        .unwrap();
        gamm.create_basic_pool(
            &[
                Coin::new(100_000_000, "uatom"),
                Coin::new(100_000_000, "uion"),
            ],
            &owner,
        )
        .unwrap();

        let code_id = wasm
            .store_code(&get_wasm(), None, &owner)
            .unwrap()
            .data
            .code_id;

        let contract_address = wasm
            .instantiate(
                code_id,
                &InstantiateMsg {
                    owner: owner.address(),
                },
                Some(&owner.address()),
                None,
                &[],
                &owner,
            )
            .unwrap()
            .data
            .address;

        TestEnv {
            app,
            contract_address,
            owner,
        }
    }
}

fn get_wasm() -> Vec<u8> {
    let wasm_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("..")
        .join("..")
        .join("..")
        .join("tests")
        .join("ibc-hooks")
        .join("bytecode")
        .join("swaprouter.wasm");
    std::fs::read(wasm_path).unwrap()
}
