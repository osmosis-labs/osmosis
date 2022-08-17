use cosmwasm_std::{Addr, Timestamp};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use std::cmp;

use cw_storage_plus::{Item, Map};

use crate::msg::QuotaMsg;

pub const RESET_TIME_DAILY: u64 = 60 * 60 * 24;
pub const RESET_TIME_WEEKLY: u64 = 60 * 60 * 24 * 7;
pub const RESET_TIME_MONTHLY: u64 = 60 * 60 * 24 * 30;

/// This represents the key for our rate limiting tracker. A tuple of a denom and
/// a channel. When interactic with storage, it's preffered to use this struct
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

#[derive(Debug, Clone)]
pub enum FlowType {
    In,
    Out,
}

/// A Flow represents the transfer of value for a denom through an IBC channel
/// during a time window.
///
/// It tracks inflows (transfers into osmosis) and outflows (transfers out of
/// osmosis).
///
/// The period_end represents the last point in time for which this Flow is
/// tracking the value transfer.
///
/// Periods are discrete repeating windows. A period only starts when a contract
/// call to update the Flow (SendPacket/RecvPackt) is made, and not right after
/// the period ends. This means that if no calls happen after a period expires,
/// the next period will begin at the time of the next call and be valid for the
/// specified duration for the quota.
///
/// This is a design decision to avoid the period calculations and thus reduce gas consumption
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema, Copy)]
pub struct Flow {
    // Q: Do we have edge case issues with inflow/outflow being u128, e.g. what if a token has super high precision.
    pub inflow: u128,
    pub outflow: u128,
    pub period_end: Timestamp,
}

impl Flow {
    pub fn new(
        inflow: impl Into<u128>,
        outflow: impl Into<u128>,
        now: Timestamp,
        duration: u64,
    ) -> Self {
        Self {
            inflow: inflow.into(),
            outflow: outflow.into(),
            period_end: now.plus_seconds(duration),
        }
    }

    /// The balance of a flow is how much absolute value for the denom has moved
    /// through the channel before period_end.
    pub fn balance(&self) -> u128 {
        self.inflow.abs_diff(self.outflow)
    }

    /// If now is greater than the period_end, the Flow is considered expired.
    pub fn is_expired(&self, now: Timestamp) -> bool {
        self.period_end < now
    }

    // Mutating methods

    /// Expire resets the Flow to start tracking the value transfer from the
    /// moment this method is called.
    pub fn expire(&mut self, now: Timestamp, duration: u64) {
        self.inflow = 0;
        self.outflow = 0;
        self.period_end = now.plus_seconds(duration);
    }

    /// Updates the current flow with a transfer of value.
    pub fn add_flow(&mut self, direction: FlowType, value: u128) {
        match direction {
            FlowType::In => self.inflow = self.inflow.saturating_add(value),
            FlowType::Out => self.outflow = self.outflow.saturating_add(value),
        }
    }
}

/// A Quota is the percentage of the denom's total value that can be transfered
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
}

impl Quota {
    /// Calculates the max capacity (absolute value in the same unit as
    /// total_value) in a given direction based on the total value of the denom
    /// in the channel.
    pub fn capacity_at(&self, total_value: &u128, direction: &FlowType) -> u128 {
        let max_percentage = match direction {
            FlowType::In => self.max_percentage_recv,
            FlowType::Out => self.max_percentage_send,
        };
        total_value * (max_percentage as u128) / 100_u128
    }
}

impl From<&QuotaMsg> for Quota {
    fn from(msg: &QuotaMsg) -> Self {
        let send_recv = (
            cmp::min(msg.send_recv.0, 100),
            cmp::min(msg.send_recv.1, 100),
        );
        Quota {
            name: msg.name.clone(),
            max_percentage_send: send_recv.0,
            max_percentage_recv: send_recv.1,
            duration: msg.duration,
        }
    }
}

/// RateLimit is the main structure tracked for each channel/denom pair. Its quota
/// represents rate limit configuration, and the flow its
/// current state (i.e.: how much value has been transfered in the current period)
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct RateLimit {
    pub quota: Quota,
    pub flow: Flow,
}

/// Only this address can manage the contract. This will likely be the
/// governance module, but could be set to something else if needed
pub const GOVMODULE: Item<Addr> = Item::new("gov_module");
/// Only this address can execute transfers. This will likely be the
/// IBC transfer module, but could be set to something else if needed
pub const IBCMODULE: Item<Addr> = Item::new("ibc_module");

/// RATE_LIMIT_TRACKERS is the main state for this contract. It maps a path (IBC
/// Channel + denom) to a vector of `RateLimit`s.
///
/// The `RateLimit` struct contains the information about how much value of a
/// denom has moved through the channel during the currently active time period
/// (channel_flow.flow) and what percentage of the denom's value we are
/// allowing to flow through that channel in a specific duration (quota)
///
/// For simplicity, the channel in the map keys refers to the "host" channel on
/// the osmosis side. This means that on PacketSend it will refer to the source
/// channel while on PacketRecv it refers to the destination channel.
///
/// It is the responsibility of the go module to pass the appropriate channel
/// when sending the messages
///
/// The map key (String, String) represents (channel_id, denom). We use
/// composite keys instead of a struct to avoid having to implement the
/// PrimaryKey trait
pub const RATE_LIMIT_TRACKERS: Map<(String, String), Vec<RateLimit>> = Map::new("flow");

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn flow() {
        let epoch = Timestamp::from_seconds(0);
        let mut flow = Flow::new(0_u32, 0_u32, epoch, RESET_TIME_WEEKLY);

        assert!(!flow.is_expired(epoch));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_DAILY)));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY)));
        assert!(flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY).plus_nanos(1)));

        assert_eq!(flow.balance(), 0_u128);
        flow.add_flow(FlowType::In, 5);
        assert_eq!(flow.balance(), 5_u128);
        flow.add_flow(FlowType::Out, 2);
        assert_eq!(flow.balance(), 3_u128);
        // Adding flow doesn't affect expiration
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_DAILY)));

        flow.expire(epoch.plus_seconds(RESET_TIME_WEEKLY), RESET_TIME_WEEKLY);
        assert_eq!(flow.balance(), 0_u128);
        assert_eq!(flow.inflow, 0_u128);
        assert_eq!(flow.outflow, 0_u128);
        assert_eq!(flow.period_end, epoch.plus_seconds(RESET_TIME_WEEKLY * 2));

        // Expiration has moved
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY).plus_nanos(1)));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY * 2)));
        assert!(flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY * 2).plus_nanos(1)));
    }
}
