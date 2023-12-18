use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

#[cw_serde]
pub struct PubkeysResponse {
    pub pubkeys: Vec<Binary>,
}

#[cw_serde]
pub struct Signature {
    pub salt: Binary,
    pub signature: Binary,
}
