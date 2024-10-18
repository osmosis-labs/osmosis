use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {
    pub count: i32,
}

#[cw_serde]
pub enum ExecuteMsg {
    Increment {},
    Reset { count: i32 },
}

#[cw_serde]
pub enum SudoMsg {
    Callback {
        job_id: u64,
    },
    Error {
        module_name: String,
        error_code: u32,
        contract_address: String,
        input_payload: String,
        error_message: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    // GetCount returns the current count as a json-encoded number
    #[returns(GetCountResponse)]
    GetCount {},
}

#[cw_serde]
pub struct GetCountResponse {
    pub count: i32,
}
