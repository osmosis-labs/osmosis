#[cfg(test)]
mod tests {
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coin, coins, from_binary, Coin, Deps, DepsMut, Uint128};

    use crate::contract::{execute, instantiate, query};
    use crate::error::ContractError;
    use crate::msg::{
        ExecuteMsg, InstantiateMsg, QueryMsg, ResolveRecordResponse, ReverseResolveRecordResponse,
    };
    use crate::state::Config;

    fn assert_name_owner(deps: Deps, name: &str, owner: &str) {
        let res = query(
            deps,
            mock_env(),
            QueryMsg::ResolveRecord {
                name: name.to_string(),
            },
        )
        .unwrap();

        let value: ResolveRecordResponse = from_binary(&res).unwrap();
        assert_eq!(Some(owner.to_string()), value.address);
    }

    fn assert_address_resolves_to(deps: Deps, name: &str, owner: &str) {
        let res = query(
            deps,
            mock_env(),
            QueryMsg::ReverseResolveRecord {
                address: deps.api.addr_validate(owner).expect("Invalid address"),
            },
        )
        .unwrap();

        let value: ReverseResolveRecordResponse = from_binary(&res).unwrap();
        assert_eq!(Some(name.to_string()), value.name);
    }

    fn assert_config_state(deps: Deps, expected: Config) {
        let res = query(deps, mock_env(), QueryMsg::Config {}).unwrap();
        let value: Config = from_binary(&res).unwrap();
        assert_eq!(value, expected);
    }

    fn mock_init_with_price(
        deps: DepsMut,
        required_denom: impl Into<String>,
        purchase_price: Uint128,
        transfer_price: Uint128,
        annual_rent_bps: Uint128,
    ) {
        let msg = InstantiateMsg {
            required_denom: required_denom.into(),
            purchase_price: purchase_price,
            transfer_price: transfer_price,
            annual_rent_bps: annual_rent_bps,
        };

        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps, mock_env(), info, msg)
            .expect("contract successfully handles InstantiateMsg");
    }

    fn mock_init_no_price(deps: DepsMut) {
        let msg = InstantiateMsg {
            required_denom: "token".to_string(),
            purchase_price: Uint128::from(0 as u128),
            transfer_price: Uint128::from(0 as u128),
            annual_rent_bps: Uint128::from(0 as u128),
        };

        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps, mock_env(), info, msg)
            .expect("contract successfully handles InstantiateMsg");
    }

    fn mock_alice_registers_name(deps: DepsMut, sent: &[Coin]) {
        // alice can register an available name
        let info = mock_info("alice_key", sent);
        let msg = ExecuteMsg::Register {
            name: "alice.ibc".to_string(),
            years: Uint128::from(2 as u128),
        };
        let _res = execute(deps, mock_env(), info, msg)
            .expect("contract successfully handles Register message");
    }

    #[test]
    fn proper_init_no_fees() {
        let mut deps = mock_dependencies();

        mock_init_no_price(deps.as_mut());

        assert_config_state(
            deps.as_ref(),
            Config {
                required_denom: "token".to_string(),
                purchase_price: Uint128::from(0 as u128),
                transfer_price: Uint128::from(0 as u128),
                annual_rent_bps: Uint128::from(0 as u128),
            },
        );
    }

    #[test]
    fn proper_init_with_fees() {
        let mut deps = mock_dependencies();

        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(3 as u128),
            Uint128::from(4 as u128),
            Uint128::from(100 as u128),
        );

        assert_config_state(
            deps.as_ref(),
            Config {
                required_denom: "token".to_string(),
                purchase_price: Uint128::from(3 as u128),
                transfer_price: Uint128::from(4 as u128),
                annual_rent_bps: Uint128::from(100 as u128),
            },
        );
    }

    #[test]
    fn register_available_name_and_query_works() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // querying for name resolves to correct address
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
    }

    #[test]
    fn register_available_name_and_query_works_with_fees_no_rent() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(2 as u128),
            Uint128::from(2 as u128),
            Uint128::from(0 as u128),
        );

        mock_alice_registers_name(deps.as_mut(), &coins(2, "token"));

        // anyone can register an available name with more fees than needed
        let info = mock_info("bob_key", &coins(5, "token"));
        let msg = ExecuteMsg::Register {
            name: "bob.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };

        let _res = execute(deps.as_mut(), mock_env(), info, msg)
            .expect("contract successfully handles Register message");

        // querying for name resolves to correct address
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
        assert_name_owner(deps.as_ref(), "bob.ibc", "bob_key");
    }

    #[test]
    fn register_available_name_and_query_works_with_fees_and_rent() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(200 as u128),
            Uint128::from(200 as u128),
            Uint128::from(100 as u128),
        );

        mock_alice_registers_name(deps.as_mut(), &coins(204, "token"));

        // anyone can register an available name with more fees than needed
        let info = mock_info("bob_key", &coins(500, "token"));
        let msg = ExecuteMsg::Register {
            name: "bob.ibc".to_string(),
            years: Uint128::from(3 as u128),
        };

        let _res = execute(deps.as_mut(), mock_env(), info, msg)
            .expect("contract successfully handles Register message");

        // querying for name resolves to correct address
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
        assert_name_owner(deps.as_ref(), "bob.ibc", "bob_key");
    }

    #[test]
    fn fails_on_register_already_taken_name() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // bob can't register the same name
        let info = mock_info("bob_key", &coins(2, "token"));
        let msg = ExecuteMsg::Register {
            name: "alice.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };
        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameTaken { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }
        // alice can't register the same name again
        let info = mock_info("alice_key", &coins(2, "token"));
        let msg = ExecuteMsg::Register {
            name: "alice.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };
        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameTaken { .. }) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        }
    }

    #[test]
    fn register_available_name_fails_with_invalid_name() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        let info = mock_info("bob_key", &coins(2, "token"));

        // hi is too short, only two characters
        let msg = ExecuteMsg::Register {
            name: "hi.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };
        match execute(deps.as_mut(), mock_env(), info.clone(), msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameTooShort { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }

        // needs .ibc suffix
        let msg = ExecuteMsg::Register {
            name: "alice".to_string(),
            years: Uint128::from(1 as u128),
        };
        match execute(deps.as_mut(), mock_env(), info.clone(), msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameNeedsSuffix { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }

        // no other suffix for now
        let msg = ExecuteMsg::Register {
            name: "alice.eth".to_string(),
            years: Uint128::from(1 as u128),
        };
        match execute(deps.as_mut(), mock_env(), info.clone(), msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameNeedsSuffix { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }

        // 65 chars is too long
        let msg = ExecuteMsg::Register {
            years: Uint128::from(1 as u128),
            name: "01234567890123456789012345678901234567890123456789012345678901234.ibc"
                .to_string(),
        };
        match execute(deps.as_mut(), mock_env(), info.clone(), msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameTooLong { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }

        // no upper case...
        let msg = ExecuteMsg::Register {
            name: "LOUD.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };
        match execute(deps.as_mut(), mock_env(), info.clone(), msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::InvalidCharacter { c }) => assert_eq!(c, 'L'),
            Err(_) => panic!("Unknown error"),
        }
        // ... or spaces
        let msg = ExecuteMsg::Register {
            name: "two words.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };
        match execute(deps.as_mut(), mock_env(), info, msg) {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::InvalidCharacter { .. }) => {}
            Err(_) => panic!("Unknown error"),
        }
    }

    #[test]
    fn fails_on_register_insufficient_fees() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(2 as u128),
            Uint128::from(2 as u128),
            Uint128::from(0 as u128),
        );

        // anyone can register an available name with sufficient fees
        let info = mock_info("alice_key", &[]);
        let msg = ExecuteMsg::Register {
            name: "alice.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };

        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("register call should fail with insufficient fees"),
            Err(ContractError::InsufficientFundsSent {}) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        }
    }

    #[test]
    fn fails_on_register_wrong_fee_denom() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(2 as u128),
            Uint128::from(2 as u128),
            Uint128::from(0 as u128),
        );

        // anyone can register an available name with sufficient fees
        let info = mock_info("alice_key", &coins(2, "earth"));
        let msg = ExecuteMsg::Register {
            name: "alice.ibc".to_string(),
            years: Uint128::from(1 as u128),
        };

        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("register call should fail with insufficient fees"),
            Err(ContractError::InsufficientFundsSent {}) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        }
    }

    #[test]
    fn transfer_works() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // alice can transfer her name successfully to bob
        let info = mock_info("alice_key", &[]);
        let msg = ExecuteMsg::Transfer {
            name: "alice.ibc".to_string(),
            to: "bob_key".to_string(),
        };

        let _res = execute(deps.as_mut(), mock_env(), info, msg)
            .expect("contract successfully handles Transfer message");
        // querying for name resolves to correct address (bob_key)
        assert_name_owner(deps.as_ref(), "alice.ibc", "bob_key");
    }

    #[test]
    fn transfer_works_with_fees() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(2 as u128),
            Uint128::from(2 as u128),
            Uint128::from(0 as u128),
        );
        mock_alice_registers_name(deps.as_mut(), &coins(2, "token"));

        // alice can transfer her name successfully to bob
        let info = mock_info("alice_key", &[coin(1, "earth"), coin(2, "token")]);
        let msg = ExecuteMsg::Transfer {
            name: "alice.ibc".to_string(),
            to: "bob_key".to_string(),
        };

        let _res = execute(deps.as_mut(), mock_env(), info, msg)
            .expect("contract successfully handles Transfer message");
        // querying for name resolves to correct address (bob_key)
        assert_name_owner(deps.as_ref(), "alice.ibc", "bob_key");
    }

    #[test]
    fn fails_on_transfer_non_existent() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // alice can transfer her name successfully to bob
        let info = mock_info("frank_key", &coins(2, "token"));
        let msg = ExecuteMsg::Transfer {
            name: "alice42".to_string(),
            to: "bob_key".to_string(),
        };

        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::NameNotExists { name }) => assert_eq!(name, "alice42"),
            Err(e) => panic!("Unexpected error: {:?}", e),
        }

        // querying for name resolves to correct address (alice_key)
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
    }

    #[test]
    fn fails_on_transfer_from_nonowner() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // alice can transfer her name successfully to bob
        let info = mock_info("frank_key", &coins(2, "token"));
        let msg = ExecuteMsg::Transfer {
            name: "alice.ibc".to_string(),
            to: "bob_key".to_string(),
        };

        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("Must return error"),
            Err(ContractError::Unauthorized { .. }) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        }

        // querying for name resolves to correct address (alice_key)
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
    }

    #[test]
    fn fails_on_transfer_insufficient_fees() {
        let mut deps = mock_dependencies();
        mock_init_with_price(
            deps.as_mut(),
            "token",
            Uint128::from(2 as u128),
            Uint128::from(5 as u128),
            Uint128::from(0 as u128),
        );
        mock_alice_registers_name(deps.as_mut(), &coins(2, "token"));

        // alice can transfer her name successfully to bob
        let info = mock_info("alice_key", &[coin(1, "earth"), coin(2, "token")]);
        let msg = ExecuteMsg::Transfer {
            name: "alice.ibc".to_string(),
            to: "bob_key".to_string(),
        };

        let res = execute(deps.as_mut(), mock_env(), info, msg);

        match res {
            Ok(_) => panic!("register call should fail with insufficient fees"),
            Err(ContractError::InsufficientFundsSent {}) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        }

        // querying for name resolves to correct address (bob_key)
        assert_name_owner(deps.as_ref(), "alice.ibc", "alice_key");
    }

    #[test]
    fn returns_empty_on_query_unregistered_name() {
        let mut deps = mock_dependencies();

        mock_init_no_price(deps.as_mut());

        // querying for unregistered name results in NotFound error
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::ResolveRecord {
                name: "alice.ibc".to_string(),
            },
        )
        .unwrap();
        let value: ResolveRecordResponse = from_binary(&res).unwrap();
        assert_eq!(None, value.address);
    }

    #[test]
    fn set_name_works() {
        let mut deps = mock_dependencies();
        mock_init_no_price(deps.as_mut());
        mock_alice_registers_name(deps.as_mut(), &[]);

        // alice can successfully set her name
        let info = mock_info("alice_key", &[]);
        let msg = ExecuteMsg::SetName {
            name: "alice.ibc".to_string(),
        };

        let _res = execute(deps.as_mut(), mock_env(), info, msg)
            .expect("contract successfully handles SetName message");
        // querying for address (alice_key) resolves to correct name (alice)
        assert_address_resolves_to(deps.as_ref(), "alice.ibc", "alice_key");
    }
}
