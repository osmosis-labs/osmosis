use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::ContractError;

use super::{flow::{Flow, FlowType}, quota::Quota, path::Path};



/// RateLimit is the main structure tracked for each channel/denom pair. Its quota
/// represents rate limit configuration, and the flow its
/// current state (i.e.: how much value has been transfered in the current period)
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct RateLimit {
    pub quota: Quota,
    pub flow: Flow,
}

impl RateLimit {
    /// Checks if a transfer is allowed and updates the data structures
    /// accordingly.
    ///
    /// If the transfer is not allowed, it will return a RateLimitExceeded error.
    ///
    /// Otherwise it will return a RateLimitResponse with the updated data structures
    pub fn allow_transfer(
        &mut self,
        path: &Path,
        direction: &FlowType,
        funds: Uint256,
        channel_value: Uint256,
        now: Timestamp,
    ) -> Result<Self, ContractError> {
        // Flow used before this transaction is applied.
        // This is used to make error messages more informative
        let initial_flow = self.flow.balance_on(direction);

        // Apply the transfer. From here on, we will updated the flow with the new transfer
        // and check if  it exceeds the quota at the current time

        let expired = self.flow.apply_transfer(direction, funds, now, &self.quota);
        // Cache the channel value if it has never been set or it has expired.
        if self.quota.channel_value.is_none() || expired {
            self.quota.channel_value = Some(calculate_channel_value(
                channel_value,
                &path.denom,
                funds,
                direction,
            ))
        }

        let (max_in, max_out) = self.quota.capacity();
        // Return the effects of applying the transfer or an error.
        match self.flow.exceeds(direction, max_in, max_out) {
            true => Err(ContractError::RateLimitExceded {
                channel: path.channel.to_string(),
                denom: path.denom.to_string(),
                amount: funds,
                quota_name: self.quota.name.to_string(),
                used: initial_flow,
                max: self.quota.capacity_on(direction),
                reset: self.flow.period_end,
            }),
            false => Ok(RateLimit {
                quota: self.quota.clone(), // Cloning here because self.quota.name (String) does not allow us to implement Copy
                flow: self.flow, // We can Copy flow, so this is slightly more efficient than cloning the whole RateLimit
            }),
        }
    }
}



// The channel value on send depends on the amount on escrow. The ibc transfer
// module modifies the escrow amount by "funds" on sends before calling the
// contract. This function takes that into account so that the channel value
// that we track matches the channel value at the moment when the ibc
// transaction started executing
fn calculate_channel_value(
    channel_value: Uint256,
    denom: &str,
    funds: Uint256,
    direction: &FlowType,
) -> Uint256 {
    match direction {
        FlowType::Out => {
            if denom.contains("ibc") {
                channel_value + funds // Non-Native tokens get removed from the supply on send. Add that amount back
            } else {
                // The commented-out code in the golang calculate channel value is what we want, but we're currently using the whole supply temporarily for efficiency. see rate_limit.go/CalculateChannelValue(..)
                //channel_value - funds // Native tokens increase escrow amount on send. Remove that amount here
                channel_value
            }
        }
        FlowType::In => channel_value,
    }
}