mod error;
mod registry;

pub use crate::registry::derive_wasmhooks_sender;
pub use crate::registry::Registry;

pub use error::RegistryError;

pub mod msg;
pub mod proto;
pub mod utils;
