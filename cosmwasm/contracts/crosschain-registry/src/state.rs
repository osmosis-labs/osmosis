use cw_storage_plus::Map;

pub const CONTRACT_MAP: Map<&str, String> = Map::new("contract_map");
