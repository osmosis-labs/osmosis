use std::path::PathBuf;

use cosmwasm_std::Coin;
use crosschain_swaps::msg::InstantiateMsg as CrosschainInstantiate;
use osmosis_testing::{Account, OsmosisTestApp, SigningAccount};
use osmosis_testing::{Gamm, Module, Wasm};
use serde::Serialize;
use swaprouter::msg::InstantiateMsg as SwapRouterInstantiate;

pub struct TestEnv {
    pub app: OsmosisTestApp,
    pub swaprouter_address: String,
    pub crosschain_address: String,
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
                Coin::new(100_000_000, "uosmo"),
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

        // Deploy the swaprouter
        println!("Deploying the swaprouter contract");
        let (_, swaprouter_address) = deploy_contract(
            &wasm,
            &owner,
            get_swaprouter_wasm(),
            &SwapRouterInstantiate {
                owner: owner.address(),
            },
        );

        println!("Deploying the crosschain swaps contract");
        let (_, crosschain_address) = deploy_contract(
            &wasm,
            &owner,
            get_crosschain_swaps_wasm(),
            &CrosschainInstantiate {
                swap_contract: swaprouter_address.clone(),
                channels: vec![("osmo".to_string(), "channel-0".to_string())],
                governor: owner.address(),
            },
        );

        TestEnv {
            app,
            swaprouter_address,
            crosschain_address,
            owner,
        }
    }
}

fn deploy_contract<M>(
    wasm: &Wasm<OsmosisTestApp>,
    owner: &SigningAccount,
    code: Vec<u8>,
    instantiate_msg: &M,
) -> (u64, String)
where
    M: ?Sized + Serialize,
{
    let code_id = wasm.store_code(&code, None, owner).unwrap().data.code_id;

    let contract_address = wasm
        .instantiate(
            code_id,
            instantiate_msg,
            Some(&owner.address()),
            None,
            &[],
            owner,
        )
        .unwrap()
        .data
        .address;
    (code_id, contract_address)
}

fn get_swaprouter_wasm() -> Vec<u8> {
    let wasm_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("..")
        .join("..")
        .join("..")
        .join("tests")
        .join("ibc-hooks")
        .join("bytecode")
        .join("swaprouter.wasm");
    println!("reading swaprouter wasm: {wasm_path:?}");
    std::fs::read(wasm_path).unwrap()
}

fn get_crosschain_swaps_wasm() -> Vec<u8> {
    let wasm_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("..")
        .join("..")
        .join("..")
        .join("tests")
        .join("ibc-hooks")
        .join("bytecode")
        .join("crosschain_swaps.wasm");
    println!("reading crosschain swaps wasm: {wasm_path:?}");
    std::fs::read(wasm_path).unwrap()
}
