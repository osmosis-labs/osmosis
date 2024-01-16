use cosmwasm_std::{DepsMut, Env, Response};
use osmosis_authenticators::{AuthenticationRequest, AuthenticationResult};

use crate::bank::query_account_balances;
use crate::state::{AuthenticatorParams, SpendLimit, SPEND_LIMITS, USDC_DENOM};
use crate::ContractError;

pub fn sudo_authenticate(
    deps: DepsMut,
    env: Env,
    auth_request: AuthenticationRequest,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!("auth_request {:?}", auth_request));
    let account_address = auth_request.account;
    let balances = query_account_balances(deps.as_ref(), &account_address)?;

    // Load spend limits if they exist
    if let Some(mut spend_limit) =
        SPEND_LIMITS.may_load(deps.storage, account_address.to_string())?
    {
        spend_limit.balance = balances.clone();
        if env.block.height - spend_limit.block_of_last_tx > spend_limit.number_of_blocks_active {
            // XXX: should we remove the spend_limit
            // SPEND_LIMITS.remove(deps.storage, account_address.to_string());
            return Ok(Response::new().set_data(AuthenticationResult::NotAuthenticated {}));
        }
        SPEND_LIMITS.save(deps.storage, account_address.to_string(), &spend_limit)?;
    }

    // Handle new authentication with authenticator_params
    if let Some(params_binary) = &auth_request.authenticator_params {
        let authenticator_params =
            serde_json_wasm::from_slice::<AuthenticatorParams>(params_binary)
                .map_err(|_| ContractError::InvalidAuthenticatorParams {})?;

        let spend_limit = SpendLimit {
            id: authenticator_params.id.clone(),
            denom: String::from(USDC_DENOM),
            amount_left: authenticator_params.limit,
            balance: balances,
            block_of_last_tx: env.block.height,
            number_of_blocks_active: authenticator_params.duration,
        };
        SPEND_LIMITS.save(deps.storage, account_address.to_string(), &spend_limit)?;
        return Ok(Response::new().set_data(AuthenticationResult::Authenticated {}));
    }

    Ok(Response::new().set_data(AuthenticationResult::NotAuthenticated {}))
}

#[cfg(test)]
mod tests {
    use crate::authenticate::sudo_authenticate;
    use crate::contract::{instantiate, query_spend_limit};
    use crate::msg::InstantiateMsg;
    use crate::state::AuthenticatorParams;

    use cosmwasm_std::testing::{
        mock_dependencies_with_balances, mock_env, mock_info, MockQuerier,
    };
    use cosmwasm_std::{Addr, Binary, Coin, Response};
    use osmosis_authenticators::{
        Any, AuthenticationRequest, AuthenticationResult, SignModeTxData, SignatureData, TxData,
    };

    #[test]
    fn test_successful_authentication_init_params() {
        let mut deps = mock_dependencies_with_balances(&[(
            "mock_account",
            &[
                Coin::new(500, "uosmo"),
                // ATOM
                Coin::new(
                    200,
                    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
                ),
                // USDC
                Coin::new(
                    200,
                    "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
                ),
                Coin::new(200, "memecoin"),
            ],
        )]);
        let env = mock_env();
        let info = mock_info("mock_signer", &[]);
        instantiate(deps.as_mut(), env.clone(), info.clone(), InstantiateMsg {}).unwrap();

        let auth_request = create_mock_authentication_request();
        let response = sudo_authenticate(deps.as_mut(), env.clone(), auth_request.clone()).unwrap();

        // Check if the authentication is successful
        assert_eq!(
            response,
            Response::new().set_data(AuthenticationResult::Authenticated {})
        );
        let query_results =
            query_spend_limit(deps.as_ref(), Addr::unchecked("mock_account")).unwrap();
        if let Some(params_binary) = &auth_request.authenticator_params {
            let authenticator_params =
                serde_json_wasm::from_slice::<AuthenticatorParams>(params_binary).unwrap();

            assert_eq!(query_results.spend_limit_data.id, authenticator_params.id);
            assert_eq!(
                query_results.spend_limit_data.number_of_blocks_active,
                authenticator_params.duration
            );
            assert_eq!(
                query_results.spend_limit_data.block_of_last_tx,
                env.block.height,
            );
        }
        dbg!(query_results);
    }

    #[test]
    fn test_successful_authentication_flow() {
        let mut deps = mock_dependencies_with_balances(&[(
            "mock_account",
            &[
                Coin::new(500, "uosmo"),
                // ATOM
                Coin::new(
                    200,
                    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
                ),
                // USDC
                Coin::new(
                    200,
                    "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
                ),
                Coin::new(200, "memecoin"),
            ],
        )]);
        let mut env = mock_env();
        let info = mock_info("mock_signer", &[]);
        instantiate(deps.as_mut(), env.clone(), info.clone(), InstantiateMsg {}).unwrap();

        let auth_request = create_mock_authentication_request();
        let response = sudo_authenticate(deps.as_mut(), env.clone(), auth_request.clone()).unwrap();

        // Check if the authentication is successful
        assert_eq!(
            response,
            Response::new().set_data(AuthenticationResult::Authenticated {})
        );

        let auth_request = create_mock_authentication_request();
        let response = sudo_authenticate(deps.as_mut(), env.clone(), auth_request.clone()).unwrap();

        // Modify balances directly in the storage
        deps.querier = MockQuerier::new(&[(
            "mock_account",
            &[
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
            ],
        )]);
        env.block.height += 100;

        let auth_request = create_mock_authentication_request();
        let response = sudo_authenticate(deps.as_mut(), env.clone(), auth_request.clone()).unwrap();

        let query_results =
            query_spend_limit(deps.as_ref(), Addr::unchecked("mock_account")).unwrap();
        if let Some(params_binary) = &auth_request.authenticator_params {
            let authenticator_params =
                serde_json_wasm::from_slice::<AuthenticatorParams>(params_binary).unwrap();

            assert_eq!(query_results.spend_limit_data.id, authenticator_params.id);
            assert_eq!(
                query_results.spend_limit_data.number_of_blocks_active,
                authenticator_params.duration
            );
            assert_eq!(
                query_results.spend_limit_data.block_of_last_tx,
                env.block.height,
            );
        }
        dbg!(query_results);
    }

    //fn test_sudo_authenticate_init_params() {
    //    // Arrange
    //    let mut deps = mock_dependencies_with_balances(&[(
    //        "mock_account",
    //        &[
    //            Coin::new(500, "uosmo"),
    //            // ATOM
    //            Coin::new(
    //                200,
    //                "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
    //            ),
    //            // USDC
    //            Coin::new(
    //                200,
    //                "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",
    //            ),
    //            Coin::new(200, "memecoin"),
    //        ],
    //    )]);
    //    let env = mock_env();
    //    let info = mock_info("mock_signer", &[Coin::new(500, "uosmo")]);

    //    let msg = InstantiateMsg {};
    //    let resp = instantiate(deps.as_mut(), env.clone(), info, msg);
    //    match resp {
    //        Ok(response) => {
    //            dbg!(response);
    //        }
    //        Err(err) => {
    //            // Handle the error if needed
    //            panic!("Unexpected error: {:?}", err);
    //        }
    //    }

    //    let auth_request = create_mock_authentication_request();

    //    let result = sudo_authenticate(deps.as_mut(), env.clone(), auth_request);
    //    match result {
    //        Ok(response) => {
    //            dbg!(response);
    //        }
    //        Err(err) => {
    //            // Handle the error if needed
    //            panic!("Unexpected error: {:?}", err);
    //        }
    //    }

    //    let query_results = query_spend_limit(deps.as_ref(), Addr::unchecked("mock_account"));
    //    dbg!(query_results);
    //}

    fn create_mock_authentication_request() -> AuthenticationRequest {
        AuthenticationRequest {
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
        }
    }
}
