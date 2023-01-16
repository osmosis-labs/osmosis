use osmosis_std_derive::CosmwasmExt;
/// Params holds parameters for the incentives module
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
#[proto_message(type_url = "/osmosis.incentives.Params")]
pub struct Params {
    /// distr_epoch_identifier is what epoch type distribution will be triggered by
    /// (day, week, etc.)
    #[prost(string, tag = "1")]
    pub distr_epoch_identifier: ::prost::alloc::string::String,
}
/// Gauge is an object that stores and distributes yields to recipients who
/// satisfy certain conditions. Currently gauges support conditions around the
/// duration for which a given denom is locked.
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
#[proto_message(type_url = "/osmosis.incentives.Gauge")]
pub struct Gauge {
    /// id is the unique ID of a Gauge
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub id: u64,
    /// is_perpetual is a flag to show if it's a perpetual or non-perpetual gauge
    /// Non-perpetual gauges distribute their tokens equally per epoch while the
    /// gauge is in the active period. Perpetual gauges distribute all their tokens
    /// at a single time and only distribute their tokens again once the gauge is
    /// refilled, Intended for use with incentives that get refilled daily.
    #[prost(bool, tag = "2")]
    pub is_perpetual: bool,
    /// distribute_to is where the gauge rewards are distributed to.
    /// This is queried via lock duration or by timestamp
    #[prost(message, optional, tag = "3")]
    pub distribute_to: ::core::option::Option<super::lockup::QueryCondition>,
    /// coins is the total amount of coins that have been in the gauge
    /// Can distribute multiple coin denoms
    #[prost(message, repeated, tag = "4")]
    pub coins: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
    /// start_time is the distribution start time
    #[prost(message, optional, tag = "5")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
    /// num_epochs_paid_over is the number of total epochs distribution will be
    /// completed over
    #[prost(uint64, tag = "6")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub num_epochs_paid_over: u64,
    /// filled_epochs is the number of epochs distribution has been completed on
    /// already
    #[prost(uint64, tag = "7")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub filled_epochs: u64,
    /// distributed_coins are coins that have been distributed already
    #[prost(message, repeated, tag = "8")]
    pub distributed_coins: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
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
#[proto_message(type_url = "/osmosis.incentives.LockableDurationsInfo")]
pub struct LockableDurationsInfo {
    /// List of incentivised durations that gauges will pay out to
    #[prost(message, repeated, tag = "1")]
    pub lockable_durations: ::prost::alloc::vec::Vec<crate::shim::Duration>,
}
/// GenesisState defines the incentives module's various parameters when first
/// initialized
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
#[proto_message(type_url = "/osmosis.incentives.GenesisState")]
pub struct GenesisState {
    /// params are all the parameters of the module
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
    /// gauges are all gauges that should exist at genesis
    #[prost(message, repeated, tag = "2")]
    pub gauges: ::prost::alloc::vec::Vec<Gauge>,
    /// lockable_durations are all lockup durations that gauges can be locked for
    /// in order to recieve incentives
    #[prost(message, repeated, tag = "3")]
    pub lockable_durations: ::prost::alloc::vec::Vec<crate::shim::Duration>,
    /// last_gauge_id is what the gauge number will increment from when creating
    /// the next gauge after genesis
    #[prost(uint64, tag = "4")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub last_gauge_id: u64,
}
/// MsgCreateGauge creates a gague to distribute rewards to users
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
#[proto_message(type_url = "/osmosis.incentives.MsgCreateGauge")]
pub struct MsgCreateGauge {
    /// is_perpetual shows if it's a perpetual or non-perpetual gauge
    /// Non-perpetual gauges distribute their tokens equally per epoch while the
    /// gauge is in the active period. Perpetual gauges distribute all their tokens
    /// at a single time and only distribute their tokens again once the gauge is
    /// refilled
    #[prost(bool, tag = "1")]
    pub is_perpetual: bool,
    /// owner is the address of gauge creator
    #[prost(string, tag = "2")]
    pub owner: ::prost::alloc::string::String,
    /// distribute_to show which lock the gauge should distribute to by time
    /// duration or by timestamp
    #[prost(message, optional, tag = "3")]
    pub distribute_to: ::core::option::Option<super::lockup::QueryCondition>,
    /// coins are coin(s) to be distributed by the gauge
    #[prost(message, repeated, tag = "4")]
    pub coins: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
    /// start_time is the distribution start time
    #[prost(message, optional, tag = "5")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
    /// num_epochs_paid_over is the number of epochs distribution will be completed
    /// over
    #[prost(uint64, tag = "6")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub num_epochs_paid_over: u64,
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
#[proto_message(type_url = "/osmosis.incentives.MsgCreateGaugeResponse")]
pub struct MsgCreateGaugeResponse {}
/// MsgAddToGauge adds coins to a previously created gauge
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
#[proto_message(type_url = "/osmosis.incentives.MsgAddToGauge")]
pub struct MsgAddToGauge {
    /// owner is the gauge owner's address
    #[prost(string, tag = "1")]
    pub owner: ::prost::alloc::string::String,
    /// gauge_id is the ID of gauge that rewards are getting added to
    #[prost(uint64, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub gauge_id: u64,
    /// rewards are the coin(s) to add to gauge
    #[prost(message, repeated, tag = "3")]
    pub rewards: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
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
#[proto_message(type_url = "/osmosis.incentives.MsgAddToGaugeResponse")]
pub struct MsgAddToGaugeResponse {}
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
#[proto_message(type_url = "/osmosis.incentives.ModuleToDistributeCoinsRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/ModuleToDistributeCoins",
    response_type = ModuleToDistributeCoinsResponse
)]
pub struct ModuleToDistributeCoinsRequest {}
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
#[proto_message(type_url = "/osmosis.incentives.ModuleToDistributeCoinsResponse")]
pub struct ModuleToDistributeCoinsResponse {
    /// Coins that have yet to be distributed
    #[prost(message, repeated, tag = "1")]
    pub coins: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
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
#[proto_message(type_url = "/osmosis.incentives.GaugeByIDRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/GaugeByID",
    response_type = GaugeByIdResponse
)]
pub struct GaugeByIdRequest {
    /// Gague ID being queried
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub id: u64,
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
#[proto_message(type_url = "/osmosis.incentives.GaugeByIDResponse")]
pub struct GaugeByIdResponse {
    /// Gauge that corresponds to provided gague ID
    #[prost(message, optional, tag = "1")]
    pub gauge: ::core::option::Option<Gauge>,
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
#[proto_message(type_url = "/osmosis.incentives.GaugesRequest")]
#[proto_query(path = "/osmosis.incentives.Query/Gauges", response_type = GaugesResponse)]
pub struct GaugesRequest {
    /// Pagination defines pagination for the request
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
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
#[proto_message(type_url = "/osmosis.incentives.GaugesResponse")]
pub struct GaugesResponse {
    /// Upcoming and active gauges
    #[prost(message, repeated, tag = "1")]
    pub data: ::prost::alloc::vec::Vec<Gauge>,
    /// Pagination defines pagination for the response
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageResponse>,
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
#[proto_message(type_url = "/osmosis.incentives.ActiveGaugesRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/ActiveGauges",
    response_type = ActiveGaugesResponse
)]
pub struct ActiveGaugesRequest {
    /// Pagination defines pagination for the request
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
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
#[proto_message(type_url = "/osmosis.incentives.ActiveGaugesResponse")]
pub struct ActiveGaugesResponse {
    /// Active gagues only
    #[prost(message, repeated, tag = "1")]
    pub data: ::prost::alloc::vec::Vec<Gauge>,
    /// Pagination defines pagination for the response
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageResponse>,
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
#[proto_message(type_url = "/osmosis.incentives.ActiveGaugesPerDenomRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/ActiveGaugesPerDenom",
    response_type = ActiveGaugesPerDenomResponse
)]
pub struct ActiveGaugesPerDenomRequest {
    /// Desired denom when querying active gagues
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    /// Pagination defines pagination for the request
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
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
#[proto_message(type_url = "/osmosis.incentives.ActiveGaugesPerDenomResponse")]
pub struct ActiveGaugesPerDenomResponse {
    /// Active gagues that match denom in query
    #[prost(message, repeated, tag = "1")]
    pub data: ::prost::alloc::vec::Vec<Gauge>,
    /// Pagination defines pagination for the response
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageResponse>,
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
#[proto_message(type_url = "/osmosis.incentives.UpcomingGaugesRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/UpcomingGauges",
    response_type = UpcomingGaugesResponse
)]
pub struct UpcomingGaugesRequest {
    /// Pagination defines pagination for the request
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
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
#[proto_message(type_url = "/osmosis.incentives.UpcomingGaugesResponse")]
pub struct UpcomingGaugesResponse {
    /// Gauges whose distribution is upcoming
    #[prost(message, repeated, tag = "1")]
    pub data: ::prost::alloc::vec::Vec<Gauge>,
    /// Pagination defines pagination for the response
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageResponse>,
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
#[proto_message(type_url = "/osmosis.incentives.UpcomingGaugesPerDenomRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/UpcomingGaugesPerDenom",
    response_type = UpcomingGaugesPerDenomResponse
)]
pub struct UpcomingGaugesPerDenomRequest {
    /// Filter for upcoming gagues that match specific denom
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    /// Pagination defines pagination for the request
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
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
#[proto_message(type_url = "/osmosis.incentives.UpcomingGaugesPerDenomResponse")]
pub struct UpcomingGaugesPerDenomResponse {
    /// Upcoming gagues that match denom in query
    #[prost(message, repeated, tag = "1")]
    pub upcoming_gauges: ::prost::alloc::vec::Vec<Gauge>,
    /// Pagination defines pagination for the response
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageResponse>,
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
#[proto_message(type_url = "/osmosis.incentives.RewardsEstRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/RewardsEst",
    response_type = RewardsEstResponse
)]
pub struct RewardsEstRequest {
    /// Address that is being queried for future estimated rewards
    #[prost(string, tag = "1")]
    pub owner: ::prost::alloc::string::String,
    /// Lock IDs included in future reward estimation
    #[prost(uint64, repeated, tag = "2")]
    pub lock_ids: ::prost::alloc::vec::Vec<u64>,
    /// Upper time limit of reward estimation
    /// Lower limit is current epoch
    #[prost(int64, tag = "3")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub end_epoch: i64,
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
#[proto_message(type_url = "/osmosis.incentives.RewardsEstResponse")]
pub struct RewardsEstResponse {
    /// Estimated coin rewards that will be recieved at provided address
    /// from specified locks between current time and end epoch
    #[prost(message, repeated, tag = "1")]
    pub coins: ::prost::alloc::vec::Vec<super::super::cosmos::base::v1beta1::Coin>,
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
#[proto_message(type_url = "/osmosis.incentives.QueryLockableDurationsRequest")]
#[proto_query(
    path = "/osmosis.incentives.Query/LockableDurations",
    response_type = QueryLockableDurationsResponse
)]
pub struct QueryLockableDurationsRequest {}
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
#[proto_message(type_url = "/osmosis.incentives.QueryLockableDurationsResponse")]
pub struct QueryLockableDurationsResponse {
    /// Time durations that users can lock coins for in order to recieve rewards
    #[prost(message, repeated, tag = "1")]
    pub lockable_durations: ::prost::alloc::vec::Vec<crate::shim::Duration>,
}
pub struct IncentivesQuerier<'a, Q: cosmwasm_std::CustomQuery> {
    querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>,
}
impl<'a, Q: cosmwasm_std::CustomQuery> IncentivesQuerier<'a, Q> {
    pub fn new(querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>) -> Self {
        Self { querier }
    }
    pub fn module_to_distribute_coins(
        &self,
    ) -> Result<ModuleToDistributeCoinsResponse, cosmwasm_std::StdError> {
        ModuleToDistributeCoinsRequest {}.query(self.querier)
    }
    pub fn gauge_by_id(&self, id: u64) -> Result<GaugeByIdResponse, cosmwasm_std::StdError> {
        GaugeByIdRequest { id }.query(self.querier)
    }
    pub fn gauges(
        &self,
        pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
    ) -> Result<GaugesResponse, cosmwasm_std::StdError> {
        GaugesRequest { pagination }.query(self.querier)
    }
    pub fn active_gauges(
        &self,
        pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
    ) -> Result<ActiveGaugesResponse, cosmwasm_std::StdError> {
        ActiveGaugesRequest { pagination }.query(self.querier)
    }
    pub fn active_gauges_per_denom(
        &self,
        denom: ::prost::alloc::string::String,
        pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
    ) -> Result<ActiveGaugesPerDenomResponse, cosmwasm_std::StdError> {
        ActiveGaugesPerDenomRequest { denom, pagination }.query(self.querier)
    }
    pub fn upcoming_gauges(
        &self,
        pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
    ) -> Result<UpcomingGaugesResponse, cosmwasm_std::StdError> {
        UpcomingGaugesRequest { pagination }.query(self.querier)
    }
    pub fn upcoming_gauges_per_denom(
        &self,
        denom: ::prost::alloc::string::String,
        pagination: ::core::option::Option<super::super::cosmos::base::query::v1beta1::PageRequest>,
    ) -> Result<UpcomingGaugesPerDenomResponse, cosmwasm_std::StdError> {
        UpcomingGaugesPerDenomRequest { denom, pagination }.query(self.querier)
    }
    pub fn rewards_est(
        &self,
        owner: ::prost::alloc::string::String,
        lock_ids: ::prost::alloc::vec::Vec<u64>,
        end_epoch: i64,
    ) -> Result<RewardsEstResponse, cosmwasm_std::StdError> {
        RewardsEstRequest {
            owner,
            lock_ids,
            end_epoch,
        }
        .query(self.querier)
    }
    pub fn lockable_durations(
        &self,
    ) -> Result<QueryLockableDurationsResponse, cosmwasm_std::StdError> {
        QueryLockableDurationsRequest {}.query(self.querier)
    }
}
