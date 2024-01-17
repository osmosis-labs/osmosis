use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum SudoMsg {
    Count {
        amount: i64,
    },
}
