use std::str::Utf8Error;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum RunnerError {
    #[error("unable to encode request")]
    EncodeError(#[from] EncodeError),

    #[error("unable to decode response")]
    DecodeError(#[from] DecodeError),

    #[error("query error: {}", .msg)]
    QueryError { msg: String },

    #[error("execute error: {}", .msg)]
    ExecuteError { msg: String },
}

#[derive(Error, Debug)]
pub enum DecodeError {
    #[error("invalid utf8 bytes")]
    Utf8Error(#[from] Utf8Error),

    #[error("invalid protobuf")]
    ProtoDecodeError(#[from] prost::DecodeError),

    #[error("invalid json")]
    JsonDecodeError(#[from] serde_json::Error),

    #[error("invalid base64")]
    Base64DecodeError(#[from] base64::DecodeError),

    #[error("invalid signing key")]
    SigningKeyDecodeError { msg: String },
}

impl PartialEq for DecodeError {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (DecodeError::Utf8Error(a), DecodeError::Utf8Error(b)) => a == b,
            (DecodeError::ProtoDecodeError(a), DecodeError::ProtoDecodeError(b)) => a == b,
            (DecodeError::JsonDecodeError(a), DecodeError::JsonDecodeError(b)) => {
                a.to_string() == b.to_string()
            }
            (DecodeError::Base64DecodeError(a), DecodeError::Base64DecodeError(b)) => a == b,
            (
                DecodeError::SigningKeyDecodeError { msg: a },
                DecodeError::SigningKeyDecodeError { msg: b },
            ) => a == b,
            _ => false,
        }
    }
}

#[derive(Error, Debug)]
pub enum EncodeError {
    #[error("invalid protobuf")]
    ProtoEncodeError(#[from] prost::EncodeError),

    #[error("unable to encode json")]
    JsonEncodeError(#[from] serde_json::Error),
}

impl PartialEq for EncodeError {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (EncodeError::ProtoEncodeError(a), EncodeError::ProtoEncodeError(b)) => a == b,
            (EncodeError::JsonEncodeError(a), EncodeError::JsonEncodeError(b)) => {
                a.to_string() == b.to_string()
            }
            _ => false,
        }
    }
}
