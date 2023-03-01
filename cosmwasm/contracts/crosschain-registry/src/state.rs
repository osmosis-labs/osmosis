use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

// CONTRACT_ALIAS_MAP is a map from a contract alias to a contract address
pub const CONTRACT_ALIAS_MAP: Map<&str, String> = Map::new("contract_alias_map");

// CHAIN_TO_CHAIN_CHANNEL_MAP is a map from a source<>destination chain pair to its respective channel id
pub const CHAIN_TO_CHAIN_CHANNEL_MAP: Map<(&str, &str), String> =
    Map::new("chain_to_chain_channel_map");

// CHANNEL_TO_CHAIN_CHAIN_MAP is a map from a channel id on a source chain to its respective destination chain
pub const CHANNEL_ON_CHAIN_CHAIN_MAP: Map<(&str, &str), String> =
    Map::new("channel_to_chain_chain_map");

pub const STATE: Item<State> = Item::new("state");

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct State {
    pub owner: Addr,
}
