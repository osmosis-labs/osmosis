use cw_storage_plus::Map;

use crate::ibc;

// (channel, sequence) -> packet
pub const INFLIGHT_PACKETS: Map<(&str, u64), ibc::Packet> = Map::new("inflight");
