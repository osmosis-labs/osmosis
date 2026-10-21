use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::packet::Packet;

/// Clock skew allowed between this chain and a counterparty when deciding
/// that a packet's timeout timestamp has passed.
pub const TIMEOUT_SKEW_SECONDS: u64 = 60 * 60;

/// Longest a record is kept past its last window end, whatever the packet's
/// timeout says. It applies when the packet has no timeout timestamp
/// (height-only timeout), since the contract cannot tell when a counterparty
/// height will be reached, and it caps the retention of a timestamp far in
/// the future, which would otherwise make the record permanent. Past the cap
/// the same residual applies in both cases: an identical packet could still
/// be sent, and the earlier failure would settle it.
pub const MAX_RETENTION_SECONDS: u64 = 30 * 24 * 60 * 60;

/// A send that passed the rate limit and has not been settled yet.
///
/// It records, for every quota the send was counted against, the window it
/// was counted in. A send that later fails (error acknowledgement or timeout)
/// is refunded to a quota only while that quota is still in the same window,
/// so a refund can only ever restore what the same window charged.
///
/// Records are keyed by packet content (Packet::send_key), which is not a
/// unique identity: the same content can be sent again as long as the
/// packet's timeout has not passed on the counterparty. One record exists per
/// content key; identical sends counted in the same windows share it and
/// `count` says how many are outstanding, so that every one of them can be
/// refunded if it fails. A record is retained until both its windows have
/// ended and the packet's timeout has passed by the clock-skew margin, and an
/// identical send that would be counted in a different window while the
/// record is retained is rejected, so a settlement can never credit a window
/// other than the one that was charged.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct PendingSend {
    pub funds: Uint256,
    /// (quota name, period_end) for each quota on the (channel, denom) path
    pub channel_windows: Vec<(String, Timestamp)>,
    /// (quota name, period_end) for each quota on the ("any", denom) path
    pub any_windows: Vec<(String, Timestamp)>,
    /// When the record may be removed without settlement, see `retention_end`
    pub retain_until: Timestamp,
    /// How many identical sends this record stands for. Each settlement
    /// refunds `funds` once and lowers it by one; the record goes at zero.
    #[serde(default = "one")]
    pub count: u32,
}

fn one() -> u32 {
    1
}

impl PendingSend {
    /// When a record for `packet` may be dropped unsettled: after the last
    /// window it was counted in has ended, and after the packet's timeout
    /// timestamp has passed, each plus the clock-skew margin, but never more
    /// than MAX_RETENTION_SECONDS past the window end. A height-only timeout
    /// gets the maximum.
    pub fn retention_end(latest_period_end: Timestamp, packet: &Packet) -> Timestamp {
        let cap = saturating_plus_seconds(latest_period_end, MAX_RETENTION_SECONDS);
        match packet.timeout_timestamp {
            Some(timeout) if timeout > 0 => {
                let timeout = Timestamp::from_nanos(timeout);
                let latest = if timeout > latest_period_end {
                    timeout
                } else {
                    latest_period_end
                };
                let until = saturating_plus_seconds(latest, TIMEOUT_SKEW_SECONDS);
                if until > cap {
                    cap
                } else {
                    until
                }
            }
            _ => cap,
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
