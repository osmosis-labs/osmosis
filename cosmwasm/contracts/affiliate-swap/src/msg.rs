use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Coin, Uint128};
use osmosis_std::types::osmosis::poolmanager::v1beta1::SwapAmountInRoute;

#[cw_serde]
pub struct InstantiateMsg {
    pub owner: String,
    pub affiliate_addr: String,
    pub affiliate_bps: u16, // out of 10_000 (basis points)
}

#[cw_serde]
pub enum ExecuteMsg {
    SwapWithFee {
        input_coin: Coin,
        output_denom: String,
        min_output_amount: Uint128,
        route: Vec<SwapAmountInRoute>,
    },
    UpdateAffiliate {
        affiliate_addr: String,
        affiliate_bps: u16,
    },
    TransferOwnership { new_owner: String },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ConfigResponse)]
    Config {},
}

#[cw_serde]
pub struct ConfigResponse {
    pub owner: String,
    pub affiliate_addr: String,
    pub affiliate_bps: u16,
}

#[cw_serde]
pub struct SwapResponse {
    pub original_sender: String,
    pub token_out_denom: String,
    pub amount_sent_to_user: Uint128,
    pub amount_sent_to_affiliate: Uint128,
}
