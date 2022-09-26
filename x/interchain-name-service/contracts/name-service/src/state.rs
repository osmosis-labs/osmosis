use std::collections::BinaryHeap;

use cw_utils::{Duration, Expiration};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Storage, Uint128};
use cosmwasm_storage::{
    bucket, bucket_read, singleton, singleton_read, Bucket, ReadonlyBucket, ReadonlySingleton,
    Singleton,
};

pub static NAME_RESOLVER_KEY: &[u8] = b"nameresolver";
pub static ADDRESS_RESOLVER_KEY: &[u8] = b"addressresolver";
pub static CONFIG_KEY: &[u8] = b"config";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    // Denom for all protocol transactions
    pub required_denom: String,
    // Price to intially purchase a name
    pub mint_price: Uint128,
    // Amount of tax paid to protocol annually (as basis points of current price)
    pub annual_tax_bps: Uint128,
    // Amount of time annually where owner can match competing bids to keep the domain
    pub owner_grace_period: Duration,
}

pub fn config(storage: &mut dyn Storage) -> Singleton<Config> {
    singleton(storage, CONFIG_KEY)
}

pub fn config_read(storage: &dyn Storage) -> ReadonlySingleton<Config> {
    singleton_read(storage, CONFIG_KEY)
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize, Clone, Debug, JsonSchema)]
pub struct NameBid {
    pub bidder: Addr,
    pub price: Uint128,
    pub years: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, JsonSchema)]
pub struct NameRecord {
    pub owner: Addr,
    pub expiry: Expiration,
    pub bids: BinaryHeap<NameBid>,
    pub remaining_escrow: Uint128,
    pub current_valuation: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct AddressRecord {
    pub name: String,
    pub expiry: Expiration,
}

pub fn resolver(storage: &mut dyn Storage) -> Bucket<NameRecord> {
    bucket(storage, NAME_RESOLVER_KEY)
}

pub fn resolver_read(storage: &dyn Storage) -> ReadonlyBucket<NameRecord> {
    bucket_read(storage, NAME_RESOLVER_KEY)
}

pub fn reverse_resolver(storage: &mut dyn Storage) -> Bucket<AddressRecord> {
    bucket(storage, ADDRESS_RESOLVER_KEY)
}

pub fn reverse_resolver_read(storage: &dyn Storage) -> ReadonlyBucket<AddressRecord> {
    bucket_read(storage, ADDRESS_RESOLVER_KEY)
}
