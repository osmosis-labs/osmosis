use std::collections::BinaryHeap;

// use chrono::{Datelike, TimeZone, Utc};
use cosmwasm_std::{
    coin, entry_point, to_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response,
    StdError, StdResult, Timestamp, Uint128,
};
use cw_utils::Expiration;

use crate::error::ContractError;
use crate::helpers::{assert_matches_denom, assert_sent_sufficient_coin};
use crate::msg::{
    ExecuteMsg, InstantiateMsg, QueryMsg, ResolveRecordResponse, ReverseResolveRecordResponse,
};
use crate::state::{
    config, config_read, resolver, resolver_read, reverse_resolver, reverse_resolver_read,
    AddressRecord, Config, NameBid, NameRecord, AVERAGE_SECONDS_PER_YEAR, IBC_SUFFIX,
};

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
        required_denom: msg.required_denom,
        mint_price: msg.purchase_price,
        transfer_price: msg.transfer_price,
        annual_tax_bps: msg.annual_rent_bps,
        owner_grace_period: msg.owner_grace_period,
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
        ExecuteMsg::SetName { name } => execute_set_name(deps, env, info, name),
        ExecuteMsg::AddBid { name } => execute_bid(deps, env, info, name),
    }
}

pub fn execute_register(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
    years: Uint128,
) -> Result<Response, ContractError> {
    if years.u128() <= 0 {
        return Err(ContractError::YearsMustBePositive {});
    }
    validate_name(&name)?;

    let config_state = config(deps.storage).load()?;
    let rent_per_year =
        config_state.annual_tax_bps * config_state.mint_price / Uint128::from(10_000 as u128);
    // Calculate required payment including rent
    let required = {
        let total_rent = rent_per_year * years;
        let amount = config_state.mint_price + total_rent;
        Some(coin(amount.u128(), config_state.required_denom))
    };
    assert_sent_sufficient_coin(&info.funds, required)?;

    let key = name.as_bytes();
    let expiry = {
        let now_ts = Timestamp::from_nanos(env.block.time.nanos());
        let expiry_ts = now_ts.plus_seconds(years.u128() as u64 * AVERAGE_SECONDS_PER_YEAR);

        Expiration::AtTime(expiry_ts)
    };

    let record = NameRecord {
        owner: info.sender,
        expiry,
        bids: BinaryHeap::new(),
        current_tax: rent_per_year,
    };

    if let Some(existing_record) = resolver(deps.storage).may_load(key)? {
        // name is already taken and expiry still not past
        if !existing_record.expiry.is_expired(&env.block) {
            return Err(ContractError::NameTaken { name });
        }
    }

    // name is available
    resolver(deps.storage).save(key, &record)?;

    Ok(Response::default())
}

pub fn execute_set_name(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
) -> Result<Response, ContractError> {
    // Check we own the name
    let expiry = match resolve_name(&deps, env, &name) {
        Some(record) => record.expiry,
        None => return Err(ContractError::NameNotExists { name }),
    };

    let addr_string = Addr::to_string(&info.sender);
    let addr_key = addr_string.as_bytes();
    let addr_record = AddressRecord { name, expiry };
    reverse_resolver(deps.storage).save(addr_key, &addr_record)?;

    Ok(Response::default())
}

fn resolve_name(deps: &DepsMut, env: Env, name: &String) -> Option<NameRecord> {
    let key = name.as_bytes();

    match resolver_read(deps.storage).may_load(key) {
        Ok(Some(record)) => {
            if record.expiry.is_expired(&env.block) {
                None
            } else {
                Some(record)
            }
        }
        _ => None,
    }
}

pub fn execute_transfer(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    name: String,
    to: String,
) -> Result<Response, ContractError> {
    let config_state = config(deps.storage).load()?;
    assert_sent_sufficient_coin(
        &info.funds,
        Some(coin(
            config_state.transfer_price.u128(),
            config_state.required_denom,
        )),
    )?;

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

pub fn execute_bid(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
) -> Result<Response, ContractError> {
    let mut bids = match resolve_name(&deps, env, &name) {
        Some(record) => record.bids,
        None => return Err(ContractError::NameNotExists { name }),
    };

    let config_state = config(deps.storage).load()?;
    assert_matches_denom(&info.funds, &config_state.required_denom)?;

    for coin in info.funds {
        let name_bid = NameBid {
            amount: coin.amount,
            bidder: info.sender.clone(),
        };
        bids.push(name_bid);
    }

    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::ResolveRecord { name } => query_resolver(deps, env, name),
        QueryMsg::ReverseResolveRecord { address } => query_reverse_resolver(deps, env, address),
        QueryMsg::Config {} => to_binary(&config_read(deps.storage).load()?),
    }
}

fn query_resolver(deps: Deps, env: Env, name: String) -> StdResult<Binary> {
    let key = name.as_bytes();

    let address = match resolver_read(deps.storage).may_load(key)? {
        Some(record) => {
            if record.expiry.is_expired(&env.block) {
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

fn query_reverse_resolver(deps: Deps, env: Env, address: Addr) -> StdResult<Binary> {
    let key = address.as_bytes();
    let name = match reverse_resolver_read(deps.storage).may_load(key)? {
        Some(record) => {
            if record.expiry.is_expired(&env.block) {
                None
            } else {
                Some(String::from(&record.name))
            }
        }
        None => None,
    };
    let resp = ReverseResolveRecordResponse { name };

    to_binary(&resp)
}

// let's not import a regexp library and just do these checks by hand
fn invalid_char(c: char) -> bool {
    let is_valid = c.is_digit(10) || c.is_ascii_lowercase() || (c == '-' || c == '_');
    !is_valid
}

/// validate_name returns an error if the name is invalid
/// (we require 3-64 lowercase ascii letters , numbers, or "-" "_")
/// ends in `.ibc` suffix, no other periods are allowed
fn validate_name(name_with_suffix: &str) -> Result<(), ContractError> {
    let length = name_with_suffix.len() as u64;
    let (name, suffix) = {
        let full_length = name_with_suffix.len();
        name_with_suffix.split_at(full_length - IBC_SUFFIX.len())
    };

    if suffix != IBC_SUFFIX {
        return Err(ContractError::NameNeedsSuffix {
            suffix: IBC_SUFFIX.to_string(),
        });
    }

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
