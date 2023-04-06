use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

// Enum to map short strings to storage keys
enum StorageKey {
    ContractAliasMap,
    ChainToChainChannelMap,
    ChannelOnChainChainMap,
    ChainToBech32PrefixMap,
    ChainToBech32PrefixReverseMap,
    Config,
    GlobalAdminMap,
    ChainAdminMap,
    ChainMaintainerMap,
}

// Implement the `StorageKey` enum to a string conversion.
impl StorageKey {
    const fn to_string(&self) -> &'static str {
        match self {
            StorageKey::ContractAliasMap => "cam",
            StorageKey::ChainToChainChannelMap => "ctccm",
            StorageKey::ChannelOnChainChainMap => "cotccm",
            StorageKey::ChainToBech32PrefixMap => "ctbpm",
            StorageKey::ChainToBech32PrefixReverseMap => "ctbprm",
            StorageKey::Config => "cfg",
            StorageKey::GlobalAdminMap => "gam",
            StorageKey::ChainAdminMap => "cam",
            StorageKey::ChainMaintainerMap => "cmm",
        }
    }
}

#[cw_serde]
pub struct RegistryValue {
    pub value: String,
    pub enabled: bool,
}

impl<T: AsRef<str>> From<(T, bool)> for RegistryValue {
    fn from((value, enabled): (T, bool)) -> Self {
        Self {
            value: value.as_ref().to_string(),
            enabled,
        }
    }
}

// CONTRACT_ALIAS_MAP is a map from a contract alias to a contract address
pub const CONTRACT_ALIAS_MAP: Map<&str, String> =
    Map::new(StorageKey::ContractAliasMap.to_string());

// CHAIN_TO_CHAIN_CHANNEL_MAP is a map from source<>destination chain pair to its respective channel id.
// The boolean value indicates whether the mapping is enabled or not.
// (SOURCE_CHAIN_ID, DESTINATION_CHAIN_ID) -> (CHANNEL_ID, ENABLED)
pub const CHAIN_TO_CHAIN_CHANNEL_MAP: Map<(&str, &str), RegistryValue> =
    Map::new(StorageKey::ChainToChainChannelMap.to_string());

// CHANNEL_TO_CHAIN_CHAIN_MAP is a map from a channel id on a source chain to its respective destination chain.
// The boolean value indicates whether the mapping is enabled or not.
// (CHANNEL_ID, SOURCE_CHAIN_ID) -> (DESTINATION_CHAIN_ID, ENABLED)
pub const CHANNEL_ON_CHAIN_CHAIN_MAP: Map<(&str, &str), RegistryValue> =
    Map::new(StorageKey::ChannelOnChainChainMap.to_string());

// CHAIN_TO_BECH32_PREFIX_MAP is a map from a chain id to its respective bech32 prefix.
// The boolean value indicates whether the mapping is enabled or not.
// CHAIN_ID -> (BECH32_PREFIX, ENABLED)
pub const CHAIN_TO_BECH32_PREFIX_MAP: Map<&str, RegistryValue> =
    Map::new(StorageKey::ChainToBech32PrefixMap.to_string());

// CHAIN_TO_BECH32_PREFIX_MAP is a map from a chain id to its respective bech32 prefix.
// The boolean value indicates whether the mapping is enabled or not.
// CHAIN_ID -> (BECH32_PREFIX, ENABLED)
pub const CHAIN_TO_BECH32_PREFIX_REVERSE_MAP: Map<&str, Vec<String>> =
    Map::new(StorageKey::ChainToBech32PrefixReverseMap.to_string());

// CONFIG stores the contract owner
pub const CONFIG: Item<Config> = Item::new(StorageKey::Config.to_string());

// CHAIN_ADMIN_MAP is a map from a source chain to the address that is authorized to add, update, or remove channels for that source chain
pub const GLOBAL_ADMIN_MAP: Map<&str, Addr> = Map::new(StorageKey::GlobalAdminMap.to_string());

// CHAIN_ADMIN_MAP is a map from a source chain to the address that is authorized to add, update, or remove channels for that source chain
pub const CHAIN_ADMIN_MAP: Map<&str, Addr> = Map::new(StorageKey::ChainAdminMap.to_string());

// CHAIN_MAINTAINER_MAP is a map from a source chain to the address that is authorized add, enable, or disable channels for that source chain
pub const CHAIN_MAINTAINER_MAP: Map<&str, Addr> =
    Map::new(StorageKey::ChainMaintainerMap.to_string());

#[cw_serde]
pub struct Config {
    pub owner: Addr,
}
