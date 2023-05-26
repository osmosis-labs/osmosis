use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Timestamp};
use cw_storage_plus::{Item, Map};
use registry::msg::SerializableJson;
use swaprouter::msg::ExecuteMsg as SwapRouterExecute;

use crate::msg::FailedDeliveryAction;

#[cw_serde]
pub struct Config {
    pub governor: Addr,
    pub swap_contract: Addr,
}

#[cw_serde]
pub struct ForwardTo {
    pub chain: String,
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

pub mod ibc {
    use super::*;

    #[cw_serde]
    pub enum PacketLifecycleStatus {
        Sent,
        AckSuccess,
        AckFailure,
        TimedOut,
    }

    /// A transfer packet sent by this contract that is expected to be received but
    /// needs to be tracked in case the receive fails or times-out
    #[cw_serde]
    pub struct IBCTransfer {
        pub recovery_addr: Addr,
        pub channel_id: String,
        pub sequence: u64,
        pub amount: u128,
        pub denom: String,
        pub status: PacketLifecycleStatus,
    }
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const SWAP_REPLY_STATE: Item<SwapMsgReplyState> = Item::new("swap_reply_states");
pub const FORWARD_REPLY_STATE: Item<ForwardMsgReplyState> = Item::new("forward_reply_states");

/// In-Flight packets by (source_channel_id, sequence)
pub const INFLIGHT_PACKETS: Map<(&str, u64), ibc::IBCTransfer> = Map::new("inflight");

/// Recovery. This tracks any recovery that an addr can execute.
pub const RECOVERY_STATES: Map<&Addr, Vec<ibc::IBCTransfer>> = Map::new("recovery");
