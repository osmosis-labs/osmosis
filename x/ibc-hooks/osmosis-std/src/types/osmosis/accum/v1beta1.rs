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
#[proto_message(type_url = "/osmosis.accum.v1beta1.AccumulatorContent")]
pub struct AccumulatorContent {
    #[prost(message, repeated, tag = "1")]
    pub accum_value: ::prost::alloc::vec::Vec<super::super::super::cosmos::base::v1beta1::DecCoin>,
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
#[proto_message(type_url = "/osmosis.accum.v1beta1.Options")]
pub struct Options {}
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
#[proto_message(type_url = "/osmosis.accum.v1beta1.Record")]
pub struct Record {
    #[prost(string, tag = "1")]
    pub num_shares: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub init_accum_value:
        ::prost::alloc::vec::Vec<super::super::super::cosmos::base::v1beta1::DecCoin>,
    #[prost(message, repeated, tag = "3")]
    pub unclaimed_rewards:
        ::prost::alloc::vec::Vec<super::super::super::cosmos::base::v1beta1::DecCoin>,
    #[prost(message, optional, tag = "4")]
    pub options: ::core::option::Option<Options>,
}
