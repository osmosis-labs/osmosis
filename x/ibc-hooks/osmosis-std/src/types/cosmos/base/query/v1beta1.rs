use osmosis_std_derive::CosmwasmExt;
/// PageRequest is to be embedded in gRPC request messages for efficient
/// pagination. Ex:
///
///  message SomeRequest {
///          Foo some_parameter = 1;
///          PageRequest pagination = 2;
///  }
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
#[proto_message(type_url = "/cosmos.base.query.v1beta1.PageRequest")]
pub struct PageRequest {
    /// key is a value returned in PageResponse.next_key to begin
    /// querying the next page most efficiently. Only one of offset or key
    /// should be set.
    #[prost(bytes = "vec", tag = "1")]
    pub key: ::prost::alloc::vec::Vec<u8>,
    /// offset is a numeric offset that can be used when key is unavailable.
    /// It is less efficient than using key. Only one of offset or key should
    /// be set.
    #[prost(uint64, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub offset: u64,
    /// limit is the total number of results to be returned in the result page.
    /// If left empty it will default to a value to be set by each app.
    #[prost(uint64, tag = "3")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub limit: u64,
    /// count_total is set to true  to indicate that the result set should include
    /// a count of the total number of items available for pagination in UIs.
    /// count_total is only respected when offset is used. It is ignored when key
    /// is set.
    #[prost(bool, tag = "4")]
    pub count_total: bool,
    /// reverse is set to true if results are to be returned in the descending order.
    ///
    /// Since: cosmos-sdk 0.43
    #[prost(bool, tag = "5")]
    pub reverse: bool,
}
/// PageResponse is to be embedded in gRPC response messages where the
/// corresponding request message has used PageRequest.
///
///  message SomeResponse {
///          repeated Bar results = 1;
///          PageResponse page = 2;
///  }
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
#[proto_message(type_url = "/cosmos.base.query.v1beta1.PageResponse")]
pub struct PageResponse {
    /// next_key is the key to be passed to PageRequest.key to
    /// query the next page most efficiently
    #[prost(bytes = "vec", tag = "1")]
    pub next_key: ::prost::alloc::vec::Vec<u8>,
    /// total is total number of results available if PageRequest.count_total
    /// was set, its value is undefined otherwise
    #[prost(uint64, tag = "2")]
    #[serde(
        serialize_with = "crate::serde::as_str::serialize",
        deserialize_with = "crate::serde::as_str::deserialize"
    )]
    pub total: u64,
}
