use crate::state::SpendLimit;
use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum SudoMsg {
    Authenticate(osmosis_authenticators::AuthenticationRequest),
    Track(osmosis_authenticators::TrackRequest),
    ConfirmExecution(osmosis_authenticators::ConfirmExecutionRequest),
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(SpendLimitDataResponse)]
    GetSpendLimitData { account: Addr },
}

#[cw_serde]
pub struct SpendLimitDataResponse {
    pub spend_limit_data: SpendLimit,
}
