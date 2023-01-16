pub mod v1beta1;
use osmosis_std_derive::CosmwasmExt;
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
#[proto_message(type_url = "/osmosis.concentratedliquidity.Params")]
pub struct Params {
    /// authorized_tick_spacing is an array of uint64s that represents the tick
    /// spacing values concentrated-liquidity pools can be created with. For
    /// example, an authorized_tick_spacing of [1, 10, 30] allows for pools
    /// to be created with tick spacing of 1, 10, or 30.
    #[prost(uint64, repeated, packed = "false", tag = "1")]
    pub authorized_tick_spacing: ::prost::alloc::vec::Vec<u64>,
}
