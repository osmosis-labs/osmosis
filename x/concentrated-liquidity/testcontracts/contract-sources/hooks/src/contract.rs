#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{coins, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, BankMsg, };
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{InstantiateMsg, SudoMsg};
use crate::state::{State, STATE};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:infinite-track-beforesend";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Handling contract instantiation
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    let state = State {
        count: 0,
    };
    STATE.save(deps.storage, &state)?;

    // With `Response` type, it is possible to dispatch message to invoke external logic.
    // See: https://github.com/CosmWasm/cosmwasm/blob/main/SEMANTICS.md#dispatching-messages
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: (),
) -> Result<Response, ContractError> {
    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: ()) -> StdResult<Binary> {
    Ok(Binary::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(_deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::BeforeCreatePosition {
            pool_id: _,
            owner,
            tokens_provided: _,
            amount_0_min: _,
            amount_1_min: _,
            lower_tick: _,
            upper_tick: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "beforeCreatePosition" to owner
            let coins = coins(1, "beforeCreatePosition");
            let msg = BankMsg::Send {
                to_address: owner.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::AfterCreatePosition {
            pool_id: _,
            owner,
            tokens_provided: _,
            amount_0_min: _,
            amount_1_min: _,
            lower_tick: _,
            upper_tick: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "afterCreatePosition" to owner
            let coins = coins(1, "afterCreatePosition");
            let msg = BankMsg::Send {
                to_address: owner.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::BeforeWithdrawPosition {
            pool_id: _,
            owner,
            position_id: _,
            amount_to_withdraw: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "beforeWithdrawPosition" to owner
            let coins = coins(1, "beforeWithdrawPosition");
            let msg = BankMsg::Send {
                to_address: owner.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::AfterWithdrawPosition {
            pool_id: _,
            owner,
            position_id: _,
            amount_to_withdraw: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "afterWithdrawPosition" to owner
            let coins = coins(1, "afterWithdrawPosition");
            let msg = BankMsg::Send {
                to_address: owner.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::BeforeSwapExactAmountIn {
            pool_id: _,
            sender,
            token_in: _,
            token_out_denom: _,
            token_out_min_amount: _,
            spread_factor: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "beforeSwapExactAmountIn" to sender
            let coins = coins(1, "beforeSwapExactAmountIn");
            let msg = BankMsg::Send {
                to_address: sender.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::AfterSwapExactAmountIn {
            pool_id: _,
            sender,
            token_in: _,
            token_out_denom: _,
            token_out_min_amount: _,
            spread_factor: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "afterSwapExactAmountIn" to sender
            let coins = coins(1, "afterSwapExactAmountIn");
            let msg = BankMsg::Send {
                to_address: sender.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::BeforeSwapExactAmountOut {
            pool_id: _,
            sender,
            token_in_denom: _,
            token_in_max_amount: _,
            token_out: _,
            spread_factor: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "beforeSwapExactAmountOut" to sender
            let coins = coins(1, "beforeSwapExactAmountOut");
            let msg = BankMsg::Send {
                to_address: sender.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
        SudoMsg::AfterSwapExactAmountOut {
            pool_id: _,
            sender,
            token_in_denom: _,
            token_in_max_amount: _,
            token_out: _,
            spread_factor: _,
        } => {
            let mut response = Response::new();

            // mint coins with denom "afterSwapExactAmountOut" to sender
            let coins = coins(1, "afterSwapExactAmountOut");
            let msg = BankMsg::Send {
                to_address: sender.clone(),
                amount: coins,
            };
            response = response.add_message(msg);
            
            Ok(response.add_attribute("method", "sudo"))
        },
    }
}

// #[cfg(test)]
// mod tests {
//     use super::*;
//     use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
//     use cosmwasm_std::{coins, Coin, Uint128, attr, coin};

//     #[test]
//     fn proper_initialization() {
//         let mut deps = mock_dependencies();

//         let msg = InstantiateMsg {};
//         let info = mock_info("creator", &coins(2, "token"));

//         // we can just call .unwrap() to assert this was a success
//         let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
//     }

//     #[test]
//     fn test_before_create_position_sudo_message() {
//         let mut deps = mock_dependencies();

//         // Fund contract with 1 token with denom "beforeCreatePosition"
//         let fund_msg = BankMsg::Send {
//             to_address: "contract".to_string(),
//             amount: vec![Coin { denom: "beforeCreatePosition".to_string(), amount: Uint128::from(1u128) }],
//         };
//         let res = deps.querier.execute(BankMsg::Send(fund_msg)).unwrap();
//         assert_eq!(res, Ok(()));

//         let msg = SudoMsg::BeforeCreatePosition {
//             pool_id: 1,
//             owner: "owner".to_string(),
//             tokens_provided: vec![Coin { denom: "token".to_string(), amount: Uint128::from(1000u128) }],
//             amount_0_min: "10".to_string(),
//             amount_1_min: "20".to_string(),
//             lower_tick: -10,
//             upper_tick: 10,
//         };

//         let _res = sudo(deps.as_mut(), mock_env(), msg).unwrap();
     
//         // Ensure the owner has received 1 token with denom "beforeCreatePosition"
//         let owner_balance = deps.querier.update_balance("owner", vec![coin(1, "beforeCreatePosition")]);
//         assert_eq!(owner_balance, Some(vec![coin(1, "beforeCreatePosition")]));
//     }
    
// }
