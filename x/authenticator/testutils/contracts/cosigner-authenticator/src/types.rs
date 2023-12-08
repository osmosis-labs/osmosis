use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

#[cw_serde]
pub enum Pubkey {
    ByName(String),
    Raw(Binary),
}

#[cw_serde]
pub struct PubkeysResponse {
    pub pubkeys: Vec<Pubkey>,
}
