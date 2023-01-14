use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Coin, Decimal, Uint128};
use osmosis_std::types::osmosis::gamm::v1beta1::SwapAmountInRoute;

#[cw_serde]
pub struct InstantiateMsg {
    pub owner: String,
}

#[cw_serde]
pub enum Slippage {
    Twap {
        window_seconds: Option<u64>,
        slippage_percentage: Decimal,
    },
    MinOutputAmount(Uint128),
}

#[cw_serde]
pub enum ExecuteMsg {
    /// The contract's owner determines how can update the routes. This method
    /// allows the owner to be transfered to someone else.
    TransferOwnership { new_owner: String },
    SetRoute {
        input_denom: String,
        output_denom: String,
        pool_route: Vec<SwapAmountInRoute>,
    },
    Swap {
        input_coin: Coin,
        output_denom: String,
        slippage: Slippage,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetOwnerResponse)]
    GetOwner {},
    #[returns(GetRouteResponse)]
    GetRoute {
        input_denom: String,
        output_denom: String,
    },
}

// Response for GetOwner query
#[cw_serde]
pub struct GetOwnerResponse {
    pub owner: String,
}

// Response for GetRoute query
#[cw_serde]
pub struct GetRouteResponse {
    pub pool_route: Vec<SwapAmountInRoute>,
}

// Response for Swap
#[cw_serde]
pub struct SwapResponse {
    pub original_sender: String,
    pub token_out_denom: String,
    pub amount: Uint128,
}
