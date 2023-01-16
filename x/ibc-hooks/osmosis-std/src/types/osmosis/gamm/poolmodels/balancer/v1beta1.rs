use osmosis_std_derive::CosmwasmExt;
/// ===================== MsgCreatePool
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
#[proto_message(type_url = "/osmosis.gamm.poolmodels.balancer.v1beta1.MsgCreateBalancerPool")]
pub struct MsgCreateBalancerPool {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub pool_params: ::core::option::Option<super::super::super::v1beta1::PoolParams>,
    #[prost(message, repeated, tag = "3")]
    pub pool_assets: ::prost::alloc::vec::Vec<super::super::super::v1beta1::PoolAsset>,
    #[prost(string, tag = "4")]
    pub future_pool_governor: ::prost::alloc::string::String,
}
/// Returns the poolID
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
    type_url = "/osmosis.gamm.poolmodels.balancer.v1beta1.MsgCreateBalancerPoolResponse"
)]
pub struct MsgCreateBalancerPoolResponse {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
}
