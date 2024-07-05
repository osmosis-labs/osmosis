//! storage variables

use std::collections::{BTreeSet, HashSet};

use cosmwasm_std::Addr;
use cw_storage_plus::{Deque, Item, Map};

use super::{rate_limit::RateLimit, rbac::{QueuedMessage, Roles}};



/// Only this address can manage the contract. This will likely be the
/// governance module, but could be set to something else if needed
pub const GOVMODULE: Item<Addr> = Item::new("gov_module");
/// Only this address can execute transfers. This will likely be the
/// IBC transfer module, but could be set to something else if needed
pub const IBCMODULE: Item<Addr> = Item::new("ibc_module");

/// RATE_LIMIT_TRACKERS is the main state for this contract. It maps a path (IBC
/// Channel + denom) to a vector of `RateLimit`s.
///
/// The `RateLimit` struct contains the information about how much value of a
/// denom has moved through the channel during the currently active time period
/// (channel_flow.flow) and what percentage of the denom's value we are
/// allowing to flow through that channel in a specific duration (quota)
///
/// For simplicity, the channel in the map keys refers to the "host" channel on
/// the osmosis side. This means that on PacketSend it will refer to the source
/// channel while on PacketRecv it refers to the destination channel.
///
/// It is the responsibility of the go module to pass the appropriate channel
/// when sending the messages
///
/// The map key (String, String) represents (channel_id, denom). We use
/// composite keys instead of a struct to avoid having to implement the
/// PrimaryKey trait
pub const RATE_LIMIT_TRACKERS: Map<(String, String), Vec<RateLimit>> = Map::new("flow");

/// Maps address -> delay, automatically applying a timelock delay to all 
/// messages submitted by a specific address
pub const TIMELOCK_DELAY: Map<String, u64> = Map::new("timelock_delay");

/// Storage variable which is used to queue messages for execution that are the result of a successful dao message.
/// In order for the message to be processed, X hours must past from QueuedMessage::submited_at
pub const MESSAGE_QUEUE: Deque<QueuedMessage> = Deque::new("queued_messages");

/// Storage variable that is used to map signing addresses and the permissions they have been granted
pub const RBAC_PERMISSIONS: Map<String, BTreeSet<Roles>> = Map::new("rbac");
