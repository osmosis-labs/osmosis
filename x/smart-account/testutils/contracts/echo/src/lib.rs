use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    from_json, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError,
};
use cw_storage_plus::Item;
use osmosis_authenticators::{
    AuthenticationRequest, ConfirmExecutionRequest, OnAuthenticatorAddedRequest,
    OnAuthenticatorRemovedRequest, TrackRequest,
};

#[cw_serde]
pub struct InstantiateMsg {
    pubkey: Binary,
}

#[cw_serde]
pub enum SudoMsg {
    OnAuthenticatorAdded(OnAuthenticatorAddedRequest),
    OnAuthenticatorRemoved(OnAuthenticatorRemovedRequest),
    Authenticate(AuthenticationRequest),
    Track(TrackRequest),
    ConfirmExecution(ConfirmExecutionRequest),
}

#[cw_serde]
pub enum QueryMsg {
    LatestSudoCall {},
}

// State
pub const PUBKEY: Item<Binary> = Item::new("pubkey");

// Tracking latest sudo call for testing purposes, acting like spy test double
pub const LATEST_SUDO_CALL: Item<SudoMsg> = Item::new("latest_sudo_call");

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
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, StdError> {
    // track latest sudo call for testing purposes
    LATEST_SUDO_CALL.save(deps.storage, &msg)?;

    match msg {
        SudoMsg::OnAuthenticatorAdded(on_authenticator_added_request) => {
            on_authenticator_added(deps, on_authenticator_added_request)
        }
        SudoMsg::OnAuthenticatorRemoved(on_authenticator_removed_request) => {
            on_authenticator_removed(deps, on_authenticator_removed_request)
        }
        SudoMsg::Authenticate(auth_request) => authenticate(deps, auth_request),
        SudoMsg::Track(track_request) => track(deps, track_request),
        SudoMsg::ConfirmExecution(_) => Ok(Response::new()),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(
    deps: Deps,
    _env: Env,
    QueryMsg::LatestSudoCall {}: QueryMsg,
) -> Result<Binary, StdError> {
    let sudo_msg: SudoMsg = LATEST_SUDO_CALL.load(deps.storage)?;
    to_json_binary(&sudo_msg)
}

#[cw_serde]
struct Params {
    label: String,
}

fn on_authenticator_added(
    _deps: DepsMut,
    OnAuthenticatorAddedRequest {
        account: _,
        authenticator_id: _,
        authenticator_params,
    }: OnAuthenticatorAddedRequest,
) -> Result<Response, StdError> {
    // validate params structure
    let _params: Params = from_json(
        authenticator_params
            .ok_or_else(|| StdError::generic_err("missing authenticator_params"))?
            .as_slice(),
    )?;

    Ok(Response::new())
}

fn on_authenticator_removed(
    _deps: DepsMut,
    OnAuthenticatorRemovedRequest {
        account: _,
        authenticator_id: _,
        authenticator_params,
    }: OnAuthenticatorRemovedRequest,
) -> Result<Response, StdError> {
    // validate params structure
    let _params: Params = from_json(
        authenticator_params
            .ok_or_else(|| StdError::generic_err("missing authenticator_params"))?
            .as_slice(),
    )?;

    Ok(Response::new())
}

fn authenticate(deps: DepsMut, auth_request: AuthenticationRequest) -> Result<Response, StdError> {
    deps.api.debug(&format!("auth_request {:?}", auth_request));
    if auth_request.msg.type_url == "/cosmos.bank.v1beta1.MsgSend" {
        let send: osmosis_std::types::cosmos::bank::v1beta1::MsgSend =
            auth_request.msg.value.try_into()?;

        deps.api.debug(&format!("send {:?}", send));
    }

    let pubkey = PUBKEY.load(deps.storage)?;

    // verify the signature
    let hash = osmosis_authenticators::sha256(&auth_request.sign_mode_tx_data.sign_mode_direct);
    deps.api
        .secp256k1_verify(&hash, &auth_request.signature, &pubkey)
        .or_else(|e| {
            deps.api.debug(&format!("error {:?}", e));
            Err(StdError::generic_err("Failed to verify signature"))
        })?;

    Ok(Response::new())
}

// Track is a no-op
fn track(_deps: DepsMut, _track_request: TrackRequest) -> Result<Response, StdError> {
    Ok(Response::new())
}
// Test that SudoMsg can be deserialized from an expected json
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deserialize_sudo_msg() {
        let json = r#"{"authenticate":{"authenticator_id": "0", "account":"osmo1cj09htxwky4lgcjwsk669v5rw8z7vhsw07a6m0","msg":{"type_url":"/cosmos.bank.v1beta1.MsgSend", "value":"Citvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wEitvc21vMWZuemU3ZjIyZzhoaHBoNnkzN3VjbTNsM2N6YXJ2bDJkcXNoNXpsGgwKBG9zbW8SBDI1MDA="},"signature":"5pfbWlX1edoW7Yx7EtDuwr0V40WCyber3mOM2Do6tlwL8IYy9Vek9Y+YVc8a8rpabHwT+DXosF0Juj5AdmKVOw==","sign_mode_tx_data":{"sign_mode_direct":"CuwBCogBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmgKK29zbW8xY2owOWh0eHdreTRsZ2Nqd3NrNjY5djVydzh6N3Zoc3cwN2E2bTASK29zbW8xZm56ZTdmMjJnOGhocGg2eTM3dWNtM2wzY3phcnZsMmRxc2g1emwaDAoEb3NtbxIEMjUwMBJfamtXbmtyRFJESVJ1eUl2TmtTdHJSemZlcUN2TmZnYWlia3htWGJXYkxnZ09jalFEU01nZU1WZ1hpVHZycWdlVElFTFBpaEpRUWJtUFJXVHhsZnJLclRzT3JHdFdLV2UStAEKTgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQOPplsdZhgi/PR8TuFqrikSqCzFa6tjORiNJigrhFanshIECgIIAQpOCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohA7ueR+LxfTrjeBPX1+iV6mkjP3Vsu27WBAiHjHfnouKSEgQKAggBEhIKDAoEb3NtbxIEMjUwMBDgpxI=","sign_mode_textual":""},"tx_data":{"chain_id":"","account_number":0,"sequence":0,"timeout_height":0,"msgs":[{"type_url":"/cosmos.bank.v1beta1.MsgSend","value":"Citvc21vMWNqMDlodHh3a3k0bGdjandzazY2OXY1cnc4ejd2aHN3MDdhNm0wEitvc21vMWZuemU3ZjIyZzhoaHBoNnkzN3VjbTNsM2N6YXJ2bDJkcXNoNXpsGgwKBG9zbW8SBDI1MDA="}],"memo":"jkWnkrDRDIRuyIvNkStrRzfeqCvNfgaibkxmXbWbLggOcjQDSMgeMVgXiTvrqgeTIELPihJQQbmPRWTxlfrKrTsOrGtWKWe"},"signature_data":{"signers":["osmo1cj09htxwky4lgcjwsk669v5rw8z7vhsw07a6m0"],"signatures":["5pfbWlX1edoW7Yx7EtDuwr0V40WCyber3mOM2Do6tlwL8IYy9Vek9Y+YVc8a8rpabHwT+DXosF0Juj5AdmKVOw=="]},"simulate":false}}"#;
        let sudo_msg: super::SudoMsg = serde_json_wasm::from_str(json).unwrap();
        if let SudoMsg::Authenticate(auth_request) = sudo_msg {
            assert_eq!(
                auth_request.msg.type_url,
                "/cosmos.bank.v1beta1.MsgSend".to_string()
            );
            println!("{:?}", auth_request);
        } else {
            panic!("unexpected sudo_msg");
        }
    }
}
