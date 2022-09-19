// use chrono::{Datelike, TimeZone, Utc};
use cosmwasm_std::{
    coin, entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError,
    StdResult, Uint128,
};

use crate::error::ContractError;
use crate::helpers::assert_sent_sufficient_coin;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, ResolveRecordResponse};
use crate::state::{config, config_read, resolver, resolver_read, Config, NameRecord};

const MIN_NAME_LENGTH: u64 = 3;
const MAX_NAME_LENGTH: u64 = 64;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, StdError> {
    let config_state = Config {
        purchase_price: msg.purchase_price,
        transfer_price: msg.transfer_price,
        annual_rent_amount: msg.annual_rent_amount,
    };

    config(deps.storage).save(&config_state)?;

    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Register { name, years } => execute_register(deps, env, info, name, years),
        ExecuteMsg::Transfer { name, to } => execute_transfer(deps, env, info, name, to),
    }
}

pub fn execute_register(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
    years: Uint128,
) -> Result<Response, ContractError> {
    // TODO: Validate rent years
    validate_name(&name)?;
    let config_state = config(deps.storage).load()?;
    let required = match config_state.purchase_price {
        Some(purchase_price) => {
            let amount = purchase_price.amount + config_state.annual_rent_amount * years;
            Some(coin(amount.u128(), purchase_price.denom))
        }
        None => None,
    };
    assert_sent_sufficient_coin(&info.funds, required)?;

    let key = name.as_bytes();
    let now = env.block.time.nanos() as u128;
    let expiry = now + 31_536_000 * years.u128();

    let record = NameRecord {
        owner: info.sender,
        expiry,
    };

    if let Some(existing_record) = resolver(deps.storage).may_load(key)? {
        // name is already taken and expiry still not past
        if existing_record.expiry > now {
            return Err(ContractError::NameTaken { name });
        }
    }

    // name is available
    resolver(deps.storage).save(key, &record)?;

    Ok(Response::default())
}

pub fn execute_transfer(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    name: String,
    to: String,
) -> Result<Response, ContractError> {
    let config_state = config(deps.storage).load()?;
    assert_sent_sufficient_coin(&info.funds, config_state.transfer_price)?;

    let new_owner = deps.api.addr_validate(&to)?;
    let key = name.as_bytes();
    resolver(deps.storage).update(key, |record| {
        if let Some(mut record) = record {
            if info.sender != record.owner {
                return Err(ContractError::Unauthorized {});
            }

            record.owner = new_owner.clone();
            Ok(record)
        } else {
            Err(ContractError::NameNotExists { name: name.clone() })
        }
    })?;
    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::ResolveRecord { name } => query_resolver(deps, env, name),
        QueryMsg::Config {} => to_binary(&config_read(deps.storage).load()?),
    }
}

fn query_resolver(deps: Deps, env: Env, name: String) -> StdResult<Binary> {
    let key = name.as_bytes();
    let now = env.block.time.nanos() as u128;

    let address = match resolver_read(deps.storage).may_load(key)? {
        Some(record) => {
            if now >= record.expiry {
                None
            } else {
                Some(String::from(&record.owner))
            }
        }
        None => None,
    };
    let resp = ResolveRecordResponse { address };

    to_binary(&resp)
}

// let's not import a regexp library and just do these checks by hand
fn invalid_char(c: char) -> bool {
    let is_valid = c.is_digit(10) || c.is_ascii_lowercase() || (c == '.' || c == '-' || c == '_');
    !is_valid
}

/// validate_name returns an error if the name is invalid
/// (we require 3-64 lowercase ascii letters, numbers, or . - _)
fn validate_name(name: &str) -> Result<(), ContractError> {
    let length = name.len() as u64;
    if (name.len() as u64) < MIN_NAME_LENGTH {
        Err(ContractError::NameTooShort {
            length,
            min_length: MIN_NAME_LENGTH,
        })
    } else if (name.len() as u64) > MAX_NAME_LENGTH {
        Err(ContractError::NameTooLong {
            length,
            max_length: MAX_NAME_LENGTH,
        })
    } else {
        match name.find(invalid_char) {
            None => Ok(()),
            Some(bytepos_invalid_char_start) => {
                let c = name[bytepos_invalid_char_start..].chars().next().unwrap();
                Err(ContractError::InvalidCharacter { c })
            }
        }
    }
}
