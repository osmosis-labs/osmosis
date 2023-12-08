use crate::contract::sv::multitest_utils::CodeId;
use crate::contract::SudoMsg;
use crate::test_utils::{generate_keys_and_sign, sign_message};
use crate::types::Pubkey;
use cosmwasm_std::{from_json, to_json_binary, Addr, Binary};
use osmosis_authenticators as oa;
use sylvia::multitest::App;

const OWNER: &str = "owner";

#[test]
fn test_basic_instantiation() {
    let app = App::default();
    let code_id = CodeId::store_code(&app);

    let contract = code_id
        .instantiate(vec![])
        .with_label("Contract")
        .with_admin(Some(OWNER))
        .call(OWNER)
        .unwrap();

    let keys = contract.pubkeys().unwrap().pubkeys;
    assert_eq!(keys.len(), 0);
}

#[test]
fn test_instantiation_raw() {
    let app = App::default();
    let code_id = CodeId::store_code(&app);

    let pubkeys = vec![
        Pubkey::Raw(Binary::from("pubkey1".as_bytes())),
        Pubkey::Raw(Binary::from("pubkey2".as_bytes())),
    ];

    let contract = code_id
        .instantiate(pubkeys.clone())
        .with_label("Contract")
        .call(OWNER)
        .unwrap();

    let keys = contract.pubkeys().unwrap().pubkeys;
    assert_eq!(keys, pubkeys);
}

#[test]
fn test_instantiation_mixed() {
    let app = App::default();
    let code_id = CodeId::store_code(&app);

    let pubkeys = vec![
        Pubkey::ByName("cosigner".to_string()),
        Pubkey::Raw(Binary::from("pubkey2".as_bytes())),
    ];

    let contract = code_id
        .instantiate(pubkeys.clone())
        .with_label("Contract")
        .call(OWNER)
        .unwrap();

    let keys = contract.pubkeys().unwrap().pubkeys;
    assert_eq!(keys, pubkeys);
}

#[test]
fn test_successful_authentication() {
    let app = App::default();
    let code_id = CodeId::store_code(&app);

    let message = "message";
    let (_, pubkey1, sig1) = generate_keys_and_sign(&[0x1; 32], message);
    let (_, pubkey2, sig2) = generate_keys_and_sign(&[0x2; 32], message);

    let pubkeys = vec![Pubkey::Raw(pubkey1), Pubkey::Raw(pubkey2)];

    let contract = code_id
        .instantiate(pubkeys.clone())
        .with_label("Contract")
        .call(OWNER)
        .unwrap();

    let sigs: Vec<Binary> = vec![sig1.clone(), sig2.clone()];
    let compound_sig: Binary = to_json_binary(&sigs).unwrap();

    let mut auth_request = oa::AuthenticationRequest {
        signature: compound_sig.clone(),
        sign_mode_tx_data: oa::SignModeTxData {
            sign_mode_direct: Binary::from(message.as_bytes()),
            sign_mode_textual: None,
        },
        account: Addr::unchecked("account"),
        msg: oa::Any {
            type_url: "".to_string(),
            value: Binary::from("msg".as_bytes()),
        },
        tx_data: oa::TxData {
            chain_id: "chain_id".to_string(),
            account_number: 1,
            sequence: 1,
            timeout_height: 1,
            msgs: vec![],
            memo: "".to_string(),
        },
        signature_data: oa::SignatureData {
            signers: vec![],
            signatures: vec![compound_sig.to_string()],
        },
        simulate: false,
    };

    //let result = contract.sudo(SudoMsg::Authenticate(auth_request)).unwrap();
    let msg = SudoMsg::Authenticate(auth_request.clone());
    let result = contract
        .app
        .app_mut()
        .wasm_sudo(contract.contract_addr.clone(), &msg)
        .unwrap();

    let auth_result: oa::AuthenticationResult = from_json(result.data.unwrap()).unwrap();
    println!("result: {:?}", auth_result);

    // Check if the result is Authenticated
    assert!(matches!(
        auth_result,
        oa::AuthenticationResult::Authenticated {}
    ));

    // Modify the signatures to be something invalid
    let sigs: Vec<Binary> = vec![sig1.clone(), sig1.clone()];
    let compound_sig: Binary = to_json_binary(&sigs).unwrap();
    auth_request.signature = to_json_binary(&sigs).unwrap();
    auth_request.signature_data.signatures = vec![compound_sig.to_string()];

    let msg = SudoMsg::Authenticate(auth_request.clone());
    let result = contract
        .app
        .app_mut()
        .wasm_sudo(contract.contract_addr.clone(), &msg)
        .unwrap();

    let auth_result: oa::AuthenticationResult = from_json(result.data.unwrap()).unwrap();
    println!("result: {:?}", auth_result);

    // Check if the result is Authenticated
    assert!(matches!(
        auth_result,
        oa::AuthenticationResult::NotAuthenticated {}
    ));

    // Now let's use an invalid signature. This should be an error.
    let sigs: Vec<Binary> = vec![sig1.clone(), "invalid".as_bytes().into()];
    let compound_sig: Binary = to_json_binary(&sigs).unwrap();
    auth_request.signature = to_json_binary(&sigs).unwrap();
    auth_request.signature_data.signatures = vec![compound_sig.to_string()];

    let msg = SudoMsg::Authenticate(auth_request.clone());
    contract
        .app
        .app_mut()
        .wasm_sudo(contract.contract_addr.clone(), &msg)
        .unwrap_err();
}
