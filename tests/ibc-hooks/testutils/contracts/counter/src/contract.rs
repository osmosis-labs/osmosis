use std::collections::HashMap;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_binary, Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128,
};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::*;
use crate::state::{Counter, COUNTERS};

// version info for migration info
const CONTRACT_NAME: &str = "osmosis:permissioned_counter";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    let initial_counter = Counter {
        count: msg.count,
        total_funds: vec![],
        owner: info.sender.clone(),
    };
    COUNTERS.save(deps.storage, info.sender.clone(), &initial_counter)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender)
        .add_attribute("count", msg.count.to_string()))
}

pub mod utils {
    use cosmwasm_std::Addr;

    use super::*;

    pub fn update_counter(
        deps: DepsMut,
        sender: Addr,
        update_counter: &dyn Fn(&Option<Counter>) -> i32,
        update_funds: &dyn Fn(&Option<Counter>) -> Vec<Coin>,
    ) -> Result<bool, ContractError> {
        COUNTERS
            .update(
                deps.storage,
                sender.clone(),
                |state| -> Result<_, ContractError> {
                    match state {
                        None => Ok(Counter {
                            count: update_counter(&None),
                            total_funds: update_funds(&None),
                            owner: sender,
                        }),
                        Some(counter) => Ok(Counter {
                            count: update_counter(&Some(counter.clone())),
                            total_funds: update_funds(&Some(counter)),
                            owner: sender,
                        }),
                    }
                },
            )
            .map(|_r| true)
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Increment {} => execute::increment(deps, info),
        ExecuteMsg::Reset { count } => execute::reset(deps, info, count),
    }
}

pub mod execute {
    use super::*;

    pub fn increment(deps: DepsMut, info: MessageInfo) -> Result<Response, ContractError> {
        utils::update_counter(
            deps,
            info.sender,
            &|counter| match counter {
                None => 0,
                Some(counter) => counter.count + 1,
            },
            &|counter| match counter {
                None => info.funds.clone(),
                Some(counter) => naive_add_coins(&info.funds, &counter.total_funds),
            },
        )?;
        Ok(Response::new().add_attribute("action", "increment"))
    }

    pub fn reset(deps: DepsMut, info: MessageInfo, count: i32) -> Result<Response, ContractError> {
        utils::update_counter(deps, info.sender, &|_counter| count, &|_counter| vec![])?;
        Ok(Response::new().add_attribute("action", "reset"))
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::IBCLifecycleComplete(IBCLifecycleComplete::IBCAck {
            channel: _,
            sequence: _,
            ack: _,
            success,
        }) => sudo::receive_ack(deps, env.contract.address, success),
        SudoMsg::IBCLifecycleComplete(IBCLifecycleComplete::IBCTimeout {
            channel: _,
            sequence: _,
        }) => sudo::ibc_timeout(deps, env.contract.address),
    }
}

pub mod sudo {
    use cosmwasm_std::Addr;

    use super::*;

    pub fn receive_ack(
        deps: DepsMut,
        contract: Addr,
        _success: bool,
    ) -> Result<Response, ContractError> {
        utils::update_counter(
            deps,
            contract,
            &|counter| match counter {
                None => 1,
                Some(counter) => counter.count + 1,
            },
            &|_counter| vec![],
        )?;
        Ok(Response::new().add_attribute("action", "ack"))
    }

    pub(crate) fn ibc_timeout(deps: DepsMut, contract: Addr) -> Result<Response, ContractError> {
        utils::update_counter(
            deps,
            contract,
            &|counter| match counter {
                None => 10,
                Some(counter) => counter.count + 10,
            },
            &|_counter| vec![],
        )?;
        Ok(Response::new().add_attribute("action", "timeout"))
    }
}

pub fn naive_add_coins(lhs: &Vec<Coin>, rhs: &Vec<Coin>) -> Vec<Coin> {
    // This is a naive, inneficient  implementation of Vec<Coin> addition.
    // This shouldn't be used in production but serves our purpose for this
    // testing contract
    let mut coins: HashMap<String, Uint128> = HashMap::new();
    for coin in lhs {
        coins.insert(coin.denom.clone(), coin.amount);
    }

    for coin in rhs {
        coins
            .entry(coin.denom.clone())
            .and_modify(|e| *e += coin.amount)
            .or_insert(coin.amount);
    }
    coins.iter().map(|(d, &a)| Coin::new(a.into(), d)).collect()
}

#[test]
fn coin_addition() {
    let c1 = vec![Coin::new(1, "a"), Coin::new(2, "b")];
    let c2 = vec![Coin::new(7, "a"), Coin::new(2, "c")];

    let mut sum = naive_add_coins(&c1, &c1);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(sum, vec![Coin::new(2, "a"), Coin::new(4, "b")]);

    let mut sum = naive_add_coins(&c1, &c2);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(
        sum,
        vec![Coin::new(8, "a"), Coin::new(2, "b"), Coin::new(2, "c"),]
    );

    let mut sum = naive_add_coins(&c2, &c2);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(sum, vec![Coin::new(14, "a"), Coin::new(4, "c"),]);

    let mut sum = naive_add_coins(&c2, &c1);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(
        sum,
        vec![Coin::new(8, "a"), Coin::new(2, "b"), Coin::new(2, "c"),]
    );

    let mut sum = naive_add_coins(&vec![], &c2);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(sum, c2);

    let mut sum = naive_add_coins(&c2, &vec![]);
    sum.sort_by(|a, b| a.denom.cmp(&b.denom));
    assert_eq!(sum, c2);
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount { addr } => to_binary(&query::count(deps, addr)?),
        QueryMsg::GetTotalFunds { addr } => to_binary(&query::total_funds(deps, addr)?),
    }
}

pub mod query {
    use cosmwasm_std::Addr;

    use super::*;

    pub fn count(deps: Deps, addr: Addr) -> StdResult<GetCountResponse> {
        let state = COUNTERS.load(deps.storage, addr)?;
        Ok(GetCountResponse { count: state.count })
    }

    pub fn total_funds(deps: Deps, addr: Addr) -> StdResult<GetTotalFundsResponse> {
        let state = COUNTERS.load(deps.storage, addr)?;
        Ok(GetTotalFundsResponse {
            total_funds: state.total_funds,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::Addr;
    use cosmwasm_std::{coins, from_binary};

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg { count: 17 };
        let info = mock_info("creator", &coins(1000, "earth"));

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // it worked, let's query the state
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetCount {
                addr: Addr::unchecked("creator"),
            },
        )
        .unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(17, value.count);
    }

    #[test]
    fn increment() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg { count: 17 };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let msg = InstantiateMsg { count: 17 };
        let info = mock_info("someone-else", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let info = mock_info("creator", &coins(2, "token"));
        let msg = ExecuteMsg::Increment {};
        let _res = execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // should increase counter by 1
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetCount {
                addr: Addr::unchecked("creator"),
            },
        )
        .unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(18, value.count);

        // Counter for someone else is not incremented
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetCount {
                addr: Addr::unchecked("someone-else"),
            },
        )
        .unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(17, value.count);
    }

    #[test]
    fn reset() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg { count: 17 };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // beneficiary can release it
        let unauth_info = mock_info("anyone", &coins(2, "token"));
        let msg = ExecuteMsg::Reset { count: 7 };
        let _res = execute(deps.as_mut(), mock_env(), unauth_info, msg);

        // should be 7
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetCount {
                addr: Addr::unchecked("anyone"),
            },
        )
        .unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(7, value.count);

        // only the original creator can reset the counter
        let auth_info = mock_info("creator", &coins(2, "token"));
        let msg = ExecuteMsg::Reset { count: 5 };
        let _res = execute(deps.as_mut(), mock_env(), auth_info, msg).unwrap();

        // should now be 5
        let res = query(
            deps.as_ref(),
            mock_env(),
            QueryMsg::GetCount {
                addr: Addr::unchecked("creator"),
            },
        )
        .unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(5, value.count);
    }

    #[test]
    fn acks() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let get_msg = QueryMsg::GetCount {
            addr: Addr::unchecked(env.clone().contract.address),
        };

        // No acks
        query(deps.as_ref(), env.clone(), get_msg.clone()).unwrap_err();

        let msg = SudoMsg::ReceiveAck {
            channel: format!("channel-0"),
            sequence: 1,
            ack: String::new(),
            success: true,
        };
        let _res = sudo(deps.as_mut(), env.clone(), msg).unwrap();

        // should increase counter by 1
        let res = query(deps.as_ref(), env.clone(), get_msg.clone()).unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(1, value.count);

        let msg = SudoMsg::ReceiveAck {
            channel: format!("channel-0"),
            sequence: 1,
            ack: String::new(),
            success: true,
        };
        let _res = sudo(deps.as_mut(), env.clone(), msg).unwrap();

        // should increase counter by 1
        let res = query(deps.as_ref(), env, get_msg).unwrap();
        let value: GetCountResponse = from_binary(&res).unwrap();
        assert_eq!(2, value.count);
    }
}
