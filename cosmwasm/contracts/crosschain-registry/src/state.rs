use cw_storage_plus::Map;

pub const CONTRACT_MAP: Map<&str, String> = Map::new("contract_map");
pub const CHAIN_CHANNEL_MAP: Map<&str, String> = Map::new("chain_channel_map");
pub const ASSET_MAP: Map<&str, String> = Map::new("asset_map");
