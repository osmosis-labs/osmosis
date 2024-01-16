use cosmwasm_schema::cw_serde;
use osmosis_std::types::osmosis::poolmanager::v1beta1::SwapAmountInRoute;

use cosmwasm_std::AllBalanceResponse;
use cw_storage_plus::Map;

pub const USDC_DENOM: &str = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4";
pub const TRACKED_DENOMS_IN_MEMORY: &str = "TBD";

pub const SPEND_LIMITS: Map<String, SpendLimit> = Map::new("sls");

pub const TRACKED_DENOMS: Map<Denom, TrackedDenom> = Map::new("tds");

#[cw_serde]
pub struct SpendLimit {
    pub id: String,
    pub denom: String,
    pub balance: AllBalanceResponse,
    pub amount_left: u128,
    pub block_of_last_tx: u64,
    pub number_of_blocks_active: u64,
}

#[cw_serde]
pub struct TrackedDenom {
    pub denom: Denom,
    pub path: Path,
}

#[cw_serde]
pub struct AuthenticatorParams {
    pub id: String,
    pub duration: u64,
    pub limit: u128,
}

pub type Denom = String;
pub type Path = Vec<SwapAmountInRoute>;
