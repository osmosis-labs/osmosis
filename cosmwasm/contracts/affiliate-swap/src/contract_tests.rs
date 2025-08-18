use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
use cosmwasm_std::{from_binary, Coin, Uint128};

use crate::contract::{execute, instantiate, query, reply};
use crate::execute::SWAP_REPLY_ID;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use osmosis_std::types::osmosis::poolmanager::v1beta1::{MsgSwapExactAmountInResponse, SwapAmountInRoute};
use prost::Message as _;

fn mock_instantiate(deps: &mut cosmwasm_std::OwnedDeps<_, _, _>) {
    let msg = InstantiateMsg {
        owner: "owner".to_string(),
        affiliate_addr: "affiliate".to_string(),
        affiliate_bps: 250, // 2.5%
    };
    let info = mock_info("owner", &[]);
    instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
}

#[test]
fn test_config_query() {
    let mut deps = mock_dependencies();
    mock_instantiate(&mut deps);
    let bin = query(deps.as_ref(), mock_env(), QueryMsg::Config {}).unwrap();
    let resp: crate::msg::ConfigResponse = from_binary(&bin).unwrap();
    assert_eq!(resp.affiliate_bps, 250);
}

#[test]
fn test_swap_reply_splits() {
    let mut deps = mock_dependencies();
    mock_instantiate(&mut deps);

    let route = vec![SwapAmountInRoute { pool_id: 1, token_out_denom: "uosmo".to_string() }];
    let msg = ExecuteMsg::SwapWithFee {
        input_coin: Coin::new(1000, "uion"),
        output_denom: "uosmo".to_string(),
        min_output_amount: Uint128::new(1),
        route,
    };
    let info = mock_info("trader", &[Coin::new(1000, "uion")]);
    let resp = execute(deps.as_mut(), mock_env(), info, msg).unwrap();
    assert_eq!(resp.messages.len(), 1);
    assert_eq!(resp.messages[0].id, Some(SWAP_REPLY_ID));

    // Simulate a successful reply with 1000 out
    let resp_msg = MsgSwapExactAmountInResponse { token_out_amount: "1000".to_string() };
    let mut data = Vec::new();
    prost::Message::encode(&resp_msg, &mut data).unwrap();
    let reply_msg = cosmwasm_std::Reply { id: SWAP_REPLY_ID, result: cosmwasm_std::SubMsgResult::Ok(cosmwasm_std::SubMsgResponse { data: Some(cosmwasm_std::Binary::from(data)), events: vec![] }) };
    let resp = reply(deps.as_mut(), mock_env(), reply_msg).unwrap();
    // Expect two bank messages: 2.5% to affiliate (25), 975 to trader
    assert_eq!(resp.messages.len(), 2);
}

