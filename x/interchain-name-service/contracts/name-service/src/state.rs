use std::collections::BinaryHeap;

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
pub static IBC_SUFFIX: &str = ".ibc";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub required_denom: String,
    pub purchase_price: Uint128,
    pub transfer_price: Uint128,
    pub annual_rent_amount: Uint128,
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
    pub amount: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, JsonSchema)]
pub struct NameRecord {
    pub owner: Addr,
    pub expiry: u128,
    pub bids: BinaryHeap<NameBid>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct AddressRecord {
    pub name: String,
    pub expiry: u128,
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
