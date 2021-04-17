use cosmwasm_std::{
    Coin,
}

use cosmwasm_storage::{
    Singleton, singleton,
}

pub const PREFIX_RECORD: &[u8] = b"record";

fn key_record(id: String) -> &[u8] {

}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Record {
    pub weight: Uint128,
    pub token: Coin,
}

// TODO: find out more fluent way to access params by both module and contract
// maybe separate gov and swap contracts?
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct PoolParams {
    lock: bool,
    swap_fee: Uint128,
    exit_fee: Uint128,
    swap_fee_governor: str,
}

// Default
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Pool {
    pub params: PoolParams,
    pub total_weight: Uint128,
    pub total_share: Coin,
    pub records: Vec<Record>,
}

pub fn pool(storage: &mut dyn storage, id: Uint128) -> Singleton<Pool> {
    singleton(storage, key_pool(id))
}
