use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Empty, Timestamp};
use cw_storage_plus::{Item, Map};
use osmosis_swap::crosschain_swaps::ibc;
use osmosis_swap::swaprouter::ExecuteMsg as SwapRouterExecute;

use osmosis_swap::crosschain_swaps::{FailedDeliveryAction, SerializableJson};

#[cw_serde]
pub struct Config {
    pub governor: Addr,
    pub swap_contract: Addr,
}

#[cw_serde]
pub struct ForwardTo {
    pub channel: String,
    pub receiver: Addr,
    pub next_memo: Option<SerializableJson>,
    pub on_failed_delivery: FailedDeliveryAction,
}

#[cw_serde]
pub struct SwapMsgReplyState {
    pub swap_msg: SwapRouterExecute,
    pub contract_addr: Addr,
    pub block_time: Timestamp,
    pub forward_to: ForwardTo,
}

#[cw_serde]
pub struct ForwardMsgReplyState {
    pub channel_id: String,
    pub to_address: String,
    pub amount: u128,
    pub denom: String,
    pub on_failed_delivery: FailedDeliveryAction,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const SWAP_REPLY_STATE: Item<SwapMsgReplyState> = Item::new("swap_reply_states");
pub const FORWARD_REPLY_STATE: Item<ForwardMsgReplyState> = Item::new("forward_reply_states");

/// In-Flight packets by (source_channel_id, sequence)
pub const INFLIGHT_PACKETS: Map<(&str, u64), ibc::IBCTransfer> = Map::new("inflight");

/// Recovery. This tracks any recovery that an addr can execute.
pub const RECOVERY_STATES: Map<&Addr, Vec<ibc::IBCTransfer>> = Map::new("recovery");

/// A mapping of knwon IBC channels accepted by the contract. bech32_prefix => channel
pub const CHANNEL_MAP: Map<&str, String> = Map::new("chain_map");
pub const DISABLED_PREFIXES: Map<&str, Empty> = Map::new("disabled_prefixes");
