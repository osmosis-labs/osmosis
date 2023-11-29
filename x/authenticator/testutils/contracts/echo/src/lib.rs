use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, Binary, DepsMut, Env, MessageInfo, Response, StdError};
use cw_storage_plus::Item;
use sha2::{Digest, Sha256};

#[cw_serde]
pub struct InstantiateMsg {
    pubkey: Binary,
}

#[cw_serde]
pub enum SudoMsg {
    Authenticate(AuthenticationRequest),
}

// TODO: Move these definitions to a package

#[cw_serde]
pub struct Any {
    pub type_url: String,
    pub value: cosmwasm_std::Binary,
}

#[cw_serde]
pub struct AuthenticationRequest {
    pub account: Addr,
    pub msg: Any,
    pub signature: Binary,
    pub sign_mode_tx_data: SignModeTxData,
    pub tx_data: TxData,
    pub signature_data: SignatureData,
}

#[cw_serde]
pub struct SignModeTxData {
    pub sign_mode_direct: Binary,
    pub sign_mode_textual: Option<String>, // Assuming it's a string or null
}

#[cw_serde]
pub struct TxData {
    pub chain_id: String,
    pub account_number: u64,
    pub sequence: u64,
    pub timeout_height: u64,
    pub msgs: Vec<Any>,
    pub memo: String,
    pub simulate: bool,
}

#[cw_serde]
pub struct SignatureData {
    pub signers: Vec<Addr>,
    pub signatures: Vec<String>,
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

#[cw_serde]
#[serde(tag = "type", content = "content")]
enum AuthenticationResult {
    Authenticated,
    NotAuthenticated,
    Rejected { msg: String },
}

impl Into<cosmwasm_std::Binary> for AuthenticationResult {
    fn into(self) -> cosmwasm_std::Binary {
        cosmwasm_std::Binary::from(
            serde_json_wasm::to_string(&self)
                .expect("Failed to serialize AuthenticationResult")
                .as_bytes()
                .to_vec(),
        )
    }
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

    // Re-verify the signature
    let mut hasher = Sha256::new();
    hasher.update(auth_request.sign_mode_tx_data.sign_mode_direct);
    let hash = hasher.finalize();

    let pubkey = PUBKEY.load(deps.storage)?;

    deps.api.debug(&format!("hash {:?}", hash));
    deps.api
        .debug(&format!("signature {:?}", auth_request.signature));
    deps.api.debug(&format!("pubkey {:?}", pubkey));

    let valid = deps
        .api
        .secp256k1_verify(&hash, &auth_request.signature, &pubkey)
        .or_else(|e| {
            deps.api.debug(&format!("error {:?}", e));
            deps.api.debug(&format!("error {:?}", e.to_string()));
            Err(StdError::generic_err("Failed to verify signature"))
        })?;

    deps.api.debug(&format!("valid {:?}", valid));

    if !valid {
        return Ok(Response::new().set_data(AuthenticationResult::NotAuthenticated {}));
    }

    Ok(Response::new().set_data(AuthenticationResult::Authenticated {}))
}
