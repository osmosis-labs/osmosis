use cosmwasm_std::{
    AllBalanceResponse, Coin, Decimal, Deps, DepsMut, Env, Response, StdError, StdResult, Uint128,
};
use osmosis_authenticators::{ConfirmExecutionRequest, ConfirmationResult};

use crate::bank::query_account_balances;
use crate::state::{SpendLimit, SPEND_LIMITS, TRACKED_DENOMS};
use crate::twap::calculate_price_from_route;
use crate::ContractError;

pub fn sudo_confirm_execution(
    deps: DepsMut,
    env: Env,
    confirm_execution_request: ConfirmExecutionRequest,
) -> Result<Response, ContractError> {
    deps.api.debug(&format!(
        "confirm_execution_request {:?}",
        confirm_execution_request
    ));

    let account_address = &confirm_execution_request.account;
    let balances = query_account_balances(deps.as_ref(), account_address)?;

    let spend_limit = match SPEND_LIMITS.may_load(deps.storage, account_address.to_string())? {
        Some(spend_limit) => spend_limit,
        // XXX; here we return confirm because in the post handler confirm_execution is
        // called before sudo_authenticate
        None => return Ok(Response::new().set_data(ConfirmationResult::Confirm {})),
    };

    //dbg!(spend_limit.clone());
    let delta = calculate_delta(&spend_limit.balance, &balances, deps.as_ref(), &env)?;
    dbg!(delta);
    let updated_spend_limit = SpendLimit {
        id: spend_limit.id.clone(),
        denom: spend_limit.denom.clone(),
        amount_left: spend_limit.amount_left.saturating_sub(delta),
        balance: balances,
        block_of_last_tx: env.block.height + 1,
        number_of_blocks_active: spend_limit.number_of_blocks_active,
    };
    //dbg!(updated_spend_limit.clone());

    SPEND_LIMITS.save(
        deps.storage,
        account_address.to_string(),
        &updated_spend_limit,
    )?;

    Ok(Response::new().set_data(ConfirmationResult::Confirm {}))
}

fn calculate_delta(
    prev_balances: &AllBalanceResponse,
    balances: &AllBalanceResponse,
    deps: Deps,
    env: &Env,
) -> Result<u128, ContractError> {
    let delta: u128 = 0;
    for spend_coin in &prev_balances.amount {
        if let Some(balance_coin) = balances
            .amount
            .iter()
            .find(|coin| coin.denom == spend_coin.denom)
        {
            if let Some(coin_delta) = process_coin_delta(spend_coin, balance_coin, deps, env)? {
                dbg!(coin_delta);
                delta.checked_add(coin_delta).ok_or_else(|| {
                    StdError::generic_err("Delta calculation resulted in overflow")
                })?;
                dbg!(delta);
            }
        }
    }
    Ok(delta)
}

fn process_coin_delta(
    spend_coin: &Coin,
    balance_coin: &Coin,
    deps: Deps,
    env: &Env,
) -> Result<Option<u128>, ContractError> {
    if balance_coin.amount != spend_coin.amount {
        let is_tracked = TRACKED_DENOMS.has(deps.storage, spend_coin.denom.clone());
        if !is_tracked {
            return Ok(None);
        }

        let tracked_denom = TRACKED_DENOMS.load(deps.storage, spend_coin.denom.clone())?;
        let path = tracked_denom.path;

        // here we call the TwapQuerier
        let coin = calculate_price_from_route(
            deps,
            spend_coin.clone(),
            env.block.time,
            None,
            Decimal::from_ratio(Uint128::new(1), Uint128::new(8)),
            path,
        )?;

        dbg!(coin.clone());
        let coin_delta = calculate_coin_delta(&coin, spend_coin, balance_coin)?;
        Ok(Some(coin_delta))
    } else {
        Ok(None)
    }
}

fn calculate_coin_delta(coin: &Coin, spend_coin: &Coin, balance_coin: &Coin) -> StdResult<u128> {
    if balance_coin.amount < spend_coin.amount {
        Ok(coin.amount.u128())
    } else {
        Ok(0)
    }
}

#[cfg(test)]
mod tests {
    use super::sudo_confirm_execution;
    use crate::authenticate::sudo_authenticate;
    use crate::ContractError;

    use crate::contract::{instantiate, query_spend_limit};
    use crate::msg::InstantiateMsg;
    use cosmwasm_std::testing::{
        mock_dependencies_with_balances, mock_env, mock_info, BankQuerier, MockApi, MockQuerier,
        MockQuerierCustomHandlerResult, MockStorage,
    };
    use cosmwasm_std::Empty;
    use cosmwasm_std::{
        from_json, to_json_binary, Addr, Binary, Coin, ContractResult, CustomQuery, OwnedDeps,
        Querier, QuerierResult, QuerierWrapper, QueryRequest, SystemError, SystemResult,
    };
    use serde::de::DeserializeOwned;

    use core::marker::PhantomData;
    use osmosis_authenticators::{
        Any, AuthenticationRequest, AuthenticationResult, ConfirmExecutionRequest,
        ConfirmationResult, SignModeTxData, SignatureData, TxData,
    };
    use osmosis_std::types::osmosis::twap::v1beta1::{
        ArithmeticTwapRequest, ArithmeticTwapResponse, GeometricTwapRequest, TwapQuerier,
    };

    #[test]
    fn test_sudo_confirm_execution() {
        // Arrange
        let mut deps = mock_custom_dependencies_with_balances(&[(
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
        let mut info = mock_info("mock_signer", &[Coin::new(500, "uosmo")]);
        let msg = InstantiateMsg {};
        let resp = instantiate(deps.as_mut(), env.clone(), info, msg);

        let auth_request = create_mock_authentication_request();

        let result = sudo_authenticate(deps.as_mut(), env.clone(), auth_request);
        let query_results = query_spend_limit(deps.as_ref(), Addr::unchecked("mock_account"));
        dbg!(query_results);

        match result {
            Ok(response) => {
                dbg!(response);
            }
            Err(err) => {
                // Handle the error if needed
                panic!("Unexpected error: {:?}", err);
            }
        }

        // Modify balances directly in the storage
        //deps.querier
        let querier: MockTWAPQuerier<Empty> = MockTWAPQuerier::new(&[(
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

        env.block.height = 12_346;
        deps.querier = querier;

        let confirm_execution_request = ConfirmExecutionRequest {
            account: Addr::unchecked("mock_account"),
            msg: Any {
                type_url: "cosmwasm/std/Msg".to_string(),
                value: Binary::from(b"mock_msg_value".to_vec()),
            },
        };
        let result = sudo_confirm_execution(deps.as_mut(), env.clone(), confirm_execution_request);
        match result {
            Ok(response) => {
                dbg!(response);
            }
            Err(err) => {
                // Handle the error if needed
                panic!("Unexpected error: {:?}", err);
            }
        }

        let query_results = query_spend_limit(deps.as_ref(), Addr::unchecked("mock_account"));
        dbg!(query_results);

        //let resultConfirm =
        //    sudo_confirm_execution(deps.as_mut(), env.clone(), confirm_execution_request);

        //println!("{:?}", resultConfirm);
    }

    fn mock_custom_dependencies_with_balances(
        balances: &[(&str, &[Coin])],
    ) -> OwnedDeps<MockStorage, MockApi, MockTWAPQuerier> {
        OwnedDeps {
            storage: MockStorage::default(),
            api: MockApi::default(),
            querier: MockTWAPQuerier::new(balances),
            custom_query_type: PhantomData,
        }
    }

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

    struct MockTWAPQuerier<C: DeserializeOwned = Empty> {
        bank: BankQuerier,
        custom_handler: PhantomData<C>,
    }

    impl<C: DeserializeOwned> MockTWAPQuerier<C> {
        pub fn new(balances: &[(&str, &[Coin])]) -> Self {
            MockTWAPQuerier {
                bank: BankQuerier::new(balances),
                custom_handler: PhantomData,
            }
        }
    }

    impl<C: CustomQuery + DeserializeOwned> Querier for MockTWAPQuerier<C> {
        fn raw_query(&self, bin_request: &[u8]) -> QuerierResult {
            let request: QueryRequest<C> = match from_json(bin_request) {
                Ok(v) => v,
                Err(e) => {
                    return SystemResult::Err(SystemError::InvalidRequest {
                        error: format!("Parsing query request: {e}"),
                        request: bin_request.into(),
                    })
                }
            };
            self.handle_query(&request)
        }
    }

    impl<C: CustomQuery + DeserializeOwned> MockTWAPQuerier<C> {
        pub fn handle_query(&self, request: &QueryRequest<C>) -> QuerierResult {
            match &request {
                QueryRequest::Bank(bank_query) => self.bank.query(bank_query),
                QueryRequest::Stargate { path, data } => {
                    // Create the response object.
                    let response = ArithmeticTwapResponse {
                        arithmetic_twap: "2.00".to_string(),
                    };
                    SystemResult::Ok(to_json_binary(&response).into())
                }
                _ => {
                    // Handle the unmatched case
                    SystemResult::Err(SystemError::UnsupportedRequest {
                        kind: "Unsupported".to_string(),
                    })
                }
            }
        }
    }
}
