use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError};

// Messages
#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    Authenticate {
        account: Addr,
        msg: Vec<u8>,
        signature_data: SignatureData,
    },
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
    Ok(Response::new())
}
