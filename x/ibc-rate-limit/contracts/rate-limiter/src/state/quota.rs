use cosmwasm_std::Uint256;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::msg::QuotaMsg;
use crate::ContractError;

use super::flow::FlowType;

/// A Quota bounds how much of a denom may move through a path in a given
/// period of time (duration). Each direction has up to two bounds:
///
/// * a percentage of the denom's channel value (its supply on this chain,
///   cached at the start of the window), applied to the net flow so that
///   round trips do not consume quota, and
/// * an absolute amount in the denom's base units, applied to the gross flow
///   so that sending real tokens out first cannot buy room to bring more in.
///
/// A transfer must satisfy every bound that is set. The absolute bound exists
/// because a percentage scales with the very supply an attacker is inflating
/// and rounds to nothing on a low-supply asset. Percentages above 100 are
/// allowed; the absolute bound is what keeps a large percentage safe.
///
/// The name of the quota is expected to be a human-readable representation of
/// the duration (i.e.: "weekly", "daily", "every-six-months", ...)
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Quota {
    pub name: String,
    /// `None` means no percentage bound in that direction. Entries written
    /// before 0.2.0 hold plain integers here and deserialize as `Some`.
    pub max_percentage_send: Option<u32>,
    pub max_percentage_recv: Option<u32>,
    pub duration: u64,
    pub channel_value: Option<Uint256>,
    /// Absolute bounds in base units. Absent from entries written before 0.2.0.
    #[serde(default)]
    pub max_absolute_send: Option<Uint256>,
    #[serde(default)]
    pub max_absolute_recv: Option<Uint256>,
}

/// The bounds that apply to one direction of a quota, resolved against the
/// cached channel value.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Capacity {
    /// Compared against the net flow in the direction
    pub percentage: Option<Uint256>,
    /// Compared against the gross flow in the direction
    pub absolute: Option<Uint256>,
}

impl Quota {
    /// Resolves the bounds in each direction against the cached channel
    /// value. The result tuple is (FlowType::In, FlowType::Out).
    ///
    /// The channel value is a bank supply, which is a 256-bit integer on the
    /// chain, so the multiplication is checked rather than left to panic.
    pub fn capacity(&self) -> Result<(Capacity, Capacity), ContractError> {
        // If the channel value is not set every percentage bound is zero and
        // disallows the transfer. This should never happen: allow_transfer
        // caches the channel value before asking for the capacity.
        let total_value = self.channel_value.unwrap_or_default();
        let percentage_of = |percentage: Option<u32>| -> Result<Option<Uint256>, ContractError> {
            percentage
                .map(|percentage| {
                    total_value
                        .checked_multiply_ratio(percentage, 100_u32)
                        .map_err(|e| ContractError::Overflow(e.to_string()))
                })
                .transpose()
        };
        Ok((
            Capacity {
                percentage: percentage_of(self.max_percentage_recv)?,
                absolute: self.max_absolute_recv,
            },
            Capacity {
                percentage: percentage_of(self.max_percentage_send)?,
                absolute: self.max_absolute_send,
            },
        ))
    }

    /// The bounds that apply to a transfer in one direction
    pub fn capacity_on(&self, direction: &FlowType) -> Result<Capacity, ContractError> {
        let (max_in, max_out) = self.capacity()?;
        Ok(match direction {
            FlowType::In => max_in,
            FlowType::Out => max_out,
        })
    }
}

impl TryFrom<&QuotaMsg> for Quota {
    type Error = ContractError;

    /// Validates a quota as submitted by governance. A quota that would bound
    /// nothing in a direction is rejected rather than stored as a silent
    /// no-op, and so are the shapes that would make it unaddressable later.
    fn try_from(msg: &QuotaMsg) -> Result<Self, Self::Error> {
        let invalid = ContractError::InvalidParameters;
        if msg.name.is_empty() {
            return Err(invalid("quota name must not be empty".to_string()));
        }
        if msg.duration == 0 {
            let reason = format!("quota {}: duration must be greater than zero", msg.name);
            return Err(invalid(reason));
        }
        let (send_percentage, recv_percentage) = msg.send_recv;
        if send_percentage.is_none() && msg.max_absolute_send.is_none() {
            let reason = format!(
                "quota {}: send needs a percentage or an absolute bound",
                msg.name
            );
            return Err(invalid(reason));
        }
        if recv_percentage.is_none() && msg.max_absolute_recv.is_none() {
            let reason = format!(
                "quota {}: recv needs a percentage or an absolute bound",
                msg.name
            );
            return Err(invalid(reason));
        }
        Ok(Quota {
            name: msg.name.clone(),
            max_percentage_send: send_percentage,
            max_percentage_recv: recv_percentage,
            duration: msg.duration,
            channel_value: None,
            max_absolute_send: msg.max_absolute_send,
            max_absolute_recv: msg.max_absolute_recv,
        })
    }
}
