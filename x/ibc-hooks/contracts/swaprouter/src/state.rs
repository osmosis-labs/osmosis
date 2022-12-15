use osmosis_std::types::osmosis::gamm::v1beta1::{MsgSwapExactAmountIn, SwapAmountInRoute};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct State {
    pub owner: Addr,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct SwapMsgReplyState {
    pub original_sender: Addr,
    pub swap_msg: MsgSwapExactAmountIn,
}

pub const STATE: Item<State> = Item::new("state");
pub const ROUTING_TABLE: Map<(&str, &str), Vec<SwapAmountInRoute>> = Map::new("routing_table");
pub const SWAP_REPLY_STATES: Map<u64, SwapMsgReplyState> = Map::new("swap_reply_states");
