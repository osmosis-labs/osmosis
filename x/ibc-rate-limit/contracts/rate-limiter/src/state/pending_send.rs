use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::packet::Packet;

/// Clock skew allowed between this chain and a counterparty when deciding
/// that a packet's timeout timestamp has passed.
pub const TIMEOUT_SKEW_SECONDS: u64 = 60 * 60;

/// How long a record is kept past its last window end when the packet has no
/// timeout timestamp (height-only timeout), since the contract cannot tell
/// when a counterparty height will be reached.
pub const HEIGHT_TIMEOUT_RETENTION_SECONDS: u64 = 30 * 24 * 60 * 60;

/// A send that passed the rate limit and has not been settled yet.
///
/// It records, for every quota the send was counted against, the window it
/// was counted in. A send that later fails (error acknowledgement or timeout)
/// is refunded to a quota only while that quota is still in the same window,
/// so a refund can only ever restore what the same window charged.
///
/// Records are keyed by packet content (Packet::send_key), which is not a
/// unique identity: the same content can be sent again as long as the
/// packet's timeout has not passed. A record is therefore retained until both
/// its windows have ended and the packet could no longer be re-sent, and only
/// one record exists per content key at a time. That way an acknowledgement
/// always settles the oldest send with that content, and a later identical
/// send can neither claim nor shadow an earlier one.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct PendingSend {
    pub funds: Uint256,
    /// (quota name, period_end) for each quota on the (channel, denom) path
    pub channel_windows: Vec<(String, Timestamp)>,
    /// (quota name, period_end) for each quota on the ("any", denom) path
    pub any_windows: Vec<(String, Timestamp)>,
    /// When the record may be removed without settlement, see `retention_end`
    pub retain_until: Timestamp,
}

impl PendingSend {
    /// When a record for `packet` may be dropped unsettled: after the last
    /// window it was counted in has ended, and after an identical packet could
    /// no longer be sent. With a timeout timestamp that is the timestamp plus
    /// clock skew; with a height-only timeout it is a fixed long retention.
    pub fn retention_end(latest_period_end: Timestamp, packet: &Packet) -> Timestamp {
        match packet.timeout_timestamp {
            Some(timeout) if timeout > 0 => {
                let timeout = Timestamp::from_nanos(timeout);
                let latest = if timeout > latest_period_end {
                    timeout
                } else {
                    latest_period_end
                };
                saturating_plus_seconds(latest, TIMEOUT_SKEW_SECONDS)
            }
            _ => saturating_plus_seconds(latest_period_end, HEIGHT_TIMEOUT_RETENTION_SECONDS),
        }
    }

    /// Whether the record may be removed without settlement.
    pub fn is_expired(&self, now: Timestamp) -> bool {
        self.retain_until < now
    }
}

/// `Timestamp::plus_seconds` aborts on overflow; a far-future timeout is
/// caller-supplied data and must not be able to do that.
fn saturating_plus_seconds(timestamp: Timestamp, seconds: u64) -> Timestamp {
    Timestamp::from_nanos(
        timestamp
            .nanos()
            .saturating_add(seconds.saturating_mul(1_000_000_000)),
    )
}
