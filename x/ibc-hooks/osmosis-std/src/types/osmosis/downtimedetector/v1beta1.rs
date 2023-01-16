use osmosis_std_derive::CosmwasmExt;
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Downtime {
    Duration30s = 0,
    Duration1m = 1,
    Duration2m = 2,
    Duration3m = 3,
    Duration4m = 4,
    Duration5m = 5,
    Duration10m = 6,
    Duration20m = 7,
    Duration30m = 8,
    Duration40m = 9,
    Duration50m = 10,
    Duration1h = 11,
    Duration15h = 12,
    Duration2h = 13,
    Duration25h = 14,
    Duration3h = 15,
    Duration4h = 16,
    Duration5h = 17,
    Duration6h = 18,
    Duration9h = 19,
    Duration12h = 20,
    Duration18h = 21,
    Duration24h = 22,
    Duration36h = 23,
    Duration48h = 24,
}
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/osmosis.downtimedetector.v1beta1.GenesisDowntimeEntry")]
pub struct GenesisDowntimeEntry {
    #[prost(enumeration = "Downtime", tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub duration: i32,
    #[prost(message, optional, tag = "2")]
    pub last_downtime: ::core::option::Option<crate::shim::Timestamp>,
}
/// GenesisState defines the twap module's genesis state.
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/osmosis.downtimedetector.v1beta1.GenesisState")]
pub struct GenesisState {
    #[prost(message, repeated, tag = "1")]
    pub downtimes: ::prost::alloc::vec::Vec<GenesisDowntimeEntry>,
    #[prost(message, optional, tag = "2")]
    pub last_block_time: ::core::option::Option<crate::shim::Timestamp>,
}
/// Query for has it been at least $RECOVERY_DURATION units of time,
/// since the chain has been down for $DOWNTIME_DURATION.
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(
    type_url = "/osmosis.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest"
)]
#[proto_query(
    path = "/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength",
    response_type = RecoveredSinceDowntimeOfLengthResponse
)]
pub struct RecoveredSinceDowntimeOfLengthRequest {
    #[prost(enumeration = "Downtime", tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub downtime: i32,
    #[prost(message, optional, tag = "2")]
    pub recovery: ::core::option::Option<crate::shim::Duration>,
}
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(
    type_url = "/osmosis.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse"
)]
pub struct RecoveredSinceDowntimeOfLengthResponse {
    #[prost(bool, tag = "1")]
    pub succesfully_recovered: bool,
}
pub struct DowntimedetectorQuerier<'a, Q: cosmwasm_std::CustomQuery> {
    querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>,
}
impl<'a, Q: cosmwasm_std::CustomQuery> DowntimedetectorQuerier<'a, Q> {
    pub fn new(querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>) -> Self {
        Self { querier }
    }
    pub fn recovered_since_downtime_of_length(
        &self,
        downtime: i32,
        recovery: ::core::option::Option<crate::shim::Duration>,
    ) -> Result<RecoveredSinceDowntimeOfLengthResponse, cosmwasm_std::StdError> {
        RecoveredSinceDowntimeOfLengthRequest { downtime, recovery }.query(self.querier)
    }
}
