use cosmwasm_std::{Addr, AllBalanceResponse, BankQuery, Deps, StdResult};

pub fn query_account_balances(deps: Deps, account_address: &Addr) -> StdResult<AllBalanceResponse> {
    let balances_query = BankQuery::AllBalances {
        address: account_address.to_string(),
    };
    deps.querier.query(&balances_query.into())
}
