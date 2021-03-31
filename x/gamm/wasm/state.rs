use cosmwasm_std::{
    Coin,
}

use cosmwasm_storage::{
    bucket, bucket_read,
}

pub const PREFIX_POOL: &[u8] = b"pool";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Record {
    pub weight: Uint128,
    pub token: Coin,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Pool {
    pub id: Uint128,
    pub params: PoolParams,
    pub total_weight: Uint128,
    pub total_share: Coin,
    pub records: Vec<Record>,
}

pub fn pool(storage: &mut dyn storage) -> Bucket<Pool> {
    bucket(storage, PREFIX_POOL)
}
