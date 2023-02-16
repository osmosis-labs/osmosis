use cw_storage_plus::Map;

// CONTRACT_ALIAS_MAP is a map from a contract alias to a contract address
pub const CONTRACT_ALIAS_MAP: Map<&str, String> = Map::new("contract_alias_map");

// CHAIN_TO_CHAIN_CHANNEL_MAP is a map from a source<>destination chain pair to its respective channel id
pub const CHAIN_TO_CHAIN_CHANNEL_MAP: Map<&str, String> = Map::new("chain_to_chain_channel_map");

// CHANNEL_TO_CHAIN_CHAIN_MAP is a map from a channel id on a source chain to its respective destination chain
pub const CHANNEL_TO_CHAIN_CHAIN_MAP: Map<&str, String> = Map::new("channel_to_chain_chain_map");

// OSMOSIS_DENOM_MAP is a map from a native cosmos chain denom to its respective osmosis ibc denom
pub const OSMOSIS_DENOM_MAP: Map<&str, String> = Map::new("osmosis_denom_map");
