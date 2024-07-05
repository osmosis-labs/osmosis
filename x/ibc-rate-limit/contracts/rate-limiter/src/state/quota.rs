use cosmwasm_std::Uint256;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::msg::QuotaMsg;

use super::flow::FlowType;


/// A Quota is the percentage of the denom's total value that can be transferred
/// through the channel in a given period of time (duration)
///
/// Percentages can be different for send and recv
///
/// The name of the quota is expected to be a human-readable representation of
/// the duration (i.e.: "weekly", "daily", "every-six-months", ...)
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Quota {
    pub name: String,
    pub max_percentage_send: u32,
    pub max_percentage_recv: u32,
    pub duration: u64,
    pub channel_value: Option<Uint256>,
}

impl Quota {
    /// Calculates the max capacity (absolute value in the same unit as
    /// total_value) in each direction based on the total value of the denom in
    /// the channel. The result tuple represents the max capacity when the
    /// transfer is in directions: (FlowType::In, FlowType::Out)
    pub fn capacity(&self) -> (Uint256, Uint256) {
        match self.channel_value {
            Some(total_value) => (
                total_value * Uint256::from(self.max_percentage_recv) / Uint256::from(100_u32),
                total_value * Uint256::from(self.max_percentage_send) / Uint256::from(100_u32),
            ),
            None => (0_u32.into(), 0_u32.into()), // This should never happen, but ig the channel value is not set, we disallow any transfer
        }
    }

    /// returns the capacity in a direction. This is used for displaying cleaner errors
    pub fn capacity_on(&self, direction: &FlowType) -> Uint256 {
        let (max_in, max_out) = self.capacity();
        match direction {
            FlowType::In => max_in,
            FlowType::Out => max_out,
        }
    }
}

impl From<&QuotaMsg> for Quota {
    fn from(msg: &QuotaMsg) -> Self {
        let send_recv = (
            std::cmp::min(msg.send_recv.0, 100),
            std::cmp::min(msg.send_recv.1, 100),
        );
        Quota {
            name: msg.name.clone(),
            max_percentage_send: send_recv.0,
            max_percentage_recv: send_recv.1,
            duration: msg.duration,
            channel_value: None,
        }
    }
}
