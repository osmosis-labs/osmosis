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
    HasPacketForwardMiddleware,
    DenomAliasMap,
    DenomAliasReverseMap,
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
            StorageKey::HasPacketForwardMiddleware => "hpfm",
            StorageKey::DenomAliasMap => "dam",
            StorageKey::DenomAliasReverseMap => "darm",
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

/// ChainPFM stores the state of the packet forward middleware for a chain. Anyone can request
/// to enable the packet forward middleware for a chain, but the contract will verify that  
/// packets can properly be forwarded by the chain
#[cw_serde]
#[derive(Default)]
pub struct ChainPFM {
    /// The verification packet has been received by the chain, forwarded, and the ack has been received
    pub acknowledged: bool,
    /// The contract has validated that the received packet is as expected
    pub validated: bool,
    /// The address that initiated the propose_pfm flow
    pub initiator: Option<Addr>,
}

impl ChainPFM {
    /// Both acknowledged and validated must be true for the pfm to be enabled. This is to avoid
    /// situations in which the chain calls the contract to set validated to true but that call is
    /// not from the same packet that was forwarded by this contract.
    pub fn is_validated(&self) -> bool {
        self.acknowledged && self.validated
    }

    pub fn new(initiator: Addr) -> Self {
        Self {
            acknowledged: false,
            validated: false,
            initiator: Some(initiator),
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

// CHAIN_TO_BECH32_PREFIX_REVERSE_MAP is a map from a bech32 prefix to the chains that use that prefix
pub const CHAIN_TO_BECH32_PREFIX_REVERSE_MAP: Map<&str, Vec<String>> =
    Map::new(StorageKey::ChainToBech32PrefixReverseMap.to_string());

// CONFIG stores the contract owner
pub const CONFIG: Item<Config> = Item::new(StorageKey::Config.to_string());

// CHAIN_ADMIN_MAP is a map from a source chain to the address that is authorized to add, update, or remove channels for that source chain
// TODO: why isn't this an item?
pub const GLOBAL_ADMIN_MAP: Map<&str, Addr> = Map::new(StorageKey::GlobalAdminMap.to_string());

// CHAIN_ADMIN_MAP is a map from a source chain to the address that is authorized to add, update, or remove channels for that source chain
pub const CHAIN_ADMIN_MAP: Map<&str, Addr> = Map::new(StorageKey::ChainAdminMap.to_string());

// CHAIN_MAINTAINER_MAP is a map from a source chain to the address that is authorized add, enable, or disable channels for that source chain
pub const CHAIN_MAINTAINER_MAP: Map<&str, Addr> =
    Map::new(StorageKey::ChainMaintainerMap.to_string());

// CHAIN_PFM_MAP stores whether a chain supports the Packet Forward Middleware interface for forwarding IBC packets
pub const CHAIN_PFM_MAP: Map<&str, ChainPFM> =
    Map::new(StorageKey::HasPacketForwardMiddleware.to_string());

// DENOM_ALIAS_MAP is a map from a denom path to a denom alias
pub const DENOM_ALIAS_MAP: Map<&str, RegistryValue> =
    Map::new(StorageKey::DenomAliasMap.to_string());

// DENOM_ALIAS_REVERSE_MAP is a map from a denom alias to a denom path
pub const DENOM_ALIAS_REVERSE_MAP: Map<&str, RegistryValue> =
    Map::new(StorageKey::DenomAliasReverseMap.to_string());

#[cw_serde]
pub struct Config {
    pub owner: Addr,
}
