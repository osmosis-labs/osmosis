use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// This represents the key for our rate limiting tracker. A tuple of a denom and
/// a channel. When interacting with storage, it's preffered to use this struct
/// and call path.into() on it to convert it to the composite key of the
/// RATE_LIMIT_TRACKERS map
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Path {
    pub denom: String,
    pub channel: String,
}

impl Path {
    pub fn new(channel: impl Into<String>, denom: impl Into<String>) -> Self {
        Path {
            channel: channel.into(),
            denom: denom.into(),
        }
    }
}

impl From<Path> for (String, String) {
    fn from(path: Path) -> (String, String) {
        (path.channel, path.denom)
    }
}

impl From<&Path> for (String, String) {
    fn from(path: &Path) -> (String, String) {
        (path.channel.to_owned(), path.denom.to_owned())
    }
}
