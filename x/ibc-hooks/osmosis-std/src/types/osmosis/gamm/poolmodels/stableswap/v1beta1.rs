use osmosis_std_derive::CosmwasmExt;
/// PoolParams defined the parameters that will be managed by the pool
/// governance in the future. This params are not managed by the chain
/// governance. Instead they will be managed by the token holders of the pool.
/// The pool's token holders are specified in future_pool_governor.
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
#[proto_message(type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.PoolParams")]
pub struct PoolParams {
    #[prost(string, tag = "1")]
    pub swap_fee: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub exit_fee: ::prost::alloc::string::String,
}
/// Pool is the stableswap Pool struct
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
#[proto_message(type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.Pool")]
pub struct Pool {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub id: u64,
    #[prost(message, optional, tag = "3")]
    pub pool_params: ::core::option::Option<PoolParams>,
    /// This string specifies who will govern the pool in the future.
    /// Valid forms of this are:
    /// {token name},{duration}
    /// {duration}
    /// where {token name} if specified is the token which determines the
    /// governor, and if not specified is the LP token for this pool.duration is
    /// a time specified as 0w,1w,2w, etc. which specifies how long the token
    /// would need to be locked up to count in governance. 0w means no lockup.
    #[prost(string, tag = "4")]
    pub future_pool_governor: ::prost::alloc::string::String,
    /// sum of all LP shares
    #[prost(message, optional, tag = "5")]
    pub total_shares:
        ::core::option::Option<super::super::super::super::super::cosmos::base::v1beta1::Coin>,
    /// assets in the pool
    #[prost(message, repeated, tag = "6")]
    pub pool_liquidity:
        ::prost::alloc::vec::Vec<super::super::super::super::super::cosmos::base::v1beta1::Coin>,
    /// for calculation amognst assets with different precisions
    #[prost(uint64, repeated, packed = "false", tag = "7")]
    pub scaling_factors: ::prost::alloc::vec::Vec<u64>,
    /// scaling_factor_controller is the address can adjust pool scaling factors
    #[prost(string, tag = "8")]
    pub scaling_factor_controller: ::prost::alloc::string::String,
}
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
#[proto_message(type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.MsgCreateStableswapPool")]
pub struct MsgCreateStableswapPool {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub pool_params: ::core::option::Option<PoolParams>,
    #[prost(message, repeated, tag = "3")]
    pub initial_pool_liquidity:
        ::prost::alloc::vec::Vec<super::super::super::super::super::cosmos::base::v1beta1::Coin>,
    #[prost(uint64, repeated, packed = "false", tag = "4")]
    pub scaling_factors: ::prost::alloc::vec::Vec<u64>,
    #[prost(string, tag = "5")]
    pub future_pool_governor: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub scaling_factor_controller: ::prost::alloc::string::String,
}
/// Returns a poolID with custom poolName.
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
    type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.MsgCreateStableswapPoolResponse"
)]
pub struct MsgCreateStableswapPoolResponse {
    #[prost(uint64, tag = "1")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
}
/// Sender must be the pool's scaling_factor_governor in order for the tx to
/// succeed. Adjusts stableswap scaling factors.
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
    type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.MsgStableSwapAdjustScalingFactors"
)]
pub struct MsgStableSwapAdjustScalingFactors {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub pool_id: u64,
    #[prost(uint64, repeated, packed = "false", tag = "3")]
    pub scaling_factors: ::prost::alloc::vec::Vec<u64>,
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
    type_url = "/osmosis.gamm.poolmodels.stableswap.v1beta1.MsgStableSwapAdjustScalingFactorsResponse"
)]
pub struct MsgStableSwapAdjustScalingFactorsResponse {}
