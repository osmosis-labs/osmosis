use cosmwasm_schema::write_api;

use osmosis_swap::crosschain_swaps::{ExecuteMsg, InstantiateMsg, QueryMsg};

fn main() {
    write_api! {
        name: "crosschain-swaps",
        instantiate: InstantiateMsg,
        query: QueryMsg,
        execute: ExecuteMsg,
    };
}
