use osmosis_std_derive::CosmwasmExt;
/// Params defines the parameters for the module.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.Params")]
pub struct Params {
    /// Boolean whether the module is going to be enabled
    #[prost(bool, tag = "1")]
    pub enabled: bool,
}
/// TokenPairArbRoutes tracks all of the hot routes for a given pair of tokens
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.TokenPairArbRoutes")]
pub struct TokenPairArbRoutes {
    /// Stores all of the possible hot paths for a given pair of tokens
    #[prost(message, repeated, tag = "1")]
    pub arb_routes: ::prost::alloc::vec::Vec<Route>,
    /// Token denomination of the first asset
    #[prost(string, tag = "2")]
    pub token_in: ::prost::alloc::string::String,
    /// Token denomination of the second asset
    #[prost(string, tag = "3")]
    pub token_out: ::prost::alloc::string::String,
}
/// Route is a hot route for a given pair of tokens
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.Route")]
pub struct Route {
    /// The pool IDs that are travered in the directed cyclic graph (traversed left
    /// -> right)
    #[prost(message, repeated, tag = "1")]
    pub trades: ::prost::alloc::vec::Vec<Trade>,
}
/// Trade is a single trade in a route
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.Trade")]
pub struct Trade {
    /// The pool IDs that are travered in the directed cyclic graph (traversed left
    /// -> right)
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool: u64,
    /// The denom of token A that is traded
    #[prost(string, tag = "2")]
    pub token_in: ::prost::alloc::string::String,
    /// The denom of token B that is traded
    #[prost(string, tag = "3")]
    pub token_out: ::prost::alloc::string::String,
}
/// PoolStatistics contains the number of trades the module has executed after a
/// swap on a given pool and the profits from the trades
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.PoolStatistics")]
pub struct PoolStatistics {
    /// profits is the total profit from all trades on this pool
    #[prost(message, repeated, tag = "1")]
    pub profits: ::prost::alloc::vec::Vec<super::super::super::cosmos::base::v1beta1::Coin>,
    /// number_of_trades is the number of trades the module has executed
    #[prost(string, tag = "2")]
    pub number_of_trades: ::prost::alloc::string::String,
    /// pool_id is the id of the pool
    #[prost(uint64, tag = "3")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
}
/// SetProtoRevEnabledProposal is a gov Content type to update whether the
/// protorev module is enabled
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.SetProtoRevEnabledProposal")]
pub struct SetProtoRevEnabledProposal {
    #[prost(string, tag = "1")]
    pub title: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub description: ::prost::alloc::string::String,
    #[prost(bool, tag = "3")]
    pub enabled: bool,
}
/// SetProtoRevAdminAccountProposal is a gov Content type to set the admin
/// account that will receive permissions to alter hot routes and set the
/// developer address that will be receiving a share of profits from the module
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.SetProtoRevAdminAccountProposal")]
pub struct SetProtoRevAdminAccountProposal {
    #[prost(string, tag = "1")]
    pub title: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub description: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub account: ::prost::alloc::string::String,
}
/// GenesisState defines the protorev module's genesis state.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.GenesisState")]
pub struct GenesisState {
    /// Module Parameters
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
    /// Hot routes that are configured on genesis
    #[prost(message, repeated, tag = "2")]
    pub token_pairs: ::prost::alloc::vec::Vec<TokenPairArbRoutes>,
}
/// MsgSetHotRoutes defines the Msg/SetHotRoutes request type.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.MsgSetHotRoutes")]
pub struct MsgSetHotRoutes {
    /// admin is the account that is authorized to set the hot routes.
    #[prost(string, tag = "1")]
    pub admin: ::prost::alloc::string::String,
    /// hot_routes is the list of hot routes to set.
    #[prost(message, repeated, tag = "2")]
    pub hot_routes: ::prost::alloc::vec::Vec<TokenPairArbRoutes>,
}
/// MsgSetHotRoutesResponse defines the Msg/SetHotRoutes response type.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.MsgSetHotRoutesResponse")]
pub struct MsgSetHotRoutesResponse {}
/// MsgSetDeveloperAccount defines the Msg/SetDeveloperAccount request type.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.MsgSetDeveloperAccount")]
pub struct MsgSetDeveloperAccount {
    /// admin is the account that is authorized to set the developer account.
    #[prost(string, tag = "1")]
    pub admin: ::prost::alloc::string::String,
    /// developer_account is the account that will receive a portion of the profits
    /// from the protorev module.
    #[prost(string, tag = "2")]
    pub developer_account: ::prost::alloc::string::String,
}
/// MsgSetDeveloperAccountResponse defines the Msg/SetDeveloperAccount response
/// type.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.MsgSetDeveloperAccountResponse")]
pub struct MsgSetDeveloperAccountResponse {}
/// QueryParamsRequest is request type for the Query/Params RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryParamsRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/Params",
    response_type = QueryParamsResponse
)]
pub struct QueryParamsRequest {}
/// QueryParamsResponse is response type for the Query/Params RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryParamsResponse")]
pub struct QueryParamsResponse {
    /// params holds all the parameters of this module.
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
}
/// QueryGetProtoRevNumberOfTradesRequest is request type for the
/// Query/GetProtoRevNumberOfTrades RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevNumberOfTradesRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevNumberOfTrades",
    response_type = QueryGetProtoRevNumberOfTradesResponse
)]
pub struct QueryGetProtoRevNumberOfTradesRequest {}
/// QueryGetProtoRevNumberOfTradesResponse is response type for the
/// Query/GetProtoRevNumberOfTrades RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevNumberOfTradesResponse")]
pub struct QueryGetProtoRevNumberOfTradesResponse {
    /// number_of_trades is the number of trades the module has executed
    #[prost(string, tag = "1")]
    pub number_of_trades: ::prost::alloc::string::String,
}
/// QueryGetProtoRevProfitsByDenomRequest is request type for the
/// Query/GetProtoRevProfitsByDenom RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevProfitsByDenomRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevProfitsByDenom",
    response_type = QueryGetProtoRevProfitsByDenomResponse
)]
pub struct QueryGetProtoRevProfitsByDenomRequest {
    /// denom is the denom to query profits by
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
/// QueryGetProtoRevProfitsByDenomResponse is response type for the
/// Query/GetProtoRevProfitsByDenom RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevProfitsByDenomResponse")]
pub struct QueryGetProtoRevProfitsByDenomResponse {
    /// profit is the profits of the module by the selected denom
    #[prost(message, optional, tag = "1")]
    pub profit: ::core::option::Option<super::super::super::cosmos::base::v1beta1::Coin>,
}
/// QueryGetProtoRevAllProfitsRequest is request type for the
/// Query/GetProtoRevAllProfits RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevAllProfitsRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevAllProfits",
    response_type = QueryGetProtoRevAllProfitsResponse
)]
pub struct QueryGetProtoRevAllProfitsRequest {}
/// QueryGetProtoRevAllProfitsResponse is response type for the
/// Query/GetProtoRevAllProfits RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevAllProfitsResponse")]
pub struct QueryGetProtoRevAllProfitsResponse {
    /// profits is a list of all of the profits from the module
    #[prost(message, repeated, tag = "1")]
    pub profits: ::prost::alloc::vec::Vec<super::super::super::cosmos::base::v1beta1::Coin>,
}
/// QueryGetProtoRevStatisticsByPoolRequest is request type for the
/// Query/GetProtoRevStatisticsByPool RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevStatisticsByPoolRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevStatisticsByPool",
    response_type = QueryGetProtoRevStatisticsByPoolResponse
)]
pub struct QueryGetProtoRevStatisticsByPoolRequest {
    /// pool_id is the pool id to query statistics by
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
}
/// QueryGetProtoRevStatisticsByPoolResponse is response type for the
/// Query/GetProtoRevStatisticsByPool RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevStatisticsByPoolResponse")]
pub struct QueryGetProtoRevStatisticsByPoolResponse {
    /// statistics contains the number of trades the module has executed after a
    /// swap on a given pool and the profits from the trades
    #[prost(message, optional, tag = "1")]
    pub statistics: ::core::option::Option<PoolStatistics>,
}
/// QueryGetProtoRevAllStatisticsRequest is request type for the
/// Query/GetProtoRevAllStatistics RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevAllStatisticsRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevAllStatistics",
    response_type = QueryGetProtoRevAllStatisticsResponse
)]
pub struct QueryGetProtoRevAllStatisticsRequest {}
/// QueryGetProtoRevAllStatisticsResponse is response type for the
/// Query/GetProtoRevAllStatistics RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevAllStatisticsResponse")]
pub struct QueryGetProtoRevAllStatisticsResponse {
    /// statistics contains the number of trades the module has executed after a
    /// swap on a given pool and the profits from the trades for all pools
    #[prost(message, repeated, tag = "1")]
    pub statistics: ::prost::alloc::vec::Vec<PoolStatistics>,
}
/// QueryGetProtoRevTokenPairArbRoutesRequest is request type for the
/// Query/GetProtoRevTokenPairArbRoutes RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevTokenPairArbRoutesRequest")]
#[proto_query(
    path = "/osmosis.protorev.v1beta1.Query/GetProtoRevTokenPairArbRoutes",
    response_type = QueryGetProtoRevTokenPairArbRoutesResponse
)]
pub struct QueryGetProtoRevTokenPairArbRoutesRequest {}
/// QueryGetProtoRevTokenPairArbRoutesResponse is response type for the
/// Query/GetProtoRevTokenPairArbRoutes RPC method.
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
#[proto_message(type_url = "/osmosis.protorev.v1beta1.QueryGetProtoRevTokenPairArbRoutesResponse")]
pub struct QueryGetProtoRevTokenPairArbRoutesResponse {
    /// routes is a list of all of the hot routes that the module is currently
    /// arbitraging
    #[prost(message, repeated, tag = "1")]
    pub routes: ::prost::alloc::vec::Vec<TokenPairArbRoutes>,
}
pub struct ProtorevQuerier<'a, Q: cosmwasm_std::CustomQuery> {
    querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>,
}
impl<'a, Q: cosmwasm_std::CustomQuery> ProtorevQuerier<'a, Q> {
    pub fn new(querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>) -> Self {
        Self { querier }
    }
    pub fn params(&self) -> Result<QueryParamsResponse, cosmwasm_std::StdError> {
        QueryParamsRequest {}.query(self.querier)
    }
    pub fn get_proto_rev_number_of_trades(
        &self,
    ) -> Result<QueryGetProtoRevNumberOfTradesResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevNumberOfTradesRequest {}.query(self.querier)
    }
    pub fn get_proto_rev_profits_by_denom(
        &self,
        denom: ::prost::alloc::string::String,
    ) -> Result<QueryGetProtoRevProfitsByDenomResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevProfitsByDenomRequest { denom }.query(self.querier)
    }
    pub fn get_proto_rev_all_profits(
        &self,
    ) -> Result<QueryGetProtoRevAllProfitsResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevAllProfitsRequest {}.query(self.querier)
    }
    pub fn get_proto_rev_statistics_by_pool(
        &self,
        pool_id: u64,
    ) -> Result<QueryGetProtoRevStatisticsByPoolResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevStatisticsByPoolRequest { pool_id }.query(self.querier)
    }
    pub fn get_proto_rev_all_statistics(
        &self,
    ) -> Result<QueryGetProtoRevAllStatisticsResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevAllStatisticsRequest {}.query(self.querier)
    }
    pub fn get_proto_rev_token_pair_arb_routes(
        &self,
    ) -> Result<QueryGetProtoRevTokenPairArbRoutesResponse, cosmwasm_std::StdError> {
        QueryGetProtoRevTokenPairArbRoutesRequest {}.query(self.querier)
    }
}
