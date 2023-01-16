use osmosis_std_derive::CosmwasmExt;
/// Params defines the parameters for the bank module.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Params")]
pub struct Params {
    #[prost(message, repeated, tag = "1")]
    pub send_enabled: ::prost::alloc::vec::Vec<SendEnabled>,
    #[prost(bool, tag = "2")]
    pub default_send_enabled: bool,
}
/// SendEnabled maps coin denom to a send_enabled status (whether a denom is
/// sendable).
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.SendEnabled")]
pub struct SendEnabled {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    #[prost(bool, tag = "2")]
    pub enabled: bool,
}
/// Input models transaction input.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Input")]
pub struct Input {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub coins: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// Output models transaction outputs.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Output")]
pub struct Output {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub coins: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// Supply represents a struct that passively keeps track of the total supply
/// amounts in the network.
/// This message is deprecated now that supply is indexed by denom.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Supply")]
#[deprecated]
pub struct Supply {
    #[prost(message, repeated, tag = "1")]
    pub total: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// DenomUnit represents a struct that describes a given
/// denomination unit of the basic token.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.DenomUnit")]
pub struct DenomUnit {
    /// denom represents the string name of the given denom unit (e.g uatom).
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    /// exponent represents power of 10 exponent that one must
    /// raise the base_denom to in order to equal the given DenomUnit's denom
    /// 1 denom = 1^exponent base_denom
    /// (e.g. with a base_denom of uatom, one can create a DenomUnit of 'atom' with
    /// exponent = 6, thus: 1 atom = 10^6 uatom).
    #[prost(uint32, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub exponent: u32,
    /// aliases is a list of string aliases for the given denom
    #[prost(string, repeated, tag = "3")]
    pub aliases: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// Metadata represents a struct that describes
/// a basic token.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Metadata")]
pub struct Metadata {
    #[prost(string, tag = "1")]
    pub description: ::prost::alloc::string::String,
    /// denom_units represents the list of DenomUnit's for a given coin
    #[prost(message, repeated, tag = "2")]
    pub denom_units: ::prost::alloc::vec::Vec<DenomUnit>,
    /// base represents the base denom (should be the DenomUnit with exponent = 0).
    #[prost(string, tag = "3")]
    pub base: ::prost::alloc::string::String,
    /// display indicates the suggested denom that should be
    /// displayed in clients.
    #[prost(string, tag = "4")]
    pub display: ::prost::alloc::string::String,
    /// name defines the name of the token (eg: Cosmos Atom)
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(string, tag = "5")]
    pub name: ::prost::alloc::string::String,
    /// symbol is the token symbol usually shown on exchanges (eg: ATOM). This can
    /// be the same as the display.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(string, tag = "6")]
    pub symbol: ::prost::alloc::string::String,
}
/// SendAuthorization allows the grantee to spend up to spend_limit coins from
/// the granter's account.
///
/// Since: cosmos-sdk 0.43
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.SendAuthorization")]
pub struct SendAuthorization {
    #[prost(message, repeated, tag = "1")]
    pub spend_limit: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// GenesisState defines the bank module's genesis state.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.GenesisState")]
pub struct GenesisState {
    /// params defines all the paramaters of the module.
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
    /// balances is an array containing the balances of all the accounts.
    #[prost(message, repeated, tag = "2")]
    pub balances: ::prost::alloc::vec::Vec<Balance>,
    /// supply represents the total supply. If it is left empty, then supply will be calculated based on the provided
    /// balances. Otherwise, it will be used to validate that the sum of the balances equals this amount.
    #[prost(message, repeated, tag = "3")]
    pub supply: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
    /// denom_metadata defines the metadata of the differents coins.
    #[prost(message, repeated, tag = "4")]
    pub denom_metadata: ::prost::alloc::vec::Vec<Metadata>,
    /// supply_offsets defines the amount of supply offset.
    #[prost(message, repeated, tag = "5")]
    pub supply_offsets: ::prost::alloc::vec::Vec<GenesisSupplyOffset>,
}
/// Balance defines an account address and balance pair used in the bank module's
/// genesis state.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.Balance")]
pub struct Balance {
    /// address is the address of the balance holder.
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    /// coins defines the different coins this balance holds.
    #[prost(message, repeated, tag = "2")]
    pub coins: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// GenesisSupplyOffset encodes the supply offsets, just for genesis.
/// The offsets are serialized directly by denom in state.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.GenesisSupplyOffset")]
pub struct GenesisSupplyOffset {
    /// Denom
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    /// SupplyOffset
    #[prost(string, tag = "2")]
    pub offset: ::prost::alloc::string::String,
}
/// MsgSend represents a message to send coins from one account to another.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.MsgSend")]
pub struct MsgSend {
    #[prost(string, tag = "1")]
    pub from_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub to_address: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "3")]
    pub amount: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
}
/// MsgSendResponse defines the Msg/Send response type.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.MsgSendResponse")]
pub struct MsgSendResponse {}
/// MsgMultiSend represents an arbitrary multi-in, multi-out send message.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.MsgMultiSend")]
pub struct MsgMultiSend {
    #[prost(message, repeated, tag = "1")]
    pub inputs: ::prost::alloc::vec::Vec<Input>,
    #[prost(message, repeated, tag = "2")]
    pub outputs: ::prost::alloc::vec::Vec<Output>,
}
/// MsgMultiSendResponse defines the Msg/MultiSend response type.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.MsgMultiSendResponse")]
pub struct MsgMultiSendResponse {}
/// QueryBalanceRequest is the request type for the Query/Balance RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryBalanceRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/Balance",
    response_type = QueryBalanceResponse
)]
pub struct QueryBalanceRequest {
    /// address is the address to query balances for.
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    /// denom is the coin denom to query balances for.
    #[prost(string, tag = "2")]
    pub denom: ::prost::alloc::string::String,
}
/// QueryBalanceResponse is the response type for the Query/Balance RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryBalanceResponse")]
pub struct QueryBalanceResponse {
    /// balance is the balance of the coin.
    #[prost(message, optional, tag = "1")]
    pub balance: ::core::option::Option<super::super::base::v1beta1::Coin>,
}
/// QueryBalanceRequest is the request type for the Query/AllBalances RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryAllBalancesRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/AllBalances",
    response_type = QueryAllBalancesResponse
)]
pub struct QueryAllBalancesRequest {
    /// address is the address to query balances for.
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    /// pagination defines an optional pagination for the request.
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
}
/// QueryAllBalancesResponse is the response type for the Query/AllBalances RPC
/// method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryAllBalancesResponse")]
pub struct QueryAllBalancesResponse {
    /// balances is the balances of all the coins.
    #[prost(message, repeated, tag = "1")]
    pub balances: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
    /// pagination defines the pagination in the response.
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageResponse>,
}
/// QueryTotalSupplyRequest is the request type for the Query/TotalSupply RPC
/// method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryTotalSupplyRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/TotalSupply",
    response_type = QueryTotalSupplyResponse
)]
pub struct QueryTotalSupplyRequest {
    /// pagination defines an optional pagination for the request.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
}
/// QueryTotalSupplyResponse is the response type for the Query/TotalSupply RPC
/// method
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryTotalSupplyResponse")]
pub struct QueryTotalSupplyResponse {
    /// supply is the supply of the coins
    #[prost(message, repeated, tag = "1")]
    pub supply: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
    /// pagination defines the pagination in the response.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageResponse>,
}
/// QuerySupplyOfRequest is the request type for the Query/SupplyOf RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOfRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/SupplyOf",
    response_type = QuerySupplyOfResponse
)]
pub struct QuerySupplyOfRequest {
    /// denom is the coin denom to query balances for.
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
/// QuerySupplyOfResponse is the response type for the Query/SupplyOf RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOfResponse")]
pub struct QuerySupplyOfResponse {
    /// amount is the supply of the coin.
    #[prost(message, optional, tag = "1")]
    pub amount: ::core::option::Option<super::super::base::v1beta1::Coin>,
}
/// QueryTotalSupplyWithoutOffsetRequest is the request type for the Query/TotalSupplyWithoutOffset RPC
/// method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryTotalSupplyWithoutOffsetRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/TotalSupplyWithoutOffset",
    response_type = QueryTotalSupplyWithoutOffsetResponse
)]
pub struct QueryTotalSupplyWithoutOffsetRequest {
    /// pagination defines an optional pagination for the request.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
}
/// QueryTotalSupplyWithoutOffsetResponse is the response type for the Query/TotalSupplyWithoutOffset RPC
/// method
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryTotalSupplyWithoutOffsetResponse")]
pub struct QueryTotalSupplyWithoutOffsetResponse {
    /// supply is the supply of the coins
    #[prost(message, repeated, tag = "1")]
    pub supply: ::prost::alloc::vec::Vec<super::super::base::v1beta1::Coin>,
    /// pagination defines the pagination in the response.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageResponse>,
}
/// QuerySupplyOfWithoutOffsetRequest is the request type for the Query/SupplyOfWithoutOffset RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOfWithoutOffsetRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/SupplyOfWithoutOffset",
    response_type = QuerySupplyOfWithoutOffsetResponse
)]
pub struct QuerySupplyOfWithoutOffsetRequest {
    /// denom is the coin denom to query balances for.
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
/// QuerySupplyOfWithoutOffsetResponse is the response type for the Query/SupplyOfWithoutOffset RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOfWithoutOffsetResponse")]
pub struct QuerySupplyOfWithoutOffsetResponse {
    /// amount is the supply of the coin.
    #[prost(message, optional, tag = "1")]
    pub amount: ::core::option::Option<super::super::base::v1beta1::Coin>,
}
/// QueryParamsRequest defines the request type for querying x/bank parameters.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryParamsRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/Params",
    response_type = QueryParamsResponse
)]
pub struct QueryParamsRequest {}
/// QueryParamsResponse defines the response type for querying x/bank parameters.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryParamsResponse")]
pub struct QueryParamsResponse {
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
}
/// QueryDenomsMetadataRequest is the request type for the Query/DenomsMetadata RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryDenomsMetadataRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/DenomsMetadata",
    response_type = QueryDenomsMetadataResponse
)]
pub struct QueryDenomsMetadataRequest {
    /// pagination defines an optional pagination for the request.
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
}
/// QueryDenomsMetadataResponse is the response type for the Query/DenomsMetadata RPC
/// method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryDenomsMetadataResponse")]
pub struct QueryDenomsMetadataResponse {
    /// metadata provides the client information for all the registered tokens.
    #[prost(message, repeated, tag = "1")]
    pub metadatas: ::prost::alloc::vec::Vec<Metadata>,
    /// pagination defines the pagination in the response.
    #[prost(message, optional, tag = "2")]
    pub pagination: ::core::option::Option<super::super::base::query::v1beta1::PageResponse>,
}
/// QueryDenomMetadataRequest is the request type for the Query/DenomMetadata RPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryDenomMetadataRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/DenomMetadata",
    response_type = QueryDenomMetadataResponse
)]
pub struct QueryDenomMetadataRequest {
    /// denom is the coin denom to query the metadata for.
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
/// QueryDenomMetadataResponse is the response type for the Query/DenomMetadata RPC
/// method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryDenomMetadataResponse")]
pub struct QueryDenomMetadataResponse {
    /// metadata describes and provides all the client information for the requested token.
    #[prost(message, optional, tag = "1")]
    pub metadata: ::core::option::Option<Metadata>,
}
/// QueryBaseDenomRequest defines the request type for the BaseDenom gRPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryBaseDenomRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/BaseDenom",
    response_type = QueryBaseDenomResponse
)]
pub struct QueryBaseDenomRequest {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
/// QueryBaseDenomResponse defines the response type for the BaseDenom gRPC method.
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
#[proto_message(type_url = "/cosmos.bank.v1beta1.QueryBaseDenomResponse")]
pub struct QueryBaseDenomResponse {
    #[prost(string, tag = "1")]
    pub base_denom: ::prost::alloc::string::String,
}
pub struct BankQuerier<'a, Q: cosmwasm_std::CustomQuery> {
    querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>,
}
impl<'a, Q: cosmwasm_std::CustomQuery> BankQuerier<'a, Q> {
    pub fn new(querier: &'a cosmwasm_std::QuerierWrapper<'a, Q>) -> Self {
        Self { querier }
    }
    pub fn balance(
        &self,
        address: ::prost::alloc::string::String,
        denom: ::prost::alloc::string::String,
    ) -> Result<QueryBalanceResponse, cosmwasm_std::StdError> {
        QueryBalanceRequest { address, denom }.query(self.querier)
    }
    pub fn all_balances(
        &self,
        address: ::prost::alloc::string::String,
        pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
    ) -> Result<QueryAllBalancesResponse, cosmwasm_std::StdError> {
        QueryAllBalancesRequest {
            address,
            pagination,
        }
        .query(self.querier)
    }
    pub fn total_supply(
        &self,
        pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
    ) -> Result<QueryTotalSupplyResponse, cosmwasm_std::StdError> {
        QueryTotalSupplyRequest { pagination }.query(self.querier)
    }
    pub fn supply_of(
        &self,
        denom: ::prost::alloc::string::String,
    ) -> Result<QuerySupplyOfResponse, cosmwasm_std::StdError> {
        QuerySupplyOfRequest { denom }.query(self.querier)
    }
    pub fn total_supply_without_offset(
        &self,
        pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
    ) -> Result<QueryTotalSupplyWithoutOffsetResponse, cosmwasm_std::StdError> {
        QueryTotalSupplyWithoutOffsetRequest { pagination }.query(self.querier)
    }
    pub fn supply_of_without_offset(
        &self,
        denom: ::prost::alloc::string::String,
    ) -> Result<QuerySupplyOfWithoutOffsetResponse, cosmwasm_std::StdError> {
        QuerySupplyOfWithoutOffsetRequest { denom }.query(self.querier)
    }
    pub fn params(&self) -> Result<QueryParamsResponse, cosmwasm_std::StdError> {
        QueryParamsRequest {}.query(self.querier)
    }
    pub fn denom_metadata(
        &self,
        denom: ::prost::alloc::string::String,
    ) -> Result<QueryDenomMetadataResponse, cosmwasm_std::StdError> {
        QueryDenomMetadataRequest { denom }.query(self.querier)
    }
    pub fn denoms_metadata(
        &self,
        pagination: ::core::option::Option<super::super::base::query::v1beta1::PageRequest>,
    ) -> Result<QueryDenomsMetadataResponse, cosmwasm_std::StdError> {
        QueryDenomsMetadataRequest { pagination }.query(self.querier)
    }
    pub fn base_denom(
        &self,
        denom: ::prost::alloc::string::String,
    ) -> Result<QueryBaseDenomResponse, cosmwasm_std::StdError> {
        QueryBaseDenomRequest { denom }.query(self.querier)
    }
}
