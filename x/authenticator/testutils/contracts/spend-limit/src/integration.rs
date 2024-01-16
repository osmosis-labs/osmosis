#[cfg(test)]
mod tests {
    use cosmwasm_std::{Addr, Binary};
    use osmosis_authenticators::{
        Any, AuthenticationRequest, SignModeTxData, SignatureData, TxData,
    };

    use cosmwasm_std::Coin;
    use osmosis_test_tube::{Account, OsmosisTestApp, SigningAccount};
    use osmosis_test_tube::{Gamm, Module, Wasm};
    use serde::Serialize;

    use crate::msg::{InstantiateMsg, SudoMsg};

    #[test]
    fn test_integration() {
        let app = OsmosisTestApp::new();
        let gamm = Gamm::new(&app);
        let wasm = Wasm::new(&app);

        // setup owner account
        let initial_balance = [
            Coin::new(400, "uosmo"),
            // ATOM
            Coin::new(
                300,
                "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
            ),
            // USDC
            Coin::new(
                200,
                "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
            ),
            Coin::new(100, "memecoin"),
        ];
        let owner = app.init_account(&initial_balance).unwrap();

        // create pools
        gamm.create_basic_pool(
            &[
                Coin::new(400, "uosmo"),
                // USDC
                Coin::new(
                    200,
                    "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
                ),
            ],
            &owner,
        )
        .unwrap();
        gamm.create_basic_pool(
            &[
                // ATOM
                Coin::new(
                    300,
                    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
                ),
                // USDC
                Coin::new(
                    200,
                    "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
                ),
            ],
            &owner,
        )
        .unwrap();
        gamm.create_basic_pool(
            &[
                Coin::new(400, "uosmo"),
                // ATOM
                Coin::new(
                    300,
                    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
                ),
            ],
            &owner,
        )
        .unwrap();

        let wasm_byte_code = std::fs::read("../artifacts/spend_limit.wasm").unwrap();
        println!("Deploying the spend_limit contract");
        let (_, spendlimit_address) =
            deploy_contract(&wasm, &owner, wasm_byte_code, &InstantiateMsg {});
        println!("{}", spendlimit_address);

        let auth_request = AuthenticationRequest {
            account: Addr::unchecked("mock_account"),
            msg: Any {
                type_url: "cosmwasm/std/Msg".to_string(),
                value: Binary::from(b"mock_msg_value".to_vec()),
            },
            signature: Binary::from(b"mock_signature".to_vec()),
            sign_mode_tx_data: SignModeTxData {
                sign_mode_direct: Binary::from(b"mock_sign_mode_direct".to_vec()),
                sign_mode_textual: Some("mock_sign_mode_textual".to_string()),
            },
            tx_data: TxData {
                chain_id: "mock_chain_id".to_string(),
                account_number: 1,
                sequence: 1,
                timeout_height: 100,
                msgs: vec![Any {
                    type_url: "cosmwasm/std/Msg".to_string(),
                    value: Binary::from(b"mock_msg_value".to_vec()),
                }],
                memo: "mock_memo".to_string(),
            },
            signature_data: SignatureData {
                signers: vec![Addr::unchecked("mock_signer")],
                signatures: vec![Binary::from(b"mock_signature".to_vec())],
            },
            simulate: false,
            authenticator_params: Some(Binary::from(
                br#"{ "id": "100", "duration": 1000, "limit": 1000 }"#.to_vec(),
            )),
        };

        // XXX: osmosis_test_tube cannot handle sudo messages
        // TODO: update osmosis_test_tube and add sudo to the API, then finish these test
        // let msg = SudoMsg::Authenticate(auth_request.clone());
        // let result = wasm
        //     .execute(
        //         &spendlimit_address,
        //         &msg,
        //         &[Coin::new(2500, "uosmo")],
        //         &owner,
        //     )
        //     .expect("Sudo call failed");

        // dbg!(result);
        // println!("Authentication result: {:?}", auth_result);
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
}
