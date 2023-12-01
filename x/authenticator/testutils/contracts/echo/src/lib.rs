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
            auth_request.msg.value.try_into()?;

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
        let json = r#"{"authenticate":{"account":"osmo1ne9h72l755k8akh7hvekzgn0revy5r32d0xdvr","msg":{"type_url":"/cosmos.bank.v1beta1.MsgSend","value":"Citvc21vMW5lOWg3Mmw3NTVrOGFraDdodmVremduMHJldnk1cjMyZDB4ZHZyEitvc21vMWg4cmM0eGZqMmVkM3pubDBlejQ1cGR3bXBjcmwzMHo4eGdqcWxtGgwKBG9zbW8SBDI1MDA="},"signature":"vfHWtCEm0Qvxeutkp2GMwpAGT5NGuO/Xj3OebTr0pGdb8DsOCTzgOweXZ8ZV5ZLFgUMSwsEAY4HB67H3r5m5iA==","sign_mode_tx_data":{"sign_mode_direct":"CpABCogBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmgKK29zbW8xbmU5aDcybDc1NWs4YWtoN2h2ZWt6Z24wcmV2eTVyMzJkMHhkdnISK29zbW8xaDhyYzR4ZmoyZWQzem5sMGV6NDVwZHdtcGNybDMwejh4Z2pxbG0aDAoEb3NtbxIEMjUwMBIDc1NQEmQKTgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQMP25Z4ptCW26T2yeMBQnT7fXQohTUV1S1bQ5qYAoww8hIECgIIARISCgwKBG9zbW8SBDI1MDAQ4KcS","sign_mode_textual":""},"tx_data":{"chain_id":"","account_number":0,"sequence":0,"timeout_height":0,"msgs":[{"type_url":"/cosmos.bank.v1beta1.MsgSend","value":"Citvc21vMW5lOWg3Mmw3NTVrOGFraDdodmVremduMHJldnk1cjMyZDB4ZHZyEitvc21vMWg4cmM0eGZqMmVkM3pubDBlejQ1cGR3bXBjcmwzMHo4eGdqcWxtGgwKBG9zbW8SBDI1MDA="}],"memo":"sSP"},"signature_data":{"signers":["osmo1ne9h72l755k8akh7hvekzgn0revy5r32d0xdvr"],"signatures":["vfHWtCEm0Qvxeutkp2GMwpAGT5NGuO/Xj3OebTr0pGdb8DsOCTzgOweXZ8ZV5ZLFgUMSwsEAY4HB67H3r5m5iA=="]},"simulate":false}}"#;
        let sudo_msg: super::SudoMsg = serde_json_wasm::from_str(json).unwrap();
        let super::SudoMsg::Authenticate(auth_request) = sudo_msg;
        assert_eq!(
            auth_request.msg.type_url,
            "/cosmos.bank.v1beta1.MsgSend".to_string()
        );
        println!("{:?}", auth_request);
    }
}
