use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, DepsMut, Env, MessageInfo, Response, StdError};
use cw_storage_plus::Item;
use osmosis_authenticators::AuthenticationResult;
use sha2::{Digest, Sha256};

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
    let hash = osmosis_authenticators::sha256(auth_request.sign_mode_tx_data.sign_mode_direct);
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
