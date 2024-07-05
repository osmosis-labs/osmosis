use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use super::quota::Quota;


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
    pub inflow: Uint256,
    pub outflow: Uint256,
    pub period_end: Timestamp,
}

impl Flow {
    pub fn new(
        inflow: impl Into<Uint256>,
        outflow: impl Into<Uint256>,
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
    /// through the channel before period_end. It returns a tuple of
    /// (balance_in, balance_out) where balance_in in is how much has been
    /// transferred into the flow, and balance_out is how much value transferred
    /// out.
    pub fn balance(&self) -> (Uint256, Uint256) {
        (
            self.inflow.saturating_sub(self.outflow),
            self.outflow.saturating_sub(self.inflow),
        )
    }

    /// checks if the flow, in the current state, has exceeded a max allowance
    pub fn exceeds(&self, direction: &FlowType, max_inflow: Uint256, max_outflow: Uint256) -> bool {
        let (balance_in, balance_out) = self.balance();
        match direction {
            FlowType::In => balance_in > max_inflow,
            FlowType::Out => balance_out > max_outflow,
        }
    }

    /// returns the balance in a direction. This is used for displaying cleaner errors
    pub fn balance_on(&self, direction: &FlowType) -> Uint256 {
        let (balance_in, balance_out) = self.balance();
        match direction {
            FlowType::In => balance_in,
            FlowType::Out => balance_out,
        }
    }

    /// If now is greater than the period_end, the Flow is considered expired.
    pub fn is_expired(&self, now: Timestamp) -> bool {
        self.period_end < now
    }

    // Mutating methods

    /// Expire resets the Flow to start tracking the value transfer from the
    /// moment this method is called.
    pub fn expire(&mut self, now: Timestamp, duration: u64) {
        self.inflow = Uint256::from(0_u32);
        self.outflow = Uint256::from(0_u32);
        self.period_end = now.plus_seconds(duration);
    }

    /// Updates the current flow incrementing it by a transfer of value.
    pub fn add_flow(&mut self, direction: FlowType, value: Uint256) {
        match direction {
            FlowType::In => self.inflow = self.inflow.saturating_add(value),
            FlowType::Out => self.outflow = self.outflow.saturating_add(value),
        }
    }

    /// Updates the current flow reducing it by a transfer of value.
    pub fn undo_flow(&mut self, direction: FlowType, value: Uint256) {
        match direction {
            FlowType::In => self.inflow = self.inflow.saturating_sub(value),
            FlowType::Out => self.outflow = self.outflow.saturating_sub(value),
        }
    }

    /// Applies a transfer. If the Flow is expired (now > period_end), it will
    /// reset it before applying the transfer.
    pub(crate) fn apply_transfer(
        &mut self,
        direction: &FlowType,
        funds: Uint256,
        now: Timestamp,
        quota: &Quota,
    ) -> bool {
        let mut expired = false;
        if self.is_expired(now) {
            self.expire(now, quota.duration);
            expired = true;
        }
        self.add_flow(direction.clone(), funds);
        expired
    }
}



#[cfg(test)]
pub mod tests {
    use super::*;

    pub const RESET_TIME_DAILY: u64 = 60 * 60 * 24;
    pub const RESET_TIME_WEEKLY: u64 = 60 * 60 * 24 * 7;
    pub const RESET_TIME_MONTHLY: u64 = 60 * 60 * 24 * 30;

    #[test]
    fn flow() {
        let epoch = Timestamp::from_seconds(0);
        let mut flow = Flow::new(0_u32, 0_u32, epoch, RESET_TIME_WEEKLY);

        assert!(!flow.is_expired(epoch));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_DAILY)));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY)));
        assert!(flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY).plus_nanos(1)));

        assert_eq!(flow.balance(), (0_u32.into(), 0_u32.into()));
        flow.add_flow(FlowType::In, 5_u32.into());
        assert_eq!(flow.balance(), (5_u32.into(), 0_u32.into()));
        flow.add_flow(FlowType::Out, 2_u32.into());
        assert_eq!(flow.balance(), (3_u32.into(), 0_u32.into()));
        // Adding flow doesn't affect expiration
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_DAILY)));

        flow.expire(epoch.plus_seconds(RESET_TIME_WEEKLY), RESET_TIME_WEEKLY);
        assert_eq!(flow.balance(), (0_u32.into(), 0_u32.into()));
        assert_eq!(flow.inflow, Uint256::from(0_u32));
        assert_eq!(flow.outflow, Uint256::from(0_u32));
        assert_eq!(flow.period_end, epoch.plus_seconds(RESET_TIME_WEEKLY * 2));

        // Expiration has moved
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY).plus_nanos(1)));
        assert!(!flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY * 2)));
        assert!(flow.is_expired(epoch.plus_seconds(RESET_TIME_WEEKLY * 2).plus_nanos(1)));
    }
}
