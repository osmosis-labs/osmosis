use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
use cosmwasm_std::{from_binary, Addr, DepsMut};

use crate::contract;
use crate::msg::{GetOwnerResponse, InstantiateMsg, QueryMsg};

static CREATOR_ADDRESS: &str = "creator";

// test helper
#[allow(unused_assignments)]
fn initialize_contract(deps: DepsMut) -> Addr {
    let msg = InstantiateMsg {
        owner: String::from(CREATOR_ADDRESS),
    };
    let info = mock_info(CREATOR_ADDRESS, &[]);

    // instantiate with enough funds provided should succeed
    contract::instantiate(deps, mock_env(), info.clone(), msg).unwrap();

    info.sender
}

#[test]
fn proper_initialization() {
    let mut deps = mock_dependencies();

    let owner = initialize_contract(deps.as_mut());

    // it worked, let's query the state
    let res: GetOwnerResponse =
        from_binary(&contract::query(deps.as_ref(), mock_env(), QueryMsg::GetOwner {}).unwrap())
            .unwrap();
    assert_eq!(owner, res.owner);
}
