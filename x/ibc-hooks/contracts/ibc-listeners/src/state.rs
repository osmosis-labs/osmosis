use cosmwasm_std::Addr;
use cw_storage_plus::Map;

// Contracts listening for a packet's channel, sequenc_number, and event
pub const LISTENERS: Map<(&str, u64, &str), Vec<Addr>> = Map::new("listeners");
