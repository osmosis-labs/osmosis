use osmosis_std_derive::CosmwasmExt;
/// A TWAP record should be indexed in state by pool_id, (asset pair), timestamp
/// The asset pair assets should be lexicographically sorted.
/// Technically (pool_id, asset_0_denom, asset_1_denom, height) do not need to
/// appear in the struct however we view this as the wrong performance tradeoff
/// given SDK today. Would rather we optimize for readability and correctness,
/// than an optimal state storage format. The system bottleneck is elsewhere for
/// now.
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.TwapRecord")]
pub struct TwapRecord {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    /// Lexicographically smaller denom of the pair
    #[prost(string, tag = "2")]
    pub asset0_denom: ::prost::alloc::string::String,
    /// Lexicographically larger denom of the pair
    #[prost(string, tag = "3")]
    pub asset1_denom: ::prost::alloc::string::String,
    /// height this record corresponds to, for debugging purposes
    #[prost(int64, tag = "4")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub height: i64,
    /// This field should only exist until we have a global registry in the state
    /// machine, mapping prior block heights within {TIME RANGE} to times.
    #[prost(message, optional, tag = "5")]
    pub time: ::core::option::Option<crate::shim::Timestamp>,
    /// We store the last spot prices in the struct, so that we can interpolate
    /// accumulator values for times between when accumulator records are stored.
    #[prost(string, tag = "6")]
    pub p0_last_spot_price: ::prost::alloc::string::String,
    #[prost(string, tag = "7")]
    pub p1_last_spot_price: ::prost::alloc::string::String,
    #[prost(string, tag = "8")]
    pub p0_arithmetic_twap_accumulator: ::prost::alloc::string::String,
    #[prost(string, tag = "9")]
    pub p1_arithmetic_twap_accumulator: ::prost::alloc::string::String,
    #[prost(string, tag = "10")]
    pub geometric_twap_accumulator: ::prost::alloc::string::String,
    /// This field contains the time in which the last spot price error occured.
    /// It is used to alert the caller if they are getting a potentially erroneous
    /// TWAP, due to an unforeseen underlying error.
    #[prost(message, optional, tag = "11")]
    pub last_error_time: ::core::option::Option<crate::shim::Timestamp>,
}
/// Params holds parameters for the twap module
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.Params")]
pub struct Params {
    #[prost(string, tag = "1")]
    pub prune_epoch_identifier: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub record_history_keep_period: ::core::option::Option<crate::shim::Duration>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.GenesisState")]
pub struct GenesisState {
    /// twaps is the collection of all twap records.
    #[prost(message, repeated, tag = "1")]
    pub twaps: ::prost::alloc::vec::Vec<TwapRecord>,
    /// params is the container of twap parameters.
    #[prost(message, optional, tag = "2")]
    pub params: ::core::option::Option<Params>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ArithmeticTwapRequest")]
#[proto_query(
    path = "/osmosis.twap.v1beta1.Query/ArithmeticTwap",
    response_type = ArithmeticTwapResponse
)]
pub struct ArithmeticTwapRequest {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    #[prost(string, tag = "2")]
    pub base_asset: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub quote_asset: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "4")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
    #[prost(message, optional, tag = "5")]
    pub end_time: ::core::option::Option<crate::shim::Timestamp>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ArithmeticTwapResponse")]
pub struct ArithmeticTwapResponse {
    #[prost(string, tag = "1")]
    pub arithmetic_twap: ::prost::alloc::string::String,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ArithmeticTwapToNowRequest")]
#[proto_query(
    path = "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow",
    response_type = ArithmeticTwapToNowResponse
)]
pub struct ArithmeticTwapToNowRequest {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    #[prost(string, tag = "2")]
    pub base_asset: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub quote_asset: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "4")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ArithmeticTwapToNowResponse")]
pub struct ArithmeticTwapToNowResponse {
    #[prost(string, tag = "1")]
    pub arithmetic_twap: ::prost::alloc::string::String,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.GeometricTwapRequest")]
#[proto_query(
    path = "/osmosis.twap.v1beta1.Query/GeometricTwap",
    response_type = GeometricTwapResponse
)]
pub struct GeometricTwapRequest {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    #[prost(string, tag = "2")]
    pub base_asset: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub quote_asset: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "4")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
    #[prost(message, optional, tag = "5")]
    pub end_time: ::core::option::Option<crate::shim::Timestamp>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.GeometricTwapResponse")]
pub struct GeometricTwapResponse {
    #[prost(string, tag = "1")]
    pub geometric_twap: ::prost::alloc::string::String,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.GeometricTwapToNowRequest")]
#[proto_query(
    path = "/osmosis.twap.v1beta1.Query/GeometricTwapToNow",
    response_type = GeometricTwapToNowResponse
)]
pub struct GeometricTwapToNowRequest {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    #[prost(string, tag = "2")]
    pub base_asset: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub quote_asset: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "4")]
    pub start_time: ::core::option::Option<crate::shim::Timestamp>,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.GeometricTwapToNowResponse")]
pub struct GeometricTwapToNowResponse {
    #[prost(string, tag = "1")]
    pub geometric_twap: ::prost::alloc::string::String,
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ParamsRequest")]
#[proto_query(
    path = "/osmosis.twap.v1beta1.Query/Params",
    response_type = ParamsResponse
)]
pub struct ParamsRequest {}
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
#[proto_message(type_url = "/osmosis.twap.v1beta1.ParamsResponse")]
pub struct ParamsResponse {
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
}
pub struct TwapQuerier<'a, Q: cosmwasm_std::CustomQuery> {
    querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>,
}
impl<'a, Q: cosmwasm_std::CustomQuery> TwapQuerier<'a, Q> {
    pub fn new(querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>) -> Self {
        Self { querier }
    }
    pub fn params(&self) -> Result<ParamsResponse, cosmwasm_std::StdError> {
        ParamsRequest {}.query(self.querier)
    }
    pub fn arithmetic_twap(
        &self,
        pool_id: u64,
        base_asset: ::prost::alloc::string::String,
        quote_asset: ::prost::alloc::string::String,
        start_time: ::core::option::Option<crate::shim::Timestamp>,
        end_time: ::core::option::Option<crate::shim::Timestamp>,
    ) -> Result<ArithmeticTwapResponse, cosmwasm_std::StdError> {
        ArithmeticTwapRequest {
            pool_id,
            base_asset,
            quote_asset,
            start_time,
            end_time,
        }
        .query(self.querier)
    }
    pub fn arithmetic_twap_to_now(
        &self,
        pool_id: u64,
        base_asset: ::prost::alloc::string::String,
        quote_asset: ::prost::alloc::string::String,
        start_time: ::core::option::Option<crate::shim::Timestamp>,
    ) -> Result<ArithmeticTwapToNowResponse, cosmwasm_std::StdError> {
        ArithmeticTwapToNowRequest {
            pool_id,
            base_asset,
            quote_asset,
            start_time,
        }
        .query(self.querier)
    }
    pub fn geometric_twap(
        &self,
        pool_id: u64,
        base_asset: ::prost::alloc::string::String,
        quote_asset: ::prost::alloc::string::String,
        start_time: ::core::option::Option<crate::shim::Timestamp>,
        end_time: ::core::option::Option<crate::shim::Timestamp>,
    ) -> Result<GeometricTwapResponse, cosmwasm_std::StdError> {
        GeometricTwapRequest {
            pool_id,
            base_asset,
            quote_asset,
            start_time,
            end_time,
        }
        .query(self.querier)
    }
    pub fn geometric_twap_to_now(
        &self,
        pool_id: u64,
        base_asset: ::prost::alloc::string::String,
        quote_asset: ::prost::alloc::string::String,
        start_time: ::core::option::Option<crate::shim::Timestamp>,
    ) -> Result<GeometricTwapToNowResponse, cosmwasm_std::StdError> {
        GeometricTwapToNowRequest {
            pool_id,
            base_asset,
            quote_asset,
            start_time,
        }
        .query(self.querier)
    }
}
