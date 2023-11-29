use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, DepsMut, Env, MessageInfo, Response, StdError};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
pub enum SudoMsg {
    Authenticate(AuthenticationRequest),
}

#[cw_serde]
pub struct Any {
    pub type_url: String,
    pub value: cosmwasm_std::Binary,
}

#[cw_serde]
pub struct AuthenticationRequest {
    pub account: Addr,
    pub msg: Any,
    pub signature: String,
    pub sign_mode_tx_data: SignModeTxData,
    pub tx_data: TxData,
    pub signature_data: SignatureData,
}

#[cw_serde]
pub struct SignModeTxData {
    pub sign_mode_direct: String,
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

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, StdError> {
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
    let send: osmosis_std::types::cosmos::bank::v1beta1::MsgSend =
        auth_request.msg.value.try_into()?;

    deps.api.debug(&format!("send {:?}", send));

    Ok(Response::new().set_data(AuthenticationResult::Authenticated {}))
}
