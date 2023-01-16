use osmosis_std_derive::CosmwasmExt;
/// SetSuperfluidAssetsProposal is a gov Content type to update the superfluid
/// assets
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
#[proto_message(type_url = "/osmosis.superfluid.v1beta1.SetSuperfluidAssetsProposal")]
pub struct SetSuperfluidAssetsProposal {
    #[prost(string, tag = "1")]
    pub title: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub description: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "3")]
    pub assets: ::prost::alloc::vec::Vec<super::SuperfluidAsset>,
}
/// RemoveSuperfluidAssetsProposal is a gov Content type to remove the superfluid
/// assets by denom
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
#[proto_message(type_url = "/osmosis.superfluid.v1beta1.RemoveSuperfluidAssetsProposal")]
pub struct RemoveSuperfluidAssetsProposal {
    #[prost(string, tag = "1")]
    pub title: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub description: ::prost::alloc::string::String,
    #[prost(string, repeated, tag = "3")]
    pub superfluid_asset_denoms: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// UpdateUnpoolWhiteListProposal is a gov Content type to update the
/// allowed list of pool ids.
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
#[proto_message(type_url = "/osmosis.superfluid.v1beta1.UpdateUnpoolWhiteListProposal")]
pub struct UpdateUnpoolWhiteListProposal {
    #[prost(string, tag = "1")]
    pub title: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub description: ::prost::alloc::string::String,
    #[prost(uint64, repeated, tag = "3")]
    pub ids: ::prost::alloc::vec::Vec<u64>,
    #[prost(bool, tag = "4")]
    pub is_overwrite: bool,
}
