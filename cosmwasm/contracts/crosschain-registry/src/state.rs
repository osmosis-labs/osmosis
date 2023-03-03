use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

// Enum to map short strings to storage keys
enum StorageKey {
    ContractAliasMap,
    ChainToChainChannelMap,
    ChannelOnChainChainMap,
    Config,
}

// Implement the `StorageKey` enum to a string conversion.
impl StorageKey {
    const fn to_string(&self) -> &'static str {
        match self {
            StorageKey::ContractAliasMap => "cam",
            StorageKey::ChainToChainChannelMap => "ctccm",
            StorageKey::ChannelOnChainChainMap => "cotccm",
            StorageKey::Config => "cfg",
        }
    }
}

// CONTRACT_ALIAS_MAP is a map from a contract alias to a contract address
pub const CONTRACT_ALIAS_MAP: Map<&str, String> =
    Map::new(StorageKey::ContractAliasMap.to_string());

// CHAIN_TO_CHAIN_CHANNEL_MAP is a map from a source<>destination chain pair to its respective channel id
pub const CHAIN_TO_CHAIN_CHANNEL_MAP: Map<(&str, &str), String> =
    Map::new(StorageKey::ChainToChainChannelMap.to_string());

// CHANNEL_TO_CHAIN_CHAIN_MAP is a map from a channel id on a source chain to its respective destination chain
pub const CHANNEL_ON_CHAIN_CHAIN_MAP: Map<(&str, &str), String> =
    Map::new(StorageKey::ChannelOnChainChainMap.to_string());

pub const CONFIG: Item<Config> = Item::new(StorageKey::Config.to_string());

#[cw_serde]
pub struct Config {
    pub owner: Addr,
}
