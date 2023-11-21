use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, DepsMut, Env, MessageInfo, Response, StdError};
use schemars::JsonSchema;

// Messages
#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {}

// // Value does not implement JsonSchema, so we wrap it here. This can be removed
// // if https://github.com/CosmWasm/serde-cw-value/pull/3 gets merged
// #[derive(
//     ::cosmwasm_schema::serde::Serialize,
//     ::cosmwasm_schema::serde::Deserialize,
//     ::std::clone::Clone,
//     ::std::fmt::Debug,
//     PartialEq,
//     Eq,
// )]
// pub struct SerializableJson(pub serde_cw_value::Value);

// impl JsonSchema for SerializableJson {
//     fn schema_name() -> String {
//         "JSON".to_string()
//     }

//     fn json_schema(_gen: &mut schemars::gen::SchemaGenerator) -> schemars::schema::Schema {
//         schemars::schema::Schema::from(true)
//     }
// }

#[cw_serde]
pub enum SudoMsg {
    Authenticate {
        account: Addr,
        msg: Msg,
        signature: String,
        sign_mode_tx_data: SignModeTxData,
        tx_data: TxData,
        signature_data: SignatureData,
    },
}

#[cw_serde]
pub struct Msg {
    pub type_url: String,
    pub value: String,
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
    pub msgs: Vec<Message>,
    pub memo: String,
    pub simulate: bool,
}

#[cw_serde]
pub struct Message {
    pub from_address: Addr,
    pub to_address: Addr,
    pub amount: Vec<Coin>,
}

#[cw_serde]
pub struct Coin {
    pub denom: String,
    pub amount: String,
}

#[cw_serde]
pub struct SignatureData {
    pub signers: Vec<Addr>,
    pub signatures: Vec<String>,
}

// Instantiate
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, StdError> {
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, StdError> {
    deps.api.debug(&format!("sudo {:?}", msg));
    Ok(Response::new().set_data(format!("{:?}", msg).as_bytes()))
}
