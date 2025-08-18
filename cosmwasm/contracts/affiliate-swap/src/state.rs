use cosmwasm_std::Addr;
use cw_storage_plus::Item;
use osmosis_std::types::osmosis::poolmanager::v1beta1::MsgSwapExactAmountIn;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Config {
    pub owner: Addr,
    pub affiliate_addr: Addr,
    pub affiliate_bps: u16,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct SwapReplyState {
    pub original_sender: Addr,
    pub swap_msg: MsgSwapExactAmountIn,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const SWAP_REPLY_STATE: Item<SwapReplyState> = Item::new("swap_reply_state");
