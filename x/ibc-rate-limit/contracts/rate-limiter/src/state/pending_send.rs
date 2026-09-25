use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// A send that passed the rate limit and has not been acknowledged yet.
///
/// It records, for every quota the send was counted against, the window it
/// was counted in. A send that later fails (error acknowledgement or timeout)
/// is refunded to a quota only while that quota is still in the same window.
/// Refunding into a later window would enlarge that window's allowance by the
/// refunded amount, which is how the previous unconditional refund could be
/// used to double the outflow cap.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct PendingSend {
    pub funds: Uint256,
    /// (quota name, period_end) for each quota on the (channel, denom) path
    pub channel_windows: Vec<(String, Timestamp)>,
    /// (quota name, period_end) for each quota on the ("any", denom) path
    pub any_windows: Vec<(String, Timestamp)>,
}

impl PendingSend {
    /// The latest window end among the quotas the send was counted against.
    /// Once it has passed, a refund can no longer change any active window.
    pub fn latest_period_end(&self) -> Timestamp {
        self.channel_windows
            .iter()
            .chain(self.any_windows.iter())
            .map(|(_, period_end)| *period_end)
            .max()
            .unwrap_or_else(|| Timestamp::from_nanos(0))
    }
}
