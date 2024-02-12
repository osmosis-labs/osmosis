use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, DepsMut, Env, MessageInfo, Response, StdError};
use cw_storage_plus::Item;
use osmosis_authenticators::AuthenticationResult;

#[cw_serde]
pub struct InstantiateMsg {
    pubkey: Binary,
}

#[cw_serde]
pub enum SudoMsg {
    Authenticate(osmosis_authenticators::AuthenticationRequest),
}

// State
pub const PUBKEY: Item<Binary> = Item::new("pubkey");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, StdError> {
    PUBKEY.save(deps.storage, &msg.pubkey)?;
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(
    deps: DepsMut,
    _env: Env,
    SudoMsg::Authenticate(auth_request): SudoMsg,
) -> Result<Response, StdError> {
    deps.api.debug(&format!("auth_request {:?}", auth_request));
    if auth_request.msg.type_url == "/cosmos.bank.v1beta1.MsgSend" {
        let send: osmosis_std::types::cosmos::bank::v1beta1::MsgSend =
            auth_request.msg.bytes.try_into()?;

        deps.api.debug(&format!("send {:?}", send));
    }

    let pubkey = PUBKEY.load(deps.storage)?;

    // verify the signature
    let hash = osmosis_authenticators::sha256(&auth_request.sign_mode_tx_data.sign_mode_direct);
    let valid = deps
        .api
        .secp256k1_verify(&hash, &auth_request.signature, &pubkey)
        .or_else(|e| {
            deps.api.debug(&format!("error {:?}", e));
            Err(StdError::generic_err("Failed to verify signature"))
        })?;

    if !valid {
        return Ok(Response::new().set_data(AuthenticationResult::NotAuthenticated {}));
    }

    Ok(Response::new().set_data(AuthenticationResult::Authenticated {}))
}

// Test that SudoMsg can be deserialized from an expected json
#[cfg(test)]
mod tests {
    #[test]
    fn test_deserialize_sudo_msg() {
        let json =   r#"{"authenticate":{"account":"osmo1cj09htxwky4lgcjwsk669v5rw8z7vhsw07a6m0","msg":{"type_url":"/cosmos.bank.v1beta1.MsgSend","value":"eyJmcm9tX2FkZHJlc3MiOiJvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wIiwidG9fYWRkcmVzcyI6Im9zbW8xZm56ZTdmMjJnOGhocGg2eTM3dWNtM2wzY3phcnZsMmRxc2g1emwiLCJhbW91bnQiOlt7ImRlbm9tIjoib3NtbyIsImFtb3VudCI6IjI1MDAifV19","bytes":"Citvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wEitvc21vMWZuemU3ZjIyZzhoaHBoNnkzN3VjbTNsM2N6YXJ2bDJkcXNoNXpsGgwKBG9zbW8SBDI1MDA="},"signature":"5pfbWlX1edoW7Yx7EtDuwr0V40WCyber3mOM2Do6tlwL8IYy9Vek9Y+YVc8a8rpabHwT+DXosF0Juj5AdmKVOw==","sign_mode_tx_data":{"sign_mode_direct":"CuwBCogBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmgKK29zbW8xY2owOWh0eHdreTRsZ2Nqd3NrNjY5djVydzh6N3Zoc3cwN2E2bTASK29zbW8xZm56ZTdmMjJnOGhocGg2eTM3dWNtM2wzY3phcnZsMmRxc2g1emwaDAoEb3NtbxIEMjUwMBJfamtXbmtyRFJESVJ1eUl2TmtTdHJSemZlcUN2TmZnYWlia3htWGJXYkxnZ09jalFEU01nZU1WZ1hpVHZycWdlVElFTFBpaEpRUWJtUFJXVHhsZnJLclRzT3JHdFdLV2UStAEKTgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQOPplsdZhgi/PR8TuFqrikSqCzFa6tjORiNJigrhFanshIECgIIAQpOCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohA7ueR+LxfTrjeBPX1+iV6mkjP3Vsu27WBAiHjHfnouKSEgQKAggBEhIKDAoEb3NtbxIEMjUwMBDgpxI=","sign_mode_textual":""},"tx_data":{"chain_id":"","account_number":0,"sequence":0,"timeout_height":0,"msgs":[{"type_url":"/cosmos.bank.v1beta1.MsgSend","value":"eyJmcm9tX2FkZHJlc3MiOiJvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wIiwidG9fYWRkcmVzcyI6Im9zbW8xZm56ZTdmMjJnOGhocGg2eTM3dWNtM2wzY3phcnZsMmRxc2g1emwiLCJhbW91bnQiOlt7ImRlbm9tIjoib3NtbyIsImFtb3VudCI6IjI1MDAifV19","bytes":"Citvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wEitvc21vMWZuemU3ZjIyZzhoaHBoNnkzN3VjbTNsM2N6YXJ2bDJkcXNoNXpsGgwKBG9zbW8SBDI1MDA="}],"memo":"jkWnkrDRDIRuyIvNkStrRzfeqCvNfgaibkxmXbWbLggOcjQDSMgeMVgXiTvrqgeTIELPihJQQbmPRWTxlfrKrTsOrGtWKWe"},"signature_data":{"signers":["osmo1cj09htxwky4lgcjwsk669v5rw8z7vhsw07a6m0"],"signatures":["5pfbWlX1edoW7Yx7EtDuwr0V40WCyber3mOM2Do6tlwL8IYy9Vek9Y+YVc8a8rpabHwT+DXosF0Juj5AdmKVOw=="]},"simulate":false}}"#;
        let sudo_msg: super::SudoMsg = serde_json_wasm::from_str(json).unwrap();
        let super::SudoMsg::Authenticate(auth_request) = sudo_msg;
        assert_eq!(
            auth_request.msg.type_url,
            "/cosmos.bank.v1beta1.MsgSend".to_string()
        );
        println!("{:?}", auth_request);
    }
}
