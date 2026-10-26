use cosmwasm_std::Uint256;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::msg::QuotaMsg;
use crate::ContractError;

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
    ///
    /// The channel value is a bank supply, which is a 256-bit integer on the
    /// chain, so the multiplication is checked rather than left to panic.
    pub fn capacity(&self) -> Result<(Uint256, Uint256), ContractError> {
        let Some(total_value) = self.channel_value else {
            // This should never happen, but if the channel value is not set, we disallow any transfer
            return Ok((Uint256::zero(), Uint256::zero()));
        };
        let capacity_for = |percentage: u32| {
            total_value
                .checked_multiply_ratio(percentage, 100_u32)
                .map_err(|e| ContractError::Overflow(e.to_string()))
        };
        Ok((
            capacity_for(self.max_percentage_recv)?,
            capacity_for(self.max_percentage_send)?,
        ))
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
