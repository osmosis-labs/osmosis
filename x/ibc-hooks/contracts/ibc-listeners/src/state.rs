use cosmwasm_std::Addr;
use cw_storage_plus::Map;

// Contracts listening for a packet sequenc_number and a specific event
pub const LISTENERS: Map<(u64, &str), Vec<Addr>> = Map::new("listeners");
