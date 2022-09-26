use std::collections::BinaryHeap;

// use chrono::{Datelike, TimeZone, Utc};
use cosmwasm_std::{
    coin, entry_point, to_binary, Addr, Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response,
    StdError, StdResult, Uint128,
};

use crate::error::ContractError;
use crate::helpers::{
    assert_sent_sufficient_coin, calculate_expiry, calculate_required_escrow, send_tokens,
    validate_name,
};
use crate::msg::{
    ExecuteMsg, InstantiateMsg, QueryMsg, ResolveRecordResponse, ReverseResolveRecordResponse,
};
use crate::state::{
    config, config_read, resolver, resolver_read, reverse_resolver, reverse_resolver_read,
    AddressRecord, Config, NameBid, NameRecord,
};

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, StdError> {
    let config_state = Config {
        required_denom: msg.required_denom,
        register_price: msg.register_price,
        annual_tax_bps: msg.annual_tax_bps,
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
        ExecuteMsg::AcceptBid { name } => execute_accept_bid(deps, env, info, name),
        ExecuteMsg::SetName { name } => execute_set_name(deps, env, info, name),
        ExecuteMsg::AddBid { name, price, years } => {
            execute_add_bid(deps, env, info, name, price, years)
        }
        ExecuteMsg::RemoveBids { name } => execute_remove_bid(deps, env, info, name),
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
    let tax_per_year =
        config_state.annual_tax_bps * config_state.register_price / Uint128::from(10_000 as u128);
    // Calculate required payment including rent
    let required_amount = {
        let total_tax = tax_per_year * years;
        config_state.register_price + total_tax
    };
    assert_sent_sufficient_coin(
        &info.funds,
        Some(coin(required_amount.u128(), config_state.required_denom)),
    )?;

    let key = name.as_bytes();
    let expiry = calculate_expiry(env.block.time, years);

    let record = NameRecord {
        owner: info.sender,
        expiry,
        bids: BinaryHeap::new(),
        remaining_escrow: required_amount,
        current_valuation: config_state.register_price,
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

pub fn execute_accept_bid(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
) -> Result<Response, ContractError> {
    // TODO: Uncomment when accept bid logic is ready
    return Err(ContractError::NotImplemented {});

    let config_state = config(deps.storage).load()?;
    let key = name.as_bytes();
    let mut balance: Vec<Coin> = Vec::new();

    resolver(deps.storage).update(key, |record| {
        if let Some(mut record) = record {
            if info.sender != record.owner {
                return Err(ContractError::Unauthorized {});
            }

            match record.bids.pop() {
                Some(highest_bid) => {
                    // Track refund amount
                    let total_amount = {
                        let sale_amount = highest_bid.price;
                        sale_amount + record.remaining_escrow
                    };
                    balance.push(coin(total_amount.u128(), config_state.required_denom));

                    // Update record
                    record.owner = highest_bid.bidder.clone();
                    record.current_valuation = highest_bid.price;
                    record.expiry = calculate_expiry(env.block.time, highest_bid.years);
                    record.remaining_escrow = calculate_required_escrow(
                        highest_bid.price,
                        config_state.annual_tax_bps,
                        highest_bid.years,
                    );
                    Ok(record)
                }
                None => Err(ContractError::NameNoBids),
            }
        } else {
            Err(ContractError::NameNotExists { name: name.clone() })
        }
    })?;

    Ok(send_tokens(
        info.sender,
        balance,
        "Sale of name and refund of unpaid tax",
    ))
}

pub fn execute_add_bid(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    name: String,
    price: Uint128,
    years: Uint128,
) -> Result<Response, ContractError> {
    let mut bids = match resolve_name(&deps, env, &name) {
        Some(record) => record.bids,
        None => return Err(ContractError::NameNotExists { name }),
    };

    let config_state = config(deps.storage).load()?;
    let required_amount = calculate_required_escrow(price, config_state.annual_tax_bps, years);

    assert_sent_sufficient_coin(
        &info.funds,
        Some(coin(required_amount.u128(), config_state.required_denom)),
    )?;

    let name_bid = NameBid {
        price,
        bidder: info.sender.clone(),
        years,
    };
    bids.push(name_bid);

    Ok(Response::default())
}

// TODO: Implement RemoveBid
pub fn execute_remove_bid(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _name: String,
) -> Result<Response, ContractError> {
    Err(ContractError::NotImplemented {})
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
