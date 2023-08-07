use cosmwasm_schema::cw_serde;
use cosmwasm_std::Coin;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum SudoMsg {
    TrackBeforeSend {
        from: String,
        to: String,
        amount: Coin,
    },
}
