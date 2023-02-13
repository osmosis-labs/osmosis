use cw_storage_plus::Map;

pub const CONTRACT_NAMES: Map<&str, String> = Map::new("contract_names");
